# ToolExecutor 统一工具调用重构设计文档

## 1. 概述

本文档记录了 `AskAiWithTools` 函数从硬编码工具调用逻辑到使用 `ToolExecutor` 统一执行的重构过程。

## 2. 问题分析

### 2.1 原有问题

在 `backend/data/openai_api.go` 的 `AskAiWithTools` 函数中存在以下问题：

1. **新旧逻辑混杂**：同时存在两种工具调用方式
   - 新方式：通过 `ToolExecutor` 统一执行（第 1157-1218 行）
   - 旧方式：通过 `if-else` 硬编码匹配工具名（第 1224-1580 行）

2. **递归调用参数缺失**：第 1583 行递归调用 `AskAiWithTools` 时缺少 `executor` 参数

3. **代码冗余**：大量重复的工具调用处理逻辑

### 2.2 需要清理的旧代码

旧的硬编码逻辑处理了以下工具（均已删除）：
- `SearchStockByIndicators` - 搜索股票
- `GetStockKLine` - 获取 K 线数据
- `InteractiveAnswer` - 投资互动数据
- `GetStockResearchReport` - 股票研究报告
- `QueryBKDictInfo` - 板块字典（已注释）
- `GetIndustryResearchReport` - 行业研究报告（已注释）

## 3. 解决方案

### 3.1 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                   OpenAi API Layer                         │
├─────────────────────────────────────────────────────────────┤
│  AskAiWithTools()                                           │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  1. Receive tool_calls from LLM streaming response   │  │
│  │  2. Collect function name + arguments                │  │
│  │  3. Send "tool start" event to frontend             │  │
│  │  4. Call executor.Execute(ctx, funcName, args)      │  │
│  │  5. Add tool result to messages                      │  │
│  │  6. Recursive call AskAiWithTools() with executor   │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              ToolExecutor Layer (toolexec package)          │
├─────────────────────────────────────────────────────────────┤
│  type ToolExecutor struct {                                 │
│      tools map[string]tool.InvokableTool                   │
│  }                                                           │
│                                                              │
│  NewToolExecutor(tools InvokableToolMap)                   │
│  Execute(ctx, toolName, arguments) (string, error)        │
│  GetToolCount() int                                         │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              InvokableTool Implementations                  │
├─────────────────────────────────────────────────────────────┤
│  SearchStockByIndicatorsTool                                │
│  GetStockKLineTool                                          │
│  InteractiveAnswerTool                                      │
│  GetStockResearchReportTool                                 │
│  ... (all registered tools)                                 │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 核心修改

#### 修改 1：清理旧的 if-else 逻辑

**文件**: `backend/data/openai_api.go`

**删除内容**:
- 第 1224-1580 行的所有硬编码工具调用逻辑
- 第 1583 行缺少 executor 参数的递归调用

**保留内容**:
- 第 1157-1218 行的 ToolExecutor 统一执行逻辑

#### 修改 2：添加工具调用前端通知

在执行工具前，添加发送到前端的通知：

```go
// 发送工具调用开始通知到前端
ch <- map[string]any{
    "code":     1,
    "question": question,
    "chatId":   streamResponse.Id,
    "model":    streamResponse.Model,
    "content":  "\r\n```\r\n开始调用工具：" + funcName + "，\n参数：" + funcArguments + "\r\n```\r\n",
    "time":     time.Now().Format(time.DateTime),
}
```

#### 修改 3：确保递归调用传递 executor

```go
// 正确的递归调用
AskAiWithTools(o, err, messages, ch, question, tools, executor)
```

## 4. 数据流

### 4.1 完整工具调用流程

```
1. LLM 返回 finish_reason = "tool_calls"
   ↓
2. 收集 functions map[string]string (funcName -> arguments)
   ↓
3. 遍历每个工具调用:
   ├─ 3.1 检查 executor 是否为 nil
   ├─ 3.2 记录日志: "Executing tool via executor: {funcName}"
   ├─ 3.3 发送工具开始事件到前端
   ├─ 3.4 调用 executor.Execute(o.ctx, funcName, funcArguments)
   ├─ 3.5 处理执行结果:
   │   ├─ 成功: 添加 tool response message
   │   └─ 失败: 添加 error message
   └─ 3.6 更新 messages 数组
   ↓
4. 递归调用 AskAiWithTools() 继续对话
   ↓
5. LLM 基于工具结果生成最终回答
```

## 5. ToolExecutor 接口

### 5.1 类型定义

**文件**: `backend/toolexec/executor.go`

```go
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
```

## 6. 调用点

### 6.1 StockPickService 中

**文件**: `backend/data/stock_pick_service.go` 第 150-151 行

```go
executor := toolexec.NewToolExecutor(s.AiInvokeTools)
AskAiWithTools(openAI, nil, msg, ch, req.UserQuery, s.AiTools, executor)
```

### 6.2 NewSummaryStockNewsStreamWithTools 中

**文件**: `backend/data/openai_api.go` 第 361 行

```go
AskAiWithTools(o, errors.New(""), msg, ch, userQuestion, tools, executor)
```

### 6.3 NewChatStream 中

**文件**: `backend/data/openai_api.go` 第 872 行

```go
AskAiWithTools(o, err, msg, ch, question, tools, executor)
```

## 7. 优势

### 7.1 可扩展性

- 新增工具无需修改 `AskAiWithTools` 函数
- 只需在工具注册表中注册新工具
- 自动支持所有已注册的工具

### 7.2 代码简洁性

- 删除了约 360 行重复的硬编码逻辑
- 统一的错误处理
- 统一的消息格式化

### 7.3 可维护性

- 工具执行逻辑集中在 `toolexec` 包
- 清晰的职责分离
- 易于测试和调试

### 7.4 前端体验

- 统一的工具调用开始通知
- 显示工具名称和参数
- 保持用户界面一致性

## 8. 兼容性

### 8.1 向后兼容

- `ToolExecutor` 定义保持不变
- `AskAiWithTools` 函数签名保持不变
- 所有调用点已正确传递 executor 参数

### 8.2 工具兼容性

- 所有现有工具继续正常工作
- 工具实现无需修改
- 通过 `InvokableTool` 接口统一调用

## 9. 验证清单

- [x] 删除所有旧的 if-else 硬编码逻辑
- [x] 修复递归调用的 executor 参数缺失问题
- [x] 添加工具调用前端通知
- [x] 验证 ToolExecutor 正确初始化
- [x] 验证所有调用点正确传递 executor
- [x] 验证工具执行流程正常
- [x] 验证错误处理正确

## 10. 总结

本次重构成功实现了：

1. **清理技术债务**：删除了 360+ 行冗余的旧代码
2. **统一工具执行**：所有工具通过 `ToolExecutor` 统一执行
3. **修复 Bug**：解决了递归调用参数缺失的问题
4. **改善用户体验**：添加了工具调用开始的前端通知
5. **提升可维护性**：代码结构更清晰，易于扩展新工具

重构后的代码更加简洁、可维护，并且为未来支持更多工具（包括 MCP 动态工具）奠定了良好基础。
