# AI Invoke Tools 重构设计文档

## 1. 需求背景

### 1.1 现状问题

当前 `backend/data/openai_api.go` 中的 `AskAiWithTools` 函数通过硬编码方式处理工具调用：

```go
if funcName == "SearchStockByIndicators" {
    // 硬编码处理
} else if funcName == "GetStockKLine" {
    // 硬编码处理
}
// ... 更多硬编码
```

**问题**：
1. 难以扩展 - 新增工具需要修改 `AskAiWithTools` 函数
2. 无法支持 MCP 工具 - MCP 工具是动态加载的，无法硬编码
3. 代码重复 - 每个工具的处理逻辑都需要单独编写

### 1.2 目标

1. 利用 `app.go` 中已有的 `AiInvokeTools map[string]tool.InvokableTool` 字段
2. 在 `NewApp()` 构造时正确初始化这个 map
3. 在 `AskAiWithTools` 中通过工具名称查找对应的 `InvokableTool` 并执行
4. 支持所有内置工具和 MCP 工具的统一调用

## 2. 架构设计

### 2.1 总体架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Application Layer                           │
│  ┌─────────────┐                                                   │
│  │    App      │  ┌──────────────────────────────────────┐         │
│  │             │  │  AiInvokeTools map[string]tool.InvokableTool  │         │
│  │ - AiTools   │  │                                      │         │
│  │ - AiInvoke  │  │  "QueryMarketNews" -> InvokableTool  │         │
│  │   Tools     │  │  "GetStockKLine"   -> InvokableTool  │         │
│  │             │  │  ...                                 │         │
│  └──────┬──────┘  └──────────────────────────────────────┘         │
│         │                                                          │
│  ┌──────▼──────────────────────────────────────────────────┐      │
│  │                  Tool Registry Layer                       │      │
│  │                                                            │      │
│  │  ┌─────────────────┐      ┌─────────────────────┐        │      │
│  │  │  Built-in Tools │      │     MCP Tools       │        │      │
│  │  │                 │      │                     │        │      │
│  │  │ - QueryMarketNews       │ - Server1.Tool1     │        │      │
│  │  │ - GetStockKLine  │      │ - Server1.Tool2     │        │      │
│  │  │ - SearchStock... │      │ - Server2.Tool1     │        │      │
│  │  │ ...              │      │ ...                 │        │      │
│  │  └─────────────────┘      └─────────────────────┘        │      │
│  │                                                            │      │
│  └────────────────────────────────────────────────────────────┘      │
│                                                                       │
└───────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌───────────────────────────────────────────────────────────────────────┐
│                         AI Interaction Layer                          │
│                                                                       │
│  ┌───────────────────────────────────────────────────────────────┐   │
│  │                    AskAiWithTools Function                     │   │
│  │                                                                │   │
│  │  1. Receive tool call request from AI                         │   │
│  │  2. Lookup tool in AiInvokeTools[funcName]                     │   │
│  │  3. Execute InvokableTool with arguments                        │   │
│  │  4. Return result to AI                                         │   │
│  │                                                                │   │
│  └───────────────────────────────────────────────────────────────┘   │
│                                                                       │
└───────────────────────────────────────────────────────────────────────┘
```

### 2.2 数据流图

```
┌──────────┐     ┌─────────────┐     ┌─────────────────┐     ┌─────────────┐
│  User    │────▶│  AI Model   │────▶│ Tool Call Req   │────▶│ AskAiWith   │
│  Query   │     │             │     │ (funcName,      │     │ Tools       │
└──────────┘     └─────────────┘     │  arguments)     │     └──────┬──────┘
                                     └─────────────────┘            │
                                                                    │
                                                                    ▼
                           ┌────────────────────────────────────────────────┐
                           │  1. funcName = "SearchStockByIndicators"        │
                           │  2. Lookup: AiInvokeTools[funcName]             │
                           │  3. Found: InvokableTool instance               │
                           └────────────────────────────────────────────────┘
                                                                    │
                                                                    ▼
                           ┌────────────────────────────────────────────────┐
                           │  Execute InvokableTool.Invoke(ctx, arguments) │
                           │  ┌────────────────────────────────────────┐   │
                           │  │  Internal Tool Logic:                  │   │
                           │  │  1. Parse arguments (JSON)             │   │
                           │  │  2. Call external API or database       │   │
                           │  │  3. Format results                     │   │
                           │  │  4. Return structured output           │   │
                           │  └────────────────────────────────────────┘   │
                           └────────────────────────────────────────────────┘
                                                                    │
                                                                    ▼
                           ┌────────────────────────────────────────────────┐
                           │  Return tool result to AskAiWithTools          │
                           │  ┌────────────────────────────────────────┐   │
                           │  │  Format as tool response message:      │   │
                           │  │  {                                     │   │
                           │  │    "role": "tool",                    │   │
                           │  │    "content": "...result...",         │   │
                           │  │    "tool_call_id": "xxx"              │   │
                           │  │  }                                     │   │
                           │  └────────────────────────────────────────┘   │
                           └────────────────────────────────────────────────┘
                                                                    │
                                                                    ▼
┌──────────┐     ┌─────────────┐     ┌─────────────────┐     ┌─────────────┐
│  User    │◄────│  AI Model   │◄────│ Continue Chat   │◄────│ Tool Result │
│  See     │     │  Generate   │     │ with context  │     │ Added to    │
│  Answer  │     │  Response   │     │                 │     │ Messages    │
└──────────┘     └─────────────┘     └─────────────────┘     └─────────────┘
```

## 3. 关键接口定义

### 3.1 工具调用接口

```go
package tool

// InvokableTool 是 Eino 框架的工具接口
type InvokableTool interface {
    // Info 返回工具的元数据信息
    Info(ctx context.Context) (*ToolInfo, error)

    // Invoke 执行工具调用
    Invoke(ctx context.Context, arguments string) (string, error)
}

// ToolInfo 包含工具的元数据
type ToolInfo struct {
    Name        string          // 工具唯一名称
    Desc        string          // 工具描述
    ParamsOneOf *schema.OneOf   // 参数定义（JSON Schema）
}
```

### 3.2 全局工具注册表

```go
package main

// App 应用结构体
type App struct {
    ctx               context.Context
    cache             *freecache.Cache
    cron              *cron.Cron
    cronEntrys        map[string]cron.EntryID

    // 工具相关字段
    AiTools           []data.Tool                          // OpenAI 格式的工具定义
    AiInvokeTools     map[string]tool.InvokableTool        // 工具名称到 InvokableTool 的映射

    SponsorInfo       map[string]any
    PromptTemplateSvc *data.PromptTemplateApi
}

// NewApp 创建应用实例
func NewApp() *App {
    // ... 初始化代码

    // 从 Registry 获取所有工具
    tools := agent.GetToolRegistry(ctx).GetAllTools()

    // 初始化 AiInvokeTools map
    aiInvokeTools := make(map[string]tool.InvokableTool)
    for _, invokableTool := range tools {
        info, err := invokableTool.Info(ctx)
        if err != nil {
            logger.SugaredLogger.Errorf("Failed to get tool info: %v", err)
            continue
        }
        aiInvokeTools[info.Name] = invokableTool
    }

    // 转换为 OpenAI 格式的工具定义
    var aiTools []data.Tool
    for _, t := range ConvertInvokableToolsToDataTools(ctx, tools) {
        if t.Type == "" {
            continue
        }
        aiTools = append(aiTools, t)
    }

    return &App{
        // ... 其他字段
        AiTools:           aiTools,
        AiInvokeTools:     aiInvokeTools,
        // ... 其他字段
    }
}
```

### 3.3 统一工具执行接口

```go
package data

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/cloudwego/eino/components/tool"
)

// ToolExecutor 工具执行器
type ToolExecutor struct {
    tools map[string]tool.InvokableTool
}

// NewToolExecutor 创建工具执行器
func NewToolExecutor(tools map[string]tool.InvokableTool) *ToolExecutor {
    return &ToolExecutor{
        tools: tools,
    }
}

// Execute 执行工具调用
func (e *ToolExecutor) Execute(ctx context.Context, funcName string, arguments string) (string, error) {
    // 查找工具
    t, ok := e.tools[funcName]
    if !ok {
        return "", fmt.Errorf("tool not found: %s", funcName)
    }

    // 执行工具
    result, err := t.Invoke(ctx, arguments)
    if err != nil {
        return "", fmt.Errorf("tool execution failed: %w", err)
    }

    return result, nil
}

// GetToolInfo 获取工具信息
func (e *ToolExecutor) GetToolInfo(ctx context.Context, funcName string) (*tool.ToolInfo, error) {
    t, ok := e.tools[funcName]
    if !ok {
        return nil, fmt.Errorf("tool not found: %s", funcName)
    }

    return t.Info(ctx)
}

// GetAllToolNames 获取所有工具名称
func (e *ToolExecutor) GetAllToolNames() []string {
    names := make([]string, 0, len(e.tools))
    for name := range e.tools {
        names = append(names, name)
    }
    return names
}
```

## 4. 修改计划

### 4.1 文件修改清单

| 序号 | 文件路径 | 修改类型 | 说明 |
|------|----------|----------|------|
| 1 | `app.go` | 修复 | 修复 `AiInvokeTools` 初始化 bug |
| 2 | `backend/data/openai_api.go` | 新增 | 添加 `ToolExecutor` 定义 |
| 3 | `backend/data/openai_api.go` | 修改 | 改造 `AskAiWithTools` 使用 `ToolExecutor` |

### 4.2 详细修改内容

#### 修改 1: 修复 `app.go` 中的初始化 bug

**文件**: `app.go`
**行号**: 64-69
**修改类型**: 修复 bug
**说明**: 变量 `aiInvokeTools` 未初始化导致 panic

**当前代码**:
```go
// 第 64-69 行（当前有 bug）
var aiInvokeTools map[string]tool.InvokableTool  // nil map！
for _, invokableTool := range tools {
    info, _ := invokableTool.Info(ctx)
    aiInvokeTools[info.Name] = invokableTool  // 会 panic！
}
```

**修复后代码**:
```go
// 第 64-69 行（修复后）
aiInvokeTools := make(map[string]tool.InvokableTool)  // 正确初始化
for _, invokableTool := range tools {
    info, err := invokableTool.Info(ctx)
    if err != nil {
        logger.SugaredLogger.Errorf("Failed to get tool info: %v", err)
        continue
    }
    aiInvokeTools[info.Name] = invokableTool
}
```

---

#### 修改 2: 在 `openai_api.go` 中添加 `ToolExecutor` 定义

**文件**: `backend/data/openai_api.go`
**行号**: 在文件头部（package 声明之后）
**修改类型**: 新增代码
**说明**: 添加 `ToolExecutor` 结构体和相关方法

**新增代码**:
```go
package data

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/cloudwego/eino/components/tool"
    "go-stock/backend/logger"
)

// ToolExecutor 工具执行器 - 统一执行各种工具调用
type ToolExecutor struct {
    tools map[string]tool.InvokableTool
}

// NewToolExecutor 创建工具执行器
func NewToolExecutor(tools map[string]tool.InvokableTool) *ToolExecutor {
    return &ToolExecutor{
        tools: tools,
    }
}

// Execute 执行工具调用
// funcName: 工具名称
// arguments: JSON 格式的参数
// 返回: 工具执行结果（JSON 字符串）和错误
func (e *ToolExecutor) Execute(ctx context.Context, funcName string, arguments string) (string, error) {
    // 查找工具
    t, ok := e.tools[funcName]
    if !ok {
        return "", fmt.Errorf("tool not found: %s", funcName)
    }

    // 执行工具
    result, err := t.Invoke(ctx, arguments)
    if err != nil {
        logger.SugaredLogger.Errorf("Tool execution failed: %s, error: %v", funcName, err)
        return "", fmt.Errorf("tool execution failed: %w", err)
    }

    return result, nil
}

// GetToolInfo 获取工具信息
func (e *ToolExecutor) GetToolInfo(ctx context.Context, funcName string) (*tool.ToolInfo, error) {
    t, ok := e.tools[funcName]
    if !ok {
        return nil, fmt.Errorf("tool not found: %s", funcName)
    }

    return t.Info(ctx)
}

// GetAllToolNames 获取所有工具名称
func (e *ToolExecutor) GetAllToolNames() []string {
    names := make([]string, 0, len(e.tools))
    for name := range e.tools {
        names = append(names, name)
    }
    return names
}
```

---

#### 修改 3: 改造 `AskAiWithTools` 使用 `ToolExecutor`

**文件**: `backend/data/openai_api.go`
**行号**: 1009 行（函数签名）和第 1142-1505 行（工具调用处理逻辑）
**修改类型**: 修改函数签名和内部逻辑
**说明**: 添加 `ToolExecutor` 参数，替换硬编码的工具调用逻辑

**当前函数签名**:
```go
// 第 1009 行
func AskAiWithTools(o *OpenAi, err error, messages []map[string]interface{},
    ch chan map[string]any, question string, tools []Tool) {
```

**新函数签名**:
```go
// 第 1009 行
func AskAiWithTools(o *OpenAi, err error, messages []map[string]interface{},
    ch chan map[string]any, question string, tools []Tool, executor *ToolExecutor) {
```

**当前工具调用处理逻辑（第 1142-1505 行，简化示意）**:
```go
// 第 1142-1505 行（当前实现）
for funcName, funcArguments := range functions {
    if funcName == "SearchStockByIndicators" {
        // 硬编码处理逻辑
        res := NewSearchStockApi(words).SearchStock(...)
        // ... 格式化结果
    } else if funcName == "GetStockKLine" {
        // 硬编码处理逻辑
        K := NewStockDataApi().GetKLineData(...)
        // ... 格式化结果
    }
    // ... 更多硬编码
}
```

**新工具调用处理逻辑（第 1142-1505 行）**:
```go
// 第 1142-1505 行（新实现）
for funcName, funcArguments := range functions {
    // 使用 ToolExecutor 统一执行工具
    if executor == nil {
        logger.SugaredLogger.Errorf("ToolExecutor is nil, cannot execute tool: %s", funcName)
        continue
    }

    // 记录工具调用
    logger.SugaredLogger.Infof("Executing tool via executor: %s", funcName)

    // 通过 executor 执行工具
    result, err := executor.Execute(o.ctx, funcName, funcArguments)
    if err != nil {
        logger.SugaredLogger.Errorf("Tool execution failed: %s, error: %v", funcName, err)
        // 添加错误信息到消息列表
        messages = append(messages, map[string]interface{}{
            "role":    "assistant",
            "content": currentAIContent.String(),
            "tool_calls": []map[string]any{
                {
                    "id":           currentCallId,
                    "tool_call_id": currentCallId,
                    "type":         "function",
                    "function": map[string]string{
                        "name":       funcName,
                        "arguments":  funcArguments,
                        "parameters": funcArguments,
                    },
                },
            },
        })
        messages = append(messages, map[string]interface{}{
            "role":         "tool",
            "content":      fmt.Sprintf("Error: %v", err),
            "tool_call_id": currentCallId,
        })
        continue
    }

    // 成功执行，添加结果到消息列表
    messages = append(messages, map[string]interface{}{
        "role":    "assistant",
        "content": currentAIContent.String(),
        "tool_calls": []map[string]any{
            {
                "id":           currentCallId,
                "tool_call_id": currentCallId,
                "type":         "function",
                "function": map[string]string{
                    "name":       funcName,
                    "arguments":  funcArguments,
                    "parameters": funcArguments,
                },
            },
        },
    })
    messages = append(messages, map[string]interface{}{
        "role":         "tool",
        "content":      result,
        "tool_call_id": currentCallId,
    })

    // 发送结果到前端（可选）
    ch <- map[string]any{
        "code":     1,
        "question": question,
        "chatId":   streamResponse.Id,
        "model":    streamResponse.Model,
        "content":  fmt.Sprintf("\r\n```\r\nTool %s executed successfully\r\n```\r\n", funcName),
        "time":     time.Now().Format(time.DateTime),
    }
}
```

### 3.3 调用点更新

所有调用 `AskAiWithTools` 的地方都需要更新，传入 `ToolExecutor` 参数：

**文件**: `backend/data/openai_api.go`

1. **第 350 行**（`NewSummaryStockNewsStreamWithTools` 中）：
```go
// 原代码
AskAiWithTools(o, errors.New(""), msg, ch, userQuestion, tools)

// 新代码
executor := NewToolExecutor(app.AiInvokeTools) // 需要获取 app 实例
AskAiWithTools(o, errors.New(""), msg, ch, userQuestion, tools, executor)
```

2. **第 859-862 行**（`NewChatStream` 中）：
```go
// 原代码
if tools != nil && len(tools) > 0 {
    AskAiWithTools(openAI, err, msg, ch, req.UserQuery, tools)
} else {
    AskAi(openAI, err, msg, ch, req.UserQuery)
}

// 新代码 - 需要从 StockPickService 获取 executor
if tools != nil && len(tools) > 0 {
    executor := NewToolExecutor(s.AiInvokeTools) // s 是 StockPickService 实例
    AskAiWithTools(openAI, err, msg, ch, req.UserQuery, tools, executor)
} else {
    AskAi(openAI, err, msg, ch, req.UserQuery)
}
```

## 4. 修改计划

### 4.1 文件修改详情

#### 4.1.1 `app.go`

**修改类型**: Bug 修复
**行号**: 64-69
**修改内容**:

```go
// 修改前（有 bug）
var aiInvokeTools map[string]tool.InvokableTool
for _, invokableTool := range tools {
    info, _ := invokableTool.Info(ctx)
    aiInvokeTools[info.Name] = invokableTool  // panic: nil map
}

// 修改后（修复）
aiInvokeTools := make(map[string]tool.InvokableTool)
for _, invokableTool := range tools {
    info, err := invokableTool.Info(ctx)
    if err != nil {
        logger.SugaredLogger.Errorf("Failed to get tool info: %v", err)
        continue
    }
    aiInvokeTools[info.Name] = invokableTool
}
```

#### 4.1.2 `backend/data/openai_api.go`

**新增代码 1**: `ToolExecutor` 定义（文件顶部）

```go
// ToolExecutor 工具执行器 - 统一执行各种工具调用
type ToolExecutor struct {
    tools map[string]tool.InvokableTool
}

// NewToolExecutor 创建工具执行器
func NewToolExecutor(tools map[string]tool.InvokableTool) *ToolExecutor {
    return &ToolExecutor{
        tools: tools,
    }
}

// Execute 执行工具调用
func (e *ToolExecutor) Execute(ctx context.Context, funcName string, arguments string) (string, error) {
    t, ok := e.tools[funcName]
    if !ok {
        return "", fmt.Errorf("tool not found: %s", funcName)
    }

    result, err := t.Invoke(ctx, arguments)
    if err != nil {
        logger.SugaredLogger.Errorf("Tool execution failed: %s, error: %v", funcName, err)
        return "", fmt.Errorf("tool execution failed: %w", err)
    }

    return result, nil
}
```

**修改 2**: `AskAiWithTools` 函数签名

```go
// 修改前
func AskAiWithTools(o *OpenAi, err error, messages []map[string]interface{},
    ch chan map[string]any, question string, tools []Tool) {

// 修改后
func AskAiWithTools(o *OpenAi, err error, messages []map[string]interface{},
    ch chan map[string]any, question string, tools []Tool, executor *ToolExecutor) {
```

**修改 3**: 工具调用逻辑（第 1142-1505 行）

将硬编码的 `if funcName == "xxx"` 替换为统一调用：

```go
// 修改后（统一执行）
for funcName, funcArguments := range functions {
    if executor == nil {
        logger.SugaredLogger.Errorf("ToolExecutor is nil, cannot execute tool: %s", funcName)
        continue
    }

    result, err := executor.Execute(o.ctx, funcName, funcArguments)
    if err != nil {
        // 添加错误信息到消息列表
        messages = append(messages, map[string]interface{}{
            "role":    "assistant",
            "content": currentAIContent.String(),
            "tool_calls": []map[string]any{
                {
                    "id":           currentCallId,
                    "tool_call_id": currentCallId,
                    "type":         "function",
                    "function": map[string]string{
                        "name":       funcName,
                        "arguments":  funcArguments,
                        "parameters": funcArguments,
                    },
                },
            },
        })
        messages = append(messages, map[string]interface{}{
            "role":         "tool",
            "content":      fmt.Sprintf("Error: %v", err),
            "tool_call_id": currentCallId,
        })
        continue
    }

    // 成功执行
    messages = append(messages, map[string]interface{}{
        "role":    "assistant",
        "content": currentAIContent.String(),
        "tool_calls": []map[string]any{
            {
                "id":           currentCallId,
                "tool_call_id": currentCallId,
                "type":         "function",
                "function": map[string]string{
                    "name":       funcName,
                    "arguments":  funcArguments,
                    "parameters": funcArguments,
                },
            },
        },
    })
    messages = append(messages, map[string]interface{}{
        "role":         "tool",
        "content":      result,
        "tool_call_id": currentCallId,
    })
}
```

### 4.2 调用点更新

所有调用 `AskAiWithTools` 的地方都需要传入 `ToolExecutor`：

#### 4.2.1 `NewSummaryStockNewsStreamWithTools` 中

**文件**: `backend/data/openai_api.go`
**行号**: 350

```go
// 修改前
AskAiWithTools(o, errors.New(""), msg, ch, userQuestion, tools)

// 修改后
executor := NewToolExecutor(agent.GetEnabledTools(*ctx))  // 从 agent 获取工具
AskAiWithTools(o, errors.New(""), msg, ch, userQuestion, tools, executor)
```

#### 4.2.2 `NewChatStream` 中

**文件**: `backend/data/openai_api.go`
**行号**: 859-862

```go
// 修改前
if tools != nil && len(tools) > 0 {
    AskAiWithTools(openAI, err, msg, ch, req.UserQuery, tools)
} else {
    AskAi(openAI, err, msg, ch, req.UserQuery)
}

// 修改后
if tools != nil && len(tools) > 0 {
    executor := NewToolExecutor(agent.GetEnabledTools(*ctx))
    AskAiWithTools(openAI, err, msg, ch, req.UserQuery, tools, executor)
} else {
    AskAi(openAI, err, msg, ch, req.UserQuery)
}
```

## 5. 风险分析和回滚方案

### 5.1 风险分析

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|----------|
| 工具执行失败 | 中 | 高 | 完善错误处理，确保错误信息返回给 AI |
| 工具找不到 | 低 | 高 | 在 `NewApp` 中严格初始化所有工具 |
| 并发问题 | 低 | 中 | `ToolExecutor` 只读访问 map，安全 |
| 性能下降 | 低 | 中 | 通过反射调用略慢，但可忽略 |
| 与旧代码不兼容 | 中 | 高 | 需要更新所有调用点 |

### 5.2 回滚方案

**情况 1: 需要回滚**

1. 恢复 `app.go` 中的初始化代码（如果有问题）
2. 恢复 `AskAiWithTools` 函数签名为原版本
3. 保留 `ToolExecutor` 代码（不调用即可）

**情况 2: 临时禁用新功能**

在 `AskAiWithTools` 开头添加检查：
```go
func AskAiWithTools(o *OpenAi, err error, messages []map[string]interface{},
    ch chan map[string]any, question string, tools []Tool, executor *ToolExecutor) {

    // 临时禁用开关
    if executor == nil || os.Getenv("DISABLE_TOOL_EXECUTOR") == "true" {
        // 使用旧逻辑
        askAiWithToolsLegacy(o, err, messages, ch, question, tools)
        return
    }

    // 新逻辑...
}
```

### 5.3 测试策略

1. **单元测试**
   - 测试 `ToolExecutor.Execute` 各种场景
   - 测试工具找不到的情况
   - 测试工具执行失败的情况

2. **集成测试**
   - 测试完整 AI 对话流程
   - 测试多个工具连续调用
   - 测试 MCP 工具调用

3. **回归测试**
   - 确保原有功能不受影响
   - 验证所有工具都能正常执行

## 6. 总结

本设计文档详细描述了如何将 AI 工具调用从硬编码方式改造为基于 `ToolExecutor` 的统一调用方式。主要收益包括：

1. **可扩展性**: 新增工具无需修改 `AskAiWithTools` 函数
2. **MCP 支持**: 可以无缝支持动态加载的 MCP 工具
3. **代码简洁**: 消除了大量的 `if funcName == "xxx"` 硬编码
4. **类型安全**: 通过 Eino 框架的 `InvokableTool` 接口保证类型安全

实施后，工具调用的架构将更加清晰，更易于维护和扩展。
