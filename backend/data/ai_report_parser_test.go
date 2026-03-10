package data

import (
	"go-stock/backend/models"
	"strings"
	"testing"
)

func TestAIReportParser_ParseJSONFormat(t *testing.T) {
	parser := NewAIReportParser()

	// 测试JSON格式
	jsonContent := `{
		"technical_analysis": "技术面分析内容：股价处于上升通道，MACD金叉",
		"fundamental_analysis": "基本面分析内容：PE估值合理，营收增长稳定",
		"risk_analysis": "风险分析内容：注意市场波动风险",
		"sentiment_analysis": "舆情分析内容：近期利好消息不断，机构给予买入评级"
	}`

	result, err := parser.Parse(jsonContent)
	if err != nil {
		t.Fatalf("JSON解析失败: %v", err)
	}

	if result.TechnicalAnalysis == "" {
		t.Error("技术面分析未正确解析")
	}
	if result.FundamentalAnalysis == "" {
		t.Error("基本面分析未正确解析")
	}
	if result.RiskAnalysis == "" {
		t.Error("风险分析未正确解析")
	}
	if result.SentimentAnalysis == "" {
		t.Error("舆情分析未正确解析")
	}
}

func TestAIReportParser_ParseMarkdownFormat(t *testing.T) {
	parser := NewAIReportParser()

	// 测试Markdown格式
	markdownContent := "## 技术面分析\n股价处于上升通道，MACD金叉，KDJ指标超买。\n\n## 基本面分析\nPE估值合理，营收增长稳定，ROE持续提升。\n\n## 风险分析\n注意市场波动风险，政策风险需要关注。\n\n## 舆情动态\n近期利好消息不断，多家券商发布买入评级，投资者互动平台显示公司业务进展顺利。\n"

	result, err := parser.Parse(markdownContent)
	if err != nil {
		t.Fatalf("Markdown解析失败: %v", err)
	}

	if result.TechnicalAnalysis == "" {
		t.Error("技术面分析未正确解析")
	}
	if result.FundamentalAnalysis == "" {
		t.Error("基本面分析未正确解析")
	}
	if result.RiskAnalysis == "" {
		t.Error("风险分析未正确解析")
	}
	if result.SentimentAnalysis == "" {
		t.Error("舆情分析未正确解析")
	}
}

func TestAIReportParser_ParseKeywordFallback(t *testing.T) {
	parser := NewAIReportParser()

	// 测试关键词回退
	content := "这是一份股票分析报告。\n从技术面来看，K线形态良好，均线多头排列。\n基本面方面，公司财务状况良好，营收增长。\n风险提示：市场存在不确定性，建议谨慎操作。\n舆情方面，近期利好消息不断，多家券商发布买入评级，新闻报道公司业务进展顺利。"

	result, _ := parser.Parse(content)

	if result.TechnicalAnalysis == "" {
		t.Error("技术面分析未通过关键词提取")
	}
	if result.FundamentalAnalysis == "" {
		t.Error("基本面分析未通过关键词提取")
	}
	if result.RiskAnalysis == "" {
		t.Error("风险分析未通过关键词提取")
	}
	if result.SentimentAnalysis == "" {
		t.Error("舆情分析未通过关键词提取")
	}
}

func TestAIReportParser_ParseToRecommendationItem(t *testing.T) {
	parser := NewAIReportParser()

	content := "## 技术面分析\n技术面内容测试\n\n## 基本面分析\n基本面内容测试\n\n## 风险分析\n风险内容测试\n\n## 舆情动态\n舆情内容测试\n"

	item := &models.RecommendationItem{}
	parser.ParseToRecommendationItem(content, item)

	if item.TechnicalAnalysis == "" {
		t.Error("TechnicalAnalysis未填充")
	}
	if item.FundamentalAnalysis == "" {
		t.Error("FundamentalAnalysis未填充")
	}
	if item.RiskTips == "" {
		t.Error("RiskTips未填充")
	}
	if item.SentimentAnalysis == "" {
		t.Error("SentimentAnalysis未填充")
	}
}

func TestAIReportParser_ParseMarkdownFormat_Variations(t *testing.T) {
	parser := NewAIReportParser()

	// 测试带冒号的标题
	testCases := []struct {
		name        string
		content     string
		checkField  func(*ParsedAnalysisResult) bool
		errorMsg    string
	}{
		{
			name: "带中文冒号的标题",
			content: "## 技术面分析：\n这是技术面分析内容\n有两行\n\n## 其他章节\n其他内容",
			checkField: func(r *ParsedAnalysisResult) bool { return r.TechnicalAnalysis != "" },
			errorMsg: "带中文冒号的标题解析失败",
		},
		{
			name: "带英文冒号的标题",
			content: "## 技术面分析:\n这是技术面分析内容\n\n## 其他章节",
			checkField: func(r *ParsedAnalysisResult) bool { return r.TechnicalAnalysis != "" },
			errorMsg: "带英文冒号的标题解析失败",
		},
		{
			name: "单#号标题",
			content: "# 技术面分析\n这是技术面分析内容\n\n# 其他章节",
			checkField: func(r *ParsedAnalysisResult) bool { return r.TechnicalAnalysis != "" },
			errorMsg: "单#号标题解析失败",
		},
		{
			name: "英文标题",
			content: "## Technical Analysis\nThis is technical analysis content\n\n## Other Section",
			checkField: func(r *ParsedAnalysisResult) bool { return r.TechnicalAnalysis != "" },
			errorMsg: "英文标题解析失败",
		},
		{
			name: "混合标题格式",
			content: "## 技术面分析\n技术内容\n## 风险提示\n风险内容\n",
			checkField: func(r *ParsedAnalysisResult) bool {
				return r.TechnicalAnalysis != "" && r.RiskAnalysis != ""
			},
			errorMsg: "混合标题格式解析失败",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.Parse(tc.content)
			if err != nil {
				// 即使返回错误，也可能通过关键词回退提取到内容
				result, _ = parser.Parse(tc.content)
			}
			if !tc.checkField(result) {
				t.Error(tc.errorMsg)
			}
		})
	}
}

func TestAIReportParser_ExtractSentimentFromToolResults(t *testing.T) {
	parser := NewAIReportParser()

	// 测试数据：创建工具调用结果集合
	toolResults := &models.ToolCallResultsCollection{
		Results: []models.ToolCallResult{
			{
				ToolName:   "QueryStockNewsTool",
				Arguments:  `{"searchWords": "中芯国际"}`,
				Result:     "中芯国际发布最新财报，营收同比增长30%，机构给予买入评级。公司在先进制程领域取得重大突破。",
				StockCode:  "SH688981",
				StockName:  "中芯国际",
				IsNewsTool: true,
			},
			{
				ToolName:   "GetStockResearchReport",
				Arguments:  `{"stockCode": "SH688981"}`,
				Result:     "券商研报认为中芯国际估值合理，未来增长潜力大。行业景气度持续提升。",
				StockCode:  "SH688981",
				StockName:  "中芯国际",
				IsNewsTool: true,
			},
			{
				ToolName:   "QueryMarketNews",
				Arguments:  `{}`,
				Result:     "今日半导体板块整体上涨，市场情绪乐观。",
				StockCode:  "",
				StockName:  "",
				IsNewsTool: true,
			},
			{
				ToolName:   "GetKLineData", // 非新闻工具
				Arguments:  `{"stockCode": "SH688981"}`,
				Result:     "K线数据：MA5上穿MA10，形成金叉。",
				StockCode:  "SH688981",
				StockName:  "中芯国际",
				IsNewsTool: false,
			},
		},
	}

	// 测试1：通过股票代码匹配
	t.Run("通过股票代码匹配", func(t *testing.T) {
		result := parser.extractSentimentFromToolResults(toolResults, "SH688981", "中芯国际")
		if result == "" {
			t.Error("应该能通过股票代码匹配到舆情内容")
		}
		if !strings.Contains(result, "中芯国际") {
			t.Error("舆情内容应该包含中芯国际相关信息")
		}
		if !strings.Contains(result, "情感分析") {
			t.Error("舆情内容应该包含情感分析")
		}
	})

	// 测试2：通过股票名称匹配
	t.Run("通过股票名称匹配", func(t *testing.T) {
		// 创建一个没有股票代码但有股票名称的测试数据
		toolResults2 := &models.ToolCallResultsCollection{
			Results: []models.ToolCallResult{
				{
					ToolName:   "QueryStockNewsTool",
					Arguments:  `{"searchWords": "中芯国际"}`,
					Result:     "中芯国际利好消息不断，股价上涨。",
					StockCode:  "",
					StockName:  "中芯国际",
					IsNewsTool: true,
				},
			},
		}
		result := parser.extractSentimentFromToolResults(toolResults2, "", "中芯国际")
		if result == "" {
			t.Error("应该能通过股票名称匹配到舆情内容")
		}
	})

	// 测试3：没有匹配结果
	t.Run("没有匹配结果", func(t *testing.T) {
		result := parser.extractSentimentFromToolResults(toolResults, "SZ000001", "平安银行")
		if result != "" {
			t.Error("不应该匹配到舆情内容")
		}
	})

	// 测试4：nil 输入
	t.Run("nil 输入", func(t *testing.T) {
		result := parser.extractSentimentFromToolResults(nil, "SH688981", "中芯国际")
		if result != "" {
			t.Error("nil 输入应该返回空字符串")
		}
	})
}

func TestAIReportParser_ParseBatchWithToolResults(t *testing.T) {
	parser := NewAIReportParser()

	// 创建测试报告内容
	reportContent := `## 市场环境分析
当前市场整体处于震荡上行趋势。

## 推荐股票

### 1. 中芯国际 (SH688981)
**推荐理由:** 半导体龙头，技术领先。

**技术面:** K线形态良好，均线多头排列。

**基本面:** PE估值合理，营收增长稳定。

**舆情:** 近期利好消息不断，机构给予买入评级。

**风险提示:** 注意行业周期波动风险。
`

	// 创建工具调用结果
	toolResults := &models.ToolCallResultsCollection{
		Results: []models.ToolCallResult{
			{
				ToolName:   "QueryStockNewsTool",
				Arguments:  `{"searchWords": "中芯国际"}`,
				Result:     "中芯国际发布最新财报，营收同比增长30%，机构给予买入评级。",
				StockCode:  "SH688981",
				StockName:  "中芯国际",
				IsNewsTool: true,
			},
		},
	}

	// 创建推荐项
	items := []models.RecommendationItem{
		{
			Rank:       1,
			StockCode:  "SH688981",
			StockName:  "中芯国际",
		},
	}

	// 测试解析
	parser.ParseBatchWithToolResults(reportContent, items, toolResults)

	if items[0].SentimentAnalysis == "" {
		t.Error("SentimentAnalysis 应该被填充")
	}

	if items[0].TechnicalAnalysis == "" {
		t.Error("TechnicalAnalysis 应该被填充")
	}

	if items[0].FundamentalAnalysis == "" {
		t.Error("FundamentalAnalysis 应该被填充")
	}

	if items[0].RiskTips == "" {
		t.Error("RiskTips 应该被填充")
	}
}

func TestAIReportParser_MarkdownTableParser(t *testing.T) {
	parser := NewAIReportParser()

	// 测试Markdown表格格式的新闻数据
	newsResult := `
## 中芯国际市场资讯/新闻

| 标题 | 内容 | 时间 |
| --- | --- | --- |
| 中芯国际发布财报 | 中芯国际2024年Q3营收同比增长30%，净利润超预期。公司先进制程取得重大突破。 | 2024-10-15 |
| 机构评级上调 | 多家券商给予中芯国际买入评级，目标价上调至60元。 | 2024-10-14 |
`

	toolResults := &models.ToolCallResultsCollection{
		Results: []models.ToolCallResult{
			{
				ToolName:   "QueryStockNewsTool",
				Arguments:  `{"searchWords": "中芯国际"}`,
				Result:     newsResult,
				StockCode:  "SH688981",
				StockName:  "中芯国际",
				IsNewsTool: true,
			},
		},
	}

	result := parser.extractSentimentFromToolResults(toolResults, "SH688981", "中芯国际")
	if result == "" {
		t.Error("应该能从Markdown表格中解析出舆情内容")
	}
	if !strings.Contains(result, "舆情动态") {
		t.Error("舆情内容应该包含'舆情动态'标题")
	}
	if !strings.Contains(result, "中芯国际发布财报") {
		t.Error("舆情内容应该包含新闻标题")
	}
	if !strings.Contains(result, "情感分析") {
		t.Error("舆情内容应该包含情感分析")
	}
}

func TestAIReportParser_InteractiveAnswerParser(t *testing.T) {
	parser := NewAIReportParser()

	// 测试投资者互动问答数据
	qaResult := `
## 投资互动数据

| 投资者提问 | 上市公司回复 | 股票名称 | 发布时间 |
| --- | --- | --- | --- |
| 请问公司三季度产能利用率如何？ | 公司三季度产能利用率保持在95%以上，订单饱满。 | 中芯国际 | 2024-10-15 |
| 公司先进制程进展如何？ | 我们的7nm工艺已实现量产，良率稳步提升。 | 中芯国际 | 2024-10-10 |
`

	toolResults := &models.ToolCallResultsCollection{
		Results: []models.ToolCallResult{
			{
				ToolName:   "QueryInteractiveAnswerData",
				Arguments:  `{"page": "1", "pageSize": "10", "keyWord": "中芯国际"}`,
				Result:     qaResult,
				StockCode:  "SH688981",
				StockName:  "中芯国际",
				IsNewsTool: true,
			},
		},
	}

	result := parser.extractSentimentFromToolResults(toolResults, "SH688981", "中芯国际")
	if result == "" {
		t.Error("应该能从互动问答中解析出舆情内容")
	}
	if !strings.Contains(result, "投资者问") || !strings.Contains(result, "公司回复") {
		t.Error("舆情内容应该包含投资者问答信息")
	}
}

func TestAIReportParser_ResearchReportParser(t *testing.T) {
	parser := NewAIReportParser()

	// 测试研报数据
	reportResult := `
## 中芯国际研报

### 投资评级：买入

### 核心观点：
中芯国际作为国内晶圆代工龙头，受益于国产替代趋势。公司技术实力不断增强，7nm工艺实现突破。预计2024年净利润增长35%以上。

### 风险提示：
关注国际形势变化及行业周期性波动。
`

	toolResults := &models.ToolCallResultsCollection{
		Results: []models.ToolCallResult{
			{
				ToolName:   "GetStockResearchReport",
				Arguments:  `{"stockCode": "SH688981"}`,
				Result:     reportResult,
				StockCode:  "SH688981",
				StockName:  "中芯国际",
				IsNewsTool: true,
			},
		},
	}

	result := parser.extractSentimentFromToolResults(toolResults, "SH688981", "中芯国际")
	if result == "" {
		t.Error("应该能从研报中解析出舆情内容")
	}
	if !strings.Contains(result, "📊") {
		t.Error("舆情内容应该包含研报标记")
	}
}

func TestAIReportParser_MixedToolsParser(t *testing.T) {
	parser := NewAIReportParser()

	// 混合多种工具结果
	newsResult := `
## 中芯国际新闻

| 标题 | 内容 |
| --- | --- |
| 利好消息 | 中芯国际获得大基金二期增资。 |
`

	qaResult := `
## 投资互动

| 投资者提问 | 上市公司回复 |
| --- | --- |
| 产能情况？ | 产能满载，订单排到明年。 |
`

	toolResults := &models.ToolCallResultsCollection{
		Results: []models.ToolCallResult{
			{
				ToolName:   "QueryStockNewsTool",
				Result:     newsResult,
				StockCode:  "SH688981",
				StockName:  "中芯国际",
				IsNewsTool: true,
			},
			{
				ToolName:   "QueryInteractiveAnswerData",
				Result:     qaResult,
				StockCode:  "SH688981",
				StockName:  "中芯国际",
				IsNewsTool: true,
			},
		},
	}

	result := parser.extractSentimentFromToolResults(toolResults, "SH688981", "中芯国际")
	if result == "" {
		t.Error("应该能从混合工具结果中解析出舆情内容")
	}
	if !strings.Contains(result, "新闻") || !strings.Contains(result, "互动问答") {
		t.Error("舆情内容应该包含多种类型的信息")
	}
}

func TestAIReportParser_FallbackToOldMethod(t *testing.T) {
	parser := NewAIReportParser()

	// 测试无法解析的格式，回退到旧方法
	plainTextResult := "中芯国际近期利好消息不断，公司经营状况良好。股价表现强劲。"

	toolResults := &models.ToolCallResultsCollection{
		Results: []models.ToolCallResult{
			{
				ToolName:   "QueryStockNewsTool",
				Result:     plainTextResult,
				StockCode:  "SH688981",
				StockName:  "中芯国际",
				IsNewsTool: true,
			},
		},
	}

	result := parser.extractSentimentFromToolResults(toolResults, "SH688981", "中芯国际")
	if result == "" {
		t.Error("即使无法解析表格，也应该能回退到旧方法提取内容")
	}
	if !strings.Contains(result, "情感分析") {
		t.Error("回退方法也应该包含情感分析")
	}
}

func TestAIReportParser_MultipleStocksSentiment(t *testing.T) {
	parser := NewAIReportParser()

	// 创建多只股票的工具调用结果
	toolResults := &models.ToolCallResultsCollection{
		Results: []models.ToolCallResult{
			// 股票1: 中芯国际
			{
				ToolName:   "QueryStockNewsTool",
				Arguments:  `{"searchWords": "中芯国际"}`,
				Result:     "中芯国际发布最新财报，营收同比增长30%，机构给予买入评级。",
				StockCode:  "SH688981",
				StockName:  "中芯国际",
				IsNewsTool: true,
			},
			{
				ToolName:   "GetStockResearchReport",
				Arguments:  `{"stockCode": "SH688981"}`,
				Result:     "券商研报认为中芯国际估值合理，未来增长潜力大。",
				StockCode:  "SH688981",
				StockName:  "中芯国际",
				IsNewsTool: true,
			},
			// 股票2: 贵州茅台
			{
				ToolName:   "QueryStockNewsTool",
				Arguments:  `{"searchWords": "贵州茅台"}`,
				Result:     "贵州茅台发布年报，净利润增长20%，拟10派250元。",
				StockCode:  "SH600519",
				StockName:  "贵州茅台",
				IsNewsTool: true,
			},
			{
				ToolName:   "QueryInteractiveAnswerData",
				Arguments:  `{"keyWord": "贵州茅台"}`,
				Result:     "贵州茅台在投资者互动平台表示，公司产能稳步提升，市场需求旺盛。",
				StockCode:  "SH600519",
				StockName:  "贵州茅台",
				IsNewsTool: true,
			},
			// 股票3: 宁德时代
			{
				ToolName:   "QueryStockNewsTool",
				Arguments:  `{"searchWords": "宁德时代"}`,
				Result:     "宁德时代与特斯拉签订长期供货协议，动力电池业务持续增长。",
				StockCode:  "SZ300750",
				StockName:  "宁德时代",
				IsNewsTool: true,
			},
		},
	}

	// 测试1: 提取中芯国际的舆情
	t.Run("提取中芯国际舆情", func(t *testing.T) {
		result := parser.extractSentimentFromToolResultsEnhanced(toolResults, "SH688981", "中芯国际")
		if result == "" {
			t.Error("中芯国际的舆情应该被提取")
		}
		if !strings.Contains(result, "中芯国际") {
			t.Error("舆情内容应该包含中芯国际")
		}
		if strings.Contains(result, "贵州茅台") {
			t.Error("中芯国际的舆情不应该包含贵州茅台")
		}
		if strings.Contains(result, "宁德时代") {
			t.Error("中芯国际的舆情不应该包含宁德时代")
		}
		t.Logf("中芯国际舆情: %s", result)
	})

	// 测试2: 提取贵州茅台的舆情
	t.Run("提取贵州茅台舆情", func(t *testing.T) {
		result := parser.extractSentimentFromToolResultsEnhanced(toolResults, "SH600519", "贵州茅台")
		if result == "" {
			t.Error("贵州茅台的舆情应该被提取")
		}
		if !strings.Contains(result, "贵州茅台") {
			t.Error("舆情内容应该包含贵州茅台")
		}
		if strings.Contains(result, "中芯国际") {
			t.Error("贵州茅台的舆情不应该包含中芯国际")
		}
		if strings.Contains(result, "宁德时代") {
			t.Error("贵州茅台的舆情不应该包含宁德时代")
		}
		t.Logf("贵州茅台舆情: %s", result)
	})

	// 测试3: 提取宁德时代的舆情
	t.Run("提取宁德时代舆情", func(t *testing.T) {
		result := parser.extractSentimentFromToolResultsEnhanced(toolResults, "SZ300750", "宁德时代")
		if result == "" {
			t.Error("宁德时代的舆情应该被提取")
		}
		if !strings.Contains(result, "宁德时代") {
			t.Error("舆情内容应该包含宁德时代")
		}
		if strings.Contains(result, "中芯国际") {
			t.Error("宁德时代的舆情不应该包含中芯国际")
		}
		if strings.Contains(result, "贵州茅台") {
			t.Error("宁德时代的舆情不应该包含贵州茅台")
		}
		t.Logf("宁德时代舆情: %s", result)
	})
}

func TestAIReportParser_ParseBatchWithToolResults_MultipleStocks(t *testing.T) {
	parser := NewAIReportParser()

	// 创建工具调用结果
	toolResults := &models.ToolCallResultsCollection{
		Results: []models.ToolCallResult{
			{
				ToolName:   "QueryStockNewsTool",
				Arguments:  `{"searchWords": "中芯国际"}`,
				Result:     "中芯国际发布最新财报，营收同比增长30%。",
				StockCode:  "SH688981",
				StockName:  "中芯国际",
				IsNewsTool: true,
			},
			{
				ToolName:   "QueryStockNewsTool",
				Arguments:  `{"searchWords": "贵州茅台"}`,
				Result:     "贵州茅台发布年报，净利润增长20%。",
				StockCode:  "SH600519",
				StockName:  "贵州茅台",
				IsNewsTool: true,
			},
		},
	}

	// 创建推荐项
	items := []models.RecommendationItem{
		{
			Rank:       1,
			StockCode:  "SH688981",
			StockName:  "中芯国际",
		},
		{
			Rank:       2,
			StockCode:  "SH600519",
			StockName:  "贵州茅台",
		},
	}

	// 测试解析
	parser.ParseBatchWithToolResults("", items, toolResults)

	// 验证中芯国际的舆情
	if items[0].SentimentAnalysis == "" {
		t.Error("中芯国际的SentimentAnalysis应该被填充")
	}
	if !strings.Contains(items[0].SentimentAnalysis, "中芯国际") {
		t.Error("中芯国际的舆情应该包含中芯国际")
	}
	if strings.Contains(items[0].SentimentAnalysis, "贵州茅台") {
		t.Error("中芯国际的舆情不应该包含贵州茅台")
	}

	// 验证贵州茅台的舆情
	if items[1].SentimentAnalysis == "" {
		t.Error("贵州茅台的SentimentAnalysis应该被填充")
	}
	if !strings.Contains(items[1].SentimentAnalysis, "贵州茅台") {
		t.Error("贵州茅台的舆情应该包含贵州茅台")
	}
	if strings.Contains(items[1].SentimentAnalysis, "中芯国际") {
		t.Error("贵州茅台的舆情不应该包含中芯国际")
	}

	t.Logf("中芯国际舆情: %s", items[0].SentimentAnalysis)
	t.Logf("贵州茅台舆情: %s", items[1].SentimentAnalysis)
}

func TestAIReportParser_ExtractSentiment_FallbackLogic(t *testing.T) {
	parser := NewAIReportParser()

	// 创建工具调用结果 - 注意这里没有设置 StockCode 和 StockName
	// 这种情况会走到放宽条件的逻辑
	toolResults := &models.ToolCallResultsCollection{
		Results: []models.ToolCallResult{
			{
				ToolName:   "QueryStockNewsTool",
				Arguments:  `{"searchWords": "中芯国际"}`,
				Result:     "中芯国际发布最新财报，营收同比增长30%。",
				StockCode:  "",  // 故意留空
				StockName:  "",  // 故意留空
				IsNewsTool: true,
			},
			{
				ToolName:   "QueryStockNewsTool",
				Arguments:  `{"searchWords": "贵州茅台"}`,
				Result:     "贵州茅台发布年报，净利润增长20%。",
				StockCode:  "",  // 故意留空
				StockName:  "",  // 故意留空
				IsNewsTool: true,
			},
		},
	}

	// 测试1: 提取中芯国际的舆情
	t.Run("提取中芯国际舆情-放宽条件", func(t *testing.T) {
		result := parser.extractSentimentFromToolResultsEnhanced(toolResults, "SH688981", "中芯国际")
		if result == "" {
			t.Error("中芯国际的舆情应该被提取")
		}
		if !strings.Contains(result, "中芯国际") {
			t.Error("舆情内容应该包含中芯国际")
		}
		if strings.Contains(result, "贵州茅台") {
			t.Error("中芯国际的舆情不应该包含贵州茅台")
		}
		t.Logf("中芯国际舆情: %s", result)
	})

	// 测试2: 提取贵州茅台的舆情
	t.Run("提取贵州茅台舆情-放宽条件", func(t *testing.T) {
		result := parser.extractSentimentFromToolResultsEnhanced(toolResults, "SH600519", "贵州茅台")
		if result == "" {
			t.Error("贵州茅台的舆情应该被提取")
		}
		if !strings.Contains(result, "贵州茅台") {
			t.Error("舆情内容应该包含贵州茅台")
		}
		if strings.Contains(result, "中芯国际") {
			t.Error("贵州茅台的舆情不应该包含中芯国际")
		}
		t.Logf("贵州茅台舆情: %s", result)
	})
}

func TestAIReportParser_Sentiment_Cross_Contamination(t *testing.T) {
	parser := NewAIReportParser()

	// 这个测试模拟一种情况：工具结果中混合了有股票代码和没有股票代码的情况
	toolResults := &models.ToolCallResultsCollection{
		Results: []models.ToolCallResult{
			// 中芯国际的新闻 - 有股票代码
			{
				ToolName:   "QueryStockNewsTool",
				Arguments:  `{"searchWords": "中芯国际"}`,
				Result:     "中芯国际发布最新财报，营收同比增长30%。",
				StockCode:  "SH688981",
				StockName:  "中芯国际",
				IsNewsTool: true,
			},
			// 贵州茅台的新闻 - 没有股票代码（这种情况会走到放宽条件）
			{
				ToolName:   "QueryStockNewsTool",
				Arguments:  `{"searchWords": "贵州茅台"}`,
				Result:     "贵州茅台发布年报，净利润增长20%。",
				StockCode:  "",
				StockName:  "",
				IsNewsTool: true,
			},
		},
	}

	// 先提取贵州茅台的舆情（因为它没有股票代码，会走到放宽条件逻辑）
	t.Run("先提取贵州茅台", func(t *testing.T) {
		result := parser.extractSentimentFromToolResultsEnhanced(toolResults, "SH600519", "贵州茅台")
		if result == "" {
			t.Error("贵州茅台的舆情应该被提取")
		}
		if !strings.Contains(result, "贵州茅台") {
			t.Error("舆情内容应该包含贵州茅台")
		}
		if strings.Contains(result, "中芯国际") {
			t.Error("贵州茅台的舆情不应该包含中芯国际")
		}
		t.Logf("贵州茅台舆情: %s", result)
	})

	// 再提取中芯国际的舆情
	t.Run("再提取中芯国际", func(t *testing.T) {
		result := parser.extractSentimentFromToolResultsEnhanced(toolResults, "SH688981", "中芯国际")
		if result == "" {
			t.Error("中芯国际的舆情应该被提取")
		}
		if !strings.Contains(result, "中芯国际") {
			t.Error("舆情内容应该包含中芯国际")
		}
		if strings.Contains(result, "贵州茅台") {
			t.Error("中芯国际的舆情不应该包含贵州茅台")
		}
		t.Logf("中芯国际舆情: %s", result)
	})
}

// 这是一个关键测试 - 模拟 ParseBatchWithToolResults 被调用时的场景
func TestAIReportParser_ParseBatch_RealScenario(t *testing.T) {
	parser := NewAIReportParser()

	// 创建 Markdown 表格格式的工具结果
	smicNews := `## 中芯国际市场资讯/新闻

| 标题 | 内容 | 时间 |
| --- | --- | --- |
| 中芯国际发布财报 | 中芯国际2024年Q3营收同比增长30%，净利润超预期。 | 2024-10-15 |
`

	moutaiNews := `## 贵州茅台市场资讯/新闻

| 标题 | 内容 | 时间 |
| --- | --- | --- |
| 贵州茅台发布年报 | 贵州茅台2024年营收增长20%，拟10派250元。 | 2024-10-16 |
`

	// 创建工具调用结果
	toolResults := &models.ToolCallResultsCollection{
		Results: []models.ToolCallResult{
			{
				ToolName:   "QueryStockNewsTool",
				Arguments:  `{"searchWords": "中芯国际"}`,
				Result:     smicNews,
				StockCode:  "SH688981",
				StockName:  "中芯国际",
				IsNewsTool: true,
			},
			{
				ToolName:   "QueryStockNewsTool",
				Arguments:  `{"searchWords": "贵州茅台"}`,
				Result:     moutaiNews,
				StockCode:  "SH600519",
				StockName:  "贵州茅台",
				IsNewsTool: true,
			},
		},
	}

	// 创建推荐项
	items := []models.RecommendationItem{
		{
			Rank:       1,
			StockCode:  "SH688981",
			StockName:  "中芯国际",
		},
		{
			Rank:       2,
			StockCode:  "SH600519",
			StockName:  "贵州茅台",
		},
	}

	// 调用解析 - 这会复现问题！
	parser.ParseBatchWithToolResults("", items, toolResults)

	// 验证中芯国际的舆情
	t.Logf("=== 中芯国际舆情 ===")
	t.Logf("%s", items[0].SentimentAnalysis)
	if items[0].SentimentAnalysis == "" {
		t.Error("中芯国际的SentimentAnalysis应该被填充")
	}
	if !strings.Contains(items[0].SentimentAnalysis, "中芯国际") {
		t.Error("中芯国际的舆情应该包含中芯国际")
	}
	if strings.Contains(items[0].SentimentAnalysis, "贵州茅台") {
		t.Error("中芯国际的舆情不应该包含贵州茅台")
	}

	// 验证贵州茅台的舆情
	t.Logf("\n=== 贵州茅台舆情 ===")
	t.Logf("%s", items[1].SentimentAnalysis)
	if items[1].SentimentAnalysis == "" {
		t.Error("贵州茅台的SentimentAnalysis应该被填充")
	}
	if !strings.Contains(items[1].SentimentAnalysis, "贵州茅台") {
		t.Error("贵州茅台的舆情应该包含贵州茅台")
	}
	if strings.Contains(items[1].SentimentAnalysis, "中芯国际") {
		t.Error("贵州茅台的舆情不应该包含中芯国际")
	}
}

// 测试修复：当没有工具结果时，不应该回退到通用舆情内容
func TestAIReportParser_NoFallbackToGenericSentiment(t *testing.T) {
	parser := NewAIReportParser()

	// 创建报告内容，其中包含通用的舆情分析
	reportContent := `## 舆情动态
这是通用的舆情分析内容，不应该被分配给个股。

## 推荐股票

### 1. 中芯国际 (SH688981)
**推荐理由:** 半导体龙头

### 2. 贵州茅台 (SH600519)
**推荐理由:** 白酒龙头
`

	// 只给中芯国际创建工具结果，不给贵州茅台创建
	toolResults := &models.ToolCallResultsCollection{
		Results: []models.ToolCallResult{
			{
				ToolName:   "QueryStockNewsTool",
				Arguments:  `{"searchWords": "中芯国际"}`,
				Result:     "中芯国际发布最新财报，营收同比增长30%。",
				StockCode:  "SH688981",
				StockName:  "中芯国际",
				IsNewsTool: true,
			},
		},
	}

	// 创建推荐项
	items := []models.RecommendationItem{
		{
			Rank:       1,
			StockCode:  "SH688981",
			StockName:  "中芯国际",
		},
		{
			Rank:       2,
			StockCode:  "SH600519",
			StockName:  "贵州茅台",
		},
	}

	// 调用解析
	parser.ParseBatchWithToolResults(reportContent, items, toolResults)

	// 验证中芯国际有舆情（来自工具结果）
	if items[0].SentimentAnalysis == "" {
		t.Error("中芯国际的SentimentAnalysis应该被填充")
	}
	if !strings.Contains(items[0].SentimentAnalysis, "中芯国际") {
		t.Error("中芯国际的舆情应该包含中芯国际")
	}

	// 验证贵州茅台没有舆情（不应该回退到通用内容）
	if items[1].SentimentAnalysis != "" {
		t.Error("贵州茅台的SentimentAnalysis应该为空，不应该回退到通用内容")
		t.Logf("贵州茅台错误地获得了舆情内容: %s", items[1].SentimentAnalysis)
	}

	t.Logf("测试通过：中芯国际有自己的舆情，贵州茅台没有舆情（正确行为）")
}
