# 修复 GetEnabledTools 类型错误

## 问题描述

在 `backend/data/openai_api.go` 中有两处代码导致编译错误：

```go
// 错误代码
executor := NewToolExecutor(agent.GetEnabledTools(*ctx))
AskAiWithTools(o, errors.New(""), msg, ch, userQuestion, tools, executor)
```

**错误信息**：
```
Cannot use 'agent.GetEnabledTools(*ctx)' (type []tool.BaseTool) as the type InvokableToolMap
```

## 原因分析

1. `agent.GetEnabledTools()` 返回的是 `[]tool.BaseTool` 类型
2. `NewToolExecutor()` 需要的是 `InvokableToolMap` (即 `map[string]tool.InvokableTool`) 类型
3. 类型不匹配导致编译错误

## 解决方案

使用已经实现的**全局工具执行器**，不再需要手动创建 executor。

### 修改前

```go
// 第 386 行
executor := NewToolExecutor(agent.GetEnabledTools(*ctx))
AskAiWithTools(o, errors.New(""), msg, ch, userQuestion, tools, executor)

// 第 897 行
if tools != nil && len(tools) > 0 {
    executor := NewToolExecutor(agent.GetEnabledTools(*ctx))
    AskAiWithTools(o, err, msg, ch, question, tools, executor)
} else {
    AskAi(o, err, msg, ch, question)
}
```

### 修改后

```go
// 第 386 行
// 使用全局工具执行器
AskAiWithTools(o, errors.New(""), msg, ch, userQuestion, tools)

// 第 897 行
if tools != nil && len(tools) > 0 {
    // 使用全局工具执行器
    AskAiWithTools(o, err, msg, ch, question, tools)
} else {
    AskAi(o, err, msg, ch, question)
}
```

## 工作原理

### 全局工具执行器初始化流程

```
1. 应用启动 (app.go)
   ↓
2. NewApp() 创建 aiInvokeTools map
   ↓
3. data.SetGlobalAiInvokeTools(aiInvokeTools)
   ├─ 设置 GlobalAiInvokeTools 全局变量
   └─ 创建 globalToolExecutor 全局执行器
   ↓
4. AskAiWithTools 被调用
   ├─ 检查是否传入了 executor 参数
   ├─ 没有传入 → 使用 globalToolExecutor
   └─ 有传入 → 使用传入的 executor (向后兼容)
```

### AskAiWithTools 函数签名

```go
// 使用可变参数实现向后兼容
func AskAiWithTools(
    o *OpenAi,
    err error,
    messages []map[string]interface{},
    ch chan map[string]any,
    question string,
    tools []Tool,
    executor ...*ToolExecutor,  // 可变参数，可选
) {
    // 优先使用传入的 executor，如果没有则使用全局
    var exec *ToolExecutor
    if len(executor) > 0 && executor[0] != nil {
        exec = executor[0]
    } else if globalToolExecutor != nil {
        exec = globalToolExecutor
    }
    // ...
}
```

## 修改文件清单

| 文件 | 修改位置 | 修改内容 |
|------|----------|----------|
| `backend/data/openai_api.go` | 第 386 行 | 删除 executor 创建，直接调用 AskAiWithTools |
| `backend/data/openai_api.go` | 第 897 行 | 删除 executor 创建，直接调用 AskAiWithTools |

## 优势

1. ✅ **修复编译错误** - 类型不匹配问题解决
2. ✅ **简化代码** - 不需要每次都创建 executor
3. ✅ **统一管理** - 所有工具调用使用同一个全局执行器
4. ✅ **向后兼容** - 仍然支持传入自定义 executor
5. ✅ **性能优化** - 全局 executor 只初始化一次

## 验证

- ✅ `backend/data` 包中不再有 `GetEnabledTools` 引用
- ✅ 两处调用点都已修复
- ✅ 使用全局工具执行器，代码更简洁
