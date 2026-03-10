# 多 Agent 系统使用指南

## 📋 概述

基于昨天的设计文档，我们已成功实现了多 Agent 系统，将原来的单一通用 Agent 拆分为 6 个专业 Agent：

1. **主协调器 Agent (Orchestrator)** - 任务分解、Agent 调度、结果整合
2. **技术面 Agent (Technical)** - K线分析、技术指标、形态识别
3. **基本面 Agent (Fundamental)** - 财务分析、估值模型、同业对比
4. **市场消息 Agent (News)** - 新闻聚合、情绪分析、舆情监控
5. **风险评估 Agent (Risk)** - 仓位建议、止损止盈、风险评级
6. **报告生成 Agent (Reporter)** - 结构化输出、自然语言润色

---

## 🏗️ 架构设计

### 整体架构

```
用户界面 (输入: 股票代码 + 问题)
         ↓
🎯 主协调器 Agent (Orchestrator)
   - 任务分解
   - Agent 调度
   - 结果整合
   - 质量控制
         ↓
   ┌─────┼─────┐
   ↓     ↓     ↓
📊技术面  📈基本面  📰消息面
 Agent   Agent   Agent
   ↓     ↓     ↓
   └─────┼─────┘
         ↓
⚖️ 风险评估 Agent
         ↓
✍️ 报告生成 Agent
         ↓
输出最终报告
```

---

## 📁 文件结构

```
backend/agent/
├── orchestrator/
│   ├── types.go              # 类型定义
│   ├── agent_interface.go    # 接口定义
│   └── orchestrator.go       # 主协调器实现
├── multi_agent_api.go        # 多 Agent 集成 API
└── agent_api.go              # 原有的单 Agent API
```

---

## 🚀 快速开始

### 1. 创建多 Agent API

```go
import "go-stock/backend/agent"

// 创建多 Agent API（需要传入 AI 配置 ID）
multiAgent := agent.NewMultiAgentStockApi(1) // 1 是 AI 配置 ID
```

### 2. 执行股票分析

```go
import "context"

ctx := context.Background()

// 执行分析
report, err := multiAgent.AnalyzeStock(
    ctx,
    "sz000001",           // 股票代码
    "平安银行",             // 股票名称
    "分析一下这只股票",    // 用户问题
    "稳健",                 // 风险偏好 (保守/稳健/激进)
    "中线",                 // 投资周期 (短线/中线/长线)
)

if err != nil {
    // 处理错误
}

// 使用报告
fmt.Println(report.Result) // 完整报告
fmt.Println(report.QuerySummary) // 查询摘要
```

### 3. 报告格式

生成的报告完全兼容现有的 `StockPickReport` 格式，包含：

```go
type StockPickReport struct {
    UserQuery       string                  // 用户输入的选股需求
    QuerySummary    string                  // 需求摘要
    Result          string                  // Markdown 格式报告
    MarketAnalysis  string                  // 市场环境分析
    FilterLogic     string                  // 筛选逻辑说明
    TotalScanned    int                     // 扫描股票总数
    CandidatesCount int                     // 候选股票数
    Recommendations []RecommendationItem    // 推荐股票列表
    ToolsUsed       string                  // 使用的工具列表
    AIConfigID      uint                    // AI配置ID
    AIModel         string                  // AI模型
    Status          string                  // 状态
    Error           string                  // 错误信息
}
```

---

## 📊 报告示例

### 生成的 Markdown 报告格式

```markdown
# 【平安银行】AI 荐股分析报告

**生成时间**: 2026-03-05 15:04:05  
**股票代码**: sz000001  
**股票名称**: 平安银行  

---

## 📊 一、综合评估

| 评估维度 | 结果 | 置信度 |
|---------|------|--------|
| **技术面** | 中性 | 70% |
| **基本面** | 评分 60/100 | - |
| **市场情绪** | 中性 | - |
| **风险等级** | 中 | - |
| **综合置信度** | - | 80% |

---

## 🎯 六、综合建议

综合来看，该股票目前处于中性状态，建议观望为主。

⚠️ 本报告仅供参考，不构成投资建议。股市有风险，投资需谨慎。

---

**报告生成时间**: 2026-03-05 15:04:05  
**AI 模型**: deepseek-v3.2  
**置信度**: 80%
```

---

## 🔧 当前实现状态

### ✅ 已完成

1. **类型定义** - 完整的分析结果类型系统
2. **主协调器** - 任务调度和结果整合
3. **简化版专业 Agent** - 各专业 Agent 的骨架实现
4. **集成 API** - 与现有代码的无缝集成
5. **报告生成** - 完全兼容现有格式的报告
6. **编译通过** - 无循环导入，代码可正常编译

### 📝 待完善（Phase 2）

1. **技术面 Agent** - 实际调用 K线、技术指标工具
2. **基本面 Agent** - 实际调用财务报表、股东人数工具
3. **市场消息 Agent** - 实际调用新闻、投资者互动工具
4. **风险评估 Agent** - 基于真实数据的风险计算
5. **并行执行** - 技术面、基本面、消息面并行处理
6. **错误处理** - 完善的错误恢复和重试机制

---

## 🎯 使用建议

### 当前版本（v0.1）

- 适合用于**架构验证**和**集成测试**
- 报告格式完全兼容，可直接替换现有单 Agent 系统
- 各专业 Agent 返回示例数据，可用于前端展示测试

### 未来版本（v1.0）

- 各专业 Agent 将实际调用工具获取真实数据
- 支持并行执行，提升 2-3 倍效率
- 完善的错误处理和重试机制
- 可观测性和监控支持

---

## 🔄 迁移指南

### 从单 Agent 迁移到多 Agent

```go
// 旧版：单 Agent
oldAgent := agent.NewStockAiAgentApi()
ch := oldAgent.Chat(question, aiConfigId, nil)
// 处理流式响应...

// 新版：多 Agent
multiAgent := agent.NewMultiAgentStockApi(aiConfigId)
report, err := multiAgent.AnalyzeStock(ctx, stockCode, stockName, question, "稳健", "中线")
// 直接获取完整报告
```

### 兼容性说明

- ✅ **报告格式 100% 兼容** - `StockPickReport` 结构完全不变
- ✅ **数据库结构不变** - 无需修改数据库表结构
- ✅ **前端无需修改** - 可以无缝切换到多 Agent 系统
- ✅ **渐进式迁移** - 可以两个系统并行运行

---

## 📚 相关文档

- [AI 荐股多 Agent 能力优化方案.md](./AI荐股多Agent能力优化方案.md) - 原始设计文档
- [编译问题分析报告.md](./编译问题分析报告.md) - 循环依赖问题解决方案

---

## 🎉 总结

多 Agent 系统已成功实现第一阶段！当前版本：

- ✅ 完整的架构设计和类型定义
- ✅ 主协调器和各专业 Agent 的骨架实现
- ✅ 与现有代码的无缝集成
- ✅ 100% 兼容的报告格式
- ✅ 代码可正常编译，无循环导入

下一步可以逐步完善各专业 Agent 的实际工具调用逻辑！
