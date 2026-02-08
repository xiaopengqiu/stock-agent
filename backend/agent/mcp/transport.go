package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"go-stock/backend/logger"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
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
	baseURL string
	// Future implementation for HTTP transport
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(baseURL string) *HTTPTransport {
	return &HTTPTransport{
		baseURL: baseURL,
	}
}

// Send for HTTP transport (placeholder)
func (t *HTTPTransport) Send(ctx context.Context, request *Request) error {
	// TODO: Implement HTTP transport
	return fmt.Errorf("HTTP transport not yet implemented")
}

// Receive for HTTP transport (placeholder)
func (t *HTTPTransport) Receive(ctx context.Context) (*Response, error) {
	// TODO: Implement HTTP transport
	return nil, fmt.Errorf("HTTP transport not yet implemented")
}

// Close for HTTP transport
func (t *HTTPTransport) Close() error {
	return nil
}

// IsConnected for HTTP transport
func (t *HTTPTransport) IsConnected() bool {
	return false
}
