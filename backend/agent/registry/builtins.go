package registry

import (
	"context"
	"go-stock/backend/agent/tools"
	"github.com/cloudwego/eino/components/tool"
)

// Builtins provides all built-in tools for the application
type Builtins struct {
	tools map[string]tool.InvokableTool
}

// NewBuiltins creates and initializes all built-in tools
func NewBuiltins() *Builtins {
	b := &Builtins{
		tools: make(map[string]tool.InvokableTool),
	}

	// Register all built-in tools
	b.registerTool(tools.GetQueryEconomicDataTool())
	b.registerTool(tools.GetQueryStockPriceInfoTool())
	b.registerTool(tools.GetQueryStockCodeInfoTool())
	b.registerTool(tools.GetQueryMarketNewsTool())
	b.registerTool(tools.GetChoiceStockByIndicatorsTool())
	b.registerTool(tools.GetStockKLineTool())
	b.registerTool(tools.GetInteractiveAnswerDataTool())
	b.registerTool(tools.GetFinancialReportTool())
	b.registerTool(tools.GetQueryStockNewsTool())
	b.registerTool(tools.GetIndustryResearchReportTool())
	b.registerTool(tools.GetQueryBKDictTool())

	return b
}

// registerTool adds a built-in tool to the registry
func (b *Builtins) registerTool(t tool.InvokableTool) {
	if t == nil {
		return
	}

	// Get tool info to get the name
	// Since InvokableTool.Info requires context, we'll execute it here
	info, err := t.Info(context.Background())
	if err != nil {
		// Use a default name if info fails
		return
	}

	b.tools[info.Name] = t
}

// GetAllTools returns all built-in tools
func (b *Builtins) GetAllTools() []tool.InvokableTool {
	allTools := make([]tool.InvokableTool, 0, len(b.tools))

	// Convert map to slice
	for _, tool := range b.tools {
		if tool != nil {
			allTools = append(allTools, tool)
		}
	}

	return allTools
}
