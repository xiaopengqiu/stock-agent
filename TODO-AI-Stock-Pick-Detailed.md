# AI荐股功能优化详细TODO清单

## 📋 任务概览

基于对项目的深度分析，本清单细化了TODO-today.md中的AI荐股优化任务，结合代码结构给出具体的实现路径。

---

## 一、工具调用完善（高优先级）

### 1.1 补充已有函数工具到配置文件
**状态**: 待完成
**涉及文件**:
- `backend/data/tool_factory.go` (line 34-124, getBuiltinToolsMap函数)
- `backend/data/tool_config.go` (line 87-117, getDefaultToolConfig函数)

**当前问题分析**:
- `getBuiltinToolsMap()` 只定义了4个工具：SearchStockByIndicators、GetStockKLine、InteractiveAnswer、GetStockResearchReport
- 但 `backend/agent/tools/` 目录下还有多个已实现但未注册的工具，同时配置文件中还有些不存在的错误工具配置需要删除

**需要补充的工具清单**:

| 工具文件名 | 工具名称 | 当前是否已注册 | 是否需要补充 |
|---------|---------|--------------|------------|
| `stock_code_tool.go` | SearchStockCode | 否 | ✅ 需要 |
| `stock_price_info_tool.go` | GetStockPriceInfo | 否 | ✅ 需要 |
| `stock_k_line_data_tool.go` | GetStockKLineData | 否 | ⚠️ 与GetStockKLine重复？需检查 |
| `stock_news_tool.go` | GetStockNews | 否 | ✅ 需要 |
| `market_news_tool.go` | GetMarketNews | 否 | ✅ 需要 |
| `economic_data_tool.go` | GetEconomicData | 否 | ✅ 需要 |
| `financial_reports_tool.go` | GetFinancialReports | 否 | ✅ 需要 |
| `industry_research_report_tool.go` | GetIndustryResearchReport | 否 | ✅ 需要 |
| `interactive_answer_data_tool.go` | GetInteractiveAnswerData | 否 | ⚠️ 与InteractiveAnswer重复？需检查 |
| `bk_dict_tool.go` | GetBKDict | 否 | ✅ 需要 |

**具体实现步骤**:

1. **检查重复工具**: 对比现有4个工具与新工具的函数，确认是否有功能重复
2. **在 `getBuiltinToolsMap()` 中补充工具定义**:
   - 为每个新工具添加 Tool 结构体定义
   - 设置合适的 Name、Description、Parameters
3. **在 `getDefaultToolConfig()` 中添加默认启用配置**:
   - 每个新工具添加 ToolItem 配置
   - 根据工具重要性设置 Enabled: true/false
4. **验证工具实现存在**: 确保每个注册的工具在 `backend/agent/tools/` 中有对应的实现文件

**代码示例** (需要添加的工具定义模板):
```go
// 在 getBuiltinToolsMap() 中添加
"GetStockPriceInfo": {
    Type: "function",
    Function: ToolFunction{
        Name:        "GetStockPriceInfo",
        Description: "获取股票实时价格信息，包括当前价、涨跌幅、成交量等...",
        Parameters: FunctionParameters{
            Type: "object",
            Properties: map[string]any{
                "stockCode": map[string]any{
                    "type":        "string",
                    "description": "股票代码（A股：sh,sz开头;港股hk开头,美股：us开头）",
                },
            },
            Required: []string{"stockCode"},
        },
    },
},
```

---

### 1.2 复用现有市场快讯功能，封装市场动向分析工具
**状态**: 待完成
**涉及文件**:
- `backend/agent/tools/market_news_tool.go` (已存在，直接使用)
- `backend/data/tool_factory.go` (注册工具到配置文件)
- `backend/agent/agent_api.go` (优化荐股提示词，明确调用市场分析工具)
- `data/skills/ai-stock-pick.md` (更新提示词，要求先分析市场再选股)

**功能需求分析**:
当前AI荐股时直接进行股票筛选，缺乏对整体市场环境的判断。项目中已有`QueryMarketNews`工具（位于`backend/agent/tools/market_news_tool.go`），可以获取财联社电报、市场资讯、全球新闻等数据。**无需新建工具，直接复用现有能力**。

**需要实现的功能**:

1. **将QueryMarketNews工具注册到配置文件**:
   - 当前该工具已实现但可能未在`getBuiltinToolsMap()`中注册
   - 需要在`tool_factory.go`中添加该工具定义
   - 在`getDefaultToolConfig()`中添加默认启用配置
   - 在`data/tools/config.json`文件中添加工具配置

2. **优化AI荐股提示词** (`data/skills/ai-stock-pick.md`):
   - 明确指定荐股流程：先调用市场分析工具 → 再选股
   - 要求AI基于市场热点和趋势板块进行选股
   - 在输出中增加"市场分析"板块

**具体实现步骤**:

1. **在`tool_factory.go`中注册QueryMarketNews工具**:
```go
// 在 getBuiltinToolsMap() 中添加
"QueryMarketNews": {
    Type: "function",
    Function: ToolFunction{
        Name:        "QueryMarketNews",
        Description: "获取市场最新资讯和动态，包括财联社电报、国内外市场资讯、全球新闻等，用于分析当前市场热点和趋势",
        Parameters: FunctionParameters{
            Type:       "object",
            Properties: map[string]any{}, // 该工具无需参数
            Required:   []string{},
        },
    },
},
```

2. **在`getDefaultToolConfig()`中添加默认配置**:
```go
{
    Name:    "QueryMarketNews",
    Type:    "builtin",
    Enabled: true,  // 默认启用
    Config:  map[string]interface{}{},
},
```

3. **优化`data/skills/ai-stock-pick.md`中的提示词**:

在"工作流程"部分，将：
```
### 第一阶段：市场环境分析
1. 分析当前大盘走势（涨跌、成交量、趋势）
2. 分析热门板块资金流向
3. 识别市场热点和风格特征
4. 评估整体市场风险水平
```

修改为：
```
### 第一阶段：市场环境分析（必须调用QueryMarketNews工具）
**必须先调用QueryMarketNews工具获取最新市场资讯，然后进行分析：**

1. **调用QueryMarketNews工具**获取最新市场动态：
   - 财联社电报（实时市场快讯）
   - 国内外市场资讯
   - 全球新闻动态

2. **基于获取的资讯分析**：
   - 当前大盘走势（涨跌、成交量、趋势）
   - 热门板块和资金流向
   - 市场热点主题和概念
   - 政策面和消息面影响
   - 整体市场风险水平

3. **输出市场分析摘要**：在最终报告中展示"市场环境分析"板块
```

在"输出要求"部分，确保包含：
```
### 市场环境分析

[基于QueryMarketNews获取的资讯，分析当前大盘走势、热点板块、资金流向、政策消息、风险水平]
```

**优化后的荐股流程**:
1. AI收到荐股请求
2. AI首先调用`QueryMarketNews`工具获取市场最新资讯
3. AI基于资讯分析市场热点和趋势
4. AI调用`SearchStockByIndicators`等工具筛选符合趋势的股票
5. AI生成包含"市场分析"板块的完整荐股报告

**前端展示优化**:
- 在荐股报告展示页面，确保"市场环境分析"板块清晰可见
- 可以高亮显示市场热点主题词

---

### 1.3 优化AI荐股提示词，增加标准化买卖点建议格式
**状态**: 待完成
**涉及文件**:
- `data/skills/ai-stock-pick.md` (修改提示词，明确输出格式要求)
- `backend/data/stock_pick_service.go` (增强报告解析能力，支持新字段)
- `backend/models/stock_pick_report.go` (确认RecommendationItem字段已包含所需字段)
- `frontend/src/components/ai-stock-pick.vue` (前端展示优化，支持新格式)

**功能需求**:
当前荐股报告格式不统一，缺乏标准化的买卖点建议。需要优化提示词，要求AI返回固定格式的荐股列表，包含明确的买卖点建议字段。

**需要AI返回的标准化字段**:

| 字段名 | 说明 | 示例 |
|-------|------|------|
| 推荐时间 | 荐股时间 | 2025-03-01 10:30 |
| 板块概念 | 所属板块/概念 | 人工智能、芯片 |
| 股票名称 | 股票简称 | 中芯国际 |
| 股票代码 | 股票代码 | sh688981 |
| 最新价 | 当前价格 | 45.67 |
| 推荐时价 | 荐股时价格 | 45.20 |
| 昨收价 | 昨日收盘价 | 44.10 |
| AI建议买入价 | 建议买入价格区间 | 44.50-45.00 |
| AI建议止盈价 | 建议止盈价格 | 52.00 |
| AI建议止损价 | 建议止损价格 | 42.00 |
| 推荐理由 | 核心推荐理由 | 1.行业龙头<br>2.资金流入<br>3.技术突破 |
| 风险提示 | 主要风险提示 | 1.估值偏高<br>2.市场波动风险 |
| 备注 | 其他补充说明 | 建议分批建仓 |
| 操作 | 操作建议 | 买入/观望/卖出 |

**实现步骤**:

1. **修改 `data/skills/ai-stock-pick.md` 提示词文件**:

在"输出要求"部分，将现有的格式要求修改为：

```markdown
## 输出要求

### 完整报告格式（严格遵循）

请严格按照以下格式输出推荐报告，**无需添加额外的JSON或其他格式数据**：

1. **首先输出市场分析摘要**（基于QueryMarketNews获取的资讯）
2. **然后以表格形式输出推荐股票列表**，表格必须包含以下列：

| 推荐时间 | 板块概念 | 股票名称 | 股票代码 | 最新 | 推荐时 | 昨收 | AI建议买入价 | AI建议止盈价 | AI建议止损价 | 推荐理由 | 风险提示 | 备注 | 操作 |
|---------|---------|---------|---------|------|--------|------|-------------|-------------|-------------|---------|---------|------|------|
| YYYY-MM-DD HH:mm | XX概念/板块 | XXX | sh/sz/xx000000 | XX.XX | XX.XX | XX.XX | XX.XX-XX.XX | XX.XX | XX.XX | 1.理由1<br>2.理由2 | 1.风险1<br>2.风险2 | 备注内容 | 买入/观望 |

**表格数据要求**：
- 每行代表一支推荐股票
- 推荐理由和风险提示可用 `<br>` 换行
- 确保所有列都有值，无数据填"-"

3. **最后输出投资建议摘要**：
   - 建议仓位
   - 持有周期
   - 跟踪要素
```

2. **增强报告解析能力** (`backend/data/stock_pick_service.go`):

在 `parseReportToRecommendations()` 函数中，增强表格解析逻辑：

```go
// parseReportToRecommendations 解析荐股报告为结构化数据
func (s *StockPickService) parseReportToRecommendations(reportContent string) []models.RecommendationItem {
    var recommendations []models.RecommendationItem

    // 使用正则表达式或markdown表格解析器提取表格数据
    // 匹配 | 推荐时间 | 板块概念 | 股票名称 | ... | 操作 |

    // 解析每一行数据并填充到 RecommendationItem
    for _, row := range tableRows {
        item := models.RecommendationItem{
            StockName:       row["股票名称"],
            StockCode:       row["股票代码"],
            CurrentPrice:    parseFloat(row["最新"]),
            // ... 其他字段映射
            // 新增字段需要确认模型是否支持
        }
        recommendations = append(recommendations, item)
    }

    return recommendations
}
```

3. **确认模型字段** (`backend/models/stock_pick_report.go`):

确认 `RecommendationItem` 结构体已包含所需字段（根据之前读取的文件，该模型已包含大部分字段）：

```go
// 已有字段确认
StockName    string  // 股票名称
StockCode    string  // 股票代码
CurrentPrice float64 // 最新价（最新）
TargetPrice  float64 // 目标价（AI建议止盈价）
RiskLevel    string  // 风险等级
Reason       string  // 推荐理由
RiskTips     string  // 风险提示
TradeSuggestion string // 操作建议（操作）

// 需要新增的字段（如果模型中不存在）
RecommendedPrice float64 // 推荐时价
PreviousClose    float64 // 昨收价
BuyPriceRange    string  // AI建议买入价区间
StopLossPrice    float64 // AI建议止损价
SectorConcept    string  // 板块概念
Remarks          string  // 备注
```

4. **前端展示优化** (`frontend/src/components/ai-stock-pick.vue`):

- 修改报告展示组件，使用表格形式展示推荐股票
- 增加买卖点建议的醒目展示（绿色买入价、红色止损价）
- 支持表格排序和筛选功能

---

### 1.4 工具调用支持配置化（启用/禁用）
**状态**: 待完成
**涉及文件**:
- `backend/data/tool_config.go` (已存在，需要完善)
- `backend/data/tool_factory.go` (已存在，需要完善)
- `backend/agent/agent.go` (需要修改以支持动态工具加载)
- `frontend/src/views/settings.vue` 或新建工具配置页面

**当前状态分析**:
- 工具配置系统已经搭建 (`tool_config.go`, `tool_factory.go`)
- 配置文件路径: `data/tools/config.json`
- 支持3种工具类型: builtin、mcp、http
- 已有缓存机制 (30秒TTL)

**需要完善的功能**:

#### 后端实现

1. **在 `AddTools()` 函数中使用配置加载工具** (`backend/agent/agent.go`)，在配置文件`data/tools/config.json`中新增是否启用的字段，同步修改该配置文件:
```go
func AddTools(ctx context.Context, config *data.AIConfig) ([]tool.BaseTool, error) {
    // 从配置加载启用的工具
    toolConfig, err := data.LoadToolConfig()
    if err != nil {
        return nil, err
    }

    var tools []tool.BaseTool

    for _, toolItem := range toolConfig.Tools {
        // 只加载启用的工具
        if !toolItem.Enabled {
            continue
        }

        // 根据配置创建工具
        toolDef, err := data.CreateTool(toolItem)
        if err != nil {
            logger.SugaredLogger.Warnf("创建工具失败 %s: %v", toolItem.Name, err)
            continue
        }

        // 转换为 eino tool
        einoTool, err := convertToEinoTool(toolDef)
        if err != nil {
            continue
        }

        tools = append(tools, einoTool)
    }

    return tools, nil
}
```

2. **确保配置热更新**:
   - 当前已有30秒缓存TTL
   - 考虑添加配置变更监听器 (文件系统watch)
   - 或者提供手动刷新配置的API

3. **添加工具配置管理API** (在 `app.go` 中添加):
```go
// 获取工具配置
func (a *App) GetToolConfig() (*data.ToolConfig, error)

// 保存工具配置
func (a *App) SaveToolConfig(config *data.ToolConfig) error

// 获取所有可用工具列表（包括未启用的）
func (a *App) GetAvailableTools() []data.ToolItem

// 重置工具配置为默认
func (a *App) ResetToolConfig() error
```

#### 前端实现

1. **创建设置页面或组件** `frontend/src/views/ToolSettings.vue`:
   - 工具列表展示 (表格形式)
   - 每行显示: 工具名称、类型、描述、启用状态(开关)、操作(编辑/删除)
   - 启用/禁用切换实时保存或批量保存

2. **工具管理功能**:
   - 添加自定义工具 (支持HTTP类型)
   - 编辑工具配置
   - 删除自定义工具
   - 重置为默认配置

3. **与后端API对接**:
```typescript
// api/toolSettings.ts
export const getToolConfig = () => request.get('/tool-config')
export const saveToolConfig = (config: ToolConfig) => request.post('/tool-config', config)
export const getAvailableTools = () => request.get('/available-tools')
export const resetToolConfig = () => request.post('/tool-config/reset')
```

---

## 二、历史报告展示问题修复（高优先级）

### 2.1 修复历史报告被截断问题
**状态**: 待完成
**涉及文件**:
- `frontend/src/components/stock-pick-report.vue` (前端展示组件)
- `frontend/src/views/ai-analysis.vue` 或相关历史记录页面

**问题分析**:
根据TODO.md中的描述，当前问题包括:
1. 历史报告无法展示全部内容
2. 划不到底部 (无法滚动查看完整内容)

**可能的原因**:
1. CSS样式问题: 容器高度限制、overflow设置不当
2. 滚动容器问题: 外层容器阻止了内部滚动
3. 内容渲染问题: 动态内容加载后未更新滚动区域高度
4. 组件库限制: NaiveUI的Modal/Drawer组件的默认行为

**排查步骤**:

1. **检查组件当前实现**:
```vue
<!-- stock-pick-report.vue -->
<template>
  <n-modal v-model:show="showModal" ...>
    <!-- 内容区域 -->
    <div class="report-content">
      <!-- 报告内容 -->
    </div>
  </n-modal>
</template>
```

2. **常见修复方案**:

方案A: 确保内容区域可滚动
```css
.report-content {
  max-height: 70vh; /* 视口高度的70% */
  overflow-y: auto;
  padding: 20px;
}
```

方案B: 使用NaiveUI的scrollable属性
```vue
<n-modal
  v-model:show="showModal"
  :style="{ width: '800px' }"
  preset="card"
  title="荐股报告"
  :scrollable="true"  <!-- 关键属性 -->
>
  <div class="report-content">
    <!-- 内容 -->
  </div>
</n-modal>
```

方案C: 使用Drawer代替Modal (更适合长内容)
```vue
<n-drawer
  v-model:show="showDrawer"
  :width="800"
  placement="right"
>
  <n-drawer-content title="荐股报告" closable>
    <div class="report-content">
      <!-- 内容 -->
    </div>
  </n-drawer-content>
</n-drawer>
```

**具体实现步骤**:

1. **定位问题组件**:
   - 找到历史报告展示的组件
   - 检查当前使用的组件 (Modal/Drawer/其他)

2. **应用修复**:
   - 根据当前组件类型选择合适的修复方案
   - 确保内容区域有正确的overflow设置
   - 确保有最大高度限制 (max-height)

3. **测试验证**:
   - 打开一个长报告
   - 验证可以滚动到最底部
   - 验证所有内容都可见

---

## 三、荐股策略优化（中优先级）

### 3.1 整合市场动向到荐股流程
**状态**: 依赖1.2市场动向工具
**涉及文件**:
- `backend/agent/agent_api.go` (修改系统提示词)
- `backend/agent/agent.go` (确保工具调用顺序)

**实现要点**:
- 在系统提示词中明确指定分析顺序: 市场 → 板块 → 个股
- 确保AI优先调用市场动向工具
- 在荐股报告中增加市场分析板块

---

## 四、前端展示优化（中优先级）

### 4.1 买卖点展示组件
**状态**: 依赖1.3买卖点建议
**涉及文件**:
- `frontend/src/components/stock-pick-report.vue` (新增买卖点展示区域)
- 可能新建: `frontend/src/components/stock-suggestion-card.vue`

**设计要求**:
- 每支推荐股票使用卡片式展示
- 买点用绿色标签/箭头表示
- 卖点用红色标签/箭头表示
- 包含止损位、目标价、持仓周期等关键信息

---

## 五、配置系统完善（中优先级）

### 5.1 工具配置页面
**状态**: 依赖1.4后端API完善
**涉及文件**:
- 新建: `frontend/src/views/ToolSettings.vue`
- 修改: `frontend/src/router/index.ts` (添加路由)
- 可能修改: `frontend/src/layout/MainLayout.vue` (添加菜单入口)

**功能清单**:
- [ ] 工具列表展示 (表格)
- [ ] 启用/禁用开关 (带确认)
- [ ] 自定义HTTP工具添加表单
- [ ] 配置重置功能
- [ ] 实时保存或批量保存选项

### 5.2 将所有内置工具都注册成配置文件
将所有的工具都注册成配置文件写入`data/tools/config.json`中，通过配置文件实现动态可插拔

---

## 附录: 工具注册状态总览

| 序号 | 工具名称 | 实现状态 | 注册状态 | 优先级 |
|-----|---------|---------|---------|-------|
| 1 | SearchStockByIndicators | ✅ | ✅ | 高 |
| 2 | GetStockKLine | ✅ | ✅ | 高 |
| 3 | InteractiveAnswer | ✅ | ✅ | 中 |
| 4 | GetStockResearchReport | ✅ | ✅ | 中 |
| 5 | SearchStockCode | ✅ | ❌ | 高 |
| 6 | GetStockPriceInfo | ✅ | ❌ | 高 |
| 7 | GetStockKLineData | ✅ | ❌ | 中 (检查是否重复) |
| 8 | GetStockNews | ✅ | ❌ | 中 |
| 9 | GetMarketNews | ✅ | ❌ | 高 |
| 10 | GetEconomicData | ✅ | ❌ | 低 |
| 11 | GetFinancialReports | ✅ | ❌ | 中 |
| 12 | GetIndustryResearchReport | ✅ | ❌ | 中 |
| 13 | GetInteractiveAnswerData | ✅ | ❌ | 低 (检查是否重复) |
| 14 | GetBKDict | ✅ | ❌ | 低 |
| 15 | GetMarketTrend | ❌ | ❌ | 高 (新增) |

---

## 快速启动建议

根据优先级和依赖关系，建议按以下顺序执行:

**第1周 - 基础工具完善**:
1. ✅ 补充已有函数工具到配置文件 (1.1)
2. ✅ 修复历史报告被截断问题 (2.1)

**第2周 - 核心功能增强**:
3. ✅ 新增市场动向总结工具 (1.2)
4. ✅ 荐股报告增加买卖点建议 (1.3)

**第3周 - 配置系统完善**:
5. ✅ 工具调用支持配置化 (1.4 后端)
6. ✅ 工具配置前端页面 (5.1)

---

*本文档由Claude Code生成，基于对项目的深度分析，详细列出了AI荐股优化的具体实现步骤。建议按优先级逐步实施。*
