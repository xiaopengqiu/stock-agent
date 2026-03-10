package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/strutil"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
)

// PositionAnalysisService 持仓分析服务
type PositionAnalysisService struct {
	ctx context.Context
}

// NewPositionAnalysisService 创建持仓分析服务
func NewPositionAnalysisService(ctx context.Context) *PositionAnalysisService {
	return &PositionAnalysisService{
		ctx: ctx,
	}
}

// GetPositionAnalysis 获取持仓分析结果
func (s *PositionAnalysisService) GetPositionAnalysis(positionID uint) (*models.PositionAnalysis, error) {
	var analysis models.PositionAnalysis
	err := db.Dao.Where("position_id = ?", positionID).Order("created_at DESC").First(&analysis).Error
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}

// GetLatestPositionAnalyses 获取所有持仓的最新分析结果
func (s *PositionAnalysisService) GetLatestPositionAnalyses() ([]*models.PositionAnalysis, error) {
	// 先获取所有活跃持仓
	var positions []*models.Position
	err := db.Dao.Where("is_active = ?", true).Find(&positions).Error
	if err != nil {
		return nil, err
	}

	if len(positions) == 0 {
		return []*models.PositionAnalysis{}, nil
	}

	// 为每个持仓获取最新分析
	var analyses []*models.PositionAnalysis
	for _, pos := range positions {
		var analysis models.PositionAnalysis
		err := db.Dao.Where("position_id = ?", pos.ID).Order("created_at DESC").First(&analysis).Error
		if err == nil {
			analyses = append(analyses, &analysis)
		}
	}

	return analyses, nil
}

// PositionAnalysisPrompt 持仓分析提示词
const PositionAnalysisPrompt = `你是一位专业的股票投资顾问和风控专家。请对用户提供的持仓股票进行全面分析，并给出具体的操作建议。

## 分析要求

请从以下几个维度进行分析：

1. **技术面分析** (Technical Analysis)
   - K线形态和趋势判断
   - 主要技术指标 (MACD, KDJ, RSI等)
   - 支撑位和压力位分析
   - 成交量分析

2. **基本面分析** (Fundamental Analysis)
   - 公司主营业务和行业地位
   - 近期财务状况和业绩表现
   - 估值水平 (PE, PB等)
   - 行业前景和竞争格局

3. **风险分析** (Risk Analysis)
   - 当前持仓的风险点
   - 市场系统性风险
   - 个股特定风险
   - 止损建议

## 输出格式要求

请以JSON格式输出分析结果，结构如下：

{
  "overall_advice": "持有|加仓|减仓|清仓",
  "confidence": 0.8,
  "suggested_buy_price": 10.50,
  "suggested_sell_price": 12.80,
  "stop_loss_price": 9.20,
  "technical_analysis": "详细的技术面分析内容...",
  "fundamental_analysis": "详细的基本面分析内容...",
  "risk_analysis": "详细的风险分析内容..."
}

## 字段说明

- overall_advice: 总体建议，只能是以下四个值之一："持有"、"加仓"、"减仓"、"清仓"
- confidence: 建议的置信度，0-1之间的小数
- suggested_buy_price: 建议的补仓价位 (如果不建议加仓，可以为null)
- suggested_sell_price: 建议的止盈价位 (如果不建议卖出，可以为null)
- stop_loss_price: 建议的止损价位
- technical_analysis: 技术面分析详细内容
- fundamental_analysis: 基本面分析详细内容
- risk_analysis: 风险分析详细内容

请确保输出是严格的JSON格式，不要包含任何其他文本。`

// AnalyzePosition 分析单个持仓
func (s *PositionAnalysisService) AnalyzePosition(position *models.Position, aiConfigID uint) (*models.PositionAnalysis, error) {
	if position == nil {
		return nil, errors.New("持仓信息不能为空")
	}

	// 1. 获取AI配置
	settingConfig := GetSettingConfig()
	var aiConfig *AIConfig
	if aiConfigID > 0 {
		aiConfig, _ = lo.Find(settingConfig.AiConfigs, func(item *AIConfig) bool {
			return aiConfigID == item.ID
		})
	}
	if aiConfig == nil && len(settingConfig.AiConfigs) > 0 {
		aiConfig = settingConfig.AiConfigs[0]
	}
	if aiConfig == nil {
		return nil, errors.New("未配置AI服务")
	}

	// 2. 构建用户问题
	userQuestion := fmt.Sprintf(`请分析我的持仓股票：

股票代码：%s
股票名称：%s
持股数量：%d
买入价格：%.2f
当前价格：%.2f
持仓市值：%.2f
盈亏金额：%.2f
盈亏比例：%.2f%%

请给出详细的分析和操作建议。`,
		position.StockCode,
		position.StockName,
		position.Quantity,
		position.BuyPrice,
		position.CurrentPrice,
		position.MarketValue,
		position.ProfitLoss,
		position.ProfitLossPct,
	)

	// 3. 创建OpenAi实例
	openAI := &OpenAi{
		ctx:          s.ctx,
		BaseUrl:      aiConfig.BaseUrl,
		ApiKey:       aiConfig.ApiKey,
		Model:        aiConfig.ModelName,
		MaxTokens:    aiConfig.MaxTokens,
		Temperature:  aiConfig.Temperature,
		TimeOut:      aiConfig.TimeOut,
		Prompt:       PositionAnalysisPrompt,
		CrawlTimeOut: settingConfig.CrawlTimeOut,
		KDays:        settingConfig.KDays,
		BrowserPath:  settingConfig.BrowserPath,
	}

	if openAI.TimeOut <= 0 {
		openAI.TimeOut = 300
	}
	if openAI.CrawlTimeOut <= 0 {
		openAI.CrawlTimeOut = 60
	}
	if openAI.KDays < 30 {
		openAI.KDays = 120
	}

	// 4. 构建消息列表
	msg := []map[string]interface{}{
		{
			"role":    "system",
			"content": PositionAnalysisPrompt,
		},
		{
			"role":    "user",
			"content": "当前时间",
		},
		{
			"role":    "assistant",
			"content": "当前本地时间是:" + time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			"role":    "user",
			"content": userQuestion,
		},
	}

	// 5. 调用AI进行分析 (非流式)
	analysisResult, rawResponse, err := s.callAIAnalysis(openAI, msg)
	if err != nil {
		return nil, err
	}

	// 6. 创建分析结果记录
	analysis := &models.PositionAnalysis{
		PositionID:          position.ID,
		OverallAdvice:       analysisResult.OverallAdvice,
		Confidence:          analysisResult.Confidence,
		TechnicalAnalysis:   analysisResult.TechnicalAnalysis,
		FundamentalAnalysis: analysisResult.FundamentalAnalysis,
		RiskAnalysis:        analysisResult.RiskAnalysis,
		RawResponse:         rawResponse,
	}

	// 设置价格建议
	if analysisResult.SuggestedBuyPrice > 0 {
		analysis.SuggestedBuyPrice = &analysisResult.SuggestedBuyPrice
	}
	if analysisResult.SuggestedSellPrice > 0 {
		analysis.SuggestedSellPrice = &analysisResult.SuggestedSellPrice
	}
	if analysisResult.StopLossPrice > 0 {
		analysis.StopLossPrice = &analysisResult.StopLossPrice
	}

	// 7. 保存到数据库
	if err := db.Dao.Create(analysis).Error; err != nil {
		logger.SugaredLogger.Errorf("保存持仓分析结果失败: %v", err)
		return nil, err
	}

	logger.SugaredLogger.Infof("持仓 %d 分析完成", position.ID)
	return analysis, nil
}

// AnalyzeAllPositions 批量分析所有持仓
func (s *PositionAnalysisService) AnalyzeAllPositions(aiConfigID uint) ([]*models.PositionAnalysis, error) {
	// 获取所有活跃持仓
	var positions []*models.Position
	err := db.Dao.Where("is_active = ?", true).Find(&positions).Error
	if err != nil {
		return nil, err
	}

	if len(positions) == 0 {
		return []*models.PositionAnalysis{}, nil
	}

	var analyses []*models.PositionAnalysis
	for _, pos := range positions {
		analysis, err := s.AnalyzePosition(pos, aiConfigID)
		if err != nil {
			logger.SugaredLogger.Errorf("分析持仓 %d 失败: %v", pos.ID, err)
			continue
		}
		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

// positionAnalysisResult AI分析结果
type positionAnalysisResult struct {
	OverallAdvice       string  `json:"overall_advice"`
	Confidence          float64 `json:"confidence"`
	SuggestedBuyPrice   float64 `json:"suggested_buy_price"`
	SuggestedSellPrice  float64 `json:"suggested_sell_price"`
	StopLossPrice       float64 `json:"stop_loss_price"`
	TechnicalAnalysis   string  `json:"technical_analysis"`
	FundamentalAnalysis string  `json:"fundamental_analysis"`
	RiskAnalysis        string  `json:"risk_analysis"`
}

// callAIAnalysis 调用AI进行分析 (同步版本)
func (s *PositionAnalysisService) callAIAnalysis(openAI *OpenAi, messages []map[string]interface{}) (*positionAnalysisResult, string, error) {
	client := resty.New()
	client.SetBaseURL(strutil.Trim(openAI.BaseUrl))
	client.SetHeader("Authorization", "Bearer "+openAI.ApiKey)
	client.SetHeader("Content-Type", "application/json")
	if openAI.TimeOut <= 0 {
		openAI.TimeOut = 300
	}
	client.SetTimeout(time.Duration(openAI.TimeOut) * time.Second)

	resp, err := client.R().
		SetBody(map[string]interface{}{
			"model":       openAI.Model,
			"max_tokens":  openAI.MaxTokens,
			"temperature": openAI.Temperature,
			"stream":      false,
			"messages":    messages,
		}).
		Post("/chat/completions")

	if err != nil {
		logger.SugaredLogger.Errorf("AI调用失败: %v", err)
		return nil, "", err
	}

	logger.SugaredLogger.Infof("AI响应: %s", resp.String())

	// 解析AI响应
	var aiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(resp.Body(), &aiResp); err != nil {
		logger.SugaredLogger.Errorf("解析AI响应失败: %v", err)
		return nil, "", err
	}

	if len(aiResp.Choices) == 0 || aiResp.Choices[0].Message.Content == "" {
		return nil, "", errors.New("AI响应为空")
	}

	rawContent := aiResp.Choices[0].Message.Content
	logger.SugaredLogger.Infof("AI响应内容: %s", rawContent)

	// 提取JSON内容
	jsonStr := extractPositionAnalysisJSONFromContent(rawContent)
	if jsonStr == "" {
		return nil, rawContent, errors.New("无法从AI响应中提取JSON")
	}

	// 解析分析结果
	var result positionAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		logger.SugaredLogger.Errorf("解析分析结果失败: %v", err)
		return nil, rawContent, err
	}

	return &result, rawContent, nil
}

// extractPositionAnalysisJSONFromContent 从AI响应中提取JSON
func extractPositionAnalysisJSONFromContent(content string) string {
	// 尝试找到第一个 { 和最后一个 }
	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")
	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		// 尝试用 ```json 和 ``` 包裹
		if strings.Contains(content, "```json") {
			startIdx = strings.Index(content, "```json") + 7
			endIdx = strings.Index(content[startIdx:], "```")
			if endIdx != -1 {
				return strings.TrimSpace(content[startIdx : startIdx+endIdx])
			}
		}
		if strings.Contains(content, "```") {
			startIdx = strings.Index(content, "```") + 3
			endIdx = strings.Index(content[startIdx:], "```")
			if endIdx != -1 {
				return strings.TrimSpace(content[startIdx : startIdx+endIdx])
			}
		}
		return ""
	}
	return content[startIdx : endIdx+1]
}

// PortfolioAnalysisPrompt 整体仓位分析提示词
const PortfolioAnalysisPrompt = `你是一位专业的投资组合经理和风控专家。请对用户提供的整体持仓进行全面分析，并给出具体的仓位调整建议。

## 分析要求

请从以下几个维度进行分析：

1. **整体评估** (Overall Assessment)
   - 当前持仓的整体表现
   - 组合的优势和劣势
   - 当前市场环境下的适应性

2. **仓位分布分析** (Allocation Analysis)
   - 行业集中度分析（是否过度集中在某一行业）
   - 个股权重分析（单只股票占比是否过高）
   - 资产配置合理性评估
   - 分散化程度分析

3. **风险评估** (Risk Assessment)
   - 整体组合风险水平
   - 相关性风险（股票之间的相关性）
   - 市场风险暴露
   - 潜在的风险点

4. **调整建议** (Adjustment Suggestions)
   - 具体的仓位调整建议
   - 建议加仓的股票及理由
   - 建议减仓/清仓的股票及理由
   - 仓位重新分配方案
   - 风险控制措施

## 输出格式要求

请以JSON格式输出分析结果，结构如下：

{
  "overall_assessment": "整体评估内容...",
  "allocation_analysis": "仓位分布分析内容...",
  "risk_assessment": "风险评估内容...",
  "adjustment_suggestions": "具体调整建议内容..."
}

## 字段说明

- overall_assessment: 整体评估详细内容
- allocation_analysis: 仓位分布分析详细内容
- risk_assessment: 风险评估详细内容
- adjustment_suggestions: 具体调整建议详细内容

请确保输出是严格的JSON格式，不要包含任何其他文本。`

// portfolioAnalysisResult 整体仓位AI分析结果
type portfolioAnalysisResult struct {
	OverallAssessment      string `json:"overall_assessment"`
	AllocationAnalysis     string `json:"allocation_analysis"`
	RiskAssessment         string `json:"risk_assessment"`
	AdjustmentSuggestions  string `json:"adjustment_suggestions"`
}

// AnalyzePortfolio 分析整体仓位
func (s *PositionAnalysisService) AnalyzePortfolio(aiConfigID uint) (*models.PortfolioAnalysis, error) {
	// 1. 获取所有活跃持仓
	var positions []*models.Position
	err := db.Dao.Where("is_active = ?", true).Find(&positions).Error
	if err != nil {
		return nil, err
	}

	if len(positions) == 0 {
		return nil, errors.New("暂无持仓数据")
	}

	// 2. 获取AI配置
	settingConfig := GetSettingConfig()
	var aiConfig *AIConfig
	if aiConfigID > 0 {
		aiConfig, _ = lo.Find(settingConfig.AiConfigs, func(item *AIConfig) bool {
			return aiConfigID == item.ID
		})
	}
	if aiConfig == nil && len(settingConfig.AiConfigs) > 0 {
		aiConfig = settingConfig.AiConfigs[0]
	}
	if aiConfig == nil {
		return nil, errors.New("未配置AI服务")
	}

	// 3. 计算整体统计数据
	totalMarketValue := 0.0
	totalProfitLoss := 0.0
	for _, pos := range positions {
		totalMarketValue += pos.MarketValue
		totalProfitLoss += pos.ProfitLoss
	}

	// 4. 构建持仓详情文本
	var positionsInfo strings.Builder
	positionsInfo.WriteString(fmt.Sprintf("## 整体持仓概况\n"))
	positionsInfo.WriteString(fmt.Sprintf("- 持仓数量：%d 只\n", len(positions)))
	positionsInfo.WriteString(fmt.Sprintf("- 总市值：%.2f 元\n", totalMarketValue))
	positionsInfo.WriteString(fmt.Sprintf("- 总盈亏：%.2f 元\n\n", totalProfitLoss))

	positionsInfo.WriteString("## 各持仓详情\n\n")
	for _, pos := range positions {
		var weight float64
		if totalMarketValue > 0 {
			weight = (pos.MarketValue / totalMarketValue) * 100
		}
		positionsInfo.WriteString(fmt.Sprintf("### %s (%s)\n", pos.StockName, pos.StockCode))
		positionsInfo.WriteString(fmt.Sprintf("- 持股数量：%d 股\n", pos.Quantity))
		positionsInfo.WriteString(fmt.Sprintf("- 买入价格：%.2f 元\n", pos.BuyPrice))
		positionsInfo.WriteString(fmt.Sprintf("- 当前价格：%.2f 元\n", pos.CurrentPrice))
		positionsInfo.WriteString(fmt.Sprintf("- 持仓市值：%.2f 元\n", pos.MarketValue))
		positionsInfo.WriteString(fmt.Sprintf("- 盈亏金额：%.2f 元\n", pos.ProfitLoss))
		positionsInfo.WriteString(fmt.Sprintf("- 盈亏比例：%.2f %%\n", pos.ProfitLossPct))
		positionsInfo.WriteString(fmt.Sprintf("- 仓位占比：%.2f %%\n\n", weight))
	}

	// 5. 创建OpenAi实例
	openAI := &OpenAi{
		ctx:          s.ctx,
		BaseUrl:      aiConfig.BaseUrl,
		ApiKey:       aiConfig.ApiKey,
		Model:        aiConfig.ModelName,
		MaxTokens:    aiConfig.MaxTokens,
		Temperature:  aiConfig.Temperature,
		TimeOut:      aiConfig.TimeOut,
		Prompt:       PortfolioAnalysisPrompt,
		CrawlTimeOut: settingConfig.CrawlTimeOut,
		KDays:        settingConfig.KDays,
		BrowserPath:  settingConfig.BrowserPath,
	}

	if openAI.TimeOut <= 0 {
		openAI.TimeOut = 600
	}

	// 6. 构建消息列表
	msg := []map[string]interface{}{
		{
			"role":    "system",
			"content": PortfolioAnalysisPrompt,
		},
		{
			"role":    "user",
			"content": "当前时间",
		},
		{
			"role":    "assistant",
			"content": "当前本地时间是:" + time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			"role":    "user",
			"content": "请分析我的整体持仓：\n\n" + positionsInfo.String(),
		},
	}

	// 7. 调用AI进行分析
	analysisResult, rawResponse, err := s.callPortfolioAIAnalysis(openAI, msg)
	if err != nil {
		return nil, err
	}

	// 8. 创建分析结果记录
	analysis := &models.PortfolioAnalysis{
		OverallAssessment:     analysisResult.OverallAssessment,
		AllocationAnalysis:    analysisResult.AllocationAnalysis,
		RiskAssessment:        analysisResult.RiskAssessment,
		AdjustmentSuggestions: analysisResult.AdjustmentSuggestions,
		RawResponse:           rawResponse,
	}

	// 9. 保存到数据库
	if err := db.Dao.Create(analysis).Error; err != nil {
		logger.SugaredLogger.Errorf("保存整体仓位分析结果失败: %v", err)
		return nil, err
	}

	logger.SugaredLogger.Infof("整体仓位分析完成")
	return analysis, nil
}

// callPortfolioAIAnalysis 调用AI进行整体仓位分析
func (s *PositionAnalysisService) callPortfolioAIAnalysis(openAI *OpenAi, messages []map[string]interface{}) (*portfolioAnalysisResult, string, error) {
	client := resty.New()
	client.SetBaseURL(strutil.Trim(openAI.BaseUrl))
	client.SetHeader("Authorization", "Bearer "+openAI.ApiKey)
	client.SetHeader("Content-Type", "application/json")
	if openAI.TimeOut <= 0 {
		openAI.TimeOut = 600
	}
	client.SetTimeout(time.Duration(openAI.TimeOut) * time.Second)

	resp, err := client.R().
		SetBody(map[string]interface{}{
			"model":       openAI.Model,
			"max_tokens":  openAI.MaxTokens,
			"temperature": openAI.Temperature,
			"stream":      false,
			"messages":    messages,
		}).
		Post("/chat/completions")

	if err != nil {
		logger.SugaredLogger.Errorf("AI调用失败: %v", err)
		return nil, "", err
	}

	logger.SugaredLogger.Infof("AI响应: %s", resp.String())

	// 解析AI响应
	var aiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(resp.Body(), &aiResp); err != nil {
		logger.SugaredLogger.Errorf("解析AI响应失败: %v", err)
		return nil, "", err
	}

	if len(aiResp.Choices) == 0 || aiResp.Choices[0].Message.Content == "" {
		return nil, "", errors.New("AI响应为空")
	}

	rawContent := aiResp.Choices[0].Message.Content
	logger.SugaredLogger.Infof("AI响应内容: %s", rawContent)

	// 提取JSON内容
	jsonStr := extractPortfolioAnalysisJSONFromContent(rawContent)
	if jsonStr == "" {
		return nil, rawContent, errors.New("无法从AI响应中提取JSON")
	}

	// 解析分析结果
	var result portfolioAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		logger.SugaredLogger.Errorf("解析分析结果失败: %v", err)
		return nil, rawContent, err
	}

	return &result, rawContent, nil
}

// extractPortfolioAnalysisJSONFromContent 从AI响应中提取JSON
func extractPortfolioAnalysisJSONFromContent(content string) string {
	// 尝试找到第一个 { 和最后一个 }
	startIdx := strings.Index(content, "{")
	endIdx := strings.LastIndex(content, "}")
	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		// 尝试用 ```json 和 ``` 包裹
		if strings.Contains(content, "```json") {
			startIdx = strings.Index(content, "```json") + 7
			endIdx = strings.Index(content[startIdx:], "```")
			if endIdx != -1 {
				return strings.TrimSpace(content[startIdx : startIdx+endIdx])
			}
		}
		if strings.Contains(content, "```") {
			startIdx = strings.Index(content, "```") + 3
			endIdx = strings.Index(content[startIdx:], "```")
			if endIdx != -1 {
				return strings.TrimSpace(content[startIdx : startIdx+endIdx])
			}
		}
		return ""
	}
	return content[startIdx : endIdx+1]
}

// GetLatestPortfolioAnalysis 获取最新的整体仓位分析结果
func (s *PositionAnalysisService) GetLatestPortfolioAnalysis() (*models.PortfolioAnalysis, error) {
	var analysis models.PortfolioAnalysis
	err := db.Dao.Order("created_at DESC").First(&analysis).Error
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}
