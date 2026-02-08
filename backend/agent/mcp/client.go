package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"go-stock/backend/logger"
	"strings"
	"sync"
	"time"
)

// Client represents an MCP client connection
type Client struct {
	config          *MCPServerConfig
	transport       Transport
	transportType   TransportType
	requestID       int64
	mu              sync.Mutex
	initialized     bool
	state           ConnectionState
	tools           map[string]Tool
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewClient creates a new MCP client
func NewClient(ctx context.Context, config *MCPServerConfig) (*Client, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)

	// Determine transport type (default to stdio for backward compatibility)
	transportType := config.Transport
	if transportType == "" {
		transportType = TransportTypeStdio
	}

	// Create appropriate transport
	var transport Transport
	var err error

	switch transportType {
	case TransportTypeStdio:
		transport, err = NewStdioTransport(ctx, config)
	case TransportTypeHTTP:
		transport, err = NewHTTPTransport(ctx, config)
	default:
		cancel()
		return nil, fmt.Errorf("unsupported transport type: %s", transportType)
	}

	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create: %w", err)
	}

	client := &Client{
		config:        config,
		transport:     transport,
		transportType: transportType,
		requestID:     0,
		initialized:   false,
		state: ConnectionState{
			Status: StatusDisconnected,
		},
		tools:  make(map[string]Tool),
		ctx:    ctx,
		cancel: cancel,
	}

	return client, nil
}

// Connect initializes the connection to the MCP server
func (c *Client) Connect() error {
	c.mu.Lock()
	if c.state.Status == StatusConnected {
		c.mu.Unlock()
		return nil // Already connected
	}
	c.state.Status = StatusConnecting
	c.mu.Unlock()

	logger.SugaredLogger.Infof("Connecting to MCP server: %s (%s)", c.config.Name, c.transportType)

	// Send initialize request
	initParams := InitializeParams{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ClientCapabilities{
			Roots:    &RootsCapability{},
			Sampling: &SamplingCapability{},
		},
		ClientInfo: ClientInfo{
			Name:    "go-stock",
			Version: "1.0.0",
		},
	}

	req := c.createRequest("initialize", initParams)

	// Send request and get response (handles both sync HTTP and async stdio)
	resp, err := c.sendRequest(req, 10*time.Second)
	if err != nil {
		c.updateState(StatusError, fmt.Sprintf("initialize failed: %v", err))
		return fmt.Errorf("initialize failed: %w", err)
	}

	if resp.Error != nil {
		c.updateState(StatusError, fmt.Sprintf("initialize error: %s", resp.Error.Message))
		return fmt.Errorf("initialize failed: %s", resp.Error.Message)
	}

	// Parse initialize result
	var initResult InitializeResult
	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		c.updateState(StatusError, fmt.Sprintf("failed to marshal result: %v", err))
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	if err := json.Unmarshal(resultBytes, &initResult); err != nil {
		c.updateState(StatusError, fmt.Sprintf("invalid initialize result: %v", err))
		return fmt.Errorf("invalid initialize result: %w", err)
	}

	c.initialized = true
	c.updateState(StatusConnected, "")
	c.state.ConnectedAt = time.Now().Unix()

	logger.SugaredLogger.Infof("MCP server %s initialized: version=%s, server=%s v%s",
		c.config.Name, initResult.ProtocolVersion, initResult.ServerInfo.Name, initResult.ServerInfo.Version)

	// Load tools
	if err := c.LoadTools(); err != nil {
		logger.SugaredLogger.Warnf("failed to load tools from %s: %v", c.config.Name, err)
		// Don't fail connection if tool loading fails
	}

	return nil
}

// LoadTools retrieves available tools from the server
func (c *Client) LoadTools() error {
	if !c.initialized {
		return fmt.Errorf("client not initialized")
	}

	req := c.createRequest("tools/list", nil)

	// Send request and get response
	resp, err := c.sendRequest(req, 5*time.Second)
	if err != nil {
		return fmt.Errorf("tools/list failed: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("tools/list failed: %s", resp.Error.Message)
	}

	var toolList ToolListResult
	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	if err := json.Unmarshal(resultBytes, &toolList); err != nil {
		return fmt.Errorf("invalid tools/list result: %w", err)
	}

	c.mu.Lock()
	c.tools = make(map[string]Tool)
	for _, tool := range toolList.Tools {
		c.tools[tool.Name] = tool
	}
	c.mu.Unlock()

	logger.SugaredLogger.Infof("Loaded %d tools from MCP server %s", len(toolList.Tools), c.config.Name)
	for _, tool := range toolList.Tools {
		logger.SugaredLogger.Debugf("  - %s: %s", tool.Name, tool.Description)
	}

	return nil
}

// CallTool invokes a tool on the MCP server
func (c *Client) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (interface{}, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	// Check if tool exists
	c.mu.Lock()
	_, exists := c.tools[toolName]
	c.mu.Unlock()
	if !exists {
		return nil, fmt.Errorf("tool %s not found", toolName)
	}

	logger.SugaredLogger.Infof("Calling MCP tool %s with args: %+v", toolName, arguments)

	call := ToolCall{
		Name:      toolName,
		Arguments: arguments,
	}

	req := c.createRequest("tools/call", call)

	// Send request and get response
	resp, err := c.sendRequest(req, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("tools/call failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tools/call failed: %s", resp.Error.Message)
	}

	var callResult ToolCallResult
	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}
	if err := json.Unmarshal(resultBytes, &callResult); err != nil {
		return nil, fmt.Errorf("invalid tools/call result: %w", err)
	}

	// Extract text content
	var resultText strings.Builder
	for _, content := range callResult.Content {
		if content.Type == "text" {
			resultText.WriteString(content.Text)
		}
	}

	if resultText.Len() > 0 {
		return resultText.String(), nil
	}

	// Return full result if no text content
	return resp.Result, nil
}

// GetTools returns the list of available tools
func (c *Client) GetTools() []Tool {
	c.mu.Lock()
	defer c.mu.Unlock()

	tools := make([]Tool, 0, len(c.tools))
	for _, tool := range c.tools {
		tools = append(tools, tool)
	}

	return tools
}

// GetTool returns a specific tool by name
func (c *Client) GetTool(name string) (Tool, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tool, exists := c.tools[name]
	return tool, exists
}

// GetState returns the current connection state
func (c *Client) GetState() ConnectionState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	state := c.GetState()
	return state.Status == StatusConnected
}

// Close closes the connection to the MCP server
func (c *Client) Close() error {
	logger.SugaredLogger.Infof("Closing MCP client: %s", c.config.Name)

	c.mu.Lock()
	c.initialized = false
	c.state.Status = StatusDisconnected
	c.mu.Unlock()

	c.cancel()
	return c.transport.Close()
}

// createRequest creates a new JSON-RPC request
func (c *Client) createRequest(method string, params interface{}) *Request {
	c.mu.Lock()
	c.requestID++
	id := c.requestID
	c.mu.Unlock()

	return &Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
}

// sendRequest sends a request and receives response, handling both sync HTTP and async stdio
func (c *Client) sendRequest(req *Request, timeout time.Duration) (*Response, error) {
	if c.transportType == TransportTypeHTTP {
		// HTTP transport is synchronous, use SendRequest directly
		httpTransport, ok := c.transport.(*HTTPTransport)
		if !ok {
			return nil, fmt.Errorf("transport is not HTTPTransport")
		}
		return httpTransport.SendRequest(c.ctx, req)
	}

	// stdio transport is asynchronous, send then wait
	if err := c.transport.Send(c.ctx, req); err != nil {
		return nil, err
	}

	return c.waitForResponse(req.ID, timeout)
}

// waitForResponse waits for a response with matching ID
func (c *Client) waitForResponse(id interface{}, timeout time.Duration) (*Response, error) {
	ctx, cancel := context.WithTimeout(c.ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			resp, err := c.transport.Receive(ctx)
			if err != nil {
				return nil, err
			}

			// Check for notification (no ID)
			if resp.ID == nil {
				continue
			}

			// Check if this is the response we're waiting for
			if resp.ID == id {
				return resp, nil
			}

			// Response for another request, handle or discard
			logger.SugaredLogger.Debugf("Received response for unknown ID: %v (waiting for: %v)", resp.ID, id)
		}
	}
}

// updateState updates the connection state
func (c *Client) updateState(status ConnectionStatus, errorMsg string) {
	c.mu.Lock()
	c.state.Status = status
	c.state.LastError = errorMsg
	if status == StatusConnected && c.state.ConnectedAt == 0 {
		c.state.ConnectedAt = time.Now().Unix()
	}
	c.mu.Unlock()

	if status == StatusError {
		logger.SugaredLogger.Errorf("MCP client %s error: %s", c.config.Name, errorMsg)
	} else {
		logger.SugaredLogger.Infof("MCP client %s state: %s", c.config.Name, status)
	}
}

// Ping sends a ping to check server responsiveness (optional MCP method)
func (c *Client) Ping() error {
	if !c.initialized {
		return fmt.Errorf("client not initialized")
	}

	req := c.createRequest("ping", nil)
	if err := c.transport.Send(c.ctx, req); err != nil {
		return fmt.Errorf("failed to send ping: %w", err)
	}

	resp, err := c.waitForResponse(req.ID, 5*time.Second)
	if err != nil {
		return fmt.Errorf("ping response timeout: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("ping failed: %s", resp.Error.Message)
	}

	return nil
}
