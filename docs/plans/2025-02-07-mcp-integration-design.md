# MCP Integration Design for go-stock

## Overview

This design outlines the integration of Model Context Protocol (MCP) into the go-stock application. Currently, AI tools are implemented using a "pro-code" pattern via Eino's `InvokableTool` interface, requiring recompilation for any tool changes. The new design will enable runtime pluggable tool configuration through MCP while preserving existing built-in tools.

## Current Architecture Analysis

**Existing Tool Implementation:**
- Tools are defined in `backend/agent/tools/` (e.g., `stock_k_line_data_tool.go`)
- Each tool implements `tool.InvokableTool` interface from Eino framework
- Tool registration in `GetStockAiAgent()` is static (hardcoded in `agent.go:64-78`)
- Tools are executed via `InvokableRun()` method
- Tool definitions are constructed using `schema.ToolInfo`

**Key Components:**
- `backend/agent/agent.go` - Agent creation with hardcoded tools
- `backend/agent/tools/*.go` - Built-in tool implementations
- `backend/data/openai_api.go` - Chat streaming with tool calling

## Design Goals

1. **Preserve Existing Functionality** - Built-in tools continue working unchanged
2. **Runtime Configurability** - MCP tools loadable via configuration without recompilation
3. **MCP Protocol Support** - Implement MCP client to connect to MCP servers
4. **Unified Tool Interface** - MCP tools expose as `tool.InvokableTool` for Eino
5. **Configuration Management** - JSON-based MCP server configuration
6. **Error Resilience** - MCP failures don't break built-in tools

## Architecture Design

### 1. MCP Client Layer

**New Package: `backend/agent/mcp/`**

```
backend/agent/mcp/
├── client.go          # MCP client implementation
├── transport.go       # MCP transport (stdio/HTTP)
├── types.go          # MCP protocol types
└── tool_adapter.go   # Convert MCP tools to Eino tools
```

**MCP Client Responsibilities:**
- Connect to MCP servers (stdio for local, HTTP for remote)
- Perform MCP handshake (initialize, ping)
- List available tools (tools/list)
- Call tools (tools/call)
- Convert MCP tool schemas to Eino `schema.ToolInfo`
- Adapt MCP tool calls to `InvokableRun` interface

### 2. Configuration Structure

**New File: `data/mcp_config.json`**

```json
{
  "enabled": true,
  "servers": [
    {
      "name": "filesystem",
      "transport": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed"],
      "enabled": true
    },
    {
      "name": "brave-search",
      "transport": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "enabled": true
    },
    {
      "name": "github",
      "transport": "stdio",
      "command": "mcp-server-github",
      "args": ["--token", "${GITHUB_TOKEN}"],
      "enabled": false
    },
    {
      "name": "remote-mcp-server",
      "transport": "http",
      "url": "http://localhost:3000/mcp",
      "headers": {
        "Authorization": "Bearer your-api-token"
      },
      "enabled": false
    },
    {
      "name": "cloud-mcp-server",
      "transport": "http",
      "url": "https://api.example.com/mcp/v1",
      "headers": {
        "Authorization": "Bearer ${CLOUD_MCP_TOKEN}",
        "X-API-Version": "v1"
      },
      "enabled": false
    }
  ]
}
```

### 2.1 Transport Implementation Details

The MCP client supports two transport types for different use cases:

#### Stdio Transport (Local MCP Servers)

**Use Case:** Local MCP servers running as subprocesses
**Implementation:** `backend/agent/mcp/transport.go` - `StdioTransport`

**Characteristics:**
- Asynchronous communication pattern (send + receive)
- Process-based: MCP server runs as child process
- JSON-RPC messages separated by newlines
- Bidirectional: can receive server-initiated notifications
- Connection lifecycle: managed process (start/stop)

**Flow:**
```
Client → Start subprocess (npx @modelcontextprotocol/server-*)
Client → stdin: {"jsonrpc":"2.0","id":1,"method":"initialize",...}
Client ← stdout: {"jsonrpc":"2.0","id":1,"result":{...}}
Client → stdin: {"jsonrpc":"2.0","id":2,"method":"tools/list"}
Client ← stdout: {"jsonrpc":"2.0","id":2,"result":{"tools":[...]}}
```

**Example Config:**
```json
{
  "name": "filesystem",
  "transport": "stdio",
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-filesystem", "/allowed/path"],
  "enabled": true
}
```

#### HTTP Transport (Remote MCP Servers)

**Use Case:** Remote/cloud-hosted MCP servers
**Implementation:** `backend/agent/mcp/transport.go` - `HTTPTransport`

**Characteristics:**
- Synchronous request-response pattern
- HTTP POST to endpoint
- JSON-RPC over HTTP
- Stateless: no persistent connection
- Custom headers support (Bearer tokens, API keys)
- Network timeout: 30 seconds (configurable)

**Flow:**
```
Client → POST /mcp {"jsonrpc":"2.0","id":1,"method":"initialize",...}
Client ← 200 OK {"jsonrpc":"2.0","id":1,"result":{...}}
Client → POST /mcp {"jsonrpc":"2.0","id":2,"method":"tools/list"}
Client ← 200 OK {"jsonrpc":"2.0","id":2,"result":{"tools":[...]}}
```

**Example Config:**
```json
{
  "name": "remote-server",
  "transport": "http",
  "url": "https://api.example.com/mcp/v1",
  "headers": {
    "Authorization": "Bearer ${MCP_API_TOKEN}",
    "X-Custom-Header": "custom-value"
  },
  "enabled": true
}
```

**HTTP Protocol Specification:**

**Request:**
- Method: POST
- Content-Type: application/json
- Body: JSON-RPC 2.0 request
- Headers: Custom headers from config

**Response:**
- Status: 200 OK
- Content-Type: application/json
- Body: JSON-RPC 2.0 response

**Error Handling:**
- Non-200 status: Return error with status code and body
- Timeout: Return context timeout error
- Invalid JSON: Return unmarshal error

**Transport Interface:**

Both transports implement the same `Transport` interface:

```go
type Transport interface {
    Send(ctx context.Context, request *Request) error
    Receive(ctx context.Context) (*Response, error)
    Close() error
    IsConnected() bool
}
```

**HTTP-Specific Methods:**

```go
// SendRequest combines send and receive for synchronous HTTP pattern
func (t *HTTPTransport) SendRequest(ctx context.Context, request *Request) (*Response, error)
```

#### Client Factory Pattern

**Implementation:** `backend/agent/mcp/client.go` - `NewClient()`

The client factory automatically creates the appropriate transport based on configuration:

```go
func NewClient(ctx context.Context, config *MCPServerConfig) (*Client, error) {
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
        return nil, fmt.Errorf("unsupported transport type: %s", transportType)
    }

    // ... create client with transport
}
```

#### Request Handling Pattern

**Implementation:** `backend/agent/mcp/client.go` - `sendRequest()`

The client uses a unified request handler that adapts to transport type:

```go
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
```

**Configuration Model:**

```go
type MCPConfig struct {
    Enabled bool                `json:"enabled"`
    Servers []*MCPServerConfig `json:"servers"`
}

type TransportType string

const (
    TransportTypeStdio TransportType = "stdio"
    TransportTypeHTTP TransportType = "http"
)

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
```

### 3. Tool Registry

**New Package: `backend/agent/registry/`**

```
backend/agent/registry/
├── registry.go       # Central tool registry
└── builtins.go      # Built-in tool definitions
```

**Tool Registry Responsibilities:**
- Manage both built-in and MCP tools
- Provide unified `[]tool.InvokableTool` to agent
- Handle tool lifecycle (connect/disconnect)
- Tool name collision resolution (MCP tools get prefix: `mcp_<server>_<tool>`)
- Health checking for MCP connections

**Registry API:**

```go
type ToolRegistry struct {
    builtins map[string]tool.InvokableTool
    mcpTools map[string]tool.InvokableTool
    mcpClients map[string]*mcp.Client
}

func NewToolRegistry() *ToolRegistry
func (r *ToolRegistry) LoadBuiltins()
func (r *ToolRegistry) LoadMCPTools(config *MCPConfig) error
func (r *ToolRegistry) GetAllTools() []tool.InvokableTool
func (r *ToolRegistry) GetTool(name string) tool.InvokableTool
func (r *ToolRegistry) Shutdown()
```

### 4. Modified Agent Creation

**Updated: `backend/agent/agent.go`**

**Before:**
```go
aiTools := compose.ToolsNodeConfig{
    Tools: []tool.BaseTool{
        tools.GetQueryEconomicDataTool(),
        tools.GetQueryStockPriceInfoTool(),
        // ... hardcoded list
    },
}
```

**After:**
```go
registry := registry.NewToolRegistry()
registry.LoadBuiltins()

mcpConfig := GetMCPConfig() // Load from JSON
registry.LoadMCPTools(mcpConfig)

aiTools := compose.ToolsNodeConfig{
    Tools: registry.GetAllTools(),
}
```

### 5. MCP to Eino Tool Adapter

**Implementation: `backend/agent/mcp/tool_adapter.go`**

Convert MCP tool schema to Eino schema:

```go
func MCPToolToEinoTool(mcpTool *MCPTool) tool.InvokableTool {
    return &MCPToolAdapter{
        name: mcpTool.Name,
        desc: mcpTool.Description,
        inputSchema: mcpTool.InputSchema, // JSON Schema
        mcpClient: client,
    }
}

type MCPToolAdapter struct {
    name        string
    desc        string
    inputSchema map[string]any
    mcpClient   *mcp.Client
}

func (a *MCPToolAdapter) Info(ctx context.Context) (*schema.ToolInfo, error) {
    // Convert JSON Schema to Eino ParameterInfo
    params := convertJSONSchemaToEinoParams(a.inputSchema)
    return &schema.ToolInfo{
        Name:        a.name,
        Desc:        a.desc,
        ParamsOneOf: schema.NewParamsOneOfByParams(params),
    }, nil
}

func (a *MCPToolAdapter) InvokableRun(ctx context.Context, args string, opts ...tool.Option) (string, error) {
    // Call MCP tool
    result, err := a.mcpClient.CallTool(ctx, a.name, args)
    if err != nil {
        return "", err
    }
    // Format result as string
    return formatMCPResult(result), nil
}
```

## Data Flow

### Tool Discovery Flow

```
1. Application Startup
   ↓
2. Load MCP Config (mcp_config.json)
   ↓
3. For each enabled server:
   - Start MCP client process
   - Initialize connection
   - List available tools
   - Convert to Eino tools
   - Register in ToolRegistry
   ↓
4. Load built-in tools to ToolRegistry
   ↓
5. GetStockAiAgent() receives unified tool list
   ↓
6. Eino agent can call both built-in and MCP tools
```

### Tool Execution Flow

```
1. AI Model decides to call tool
   ↓
2. Eino invokes tool.InvokableRun()
   ↓
3. ToolRegistry routes to appropriate tool:
   - If built-in: Direct execution
   - If MCP: MCPToolAdapter → MCP Client → Server
   ↓
4. Result formatted and returned to AI
```

### Error Handling

```
MCP Server Unavailable:
- Log error
- Disable that server in registry
- Other tools continue working
- Built-in tools unaffected

Tool Call Failure:
- Return error to AI
- AI can retry or continue
- Application continues running

Config Load Error:
- Use default (no MCP tools)
- Log warning
- Built-in tools work normally
```

## Implementation Phases

### Phase 1: Core MCP Client
- Implement MCP protocol types
- Implement stdio transport
- Implement basic client (init, ping, tools/list)
- Add configuration loading

### Phase 2: Tool Adapter
- Implement JSON Schema to Eino params conversion
- Create MCPToolAdapter
- Test with simple MCP server

### Phase 3: Tool Registry
- Implement registry with built-in tools
- Add MCP tool loading
- Add health checking
- Integrate with GetStockAiAgent()

### Phase 4: Configuration UI
- Add MCP config editor in settings
- Add MCP server status display
- Add tool enable/disable toggles
- Import/export MCP config

### Phase 5: Testing & Polish
- Unit tests for MCP client
- Integration tests with real MCP servers
- Error handling validation
- Performance optimization

## Database Changes

**New Table: MCP Config (optional, could use JSON file)**

If using database for persistence:
```go
type MCPServer struct {
    gorm.Model
    Name    string
    Command string `gorm:"type:text"`
    Args    string `gorm:"type:text"` // JSON array
    Env     string `gorm:"type:text"` // JSON object
    Enabled bool
    Status  string // "connected", "disconnected", "error"
    LastError string `gorm:"type:text"`
    HealthCheckTime time.Time
}
```

## Frontend Changes

**Settings Component Updates:**
- Add "MCP Settings" tab
- List configured MCP servers
- Show connection status
- Add/remove server buttons
- Enable/disable toggles per server
- Display available tools from each server

**New Routes/Components:**
- `frontend/src/components/MCPSettings.vue` - MCP configuration UI
- `frontend/src/components/ToolList.vue` - Display all available tools

## Dependencies

**New Go Dependencies:**
```go
require (
    github.com/cloudwego/eino v0.5.7
    // MCP related (if external library available)
    // Or implement from scratch
)
```

**No new frontend dependencies** - using existing NaiveUI components.

## Security Considerations

1. **Command Injection** - Validate MCP server commands before execution
2. **Path Traversal** - Validate filesystem MCP allowed paths
3. **API Token Handling** - Store MCP tokens securely, never log
4. **Sandboxing** - Consider running MCP servers in isolated process
5. **Resource Limits** - Timeout MCP calls, limit concurrent connections

## Backward Compatibility

- Default `mcp_config.json` with `enabled: false` maintains current behavior
- Built-in tools always available regardless of MCP status
- No breaking changes to existing APIs
- Gradual rollout possible (enable MCP per user)

## Success Criteria

1. Built-in tools work unchanged when MCP disabled
2. MCP tools discoverable and callable when configured
3. MCP server failures don't crash application
4. Configuration persists across restarts
5. UI allows MCP management
6. Tool results properly formatted for AI consumption
