# 内置AI工具文档

本文档详细介绍了 go-stock 应用程序中内置的 AI 工具，这些工具通过 `backend/agent/registry/builtins.go` 进行注册和管理，实际实现位于 `backend/agent/tools/` 目录下。

## 工具列表总览

| 工具名称 | 功能描述 |
|---------|---------|
| QueryBKDictInfo | 获取所有板块/行业名称和代码 |
| QueryEconomicData | 查询宏观经济数据（GDP、CPI、PPI、PMI） |
| QueryStockPriceInfo | 批量获取实时股价数据 |
| QueryStockCodeInfo | 查询股票/指数信息 |
| QueryMarketNews | 获取国内外市场资讯/电报/会议/事件 |
| ChoiceStockByIndicators | 根据自然语言筛选股票 |
| QueryStockKLine | 获取股票K线数据 |
| QueryInteractiveAnswerData | 获取投资者与上市公司互动问答数据 |
| GetFinancialReport | 查询股票财务报表数据 |
| QueryStockNewsTool | 按关键词搜索相关市场资讯/新闻 |
| GetIndustryResearchReport | 获取行业/板块研究报告 |

---

## 工具详细说明

### 1. QueryBKDictInfo（板块/行业字典查询工具）

#### 功能描述
获取所有板块/行业名称或者代码(bkCode,bkName)。

#### 使用方法
无需参数，直接调用即可获取所有板块和行业的基础信息。

#### 代码实现
文件：`bk_dict_tool.go`

```go
func (t ToolQueryBKDict) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "QueryBKDictInfo",
        Desc: "获取所有板块/行业名称或者代码(bkCode,bkName)",
    }, nil
}
```

---

### 2. QueryEconomicData（宏观经济数据查询工具）

#### 功能描述
查询宏观经济数据，支持查询 GDP、CPI、PPI、PMI 等重要经济指标。

#### 参数说明
- `flag`（可选）：查询类型
  - `all`：查询所有宏观经济数据（默认）
  - `GDP`：国内生产总值
  - `CPI`：居民消费价格指数
  - `PPI`：工业品出厂价格指数
  - `PMI`：采购经理人指数

#### 使用示例
```json
{
  "flag": "GDP"
}
```

#### 代码实现
文件：`economic_data_tool.go`

---

### 3. QueryStockPriceInfo（实时股价查询工具）

#### 功能描述
批量获取实时股价数据。

#### 参数说明
- `stockCodes`（必填）：股票代码列表，多个代码用逗号分隔。股票代码必须以 sh、sz、hk 或 us 开头，例如：sz399001,sh600859。

#### 使用示例
```json
{
  "stockCodes": "sz300001,sh600036"
}
```

#### 代码实现
文件：`stock_price_info_tool.go`

---

### 4. QueryStockCodeInfo（股票/指数信息查询工具）

#### 功能描述
查询股票/指数的详细信息，包括股票名称、代码、拼音、首字母、交易所等。

#### 参数说明
- `searchWord`（必填）：股票搜索关键词。

#### 使用示例
```json
{
  "searchWord": "贵州茅台"
}
```

#### 代码实现
文件：`stock_code_tool.go`

---

### 5. QueryMarketNews（市场资讯查询工具）

#### 功能描述
获取国内外市场资讯/电报/会议/事件，包括财联社电报、TradingView 新闻和路透社新闻等。

#### 使用方法
无需参数，直接调用即可获取最新市场资讯。

#### 代码实现
文件：`market_news_tool.go`

---

### 6. ChoiceStockByIndicators（自然语言选股工具）

#### 功能描述
根据自然语言筛选股票，返回符合选股条件的股票相关数据。

#### 参数说明
- `words`（必填）：选股自然语言条件。支持多种选股语法，包括：
  - 直接输入股票名称，如："长电科技,上海贝岭"
  - 技术指标筛选，如："KDJ,MACD,RSI,BOLL"
  - 财务指标筛选，如："PE<30;净利润增长率>50%"
  - 综合条件筛选，如："股价在20日线上，一月之内涨停次数>=1，量比大于1"

#### 使用示例
```json
{
  "words": "股价在20日线上，一月之内涨停次数>=1，量比大于1，换手率大于3%，流通市值大于50亿小于200亿"
}
```

#### 代码实现
文件：`choice_stock_by_indicators_tool.go`

---

### 7. QueryStockKLine（股票K线数据查询工具）

#### 功能描述
获取股票K线数据，支持A股、港股和美股的K线查询。

#### 参数说明
- `stockCode`（必填）：股票代码，格式为 sh、sz、hk 或 us 开头。
- `days`（必填）：获取K线数据的天数。

#### 使用示例
```json
{
  "stockCode": "sh600519",
  "days": "90"
}
```

#### 代码实现
文件：`stock_k_line_data_tool.go`

---

### 8. QueryInteractiveAnswerData（投资者互动问答查询工具）

#### 功能描述
获取投资者与上市公司互动问答的数据，反映当前投资者关注的热点问题。

#### 参数说明
- `page`（必填）：分页号。
- `pageSize`（必填）：分页大小。
- `keyWord`（可选）：搜索关键词，多个关键词用空格隔开。

#### 使用示例
```json
{
  "page": "1",
  "pageSize": "50",
  "keyWord": "贵州茅台 业绩"
}
```

#### 代码实现
文件：`interactive_answer_data_tool.go`

---

### 9. GetFinancialReport（财务报表查询工具）

#### 功能描述
查询股票财务报表数据。

#### 参数说明
- `stockCode`（必填）：股票代码，格式为 sh、sz、hk 或 us 开头。

#### 使用示例
```json
{
  "stockCode": "sh600519"
}
```

#### 代码实现
文件：`financial_reports_tool.go`

---

### 10. QueryStockNewsTool（股票新闻查询工具）

#### 功能描述
按关键词搜索相关市场资讯/新闻。

#### 参数说明
- `searchWords`（必填）：搜索关键词，多个关键词用空格分隔。

#### 使用示例
```json
{
  "searchWords": "贵州茅台 白酒"
}
```

#### 代码实现
文件：`stock_news_tool.go`

---

### 11. GetIndustryResearchReport（行业研究报告查询工具）

#### 功能描述
获取行业/板块研究报告。

#### 参数说明
- `code`（必填）：行业/板块代码。
- `name`（可选）：行业/板块名称。

#### 使用示例
```json
{
  "code": "bk0421",
  "name": "白酒"
}
```

#### 代码实现
文件：`industry_research_report_tool.go`

---

## 工具架构与数据流程

### 工具注册与管理
所有内置工具通过 `builtins.go` 文件进行注册：

```go
func NewBuiltins() *Builtins {
    b := &Builtins{
        tools: make(map[string]tool.InvokableTool),
    }

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
```

### 数据获取与处理
工具通过调用 `backend/data/` 包中的 API 接口获取数据，主要使用以下数据服务：
- `StockDataApi`：股票数据API
- `MarketNewsApi`：市场新闻API
- `SearchStockApi`：股票搜索API

### 结果格式化
工具返回结果通常格式化为 Markdown 表格或结构化文本，方便在聊天界面中展示。

---

## 工具使用场景示例

### 场景1：查询茅台股票信息及财务数据

```json
{
  "tool": "QueryStockCodeInfo",
  "params": {
    "searchWord": "贵州茅台"
  }
}
```

```json
{
  "tool": "GetFinancialReport",
  "params": {
    "stockCode": "sh600519"
  }
}
```

### 场景2：筛选符合条件的科技股

```json
{
  "tool": "ChoiceStockByIndicators",
  "params": {
    "words": "创新药,半导体;PE<30;净利润增长率>50%"
  }
}
```

### 场景3：获取最新市场资讯

```json
{
  "tool": "QueryMarketNews",
  "params": {}
}
```

---

## 扩展与自定义

如果需要添加新的工具，可以按照以下步骤进行：

1. 在 `backend/agent/tools/` 目录下创建新的工具实现文件
2. 实现 `tool.InvokableTool` 接口
3. 在 `builtins.go` 中注册新工具
4. 在前端组件中添加对应的调用和展示逻辑

---

## 注意事项

1. 工具调用需要遵循 API 限制，避免频繁调用
2. 部分工具需要特定格式的参数，使用前请仔细阅读参数说明
3. 工具返回结果的格式可能会根据数据源变化而调整
4. 对于大量数据查询，建议使用分页功能
