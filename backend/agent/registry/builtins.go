package registry

import (
	"context"
	"github.com/cloudwego/eino/components/tool"
	"go-stock/backend/agent/tools"
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
	b.registerBuiltinTool(tools.GetQueryEconomicDataTool())
	b.registerBuiltinTool(tools.GetQueryStockPriceInfoTool())
	b.registerBuiltinTool(tools.GetQueryStockCodeInfoTool())
	b.registerBuiltinTool(tools.GetQueryMarketNewsTool())
	b.registerBuiltinTool(tools.GetChoiceStockByIndicatorsTool())
	b.registerBuiltinTool(tools.GetStockKLineTool())
	b.registerBuiltinTool(tools.GetInteractiveAnswerDataTool())
	b.registerBuiltinTool(tools.GetFinancialReportTool())
	b.registerBuiltinTool(tools.GetQueryStockNewsTool())
	b.registerBuiltinTool(tools.GetIndustryResearchReportTool())
	b.registerBuiltinTool(tools.GetQueryBKDictTool())
	b.registerBuiltinTool(tools.GetQueryHKStockPriceTool())     // 港股价格查询工具
	b.registerBuiltinTool(tools.GetQueryShareholderCountTool()) // 股东人数查询工具

	return b
}

// registerBuiltinTool adds a built-in tool to the registry
func (b *Builtins) registerBuiltinTool(t tool.InvokableTool) {
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

	builtinToolName := info.Name
	builtinToolName = builtinToolName + builtinToolSuffix
	b.tools[builtinToolName] = t
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
