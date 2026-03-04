package toolexec

import (
	"context"
	"fmt"
	"go-stock/backend/logger"

	"github.com/cloudwego/eino/components/tool"
)

// ToolExecutor 工具执行器
type ToolExecutor struct {
	tools map[string]tool.InvokableTool
}

// InvokableToolMap 工具映射类型
type InvokableToolMap map[string]tool.InvokableTool

// NewToolExecutor 创建新的工具执行器
func NewToolExecutor(tools InvokableToolMap) *ToolExecutor {
	executor := &ToolExecutor{
		tools: make(map[string]tool.InvokableTool),
	}

	for name, t := range tools {
		executor.tools[name] = t
		logger.SugaredLogger.Infof("Registered tool in executor: %s", name)
	}

	return executor
}

// Execute 执行工具调用
func (e *ToolExecutor) Execute(ctx context.Context, toolName string, arguments string) (string, error) {
	t, exists := e.tools[toolName]
	if !exists {
		return "", fmt.Errorf("tool not found: %s", toolName)
	}

	// 调用工具的InvokableRun方法
	result, err := t.InvokableRun(ctx, arguments)
	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	return result, nil
}

// GetToolCount 获取已注册的工具数量
func (e *ToolExecutor) GetToolCount() int {
	return len(e.tools)
}
