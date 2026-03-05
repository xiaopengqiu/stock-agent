package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-stock/backend/logger"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Transport is the interface for MCP communication
type Transport interface {
	Send(ctx context.Context, request *Request) error
	Receive(ctx context.Context) (*Response, error)
	Close() error
	IsConnected() bool
}

// StdioTransport implements MCP communication via stdio
type StdioTransport struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	scanner   *bufio.Scanner
	mu        sync.Mutex
	connected bool
	readChan  chan *Response
	errChan   chan error
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(ctx context.Context, config *MCPServerConfig) (*StdioTransport, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)

	// Prepare environment variables
	env := os.Environ()
	for k, v := range config.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Create command
	cmd := exec.CommandContext(ctx, config.Command, config.Args...)
	cmd.Env = env

	// Setup pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		cancel()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		cancel()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	transport := &StdioTransport{
		cmd:       cmd,
		stdin:     stdin,
		stdout:    stdout,
		stderr:    stderr,
		scanner:   nil,
		connected: false,
		readChan:  make(chan *Response, 100),
		errChan:   make(chan error, 1),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		transport.Close()
		return nil, fmt.Errorf("failed to start MCP server %s: %w", config.Command, err)
	}

	// Set connected flag after successful start
	transport.mu.Lock()
	transport.connected = true
	transport.mu.Unlock()

	// Start reading responses
	go transport.readResponses()
	go transport.readStderr()

	// Wait for command to exit
	go func() {
		err := cmd.Wait()
		if err != nil {
			logger.SugaredLogger.Errorf("MCP server %s exited with error: %v", config.Name, err)
		} else {
			logger.SugaredLogger.Infof("MCP server %s exited normally", config.Name)
		}
		transport.mu.Lock()
		transport.connected = false
		transport.mu.Unlock()
		if ctx.Err() == nil {
			transport.errChan <- fmt.Errorf("MCP server %s exited", config.Name)
		}
	}()

	return transport, nil
}

// Send sends a JSON-RPC request to the server
func (t *StdioTransport) Send(ctx context.Context, request *Request) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected && ctx.Err() == nil {
		return fmt.Errorf("transport not connected")
	}

	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// MCP protocol uses JSON messages separated by newlines
	message := string(data) + "\n"
	logger.SugaredLogger.Debugf("MCP send: %s", message)

	_, err = t.stdin.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write to stdin: %w", err)
	}

	return nil
}

// Receive receives a JSON-RPC response from the server
func (t *StdioTransport) Receive(ctx context.Context) (*Response, error) {
	select {
	case response := <-t.readChan:
		return response, nil
	case err := <-t.errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close closes the transport
func (t *StdioTransport) Close() error {
	t.cancel()

	t.mu.Lock()
	defer t.mu.Unlock()

	var errs []error

	if t.stdin != nil {
		err := t.stdin.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if t.stdout != nil {
		err := t.stdout.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if t.stderr != nil {
		err := t.stderr.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	close(t.readChan)
	close(t.errChan)

	if len(errs) > 0 {
		return fmt.Errorf("multiple errors during close: %v", errs)
	}

	return nil
}

// IsConnected returns whether the transport is connected
func (t *StdioTransport) IsConnected() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.connected
}

// readResponses continuously reads responses from stdout
func (t *StdioTransport) readResponses() {
	scanner := bufio.NewScanner(t.stdout)
	for scanner.Scan() {
		line := scanner.Text()
		logger.SugaredLogger.Debugf("MCP receive: %s", line)

		var response Response
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			logger.SugaredLogger.Errorf("failed to unmarshal MCP response: %v, data: %s", err, line)
			continue
		}

		// Check for notifications (no ID field)
		if response.ID == nil {
			logger.SugaredLogger.Debugf("MCP notification: %s", line)
			// Handle notifications if needed
			continue
		}

		t.readChan <- &response
	}

	if err := scanner.Err(); err != nil {
		logger.SugaredLogger.Errorf("error reading from MCP server stdout: %v", err)
	}
}

// readStderr reads and logs stderr output
func (t *StdioTransport) readStderr() {
	scanner := bufio.NewScanner(t.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		// Log stderr messages
		if strings.TrimSpace(line) != "" {
			logger.SugaredLogger.Warnf("MCP stderr: %s", line)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.SugaredLogger.Errorf("error reading from MCP server stderr: %v", err)
	}
}

// HTTPTransport implements MCP communication via HTTP (for remote MCP servers)
type HTTPTransport struct {
	baseURL   string
	headers   map[string]string
	client    *http.Client
	mu        sync.Mutex
	connected bool
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(ctx context.Context, config *MCPServerConfig) (*HTTPTransport, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("URL is required for HTTP transport")
	}

	transport := &HTTPTransport{
		baseURL: config.URL,
		headers: config.Headers,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		connected: false,
	}

	// Set connected flag (HTTP is stateless, but we track if URL is valid)
	transport.mu.Lock()
	transport.connected = true
	transport.mu.Unlock()

	return transport, nil
}

// Send for HTTP transport sends a JSON-RPC request and waits for immediate response
func (t *HTTPTransport) Send(ctx context.Context, request *Request) error {
	t.mu.Lock()
	if !t.connected {
		t.mu.Unlock()
		return fmt.Errorf("transport not connected")
	}
	t.mu.Unlock()

	// Marshal request
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.SugaredLogger.Debugf("MCP HTTP send: %s", string(data))

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add custom headers
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	// Send request
	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Receive for HTTP transport - for HTTP, responses are immediate in Send, so this is a no-op
// HTTP transport follows request-response pattern, not streaming like stdio
func (t *HTTPTransport) Receive(ctx context.Context) (*Response, error) {
	// HTTP transport uses synchronous request-response pattern
	// This method is not used for HTTP transport
	return nil, fmt.Errorf("HTTP transport uses synchronous pattern, Receive not applicable")
}

// Close for HTTP transport
func (t *HTTPTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.connected = false
	return nil
}

// IsConnected for HTTP transport
func (t *HTTPTransport) IsConnected() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.connected
}

// SendRequest sends a request and receives response in one call (synchronous HTTP pattern)
func (t *HTTPTransport) SendRequest(ctx context.Context, request *Request) (*Response, error) {
	t.mu.Lock()
	if !t.connected {
		t.mu.Unlock()
		return nil, fmt.Errorf("transport not connected")
	}
	t.mu.Unlock()

	// Marshal request
	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.SugaredLogger.Debugf("MCP HTTP send: %s", string(data))

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add custom headers
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	// Send request
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	logger.SugaredLogger.Debugf("MCP HTTP receive: %s", string(body))

	// Unmarshal response
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}
