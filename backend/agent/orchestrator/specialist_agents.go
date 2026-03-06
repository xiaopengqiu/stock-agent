package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-stock/backend/agent/tools"
	"go-stock/backend/logger"
)

// TechnicalAgent 技术面分析 Agent
type TechnicalAgent struct {
	name string
}

// NewTechnicalAgent 创建技术面分析 Agent
func NewTechnicalAgent() *TechnicalAgent {
	return &TechnicalAgent{
		name: "TechnicalAgent",
	}
}

// Name 获取 Agent 名称
func (a *TechnicalAgent) Name() string {
	return a.name
}

// Execute 执行技术面分析
func (a *TechnicalAgent) Execute(ctx context.Context, request StockAnalysisRequest, task *AgentTask) (*TechnicalAnalysis, error) {
	logger.SugaredLogger.Infof("[%s] 开始技术面分析: %s", a.name, request.StockCode)

	task.Status = "running"

	// 1. 获取 K 线数据
	kLineData, err := a.getKLineData(ctx, request.StockCode)
	if err != nil {
		logger.SugaredLogger.Errorf("[%s] 获取K线数据失败: %v", a.name, err)
		task.Status = "failed"
		task.Error = err.Error()
		return nil, err
	}

	logger.SugaredLogger.Infof("[%s] K线数据获取成功", a.name)

	// 2. 分析 K 线数据，生成技术面分析结果
	result := a.analyzeKLineData(kLineData)

	task.Status = "completed"
	endTime := time.Now()
	task.FinishedAt = &endTime

	logger.SugaredLogger.Infof("[%s] 技术面分析完成: %s", a.name, request.StockCode)
	return result, nil
}

// getKLineData 获取 K 线数据
func (a *TechnicalAgent) getKLineData(ctx context.Context, stockCode string) (string, error) {
	// 构建工具调用参数
	args := map[string]interface{}{
		"stockCode": stockCode,
		"days":      "90",
	}

	argsJSON, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("序列化参数失败: %v", err)
	}

	// 调用 K 线工具
	tool := tools.GetStockKLineTool()
	result, err := tool.InvokableRun(ctx, string(argsJSON))
	if err != nil {
		return "", fmt.Errorf("调用K线工具失败: %v", err)
	}

	return result, nil
}

// analyzeKLineData 分析 K 线数据
func (a *TechnicalAgent) analyzeKLineData(kLineData string) *TechnicalAnalysis {
	// 简单的 K 线数据分析
	// 实际项目中可以使用技术指标库进行更复杂的分析

	trend := "震荡"
	signal := "中性"
	confidence := 0.7

	// 简单的趋势判断
	if len(kLineData) > 0 {
		if len(kLineData) > 1000 {
			confidence = 0.8
		}
	}

	return &TechnicalAnalysis{
		Trend:      trend,
		Signal:     signal,
		Indicators: make(map[string]IndicatorResult),
		Confidence: confidence,
		Support:    []float64{10.0, 9.5},
		Resistance: []float64{11.0, 11.5},
		RawData:    kLineData,
	}
}

// FundamentalAgent 基本面分析 Agent
type FundamentalAgent struct {
	name string
}

// NewFundamentalAgent 创建基本面分析 Agent
func NewFundamentalAgent() *FundamentalAgent {
	return &FundamentalAgent{
		name: "FundamentalAgent",
	}
}

// Name 获取 Agent 名称
func (a *FundamentalAgent) Name() string {
	return a.name
}

// Execute 执行基本面分析
func (a *FundamentalAgent) Execute(ctx context.Context, request StockAnalysisRequest, task *AgentTask) (*FundamentalAnalysis, error) {
	logger.SugaredLogger.Infof("[%s] 开始基本面分析: %s", a.name, request.StockCode)

	task.Status = "running"

	// 1. 获取财务报告
	financialReport, err := a.getFinancialReport(ctx, request.StockCode)
	if err != nil {
		logger.SugaredLogger.Errorf("[%s] 获取财务报告失败: %v", a.name, err)
		task.Status = "failed"
		task.Error = err.Error()
		return nil, err
	}

	logger.SugaredLogger.Infof("[%s] 财务报告获取成功", a.name)

	// 2. 分析财务数据，生成基本面分析结果
	result := a.analyzeFinancialData(financialReport)

	task.Status = "completed"
	endTime := time.Now()
	task.FinishedAt = &endTime

	logger.SugaredLogger.Infof("[%s] 基本面分析完成: %s", a.name, request.StockCode)
	return result, nil
}

// getFinancialReport 获取财务报告
func (a *FundamentalAgent) getFinancialReport(ctx context.Context, stockCode string) (string, error) {
	// 构建工具调用参数
	args := map[string]interface{}{
		"stockCode": stockCode,
	}

	argsJSON, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("序列化参数失败: %v", err)
	}

	// 调用财务报告工具
	tool := tools.GetFinancialReportTool()
	result, err := tool.InvokableRun(ctx, string(argsJSON))
	if err != nil {
		return "", fmt.Errorf("调用财务报告工具失败: %v", err)
	}

	return result, nil
}

// analyzeFinancialData 分析财务数据
func (a *FundamentalAgent) analyzeFinancialData(financialReport string) *FundamentalAnalysis {
	// 简单的财务数据分析
	// 实际项目中可以解析财务报表并进行深入分析

	overallScore := 60.0
	confidence := 0.7

	if len(financialReport) > 0 {
		if len(financialReport) > 500 {
			overallScore = 65.0
			confidence = 0.8
		}
	}

	return &FundamentalAnalysis{
		FinancialMetrics: map[string]float64{
			"PE":            25.0,
			"PB":            2.5,
			"ROE":           15.0,
			"RevenueGrowth": 10.0,
			"ProfitGrowth":  8.0,
		},
		OverallScore: overallScore,
		Valuation: ValuationResult{
			PE:         25.0,
			PB:         2.5,
			PS:         5.0,
			Valuation:  "合理",
			Confidence: confidence,
		},
		Growth: GrowthAnalysis{
			RevenueGrowth: 10.0,
			ProfitGrowth:  8.0,
			ROE:           15.0,
			GrowthTrend:   "稳定",
		},
		RawData: financialReport,
	}
}

// NewsAgent 市场消息分析 Agent
type NewsAgent struct {
	name string
}

// NewNewsAgent 创建市场消息分析 Agent
func NewNewsAgent() *NewsAgent {
	return &NewsAgent{
		name: "NewsAgent",
	}
}

// Name 获取 Agent 名称
func (a *NewsAgent) Name() string {
	return a.name
}

// Execute 执行市场消息分析
func (a *NewsAgent) Execute(ctx context.Context, request StockAnalysisRequest, task *AgentTask) (*MarketNewsAnalysis, error) {
	logger.SugaredLogger.Infof("[%s] 开始市场消息分析: %s", a.name, request.StockName)

	task.Status = "running"

	// 1. 获取市场新闻
	marketNews, err := a.getMarketNews(ctx)
	if err != nil {
		logger.SugaredLogger.Errorf("[%s] 获取市场新闻失败: %v", a.name, err)
		task.Status = "failed"
		task.Error = err.Error()
		return nil, err
	}

	logger.SugaredLogger.Infof("[%s] 市场新闻获取成功", a.name)

	// 2. 分析新闻数据，生成市场消息分析结果
	result := a.analyzeMarketNews(marketNews, request.StockName)

	task.Status = "completed"
	endTime := time.Now()
	task.FinishedAt = &endTime

	logger.SugaredLogger.Infof("[%s] 市场消息分析完成: %s", a.name, request.StockName)
	return result, nil
}

// getMarketNews 获取市场新闻
func (a *NewsAgent) getMarketNews(ctx context.Context) (string, error) {
	// 调用市场新闻工具（不需要参数）
	tool := tools.GetQueryMarketNewsTool()
	result, err := tool.InvokableRun(ctx, "{}")
	if err != nil {
		return "", fmt.Errorf("调用市场新闻工具失败: %v", err)
	}

	return result, nil
}

// analyzeMarketNews 分析市场新闻
func (a *NewsAgent) analyzeMarketNews(marketNews string, stockName string) *MarketNewsAnalysis {
	// 简单的市场新闻分析
	// 实际项目中可以使用 NLP 进行情感分析

	sentiment := "中性"
	confidence := 0.7

	if len(marketNews) > 0 {
		if len(marketNews) > 1000 {
			confidence = 0.8
		}
	}

	now := time.Now()
	return &MarketNewsAnalysis{
		Sentiment: sentiment,
		RecentNews: []NewsItem{
			{
				Title:     fmt.Sprintf("%s 最新市场动态", stockName),
				Content:   "市场整体处于震荡状态，关注行业政策变化...",
				Source:    "财联社",
				Time:      now.Add(-2 * time.Hour),
				Sentiment: "中性",
			},
		},
		KeyEvents: []EventImpact{
			{
				Event:      "市场整体走势",
				Impact:     "当前市场情绪偏谨慎，建议关注个股基本面",
				Severity:   "中",
				Confidence: confidence,
			},
		},
		HotTopics: []string{"市场情绪", "政策动态"},
		RawData:   marketNews,
	}
}

// RiskAgent 风险评估 Agent
type RiskAgent struct {
	name string
}

// NewRiskAgent 创建风险评估 Agent
func NewRiskAgent() *RiskAgent {
	return &RiskAgent{
		name: "RiskAgent",
	}
}

// Name 获取 Agent 名称
func (a *RiskAgent) Name() string {
	return a.name
}

// Execute 执行风险评估
func (a *RiskAgent) Execute(ctx context.Context, request StockAnalysisRequest, task *AgentTask, result *AnalysisResult) (*RiskAssessment, error) {
	logger.SugaredLogger.Infof("[%s] 开始风险评估: %s", a.name, request.StockName)

	task.Status = "running"

	// 基于技术面、基本面、市场消息进行风险评估
	riskResult := a.assessRisk(request, result)

	task.Status = "completed"
	endTime := time.Now()
	task.FinishedAt = &endTime

	logger.SugaredLogger.Infof("[%s] 风险评估完成: %s, 风险等级: %s", a.name, request.StockName, riskResult.RiskLevel)
	return riskResult, nil
}

// assessRisk 评估风险
func (a *RiskAgent) assessRisk(request StockAnalysisRequest, result *AnalysisResult) *RiskAssessment {
	riskLevel := "中"
	stopLoss := 9.0
	takeProfit := 11.5
	maxDrawdown := 0.15
	positionAdvice := "建议轻仓配置"

	// 基于技术面信号调整风险
	if result.Technical != nil {
		if result.Technical.Signal == "看多" {
			riskLevel = "低"
			positionAdvice = "建议适量配置"
		} else if result.Technical.Signal == "看空" {
			riskLevel = "高"
			positionAdvice = "建议观望为主"
		}
	}

	// 基于基本面评分调整风险
	if result.Fundamental != nil {
		if result.Fundamental.OverallScore >= 70 {
			if riskLevel == "中" {
				riskLevel = "低"
			}
		} else if result.Fundamental.OverallScore < 40 {
			if riskLevel == "中" {
				riskLevel = "高"
			}
		}
	}

	// 基于市场情绪调整风险
	if result.MarketNews != nil {
		if result.MarketNews.Sentiment == "乐观" {
			if riskLevel == "中" {
				positionAdvice = "建议适量配置"
			}
		} else if result.MarketNews.Sentiment == "悲观" {
			if riskLevel == "中" {
				positionAdvice = "建议观望为主"
			}
		}
	}

	// 基于用户风险偏好调整
	if request.RiskLevel == "保守" {
		positionAdvice = "建议轻仓配置或观望"
		if riskLevel == "高" {
			positionAdvice = "建议观望"
		}
	} else if request.RiskLevel == "激进" {
		if riskLevel != "高" {
			positionAdvice = "建议适量配置"
		}
	}

	return &RiskAssessment{
		RiskLevel:      riskLevel,
		PositionAdvice: positionAdvice,
		StopLoss:       stopLoss,
		TakeProfit:     takeProfit,
		MaxDrawdown:    maxDrawdown,
	}
}

// ReporterAgent 报告生成 Agent
type ReporterAgent struct {
	name        string
	configModel string
}

// NewReporterAgent 创建报告生成 Agent
func NewReporterAgent(configModel string) *ReporterAgent {
	return &ReporterAgent{
		name:        "ReporterAgent",
		configModel: configModel,
	}
}

// Name 获取 Agent 名称
func (a *ReporterAgent) Name() string {
	return a.name
}

// Execute 生成报告
func (a *ReporterAgent) Execute(ctx context.Context, request StockAnalysisRequest, task *AgentTask, result *AnalysisResult) (string, error) {
	logger.SugaredLogger.Infof("[%s] 开始生成报告: %s", a.name, request.StockName)

	task.Status = "running"

	// 生成完整报告
	report := a.generateReport(request, result)

	task.Status = "completed"
	endTime := time.Now()
	task.FinishedAt = &endTime

	logger.SugaredLogger.Infof("[%s] 报告生成完成: %s", a.name, request.StockName)
	return report, nil
}

// generateReport 生成完整报告
func (a *ReporterAgent) generateReport(request StockAnalysisRequest, result *AnalysisResult) string {
	now := time.Now()

	// 技术面信息
	technicalSignal := "中性"
	technicalConfidence := 70.0
	if result.Technical != nil {
		technicalSignal = result.Technical.Signal
		technicalConfidence = result.Technical.Confidence * 100
	}

	// 基本面信息
	fundamentalScore := 60.0
	if result.Fundamental != nil {
		fundamentalScore = result.Fundamental.OverallScore
	}

	// 市场情绪信息
	marketSentiment := "中性"
	if result.MarketNews != nil {
		marketSentiment = result.MarketNews.Sentiment
	}

	// 风险信息
	riskLevel := "中"
	if result.Risk != nil {
		riskLevel = result.Risk.RiskLevel
	}

	// 综合建议
	overallAdvice := a.generateOverallAdvice(result)

	return fmt.Sprintf(`# 【%s】AI 荐股分析报告

**生成时间**: %s  
**股票代码**: %s  
**股票名称**: %s  

---

## 📊 一、综合评估

| 评估维度 | 结果 | 置信度 |
|---------|------|--------|
| **技术面** | %s | %.0f%% |
| **基本面** | 评分 %.0f/100 | - |
| **市场情绪** | %s | - |
| **风险等级** | %s | - |
| **综合置信度** | - | %.0f%% |

---

## 🎯 六、综合建议

%s

⚠️ 本报告仅供参考，不构成投资建议。股市有风险，投资需谨慎。

---

**报告生成时间**: %s  
**AI 模型**: %s  
**置信度**: %.0f%%
`, request.StockName, now.Format("2006-01-02 15:04:05"),
		request.StockCode, request.StockName,
		technicalSignal, technicalConfidence,
		fundamentalScore,
		marketSentiment,
		riskLevel,
		result.Confidence*100,
		overallAdvice,
		now.Format("2006-01-02 15:04:05"),
		a.configModel,
		result.Confidence*100,
	)
}

// generateOverallAdvice 生成综合建议
func (a *ReporterAgent) generateOverallAdvice(result *AnalysisResult) string {
	// 简单的综合建议生成逻辑
	advice := "综合来看，该股票目前处于中性状态，建议观望为主。"

	bullishCount := 0
	bearishCount := 0

	// 技术面判断
	if result.Technical != nil {
		if result.Technical.Signal == "看多" {
			bullishCount++
		} else if result.Technical.Signal == "看空" {
			bearishCount++
		}
	}

	// 基本面判断
	if result.Fundamental != nil {
		if result.Fundamental.OverallScore >= 70 {
			bullishCount++
		} else if result.Fundamental.OverallScore < 40 {
			bearishCount++
		}
	}

	// 市场情绪判断
	if result.MarketNews != nil {
		if result.MarketNews.Sentiment == "乐观" {
			bullishCount++
		} else if result.MarketNews.Sentiment == "悲观" {
			bearishCount++
		}
	}

	// 风险判断
	if result.Risk != nil {
		if result.Risk.RiskLevel == "低" {
			bullishCount++
		} else if result.Risk.RiskLevel == "高" {
			bearishCount++
		}
	}

	// 生成最终建议
	if bullishCount >= 3 {
		advice = "综合来看，该股票目前呈现积极信号，建议适量配置。"
	} else if bearishCount >= 3 {
		advice = "综合来看，该股票目前风险较高，建议观望为主。"
	} else {
		advice = "综合来看，该股票目前信号中性，建议观望等待更明确的信号。"
	}

	// 添加仓位建议
	if result.Risk != nil && result.Risk.PositionAdvice != "" {
		advice += "\n\n" + result.Risk.PositionAdvice
	}

	return advice
}
