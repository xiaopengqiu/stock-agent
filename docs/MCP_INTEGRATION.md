# MCP Integration

This document describes the MCP (Model Context Protocol) integration added to the go-stock application.

## Overview

MCP integration enables the application to use external tools through the Model Context Protocol, providing:
- Runtime pluggable tool configuration without code changes
- Support for multiple MCP servers (filesystem, search, GitHub, etc.)
- Preserved existing built-in tool functionality
- Configuration-based tool management

## Architecture

### Backend Components

```
backend/agent/mcp/
├── types.go          # MCP protocol types and definitions
├── transport.go      # Transport layer (stdio/HTTP)
├── client.go        # MCP client implementation
└── tool_adapter.go   # Eino tool adapter for MCP tools

backend/agent/registry/
├── registry.go       # Central tool registry
└── builtins.go      # Built-in tool definitions
```

### Frontend Components

```
frontend/src/components/
└── mcp-settings.vue  # MCP configuration UI

frontend/src/router/
└── router.js        # Added mcp-settings route
```

## Configuration

MCP configuration is stored in `data/mcp_config.json`:

```json
{
  "enabled": false,
  "servers": [
    {
      "name": "filesystem",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed"],
      "enabled": false
    },
    {
      "name": "brave-search",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "enabled": false
    }
  ]
}
```

## Usage

### Enabling MCP Tools

1. Open the application
2. Navigate to Settings or access MCP settings directly at `/mcp-settings`
3. Click "配置服务器" to open configuration editor
4. Choose an example configuration or edit JSON directly
5. Save configuration - tools will be automatically loaded

### Available MCP Servers

#### Filesystem Server
Allows AI to access local filesystem for file operations.

**Configuration:**
```json
{
  "enabled": true,
  "servers": [
    {
      "name": "filesystem",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "C:/Users"],
      "enabled": true
    }
  ]
}
```

#### Brave Search Server
Enables web search using Brave Search API.

**Configuration:**
```json
{
  "enabled": true,
  "servers": [
    {
      "name": "brave-search",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "enabled": true
    }
  ]
}
```

#### GitHub Server
Access GitHub repositories and code via MCP.

**Configuration:**
```json
{
  "enabled": true,
  "servers": [
    {
      "name": "github",
      "command": "mcp-server-github",
      "args": ["--token", "${GITHUB_TOKEN}"],
      "env": {
        "GITHUB_TOKEN": "your_github_token_here"
      },
      "enabled": false
    }
  ]
}
```

## API

### Backend (Go)

```go
// Get MCP status
func (a *App) GetMCPEnabled() bool
func (a *App) GetMCPStatus() map[string]any
func (a *App) GetMCPToolCount() int
func (a *App) GetBuiltinToolCount() int

// Manage MCP
func (a *App) ReloadMCPTools() string
func (a *App) GetMCPConfig() string
func (a *App) SetMCPConfig(config string) string
```

### Frontend (Vue)

```javascript
import {
  GetMCPEnabled,
  GetMCPStatus,
  ReloadMCPTools,
  GetMCPConfig,
  SetMCPConfig,
  GetMCPToolCount,
  GetBuiltinToolCount
} from "../../wailsjs/go/main/App";
```

## Tool Naming

MCP tools are prefixed with `mcp_<server>_<tool>` to avoid naming conflicts:

- Built-in tool: `QueryStockKLine`
- MCP tool: `mcp_brave-search_search_web`

## Built-in Tools (Preserved)

The following built-in tools continue to work without MCP:

1. QueryEconomicData - Economic data query
2. QueryStockPriceInfo - Stock price information
3. QueryStockCodeInfo - Stock code information
4. QueryMarketNews - Market news
5. ChoiceStockByIndicators - Stock screening by indicators
6. QueryStockKLine - K-line data retrieval
7. InteractiveAnswerData - Interactive Q&A data
8. FinancialReport - Financial reports
9. QueryStockNews - Stock-specific news
10. IndustryResearchReport - Industry research reports
11. QueryBKDict - Block dictionary queries

## Error Handling

- MCP server failures don't affect built-in tools
- Connection errors are logged and shown in UI
- Invalid configuration returns specific error messages
- Tool call failures are reported to AI model

## Development

### Adding New MCP Server Support

1. Ensure MCP server follows MCP protocol
2. Add server configuration to `data/mcp_config.json`
3. Restart application or click "重载工具"

### Testing MCP Integration

1. Start application with MCP enabled
2. Check connection status in MCP settings
3. Send message to AI requesting MCP tool usage
4. Verify tool execution and result formatting

## Security Considerations

- Validate MCP server commands before execution
- Sanitize filesystem paths
- Securely store API tokens
- Run MCP servers in isolated processes
- Apply resource limits (timeouts, connection limits)

## Troubleshooting

**MCP tools not loading:**
- Check Node.js/npm is available for stdio servers
- Verify server command and arguments
- Check logs for connection errors

**Configuration not saving:**
- Validate JSON syntax
- Ensure data directory is writable
- Check file permissions

**Tool execution failures:**
- Check MCP server is running
- Verify tool parameters match schema
- Review MCP server logs
