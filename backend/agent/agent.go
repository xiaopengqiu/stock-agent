package agent

import (
	"context"
	"go-stock/backend/agent/mcp"
	"go-stock/backend/agent/registry"
	"go-stock/backend/data"
	"go-stock/backend/logger"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

var (
	globalRegistry *registry.Registry
	registryOnce   sync.Once
)

// GetToolRegistry returns global tool registry instance，注册已有工具信息
func GetToolRegistry(ctx context.Context) *registry.Registry {
	registryOnce.Do(func() {
		globalRegistry = registry.NewRegistry(ctx)
		if err := globalRegistry.Initialize(); err != nil {
			logger.SugaredLogger.Errorf("Failed to initialize tool registry: %v", err)
		}
	})
	return globalRegistry
}

// GetStockAiAgent @Author spark
// @Date 2025/8/4 16:17
// @Desc
// -----------------------------------------------------------------------------------
func GetStockAiAgent(ctx *context.Context, aiConfig data.AIConfig) *react.Agent {
	logger.SugaredLogger.Infof("GetStockAiAgent aiConfig: %v", aiConfig)
	temperature := float32(aiConfig.Temperature)
	var toolableChatModel model.ToolCallingChatModel
	var err error

	switch aiConfig.BaseUrl {
	case "https://ark.cn-beijing.volces.com/api/v3":
		toolableChatModel, err = ark.NewChatModel(context.Background(), &ark.ChatModelConfig{
			BaseURL:     aiConfig.BaseUrl,
			Model:       aiConfig.ModelName,
			APIKey:      aiConfig.ApiKey,
			MaxTokens:   &aiConfig.MaxTokens,
			Temperature: &temperature,
		})

	case "https://api.deepseek.com":
		toolableChatModel, err = deepseek.NewChatModel(*ctx, &deepseek.ChatModelConfig{
			BaseURL:     aiConfig.BaseUrl,
			Model:       aiConfig.ModelName,
			APIKey:      aiConfig.ApiKey,
			Timeout:     time.Duration(aiConfig.TimeOut) * time.Second,
			MaxTokens:   aiConfig.MaxTokens,
			Temperature: temperature,
		})

	default:
		toolableChatModel, err = openai.NewChatModel(*ctx, &openai.ChatModelConfig{
			BaseURL:     aiConfig.BaseUrl,
			Model:       aiConfig.ModelName,
			APIKey:      aiConfig.ApiKey,
			Timeout:     time.Duration(aiConfig.TimeOut) * time.Second,
			MaxTokens:   intPtr(aiConfig.MaxTokens),
			Temperature: float32Ptrt(temperature),
		})
	}

	if err != nil {
		logger.SugaredLogger.Error(err.Error())
		return nil
	}

	// Get enabled tools from configuration
	baseTools := GetEnabledTools(*ctx)

	logger.SugaredLogger.Infof("Agent initialized with %d tools (built-in: %d, MCP: %d)",
		len(baseTools),
		GetBuiltinToolCount(),
		GetMCPToolCount())

	// Create tools config
	aiTools := compose.ToolsNodeConfig{
		Tools: baseTools,
	}

	// Calculate max steps based on number of tools
	maxStep := len(baseTools)*1 + 3

	// Create agent
	agent, err := react.NewAgent(*ctx, &react.AgentConfig{
		ToolCallingModel: toolableChatModel,
		ToolsConfig:      aiTools,
		MaxStep:          maxStep,
		MessageModifier: func(ctx context.Context, input []*schema.Message) []*schema.Message {
			return input
		},
	})
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
		return nil
	}

	return agent
}

// GetEnabledTools returns all enabled tools from configuration
// This includes both built-in and MCP tools that are enabled in config
func GetEnabledTools(ctx context.Context) []tool.BaseTool {
	// Load tool configuration
	toolConfig, err := data.LoadToolConfig()
	if err != nil {
		logger.SugaredLogger.Warnf("Failed to load tool config, using all available tools: %v", err)
		// Fall back to getting all tools from registry
		return getAllToolsFromRegistry(ctx)
	}

	// Get tool registry
	toolReg := GetToolRegistry(ctx)
	allRegistryTools := toolReg.GetAllTools()

	// Create a map of enabled tool names from config
	enabledTools := make(map[string]bool)
	for _, toolItem := range toolConfig.Tools {
		if toolItem.Enabled {
			enabledTools[toolItem.Name] = true
		}
	}

	// Filter tools based on configuration
	baseTools := make([]tool.BaseTool, 0)
	for _, t := range allRegistryTools {
		if baseT, ok := t.(tool.BaseTool); ok {
			// Get tool name to check if it's enabled
			info, err := baseT.Info(ctx)
			if err != nil {
				logger.SugaredLogger.Warnf("Failed to get tool info: %v", err)
				continue
			}

			// Check if tool is enabled in config
			if enabledTools[info.Name] {
				baseTools = append(baseTools, baseT)
			} else {
				logger.SugaredLogger.Debugf("Tool %s is disabled in config, skipping", info.Name)
			}
		}
	}

	logger.SugaredLogger.Infof("Loaded %d enabled tools from configuration (total available: %d)",
		len(baseTools), len(allRegistryTools))

	return baseTools
}

// getAllToolsFromRegistry returns all tools from registry without filtering
func getAllToolsFromRegistry(ctx context.Context) []tool.BaseTool {
	toolReg := GetToolRegistry(ctx)
	allTools := toolReg.GetAllTools()

	baseTools := make([]tool.BaseTool, 0, len(allTools))
	for _, t := range allTools {
		if baseT, ok := t.(tool.BaseTool); ok {
			baseTools = append(baseTools, baseT)
		}
	}

	return baseTools
}

func intPtr(num int) *int {
	return &num
}

func float32Ptrt(num float32) *float32 {
	return &num
}

// GetBuiltinTools returns all built-in tools from the registry
func GetBuiltinTools() []tool.InvokableTool {
	if globalRegistry != nil {
		return globalRegistry.GetBuiltins().GetAllTools()
	}
	return nil
}

// GetMCPTools returns all MCP tools from the registry
func GetMCPTools() map[string]tool.InvokableTool {
	if globalRegistry != nil {
		return globalRegistry.GetMCPTools()
	}
	return nil
}

// ShutdownToolRegistry cleans up tool registry
func ShutdownToolRegistry() {
	if globalRegistry != nil {
		globalRegistry.Shutdown()
	}
}

// ReloadMCPTools reloads MCP tools (reconnects to all servers)
func ReloadMCPTools() error {
	if globalRegistry != nil {
		return globalRegistry.ReloadMCPTools()
	}
	return nil
}

// GetMCPStatus returns status of all MCP connections
func GetMCPStatus() map[string]mcp.ConnectionState {
	if globalRegistry != nil {
		return globalRegistry.GetStatus()
	}
	return nil
}

// GetMCPToolCount returns number of loaded MCP tools
func GetMCPToolCount() int {
	if globalRegistry != nil {
		return len(globalRegistry.GetMCPTools())
	}
	return 0
}

// GetBuiltinToolCount returns number of built-in tools
func GetBuiltinToolCount() int {
	if globalRegistry != nil {
		return len(globalRegistry.GetBuiltins().GetAllTools())
	}
	return 0
}
