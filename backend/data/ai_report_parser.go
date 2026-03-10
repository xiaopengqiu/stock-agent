package data

import (
	"encoding/json"
	"fmt"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	"regexp"
	"strings"
	"time"
)

// AIReportParser AI报告解析器
type AIReportParser struct{}

// ParsedAnalysisResult 解析后的分析结果
type ParsedAnalysisResult struct {
	TechnicalAnalysis   string `json:"technical_analysis"`
	FundamentalAnalysis string `json:"fundamental_analysis"`
	RiskAnalysis        string `json:"risk_analysis"`
	SentimentAnalysis   string `json:"sentiment_analysis"` // 舆情动态分析
}

// NewAIReportParser 创建AI报告解析器
func NewAIReportParser() *AIReportParser {
	return &AIReportParser{}
}

// Parse 解析AI报告，提取结构化分析
func (p *AIReportParser) Parse(reportContent string) (*ParsedAnalysisResult, error) {
	logger.SugaredLogger.Infof("开始解析AI报告，长度: %d", len(reportContent))

	// 策略1: 尝试JSON格式解析
	if result, err := p.parseJSONFormat(reportContent); err == nil {
		logger.SugaredLogger.Infof("成功使用JSON格式解析报告")
		return result, nil
	}

	// 策略2: 尝试Markdown标题格式解析
	if result, err := p.parseMarkdownFormat(reportContent); err == nil {
		logger.SugaredLogger.Infof("成功使用Markdown格式解析报告")
		return result, nil
	}

	// 策略3: 关键词匹配回退策略
	result := p.parseByKeywords(reportContent)
	logger.SugaredLogger.Infof("使用关键词匹配策略解析报告")
	return result, nil
}

// parseJSONFormat 尝试JSON格式解析
func (p *AIReportParser) parseJSONFormat(content string) (*ParsedAnalysisResult, error) {
	// 尝试直接解析完整JSON
	var result ParsedAnalysisResult
	err := json.Unmarshal([]byte(content), &result)
	if err == nil {
		return &result, nil
	}

	// 尝试提取JSON部分（可能在markdown代码块中）
	jsonPattern := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
	matches := jsonPattern.FindStringSubmatch(content)
	if len(matches) >= 2 {
		err = json.Unmarshal([]byte(matches[1]), &result)
		if err == nil {
			return &result, nil
		}
	}

	return nil, fmt.Errorf("JSON解析失败")
}

// parseMarkdownFormat 尝试Markdown标题格式解析
func (p *AIReportParser) parseMarkdownFormat(content string) (*ParsedAnalysisResult, error) {
	result := &ParsedAnalysisResult{}

	// 定义各种可能的标题模式
	patterns := map[string]*string{
		"技术面分析|technical.*analysis": &result.TechnicalAnalysis,
		"基本面分析|fundamental.*analysis": &result.FundamentalAnalysis,
		"风险分析|风险提示|risk.*analysis|risk.*tips": &result.RiskAnalysis,
		"舆情动态|舆情分析|市场情绪|sentiment.*analysis|market.*sentiment": &result.SentimentAnalysis,
	}

	for pattern, target := range patterns {
		extracted := p.extractSectionByPattern(content, pattern)
		if extracted != "" {
			*target = extracted
		}
	}

	// 检查是否至少提取到一个部分
	if result.TechnicalAnalysis == "" && result.FundamentalAnalysis == "" && result.RiskAnalysis == "" {
		return nil, fmt.Errorf("Markdown格式解析失败")
	}

	return result, nil
}

// extractSectionByPattern 使用手动解析提取指定标题的内容（替代不支持的正则先行断言）
func (p *AIReportParser) extractSectionByPattern(content, pattern string) string {
	lines := strings.Split(content, "\n")
	var resultLines []string
	inSection := false
	titlePatterns := strings.Split(pattern, "|")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// 检查是否是标题行
		isHeader := strings.HasPrefix(trimmedLine, "#") || strings.HasPrefix(trimmedLine, "##")
		if isHeader {
			// 移除标题标记
			headerContent := strings.TrimSpace(strings.TrimLeft(trimmedLine, "#"))
			headerContent = strings.TrimSpace(strings.TrimRight(headerContent, "：:"))
			lowerHeader := strings.ToLower(headerContent)

			// 检查是否匹配我们要找的标题
			foundMatch := false
			for _, titlePattern := range titlePatterns {
				titlePattern = strings.TrimSpace(titlePattern)
				if strings.Contains(lowerHeader, strings.ToLower(titlePattern)) {
					foundMatch = true
					break
				}
				// 支持正则式简单的 .* 匹配
				if strings.Contains(titlePattern, ".*") {
					parts := strings.Split(titlePattern, ".*")
					matchAll := true
					for _, part := range parts {
						if part != "" && !strings.Contains(lowerHeader, part) {
							matchAll = false
							break
						}
					}
					if matchAll {
						foundMatch = true
						break
					}
				}
			}

			if foundMatch {
				inSection = true
				continue
			} else if inSection {
				// 遇到了另一个标题，停止收集
				break
			}
		}

		// 如果在目标区域内，收集内容
		if inSection {
			// 检查是否是下一个标题（不以#开头但可能是新章节的情况）
			// 这里我们保守一点，只在遇到明确的#标题时才停止
			resultLines = append(resultLines, line)
		}
	}

	// 清理结果：移除开头和结尾的空行
	var cleanedLines []string
	inContent := false
	for _, line := range resultLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			inContent = true
		}
		if inContent {
			cleanedLines = append(cleanedLines, line)
		}
	}

	// 再次从尾部移除空行
	for i := len(cleanedLines) - 1; i >= 0; i-- {
		if strings.TrimSpace(cleanedLines[i]) == "" {
			cleanedLines = cleanedLines[:i]
		} else {
			break
		}
	}

	return strings.TrimSpace(strings.Join(cleanedLines, "\n"))
}

// parseByKeywords 关键词匹配回退策略
func (p *AIReportParser) parseByKeywords(content string) *ParsedAnalysisResult {
	result := &ParsedAnalysisResult{}

	// 技术面关键词
	techKeywords := []string{
		"技术面", "K线", "均线", "MACD", "KDJ", "RSI", "布林带", "支撑位", "阻力位",
		"趋势", "突破", "回调", "震荡", "超买", "超卖", "金叉", "死叉",
	}
	result.TechnicalAnalysis = p.extractByKeywords(content, techKeywords, 10000)

	// 基本面关键词
	fundKeywords := []string{
		"基本面", "财务", "营收", "利润", "PE", "PB", "ROE", "估值", "业绩",
		"行业", "竞争", "市场份额", "成长性", "盈利能力", "资产负债",
	}
	result.FundamentalAnalysis = p.extractByKeywords(content, fundKeywords, 10000)

	// 风险关键词
	riskKeywords := []string{
		"风险", "提示", "注意", "警告", "不确定性", "下跌", "亏损", "波动",
		"政策", "监管", "竞争风险", "市场风险", "经营风险",
	}
	result.RiskAnalysis = p.extractByKeywords(content, riskKeywords, 10000)

	// 舆情关键词
	sentimentKeywords := []string{
		"舆情", "情绪", "新闻", "资讯", "研报", "券商观点", "投资者互动",
		"利好", "利空", "市场情绪", "资金流向", "机构评级", "买入评级", "卖出评级",
	}
	result.SentimentAnalysis = p.extractByKeywords(content, sentimentKeywords, 10000)

	return result
}

// extractByKeywords 根据关键词提取内容
func (p *AIReportParser) extractByKeywords(content string, keywords []string, maxLength int) string {
	lines := strings.Split(content, "\n")
	var extractedLines []string

	for _, line := range lines {
		for _, keyword := range keywords {
			if strings.Contains(line, keyword) {
				extractedLines = append(extractedLines, line)
				break
			}
		}
	}

	result := strings.Join(extractedLines, "\n")
	if len(result) > maxLength {
		result = result[:maxLength] + "..."
	}

	return strings.TrimSpace(result)
}

// ParseToRecommendationItem 解析并填充到RecommendationItem
func (p *AIReportParser) ParseToRecommendationItem(content string, item *models.RecommendationItem) {
	result, err := p.Parse(content)
	if err != nil {
		logger.SugaredLogger.Warnf("解析AI报告失败: %v", err)
		return
	}

	if result.TechnicalAnalysis != "" {
		item.TechnicalAnalysis = result.TechnicalAnalysis
	}
	if result.FundamentalAnalysis != "" {
		item.FundamentalAnalysis = result.FundamentalAnalysis
	}
	if result.RiskAnalysis != "" {
		item.RiskTips = result.RiskAnalysis
	}
	if result.SentimentAnalysis != "" {
		item.SentimentAnalysis = result.SentimentAnalysis
	}

	logger.SugaredLogger.Infof("成功解析报告到RecommendationItem")
}

// ParseBatch 批量解析多个推荐项
func (p *AIReportParser) ParseBatch(reportContent string, items []models.RecommendationItem) {
	// 尝试从完整报告中提取各部分
	result, _ := p.Parse(reportContent)

	for i := range items {
		if result.TechnicalAnalysis != "" && items[i].TechnicalAnalysis == "" {
			items[i].TechnicalAnalysis = result.TechnicalAnalysis
		}
		if result.FundamentalAnalysis != "" && items[i].FundamentalAnalysis == "" {
			items[i].FundamentalAnalysis = result.FundamentalAnalysis
		}
		if result.RiskAnalysis != "" && items[i].RiskTips == "" {
			items[i].RiskTips = result.RiskAnalysis
		}
		if result.SentimentAnalysis != "" && items[i].SentimentAnalysis == "" {
			items[i].SentimentAnalysis = result.SentimentAnalysis
		}
	}
}

// ParseBatchWithToolResults 批量解析多个推荐项，并优先使用工具调用结果
func (p *AIReportParser) ParseBatchWithToolResults(reportContent string, items []models.RecommendationItem, toolResults *models.ToolCallResultsCollection) {
	// 首先尝试从完整报告中提取各部分
	result, _ := p.Parse(reportContent)

	for i := range items {
		stockCode := items[i].StockCode
		stockName := items[i].StockName

		// 优先从工具调用结果中提取舆情分析
		if toolResults != nil && len(toolResults.Results) > 0 {
			if sentimentContent := p.extractSentimentFromToolResults(toolResults, stockCode, stockName); sentimentContent != "" {
				items[i].SentimentAnalysis = sentimentContent
				logger.SugaredLogger.Infof("从工具调用结果中提取到舆情分析: %s (%s)", stockName, stockCode)
			}
			// 注意：舆情分析不回退到通用报告内容！
			// 舆情是每只股票特有的，通用内容可能属于其他股票
		}

		// 其他分析内容仍从报告中提取
		if result.TechnicalAnalysis != "" && items[i].TechnicalAnalysis == "" {
			items[i].TechnicalAnalysis = result.TechnicalAnalysis
		}
		if result.FundamentalAnalysis != "" && items[i].FundamentalAnalysis == "" {
			items[i].FundamentalAnalysis = result.FundamentalAnalysis
		}
		if result.RiskAnalysis != "" && items[i].RiskTips == "" {
			items[i].RiskTips = result.RiskAnalysis
		}
	}
}

// extractSentimentFromToolResults 从工具调用结果中提取舆情内容
func (p *AIReportParser) extractSentimentFromToolResults(toolResults *models.ToolCallResultsCollection, stockCode, stockName string) string {
	// 使用增强版的舆情提取方法
	return p.extractSentimentFromToolResultsEnhanced(toolResults, stockCode, stockName)
}

// ============================================================================
// 工具结果解析器 - 专门用于解析各种工具返回的Markdown表格和内容
// ============================================================================

// ToolResultParser 工具结果解析器接口
type ToolResultParser interface {
	CanParse(toolName string) bool
	Parse(result string) []SentimentItem
}

// SentimentItem 解析后的舆情条目
type SentimentItem struct {
	Title       string    `json:"title"`        // 标题
	Content     string    `json:"content"`      // 内容
	Source      string    `json:"source"`       // 来源
	Time        time.Time `json:"time"`         // 时间
	Type        string    `json:"type"`         // 类型: news|qa|report
	StockCode   string    `json:"stock_code"`   // 关联股票代码
	StockName   string    `json:"stock_name"`   // 关联股票名称
	Sentiment   string    `json:"sentiment"`    // 情感倾向: positive|negative|neutral
	Importance  int       `json:"importance"`   // 重要程度: 1-5
}

// BaseToolResultParser 基础解析器
type BaseToolResultParser struct{}

// MarkdownTableParser Markdown表格解析器
type MarkdownTableParser struct {
	BaseToolResultParser
}

// NewsToolParser 新闻工具解析器 (QueryStockNewsTool)
type NewsToolParser struct {
	MarkdownTableParser
}

// InteractiveAnswerParser 投资者互动问答解析器 (QueryInteractiveAnswerData)
type InteractiveAnswerParser struct {
	MarkdownTableParser
}

// ResearchReportParser 研报解析器
type ResearchReportParser struct {
	BaseToolResultParser
}

// CompositeToolResultParser 组合解析器
type CompositeToolResultParser struct {
	parsers []ToolResultParser
}

// NewCompositeToolResultParser 创建组合解析器
func NewCompositeToolResultParser() *CompositeToolResultParser {
	return &CompositeToolResultParser{
		parsers: []ToolResultParser{
			&NewsToolParser{},
			&InteractiveAnswerParser{},
			&ResearchReportParser{},
		},
	}
}

// Parse 解析工具结果
func (c *CompositeToolResultParser) Parse(toolName, result string) []SentimentItem {
	for _, parser := range c.parsers {
		if parser.CanParse(toolName) {
			return parser.Parse(result)
		}
	}
	// 默认回退：创建一个简单的条目
	return []SentimentItem{{
		Content: result,
		Type:    "other",
		Source:  toolName,
	}}
}

// CanParse 判断是否能解析新闻工具
func (p *NewsToolParser) CanParse(toolName string) bool {
	return strings.Contains(toolName, "QueryStockNewsTool") ||
		strings.Contains(toolName, "News") ||
		strings.Contains(toolName, "Telegraph")
}

// Parse 解析新闻工具结果
func (p *NewsToolParser) Parse(result string) []SentimentItem {
	var items []SentimentItem

	// 尝试解析Markdown表格
	tableRows := p.parseMarkdownTable(result)
	if len(tableRows) > 0 {
		for _, row := range tableRows {
			item := SentimentItem{
				Type:   "news",
				Source: "财联社",
			}

			// 从表格行中提取数据
			if title, ok := row["标题"]; ok {
				item.Title = title
			} else if title, ok := row["Title"]; ok {
				item.Title = title
			}

			if content, ok := row["内容"]; ok {
				item.Content = content
			} else if content, ok := row["Content"]; ok {
				item.Content = content
			}

			if timeStr, ok := row["时间"]; ok {
				item.Time = p.parseTime(timeStr)
			} else if timeStr, ok := row["Time"]; ok {
				item.Time = p.parseTime(timeStr)
			}

			// 如果没有标题但有内容，使用内容前50字作为标题
			if item.Title == "" && item.Content != "" {
				if len(item.Content) > 50 {
					item.Title = item.Content[:50] + "..."
				} else {
					item.Title = item.Content
				}
			}

			// 分析情感
			if item.Content != "" {
				sentimentResult := AnalyzeSentiment(item.Content)
				switch sentimentResult.Category {
				case Positive:
					item.Sentiment = "positive"
				case Negative:
					item.Sentiment = "negative"
				default:
					item.Sentiment = "neutral"
				}
			}

			if item.Content != "" || item.Title != "" {
				items = append(items, item)
			}
		}
	}

	// 如果表格解析失败，尝试直接提取文本内容
	if len(items) == 0 {
		// 移除Markdown表格语法，提取纯文本
		cleanContent := p.cleanMarkdownContent(result)
		if cleanContent != "" {
			// 尝试按段落分割
			paragraphs := strings.Split(cleanContent, "\n\n")
			for _, para := range paragraphs {
				para = strings.TrimSpace(para)
				if para == "" || strings.HasPrefix(para, "##") || strings.HasPrefix(para, "|") {
					continue
				}
				item := SentimentItem{
					Content: para,
					Type:    "news",
					Source:  "财联社",
				}
				// 分析情感
				sentimentResult := AnalyzeSentiment(para)
				switch sentimentResult.Category {
				case Positive:
					item.Sentiment = "positive"
				case Negative:
					item.Sentiment = "negative"
				default:
					item.Sentiment = "neutral"
				}
				items = append(items, item)
			}
		}
	}

	return items
}

// CanParse 判断是否能解析投资者互动问答工具
func (p *InteractiveAnswerParser) CanParse(toolName string) bool {
	return strings.Contains(toolName, "QueryInteractiveAnswerData") ||
		strings.Contains(toolName, "InteractiveAnswer")
}

// Parse 解析投资者互动问答结果
func (p *InteractiveAnswerParser) Parse(result string) []SentimentItem {
	var items []SentimentItem

	// 尝试解析Markdown表格
	tableRows := p.parseMarkdownTable(result)
	if len(tableRows) > 0 {
		for _, row := range tableRows {
			item := SentimentItem{
				Type:   "qa",
				Source: "投资者互动平台",
			}

			// 提取投资者提问
			if question, ok := row["投资者提问"]; ok {
				item.Title = "投资者问: " + p.truncate(question, 40)
				item.Content = question
			} else if question, ok := row["MainContent"]; ok {
				item.Title = "投资者问: " + p.truncate(question, 40)
				item.Content = question
			}

			// 提取上市公司回复
			var replyContent string
			if reply, ok := row["上市公司回复"]; ok {
				replyContent = reply
			} else if reply, ok := row["AttachedContent"]; ok {
				replyContent = reply
			}

			if replyContent != "" {
				if item.Content != "" {
					item.Content += "\n\n公司回复: " + replyContent
				} else {
					item.Content = "公司回复: " + replyContent
					item.Title = "公司回复"
				}
			}

			// 提取股票信息
			if stockName, ok := row["股票名称"]; ok {
				item.StockName = stockName
			} else if stockName, ok := row["CompanyShortName"]; ok {
				item.StockName = stockName
			}

			if stockCode, ok := row["股票代码"]; ok {
				item.StockCode = stockCode
			} else if stockCode, ok := row["StockCode"]; ok {
				item.StockCode = stockCode
			}

			// 提取时间
			if timeStr, ok := row["发布时间"]; ok {
				item.Time = p.parseTime(timeStr)
			} else if timeStr, ok := row["PubDate"]; ok {
				item.Time = p.parseTime(timeStr)
			}

			// 分析情感
			if item.Content != "" {
				sentimentResult := AnalyzeSentiment(item.Content)
				switch sentimentResult.Category {
				case Positive:
					item.Sentiment = "positive"
				case Negative:
					item.Sentiment = "negative"
				default:
					item.Sentiment = "neutral"
				}
			}

			if item.Content != "" {
				items = append(items, item)
			}
		}
	}

	// 如果表格解析失败，尝试直接提取文本
	if len(items) == 0 {
		cleanContent := p.cleanMarkdownContent(result)
		if cleanContent != "" {
			item := SentimentItem{
				Content: cleanContent,
				Type:    "qa",
				Source:  "投资者互动平台",
			}
			sentimentResult := AnalyzeSentiment(cleanContent)
			switch sentimentResult.Category {
			case Positive:
				item.Sentiment = "positive"
			case Negative:
				item.Sentiment = "negative"
			default:
				item.Sentiment = "neutral"
			}
			items = append(items, item)
		}
	}

	return items
}

// CanParse 判断是否能解析研报工具
func (p *ResearchReportParser) CanParse(toolName string) bool {
	return strings.Contains(toolName, "GetStockResearchReport") ||
		strings.Contains(toolName, "GetIndustryResearchReport") ||
		strings.Contains(toolName, "Research") ||
		strings.Contains(toolName, "Opinion")
}

// Parse 解析研报结果
func (p *ResearchReportParser) Parse(result string) []SentimentItem {
	var items []SentimentItem

	cleanContent := p.cleanMarkdownContent(result)
	if cleanContent != "" {
		// 分析情感
		sentimentResult := AnalyzeSentiment(cleanContent)
		var sentiment string
		switch sentimentResult.Category {
		case Positive:
			sentiment = "positive"
		case Negative:
			sentiment = "negative"
		default:
			sentiment = "neutral"
		}

		// 尝试提取标题
		title := "券商研报"
		lines := strings.Split(cleanContent, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && len(line) < 100 && !strings.Contains(line, "：") && !strings.Contains(line, ":") {
				title = line
				break
			}
		}

		item := SentimentItem{
			Title:     title,
			Content:   cleanContent,
			Type:      "report",
			Source:    "券商研报",
			Sentiment: sentiment,
		}
		items = append(items, item)
	}

	return items
}

// parseMarkdownTable 解析Markdown表格
func (p *MarkdownTableParser) parseMarkdownTable(content string) []map[string]string {
	var results []map[string]string

	lines := strings.Split(content, "\n")
	var headers []string
	var inTable bool
	var separatorLineFound bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过空行
		if line == "" {
			continue
		}

		// 检查是否是表格行（以|开头和结尾）
		if !strings.HasPrefix(line, "|") || !strings.HasSuffix(line, "|") {
			// 如果已经在表格中，但遇到非表格行，结束解析
			if inTable && len(headers) > 0 {
				break
			}
			continue
		}

		// 移除首尾的|并分割
		cells := strings.Split(line[1:len(line)-1], "|")
		for i, cell := range cells {
			cells[i] = strings.TrimSpace(cell)
		}

		// 检查是否是分隔线（包含---）
		isSeparator := true
		for _, cell := range cells {
			if !strings.Contains(cell, "---") && cell != "" {
				isSeparator = false
				break
			}
		}

		if isSeparator {
			separatorLineFound = true
			continue
		}

		// 如果还没有表头，且已经找到分隔线，那么上一行就是表头
		if len(headers) == 0 && separatorLineFound {
			// 这种情况下表头应该在前面
			// 我们重新开始查找
			continue
		}

		// 如果还没有表头，这行就是表头
		if len(headers) == 0 {
			headers = cells
			inTable = true
			continue
		}

		// 这是数据行
		row := make(map[string]string)
		for i, header := range headers {
			if i < len(cells) {
				row[header] = cells[i]
			} else {
				row[header] = ""
			}
		}
		results = append(results, row)
	}

	// 另一种解析方式：查找标题后的表格
	if len(results) == 0 {
		var tableLines []string
		inTable = false
		for _, line := range lines {
			if strings.HasPrefix(line, "|") {
				inTable = true
				tableLines = append(tableLines, line)
			} else if inTable {
				break
			}
		}

		if len(tableLines) >= 3 { // 至少有表头、分隔线、一行数据
			// 解析表头
			if strings.HasPrefix(tableLines[0], "|") {
				headerCells := strings.Split(tableLines[0][1:len(tableLines[0])-1], "|")
				for i, cell := range headerCells {
					headerCells[i] = strings.TrimSpace(cell)
				}
				headers = headerCells

				// 解析数据行（跳过分隔线）
				for i := 2; i < len(tableLines); i++ {
					if strings.HasPrefix(tableLines[i], "|") {
						cells := strings.Split(tableLines[i][1:len(tableLines[i])-1], "|")
						row := make(map[string]string)
						for j, header := range headers {
							if j < len(cells) {
								row[header] = strings.TrimSpace(cells[j])
							}
						}
						results = append(results, row)
					}
				}
			}
		}
	}

	return results
}

// cleanMarkdownContent 清理Markdown内容，移除表格语法等
func (p *BaseToolResultParser) cleanMarkdownContent(content string) string {
	var result strings.Builder
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过表格行
		if strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") {
			continue
		}
		// 跳过Markdown标题
		if strings.HasPrefix(line, "##") {
			continue
		}
		// 跳过分隔线
		if strings.HasPrefix(line, "---") {
			continue
		}
		if line != "" {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return strings.TrimSpace(result.String())
}

// parseTime 解析时间字符串
func (p *BaseToolResultParser) parseTime(timeStr string) time.Time {
	if timeStr == "" {
		return time.Now()
	}

	// 尝试多种时间格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		"15:04:05",
		"2006/01/02 15:04:05",
		"2006/01/02",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, timeStr, time.Local); err == nil {
			return t
		}
	}

	return time.Now()
}

// truncate 截断字符串
func (p *BaseToolResultParser) truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ============================================================================
// 增强版舆情分析 - 使用新的解析器
// ============================================================================

// extractSentimentFromToolResultsEnhanced 增强版：从工具调用结果中提取舆情内容
func (p *AIReportParser) extractSentimentFromToolResultsEnhanced(toolResults *models.ToolCallResultsCollection, stockCode, stockName string) string {
	if toolResults == nil {
		logger.SugaredLogger.Debugf("extractSentimentFromToolResultsEnhanced: toolResults 为 nil")
		return ""
	}

	logger.SugaredLogger.Infof("extractSentimentFromToolResultsEnhanced: 开始提取舆情, 股票代码=%s, 股票名称=%s, 工具结果总数=%d",
		stockCode, stockName, len(toolResults.Results))

	// 创建解析器 - 每次都创建新的，避免状态污染
	parser := NewCompositeToolResultParser()

	// 收集所有相关的舆情条目 - 每次都重新初始化
	var allItems []SentimentItem

	// 按优先级查找工具结果
	for _, result := range toolResults.Results {
		if !result.IsNewsTool {
			continue
		}

		// 检查是否与当前股票相关
		isRelevant := false

		// 1. 通过股票代码精确匹配
		if stockCode != "" && result.StockCode == stockCode {
			isRelevant = true
			logger.SugaredLogger.Debugf("通过股票代码精确匹配: %s", result.ToolName)
		}
		// 2. 通过股票名称精确匹配
		if !isRelevant && stockName != "" && result.StockName == stockName {
			isRelevant = true
			logger.SugaredLogger.Debugf("通过股票名称精确匹配: %s", result.ToolName)
		}
		// 3. 通过结果内容中的股票代码或名称匹配（仅当工具没有关联特定股票时）
		if !isRelevant && result.StockCode == "" && result.StockName == "" {
			if (stockCode != "" && strings.Contains(result.Result, stockCode)) ||
				(stockName != "" && strings.Contains(result.Result, stockName)) {
				isRelevant = true
				logger.SugaredLogger.Debugf("通过内容匹配: %s", result.ToolName)
			}
		}

		if isRelevant {
			// 使用解析器解析工具结果
			items := parser.Parse(result.ToolName, result.Result)
			for _, item := range items {
				// 确保这个条目确实与当前股票相关
				content := item.Title + " " + item.Content
				itemIsRelevant := true

				// 如果有股票代码，检查是否匹配
				if stockCode != "" && result.StockCode == "" && result.StockName == "" {
					// 工具没有关联股票，需要检查内容
					if !strings.Contains(content, stockCode) && !strings.Contains(content, stockName) {
						itemIsRelevant = false
					}
				}

				if itemIsRelevant {
					// 设置股票信息
					newItem := item // 创建副本，避免引用问题
					if newItem.StockCode == "" {
						newItem.StockCode = stockCode
					}
					if newItem.StockName == "" {
						newItem.StockName = stockName
					}
					allItems = append(allItems, newItem)
				}
			}
			logger.SugaredLogger.Debugf("从工具 %s 解析到 %d 条舆情条目（过滤后 %d 条）",
				result.ToolName, len(items), len(allItems))
		}
	}

	// 如果没有找到，放宽条件：只要是新闻工具，就尝试解析并看看内容是否相关
	if len(allItems) == 0 {
		logger.SugaredLogger.Debugf("未找到精确匹配，尝试放宽条件")
		for _, result := range toolResults.Results {
			if !result.IsNewsTool {
				continue
			}

			items := parser.Parse(result.ToolName, result.Result)
			for _, item := range items {
				// 检查内容是否包含股票名称或代码
				content := item.Title + " " + item.Content
				if (stockCode != "" && strings.Contains(content, stockCode)) ||
					(stockName != "" && strings.Contains(content, stockName)) {
					newItem := item // 创建副本
					if newItem.StockCode == "" {
						newItem.StockCode = stockCode
					}
					if newItem.StockName == "" {
						newItem.StockName = stockName
					}
					allItems = append(allItems, newItem)
				}
			}
		}
	}

	// 生成舆情内容
	if len(allItems) > 0 {
		var sentimentContent strings.Builder
		sentimentContent.WriteString("【舆情动态】\n\n")

		// 分类统计
		newsCount := 0
		qaCount := 0
		reportCount := 0
		positiveCount := 0
		negativeCount := 0
		neutralCount := 0

		for _, item := range allItems {
			switch item.Type {
			case "news":
				newsCount++
			case "qa":
				qaCount++
			case "report":
				reportCount++
			}
			switch item.Sentiment {
			case "positive":
				positiveCount++
			case "negative":
				negativeCount++
			default:
				neutralCount++
			}
		}

		// 输出统计摘要
		sentimentContent.WriteString(fmt.Sprintf("共收集 %d 条舆情信息（新闻 %d 条，互动问答 %d 条，研报 %d 条）\n",
			len(allItems), newsCount, qaCount, reportCount))
		sentimentContent.WriteString(fmt.Sprintf("情感倾向：正面 %d 条，负面 %d 条，中性 %d 条\n\n",
			positiveCount, negativeCount, neutralCount))

		// 输出详细内容（最多显示8条）
		displayCount := 0
		for _, item := range allItems {
			if displayCount >= 8 {
				break
			}

			// 根据类型添加不同的标记
			switch item.Type {
			case "news":
				sentimentContent.WriteString("📰 ")
			case "qa":
				sentimentContent.WriteString("💬 ")
			case "report":
				sentimentContent.WriteString("📊 ")
			default:
				sentimentContent.WriteString("📄 ")
			}

			// 添加情感标记
			switch item.Sentiment {
			case "positive":
				sentimentContent.WriteString("[利好] ")
			case "negative":
				sentimentContent.WriteString("[利空] ")
			default:
				sentimentContent.WriteString("[中性] ")
			}

			// 添加标题
			if item.Title != "" {
				sentimentContent.WriteString(item.Title)
				sentimentContent.WriteString("\n")
			}

			// 添加内容（截断避免过长）
			content := item.Content
			if len(content) > 300 {
				content = content[:300] + "..."
			}
			sentimentContent.WriteString(content)
			sentimentContent.WriteString("\n\n")

			displayCount++
		}

		// 如果还有更多，显示剩余数量
		if len(allItems) > displayCount {
			sentimentContent.WriteString(fmt.Sprintf("... 还有 %d 条舆情信息\n\n", len(allItems)-displayCount))
		}

		// 添加综合情感分析
		var textParts []string
		for _, item := range allItems {
			textParts = append(textParts, item.Title+" "+item.Content)
		}
		fullText := strings.Join(textParts, " ")
		if fullText != "" {
			sentimentResult := AnalyzeSentiment(fullText)
			sentimentContent.WriteString("【综合情感分析】\n")
			sentimentContent.WriteString(fmt.Sprintf("市场情绪: %s (得分: %.2f, 正面词: %d, 负面词: %d)\n",
				GetSentimentDescription(sentimentResult.Category),
				sentimentResult.Score,
				sentimentResult.PositiveCount,
				sentimentResult.NegativeCount))
		}

		logger.SugaredLogger.Infof("成功生成增强版舆情内容，包含 %d 条条目，长度: %d",
			len(allItems), sentimentContent.Len())

		return strings.TrimSpace(sentimentContent.String())
	}

	logger.SugaredLogger.Infof("未找到任何相关舆情内容，回退到原方法")

	// 回退到原来的方法
	return p.extractSentimentFromToolResultsFallback(toolResults, stockCode, stockName)
}

// extractSentimentFromToolResultsFallback 原来的提取方法作为回退
func (p *AIReportParser) extractSentimentFromToolResultsFallback(toolResults *models.ToolCallResultsCollection, stockCode, stockName string) string {
	if toolResults == nil {
		return ""
	}

	var sentimentContent strings.Builder

	// 查找与该股票直接相关的新闻工具结果
	newsResults := toolResults.GetNewsResultsByStockCode(stockCode)

	if len(newsResults) == 0 && stockName != "" {
		allNews := toolResults.GetNewsResults()
		for _, r := range allNews {
			if strings.Contains(r.StockName, stockName) || strings.Contains(r.Result, stockName) {
				newsResults = append(newsResults, r)
			}
		}
	}

	if len(newsResults) == 0 {
		for _, r := range toolResults.Results {
			if r.IsNewsTool {
				if (stockCode != "" && strings.Contains(r.Result, stockCode)) ||
					(stockName != "" && strings.Contains(r.Result, stockName)) {
					newsResults = append(newsResults, r)
				}
			}
		}
	}

	if len(newsResults) > 0 {
		sentimentContent.WriteString(fmt.Sprintf("【相关新闻资讯】\n"))

		for idx, result := range newsResults {
			if idx >= 5 {
				break
			}

			switch result.ToolName {
			case "QueryStockNewsTool":
				sentimentContent.WriteString(fmt.Sprintf("\n■ 股票新闻:\n"))
			case "QueryMarketNews":
				sentimentContent.WriteString(fmt.Sprintf("\n■ 市场资讯:\n"))
			case "GetStockResearchReport":
				sentimentContent.WriteString(fmt.Sprintf("\n■ 研究报告:\n"))
			case "QueryInteractiveAnswerData":
				sentimentContent.WriteString(fmt.Sprintf("\n■ 投资者互动:\n"))
			case "GetNewsList2":
				sentimentContent.WriteString(fmt.Sprintf("\n■ 新闻列表:\n"))
			case "GetTelegraphListWithPaging":
				sentimentContent.WriteString(fmt.Sprintf("\n■ 财联社电报:\n"))
			case "GetSecuritiesCompanyOpinionContent":
				sentimentContent.WriteString(fmt.Sprintf("\n■ 券商观点:\n"))
			default:
				sentimentContent.WriteString(fmt.Sprintf("\n■ %s:\n", result.ToolName))
			}

			content := strings.TrimSpace(result.Result)
			if len(content) > 800 {
				content = content[:800] + "..."
			}
			sentimentContent.WriteString(content)
			sentimentContent.WriteString("\n")
		}

		fullContent := sentimentContent.String()
		if fullContent != "" {
			sentimentResult := AnalyzeSentiment(fullContent)
			sentimentContent.WriteString(fmt.Sprintf("\n【情感分析】\n"))
			sentimentContent.WriteString(fmt.Sprintf("市场情绪: %s (得分: %.2f, 正面词: %d, 负面词: %d)\n",
				GetSentimentDescription(sentimentResult.Category),
				sentimentResult.Score,
				sentimentResult.PositiveCount,
				sentimentResult.NegativeCount))
		}
	}

	return strings.TrimSpace(sentimentContent.String())
}
