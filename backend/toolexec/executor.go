package toolexec

import (
	"context"
	"fmt"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/tidwall/gjson"
)

// ToolExecutor 工具执行器
type ToolExecutor struct {
	tools          map[string]tool.InvokableTool
	results        *models.ToolCallResultsCollection
	resultCallback func(result models.ToolCallResult)
}

// InvokableToolMap 工具映射类型
type InvokableToolMap map[string]tool.InvokableTool

// NewToolExecutor 创建新的工具执行器
func NewToolExecutor(tools InvokableToolMap) *ToolExecutor {
	executor := &ToolExecutor{
		tools:   make(map[string]tool.InvokableTool),
		results: &models.ToolCallResultsCollection{Results: []models.ToolCallResult{}},
	}

	for name, t := range tools {
		executor.tools[name] = t
		logger.SugaredLogger.Infof("Registered tool in executor: %s", name)
	}

	return executor
}

// SetResultCallback 设置工具调用结果回调
func (e *ToolExecutor) SetResultCallback(callback func(result models.ToolCallResult)) {
	e.resultCallback = callback
}

// GetResults 获取所有工具调用结果
func (e *ToolExecutor) GetResults() *models.ToolCallResultsCollection {
	return e.results
}

// GetResultsJSON 获取工具调用结果的JSON字符串
func (e *ToolExecutor) GetResultsJSON() (string, error) {
	return e.results.ToJSON()
}

// isNewsTool 判断是否是新闻/舆情相关工具
func isNewsTool(toolName string) bool {
	newsToolNames := []string{
		"QueryStockNewsTool",
		"QueryMarketNews",
		"GetStockResearchReport",
		"GetIndustryResearchReport",
		"QueryInteractiveAnswerData",
	}
	for _, name := range newsToolNames {
		if strings.Contains(toolName, name) || toolName == name {
			return true
		}
	}
	return false
}

// extractStockCodeFromArgs 从工具参数中提取股票代码
func extractStockCodeFromArgs(arguments string) string {
	// 尝试从JSON参数中提取常见的股票代码字段
	stockCodeFields := []string{"stockCode", "stock_code", "code"}
	for _, field := range stockCodeFields {
		if code := gjson.Get(arguments, field).String(); code != "" {
			return code
		}
	}
	return ""
}

// extractStockNameFromArgs 从工具参数中提取股票名称或搜索关键词
func extractStockNameFromArgs(arguments string) string {
	// 尝试从JSON参数中提取常见的字段
	searchFields := []string{"searchWords", "search_words", "keyWord", "keyword", "words"}
	for _, field := range searchFields {
		if name := gjson.Get(arguments, field).String(); name != "" {
			return name
		}
	}
	return ""
}

// Execute 执行工具调用
func (e *ToolExecutor) Execute(ctx context.Context, toolName string, arguments string) (string, error) {
	t, exists := e.tools[toolName]
	if !exists {
		return "", fmt.Errorf("tool not found: %s", toolName)
	}

	callTime := time.Now()

	// 调用工具的InvokableRun方法
	result, err := t.InvokableRun(ctx, arguments)
	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	// 记录工具调用结果
	callResult := models.ToolCallResult{
		ToolName:   toolName,
		Arguments:  arguments,
		Result:     result,
		CallTime:   callTime,
		StockCode:  extractStockCodeFromArgs(arguments),
		StockName:  extractStockNameFromArgs(arguments),
		IsNewsTool: isNewsTool(toolName),
	}

	e.results.AddResult(callResult)

	// 调用回调函数
	if e.resultCallback != nil {
		e.resultCallback(callResult)
	}

	logger.SugaredLogger.Infof("Tool executed: %s, isNews: %v, stockCode: %s",
		toolName, callResult.IsNewsTool, callResult.StockCode)

	return result, nil
}

// GetToolCount 获取已注册的工具数量
func (e *ToolExecutor) GetToolCount() int {
	return len(e.tools)
}
