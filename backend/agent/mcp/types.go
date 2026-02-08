package mcp

// MCP Protocol Types

// MCP Protocol Version
const (
    ProtocolVersion = "2024-11-05"
)

// Request represents an MCP JSON-RPC request
type Request struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

// Response represents an MCP JSON-RPC response
type Response struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Result  interface{} `json:"result,omitempty"`
    Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// InitializeParams for initialize method
type InitializeParams struct {
    ProtocolVersion string                 `json:"protocolVersion"`
    Capabilities  ClientCapabilities       `json:"capabilities"`
    ClientInfo    ClientInfo               `json:"clientInfo"`
    Locale        string                  `json:"locale,omitempty"`
}

// ClientCapabilities announced to server
type ClientCapabilities struct {
    Roots    *RootsCapability    `json:"roots,omitempty"`
    Sampling *SamplingCapability `json:"sampling,omitempty"`
}

// RootsCapability for roots listing support
type RootsCapability struct {
    ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability for sampling support
type SamplingCapability struct{}

// ClientInfo sent to server
type ClientInfo struct {
    Name    string `json:"name"`
    Version string `json:"version"`
}

// InitializeResult from server
type InitializeResult struct {
    ProtocolVersion string             `json:"protocolVersion"`
    Capabilities  ServerCapabilities    `json:"capabilities"`
    ServerInfo    ServerInfo          `json:"serverInfo"`
    Instructions  string              `json:"instructions,omitempty"`
}

// ServerCapabilities from server
type ServerCapabilities struct {
    Tools          *ToolCapability      `json:"tools,omitempty"`
    Resources      *ResourceCapability  `json:"resources,omitempty"`
    Prompts        *PromptCapability   `json:"prompts,omitempty"`
}

// ToolCapability for tool support
type ToolCapability struct {
    ListChanged bool `json:"listChanged,omitempty"`
}

// ResourceCapability for resource support
type ResourceCapability struct {
    Subscribe   bool `json:"subscribe,omitempty"`
    ListChanged bool `json:"listChanged,omitempty"`
}

// PromptCapability for prompt support
type PromptCapability struct {
    ListChanged bool `json:"listChanged,omitempty"`
}

// ServerInfo from server
type ServerInfo struct {
    Name    string `json:"name"`
    Version string `json:"version"`
}

// Tool definition from server
type Tool struct {
    Name        string      `json:"name"`
    Description string      `json:"description,omitempty"`
    InputSchema interface{} `json:"inputSchema,omitempty"`
}

// ToolListResult from tools/list
type ToolListResult struct {
    Tools []Tool `json:"tools"`
}

// ToolCall for tools/call
type ToolCall struct {
    Name      string                 `json:"name"`
    Arguments map[string]interface{} `json:"arguments,omitempty"`
    ArgumentsString string               `json:"_arguments,omitempty"` // For JSON string
}

// ToolCallResult from tools/call
type ToolCallResult struct {
    Content []ToolContent `json:"content,omitempty"`
    IsError bool            `json:"isError,omitempty"`
}

// ToolContent in tool call result
type ToolContent struct {
    Type string `json:"type"`
    Text string `json:"text,omitempty"`
    Data interface{} `json:"data,omitempty"`
}

// Notification represents a server-initiated notification
type Notification struct {
    JSONRPC string      `json:"jsonrpc"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

// NotificationParams for various notifications
type NotificationParams struct {
    Progress *ProgressParams      `json:"progress,omitempty"`
    Tools   *ToolsChangedParams `json:"tools,omitempty"`
}

// ProgressParams for progress notifications
type ProgressParams struct {
    ProgressToken string  `json:"progressToken"`
    Progress      float64 `json:"progress,omitempty"`
    Total         float64 `json:"total,omitempty"`
    Message       string  `json:"message,omitempty"`
}

// ToolsChangedParams for tools/list_changed
type ToolsChangedParams struct{}

// TransportType defines the transport protocol for MCP communication
type TransportType string

const (
    // TransportTypeStdio uses stdio for local MCP servers
    TransportTypeStdio TransportType = "stdio"
    // TransportTypeHTTP uses HTTP for remote MCP servers
    TransportTypeHTTP TransportType = "http"
)

// Config related types
type MCPServerConfig struct {
    Name      string            `json:"name"`
    Transport TransportType     `json:"transport"` // "stdio" or "http"
    Command   string            `json:"command,omitempty"`  // Required for stdio
    Args      []string          `json:"args,omitempty"`     // Required for stdio
    URL       string            `json:"url,omitempty"`      // Required for http
    Headers   map[string]string `json:"headers,omitempty"`  // Optional for http
    Env       map[string]string `json:"env,omitempty"`      // For stdio
    Enabled   bool              `json:"enabled"`
}

type MCPConfig struct {
    Enabled bool                `json:"enabled"`
    Servers []*MCPServerConfig   `json:"servers"`
}

// Connection status
type ConnectionStatus string

const (
    StatusDisconnected ConnectionStatus = "disconnected"
    StatusConnecting   ConnectionStatus = "connecting"
    StatusConnected    ConnectionStatus = "connected"
    StatusError       ConnectionStatus = "error"
)

// ConnectionState tracks connection state
type ConnectionState struct {
    Status    ConnectionStatus `json:"status"`
    LastError string          `json:"lastError,omitempty"`
    ConnectedAt int64          `json:"connectedAt,omitempty"`
}
