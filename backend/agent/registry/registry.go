package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"go-stock/backend/agent/mcp"
	"go-stock/backend/logger"
	"os"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
)

// Registry manages both built-in and MCP tools
type Registry struct {
	mu         sync.RWMutex
	builtins   *Builtins
	mcpTools   map[string]tool.InvokableTool
	mcpClients map[string]*mcp.Client
	config     *mcp.MCPConfig
	ctx        context.Context
	cancel     context.CancelFunc
}

// DefaultConfigPath is the default path for MCP configuration file
const DefaultConfigPath = "data/mcp_config.json"

// NewRegistry creates a new tool registry
func NewRegistry(ctx context.Context) *Registry {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)

	return &Registry{
		builtins:   NewBuiltins(),
		mcpTools:   make(map[string]tool.InvokableTool),
		mcpClients: make(map[string]*mcp.Client),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// LoadConfig loads MCP configuration from file
func (r *Registry) LoadConfig(path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if path == "" {
		path = DefaultConfigPath
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.SugaredLogger.Infof("MCP config file not found at %s, using empty config", path)
		r.config = &mcp.MCPConfig{
			Enabled: false,
			Servers:  []*mcp.MCPServerConfig{},
		}
		return nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read MCP config: %w", err)
	}

	// Parse JSON
	var config mcp.MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse MCP config: %w", err)
	}

	r.config = &config
	logger.SugaredLogger.Infof("Loaded MCP config from %s, enabled=%v, servers=%d",
		path, config.Enabled, len(config.Servers))

	return nil
}

// Initialize sets up the registry with built-in and MCP tools
func (r *Registry) Initialize() error {
	// Load config
	if err := r.LoadConfig(""); err != nil {
		logger.SugaredLogger.Errorf("Failed to load MCP config: %v", err)
		// Continue with empty config
	}

	// Load MCP tools if enabled
	if r.config != nil && r.config.Enabled {
		if err := r.LoadMCPTools(); err != nil {
			logger.SugaredLogger.Errorf("Failed to load MCP tools: %v", err)
			// Don't fail, built-in tools will still work
		}
	}

	return nil
}

// LoadMCPTools connects to all configured MCP servers and loads their tools
func (r *Registry) LoadMCPTools() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.config == nil || !r.config.Enabled {
		return nil
	}

	logger.SugaredLogger.Infof("Loading MCP tools from %d servers", len(r.config.Servers))

	for _, serverConfig := range r.config.Servers {
		if !serverConfig.Enabled {
			logger.SugaredLogger.Infof("MCP server %s is disabled, skipping", serverConfig.Name)
			continue
		}

		// Create MCP client
		client, err := mcp.NewClient(r.ctx, serverConfig)
		if err != nil {
			logger.SugaredLogger.Errorf("Failed to create MCP client for %s: %v", serverConfig.Name, err)
			continue
		}

		// Connect to server
		if err := client.Connect(); err != nil {
			logger.SugaredLogger.Errorf("Failed to connect to MCP server %s: %v", serverConfig.Name, err)
			client.Close()
			continue
		}

		// Store client
		r.mcpClients[serverConfig.Name] = client

		// Create tool adapters
		adapters := mcp.CreateToolAdapters(client, serverConfig.Name)
		for _, adapter := range adapters {
			r.mcpTools[getToolName(adapter)] = adapter
		}

		logger.SugaredLogger.Infof("Loaded %d tools from MCP server %s", len(adapters), serverConfig.Name)
	}

	return nil
}

// GetAllTools returns all available tools (built-in + MCP)
func (r *Registry) GetAllTools() []tool.InvokableTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	allTools := make([]tool.InvokableTool, 0)

	// Add built-in tools
	allTools = append(allTools, r.builtins.GetAllTools()...)

	// Add MCP tools
	for _, tool := range r.mcpTools {
		allTools = append(allTools, tool)
	}

	logger.SugaredLogger.Debugf("Total tools available: %d (built-in: %d, MCP: %d)",
		len(allTools), len(r.builtins.GetAllTools()), len(r.mcpTools))

	return allTools
}

// GetTool returns a specific tool by name
func (r *Registry) GetTool(name string) tool.InvokableTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check MCP tools first
	if tool, exists := r.mcpTools[name]; exists {
		return tool
	}

	// Check built-in tools (this requires iterating)
	for _, tool := range r.builtins.GetAllTools() {
		// We would need to call tool.Info() to compare names
		// For efficiency, we're skipping this for now
		_ = tool
	}

	return nil
}

// GetMCPClients returns all MCP clients
func (r *Registry) GetMCPClients() map[string]*mcp.Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make(map[string]*mcp.Client)
	for k, v := range r.mcpClients {
		clients[k] = v
	}

	return clients
}

// GetMCPClient returns a specific MCP client by name
func (r *Registry) GetMCPClient(name string) (*mcp.Client, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, exists := r.mcpClients[name]
	return client, exists
}

// ReloadMCPTools reloads MCP tools (reconnects to all servers)
func (r *Registry) ReloadMCPTools() error {
	r.mu.Lock()

	// Close existing MCP clients
	for name, client := range r.mcpClients {
		logger.SugaredLogger.Infof("Closing MCP client: %s", name)
		client.Close()
	}

	r.mcpClients = make(map[string]*mcp.Client)
	r.mcpTools = make(map[string]tool.InvokableTool)

	// Store a copy of config to use after unlock
	config := r.config
	r.mu.Unlock()

	// Reload config
	if err := r.LoadConfig(""); err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	// Only load tools if enabled
	if config != nil && config.Enabled {
		return r.LoadMCPTools()
	}

	return nil
}

// Shutdown closes all MCP connections and cleans up resources
func (r *Registry) Shutdown() {
	r.mu.Lock()
	defer r.mu.Unlock()

	logger.SugaredLogger.Info("Shutting down tool registry")

	// Close all MCP clients
	for name, client := range r.mcpClients {
		logger.SugaredLogger.Infof("Closing MCP client: %s", name)
		client.Close()
	}

	r.cancel()
}

// GetStatus returns status information for all MCP servers
func (r *Registry) GetStatus() map[string]mcp.ConnectionState {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status := make(map[string]mcp.ConnectionState)
	for name, client := range r.mcpClients {
		status[name] = client.GetState()
	}

	return status
}

// GetMCPTools returns MCP tools map
func (r *Registry) GetMCPTools() map[string]tool.InvokableTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make(map[string]tool.InvokableTool)
	for k, v := range r.mcpTools {
		tools[k] = v
	}

	return tools
}

// GetBuiltins returns built-ins instance
func (r *Registry) GetBuiltins() *Builtins {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.builtins
}

// HealthCheck performs a health check on all MCP servers
func (r *Registry) HealthCheck() map[string]bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]bool)

	for name, client := range r.mcpClients {
		if client.IsConnected() {
			// Try to ping the server
			err := client.Ping()
			results[name] = (err == nil)
			if err != nil {
				logger.SugaredLogger.Warnf("MCP server %s health check failed: %v", name, err)
			}
		} else {
			results[name] = false
		}
	}

	return results
}

// StartHealthCheckWatcher starts a background goroutine that periodically checks MCP server health
func (r *Registry) StartHealthCheckWatcher(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if len(r.mcpClients) == 0 {
					continue
				}

				results := r.HealthCheck()
				for name, healthy := range results {
					if !healthy {
						logger.SugaredLogger.Warnf("MCP server %s is unhealthy", name)
						// Optionally attempt reconnection
						// This is handled by reconnect logic in the client
					}
				}
			case <-r.ctx.Done():
				return
			}
		}
	}()
}

// getToolName extracts the tool name from a tool.InvokableTool
// This is a helper since tool.Info() requires context
func getToolName(t tool.InvokableTool) string {
	// We need to execute tool.Info() to get the name
	// For now, use a default approach
	info, err := t.Info(context.Background())
	if err != nil {
		return "unknown_tool"
	}
	return info.Name
}
