# 全局工具执行器 (GlobalToolExecutor) 设计文档

## 1. 概述

本文档描述了在 `backend/data` 包中实现全局工具执行器的设计方案，解决 `AskAiWithTools` 函数参数传递复杂的问题。

## 2. 问题分析

### 2.1 原有问题

在使用 `ToolExecutor` 重构后，`AskAiWithTools` 函数存在以下问题：

1. **参数传递复杂**：每个调用点都需要手动创建并传递 `executor` 参数
2. **递归调用易错**：递归调用 `AskAiWithTools` 时容易遗漏 `executor` 参数
3. **StockPickService 重复工作**：需要在 `StockPickService` 中存储 `AiInvokeTools` 并每次创建 executor

### 2.2 调用点示例

**修改前**:
```go
// stock_pick_service.go
executor := toolexec.NewToolExecutor(s.AiInvokeTools)
AskAiWithTools(openAI, nil, msg, ch, req.UserQuery, s.AiTools, executor)

// 递归调用处 - 容易出错
AskAiWithTools(o, err, messages, ch, question, tools)  // 缺少 executor!
```

## 3. 解决方案

### 3.1 核心思路

在 `backend/data` 包中创建全局变量，在应用启动时初始化，`AskAiWithTools` 优先使用全局 executor。

### 3.2 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                    app.go (main package)                     │
├─────────────────────────────────────────────────────────────┤
│  NewApp() {                                                  │
│    aiInvokeTools := make(map[string]tool.InvokableTool)    │
│    ... 初始化工具 ...                                         │
│                                                               │
│    // 关键步骤：设置全局工具                                  │
│    data.SetGlobalAiInvokeTools(aiInvokeTools)               │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              backend/data/openai_api.go                     │
├─────────────────────────────────────────────────────────────┤
│  // 全局变量                                                │
│  var (                                                     │
│    GlobalAiInvokeTools map[string]tool.InvokableTool     │
│    globalToolExecutor     *ToolExecutor                    │
│  )                                                          │
│                                                              │
│  // 初始化函数                                               │
│  SetGlobalAiInvokeTools(tools) {                            │
│    GlobalAiInvokeTools = tools                              │
│    globalToolExecutor = NewToolExecutor(tools)              │
│  }                                                           │
│                                                              │
│  // 修改后的函数签名                                         │
│  AskAiWithTools(..., executor ...*ToolExecutor) {          │
│    // 优先使用传入的，否则使用全局                          │
│    exec := first(executor) ?? globalToolExecutor           │
│    ...                                                      │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
```

## 4. 实现细节

### 4.1 全局变量定义

**文件**: `backend/data/openai_api.go`

```go
package data

import (
    "github.com/cloudwego/eino/components/tool"
    "go-stock/backend/toolexec"
)

// ToolExecutor 统一从 toolexec 包导入
type ToolExecutor = toolexec.ToolExecutor

// NewToolExecutor 统一从 toolexec 包导入
var NewToolExecutor = toolexec.NewToolExecutor

// ============================================
// 全局工具执行器相关变量
// ============================================

// GlobalAiInvokeTools 全局的可调用工具映射
// 在应用启动时由 app.go 初始化，供 AskAiWithTools 使用
var GlobalAiInvokeTools map[string]tool.InvokableTool

// globalToolExecutor 全局工具执行器（懒加载）
var globalToolExecutor *ToolExecutor

// SetGlobalAiInvokeTools 设置全局可调用工具
// 在应用启动时调用，初始化全局工具映射
func SetGlobalAiInvokeTools(tools map[string]tool.InvokableTool) {
    GlobalAiInvokeTools = tools
    // 创建全局工具执行器
    if tools != nil {
        globalToolExecutor = NewToolExecutor(tools)
        logger.SugaredLogger.Infof(
            "GlobalToolExecutor initialized with %d tools",
            globalToolExecutor.GetToolCount(),
        )
    } else {
        logger.SugaredLogger.Warn("SetGlobalAiInvokeTools called with nil tools")
    }
}

// GetGlobalToolExecutor 获取全局工具执行器
// 如果尚未初始化，返回 nil
func GetGlobalToolExecutor() *ToolExecutor {
    return globalToolExecutor
}
```

### 4.2 AskAiWithTools 函数修改

**文件**: `backend/data/openai_api.go`

**修改前**:
```go
func AskAiWithTools(o *OpenAi, err error, messages []map[string]interface{},
    ch chan map[string]any, question string, tools []Tool, executor *ToolExecutor) {
```

**修改后**:
```go
func AskAiWithTools(o *OpenAi, err error, messages []map[string]interface{},
    ch chan map[string]any, question string, tools []Tool, executor ...*ToolExecutor) {

    // 优先使用传入的 executor，如果没有则使用全局 executor
    var exec *ToolExecutor
    if len(executor) > 0 && executor[0] != nil {
        exec = executor[0]
    } else if globalToolExecutor != nil {
        exec = globalToolExecutor
        logger.SugaredLogger.Infof(
            "Using global ToolExecutor with %d tools",
            exec.GetToolCount(),
        )
    } else {
        logger.SugaredLogger.Error(
            "No ToolExecutor available (neither parameter nor global)",
        )
    }

    // 后续使用 exec 而不是 executor
    if exec == nil {
        // 处理错误...
        return
    }

    result, execErr := exec.Execute(o.ctx, funcName, funcArguments)

    // 递归调用时传递 exec
    AskAiWithTools(o, err, messages, ch, question, tools, exec)
}
```

### 4.3 app.go 初始化

**文件**: `app.go`

```go
func NewApp() *App {
    // ... 其他初始化代码 ...

    tools := agent.GetToolRegistry(ctx).GetAllTools()
    aiInvokeTools := make(map[string]tool.InvokableTool)
    for _, invokableTool := range tools {
        info, err := invokableTool.Info(ctx)
        if err != nil {
            logger.SugaredLogger.Errorf("Failed to get tool info: %v", err)
            continue
        }
        aiInvokeTools[info.Name] = invokableTool
    }

    // ... 转换为 AiTools ...

    // ============================================
    // 关键：初始化全局工具执行器
    // ============================================
    data.SetGlobalAiInvokeTools(aiInvokeTools)

    return &App{
        // ... 其他字段 ...
        AiTools:       aiTools,
        AiInvokeTools: aiInvokeTools,
        // ...
    }
}
```

## 5. 调用点更新

### 5.1 StockPickService (推荐)

**文件**: `backend/data/stock_pick_service.go`

**修改前**:
```go
executor := toolexec.NewToolExecutor(s.AiInvokeTools)
AskAiWithTools(openAI, nil, msg, ch, req.UserQuery, s.AiTools, executor)
```

**修改后** (简化，使用全局 executor):
```go
// 直接调用，不传 executor - 将自动使用全局 executor
AskAiWithTools(openAI, nil, msg, ch, req.UserQuery, s.AiTools)
```

### 5.2 保持兼容的调用点

某些调用点（如 NewSummaryStockNewsStreamWithTools 和 NewChatStream）使用 `agent.GetEnabledTools()`，这些可以继续传递自定义 executor：

```go
// 仍然可以传递自定义 executor（保持向后兼容）
executor := NewToolExecutor(agent.GetEnabledTools(*ctx))
AskAiWithTools(o, errors.New(""), msg, ch, userQuestion, tools, executor)
```

## 6. 优势

### 6.1 简化调用

```go
// 修改前
executor := toolexec.NewToolExecutor(s.AiInvokeTools)
AskAiWithTools(openAI, nil, msg, ch, req.UserQuery, s.AiTools, executor)

// 修改后
AskAiWithTools(openAI, nil, msg, ch, req.UserQuery, s.AiTools)
```

### 6.2 避免递归调用错误

```go
// 修改前 - 容易遗漏
AskAiWithTools(o, err, messages, ch, question, tools)  // ❌ 缺少参数

// 修改后 - 即使不传也会用全局
AskAiWithTools(o, err, messages, ch, question, tools)  // ✅ 自动使用全局
```

### 6.3 向后兼容

- 支持可变参数 `executor ...*ToolExecutor`
- 旧代码继续传递 executor 仍然正常工作
- 新代码可以选择不传，使用全局

### 6.4 性能优化

- 全局 executor 只创建一次
- 避免在每次调用时重复创建 executor

## 7. 初始化流程

```
1. 程序启动
   ↓
2. NewApp() 执行
   ↓
3. 从 Registry 获取所有工具
   ↓
4. 构建 aiInvokeTools map
   ↓
5. data.SetGlobalAiInvokeTools(aiInvokeTools)
   ├─ 设置 GlobalAiInvokeTools 变量
   └─ 创建 globalToolExecutor
   ↓
6. App 初始化完成
   ↓
7. AskAiWithTools 被调用
   ├─ 有传 executor? → 使用传入的
   └─ 没传? → 使用 globalToolExecutor
```

## 8. 错误处理

### 8.1 全局 executor 未初始化

```go
func AskAiWithTools(..., executor ...*ToolExecutor) {
    var exec *ToolExecutor
    if len(executor) > 0 && executor[0] != nil {
        exec = executor[0]
    } else if globalToolExecutor != nil {
        exec = globalToolExecutor
    } else {
        // 记录错误
        logger.SugaredLogger.Error("No ToolExecutor available")
        // 发送错误到前端
        ch <- map[string]any{
            "code":    0,
            "content": "工具执行器未初始化",
        }
        return
    }
    // ...
}
```

## 9. 文件修改清单

| 文件 | 修改类型 | 说明 |
|------|----------|------|
| `backend/data/openai_api.go` | 修改 | 添加全局变量和初始化函数 |
| `backend/data/openai_api.go` | 修改 | 修改 AskAiWithTools 函数签名 |
| `backend/data/openai_api.go` | 修改 | 更新内部逻辑使用 exec 变量 |
| `app.go` | 修改 | 调用 SetGlobalAiInvokeTools |

## 10. 总结

本次实现通过引入全局工具执行器，解决了以下问题：

1. ✅ **简化调用**：不再需要每次都创建并传递 executor
2. ✅ **避免错误**：递归调用不再需要担心遗漏参数
3. ✅ **向后兼容**：仍然支持传递自定义 executor
4. ✅ **性能优化**：全局 executor 只创建一次
5. ✅ **易于维护**：初始化逻辑集中在 app.go

该方案既解决了当前的参数传递问题，又保持了良好的灵活性和向后兼容性。
