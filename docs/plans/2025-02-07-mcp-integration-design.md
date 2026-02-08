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
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed"],
      "enabled": true
    },
    {
      "name": "brave-search",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "enabled": true
    },
    {
      "name": "github",
      "command": "mcp-server-github",
      "args": ["--token", "${GITHUB_TOKEN}"],
      "enabled": false
    }
  ]
}
```

**Configuration Model:**

```go
type MCPConfig struct {
    Enabled bool                `json:"enabled"`
    Servers []*MCPServerConfig `json:"servers"`
}

type MCPServerConfig struct {
    Name    string   `json:"name"`
    Command string   `json:"command"`
    Args    []string `json:"args"`
    Env     map[string]string `json:"env"`
    Enabled bool     `json:"enabled"`
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
