package agent

import (
	"context"
	"fmt"

	"go-stock/backend/agent/orchestrator"
	"go-stock/backend/logger"
	"go-stock/backend/models"
)

// MultiAgentStockApi 多 Agent 荐股 API
type MultiAgentStockApi struct {
	aiConfigID  int
	useEnhanced bool // 是否使用增强版协调器
}

// NewMultiAgentStockApi 创建多 Agent 荐股 API（基础版）
func NewMultiAgentStockApi(aiConfigID int) *MultiAgentStockApi {
	return &MultiAgentStockApi{
		aiConfigID:  aiConfigID,
		useEnhanced: false,
	}
}

// NewEnhancedMultiAgentStockApi 创建增强版多 Agent 荐股 API
func NewEnhancedMultiAgentStockApi(aiConfigID int) *MultiAgentStockApi {
	return &MultiAgentStockApi{
		aiConfigID:  aiConfigID,
		useEnhanced: true,
	}
}

// SetUseEnhanced 设置是否使用增强版
func (m *MultiAgentStockApi) SetUseEnhanced(useEnhanced bool) {
	m.useEnhanced = useEnhanced
}

// AnalyzeStock 执行股票分析（使用多 Agent 系统）
func (m *MultiAgentStockApi) AnalyzeStock(ctx context.Context, stockCode, stockName, question string, riskLevel, timeHorizon string) (*models.StockPickReport, error) {
	logger.SugaredLogger.Infof("[多Agent系统] 开始分析股票: %s (%s)", stockName, stockCode)

	// 1. 创建请求
	request := orchestrator.StockAnalysisRequest{
		StockCode:   stockCode,
		StockName:   stockName,
		Question:    question,
		RiskLevel:   riskLevel,
		TimeHorizon: timeHorizon,
	}

	var result *orchestrator.AnalysisResult
	var err error

	if m.useEnhanced {
		// 使用增强版协调器（带重试、日志、Agent交互）
		logger.SugaredLogger.Infof("[多Agent系统] 使用增强版协调器")
		orch := orchestrator.NewEnhancedOrchestrator(m.aiConfigID)
		result, err = orch.AnalyzeWithEnhancements(ctx, request)
	} else {
		// 使用基础版协调器
		logger.SugaredLogger.Infof("[多Agent系统] 使用基础版协调器")
		orch := orchestrator.NewOrchestrator(m.aiConfigID)
		result, err = orch.Analyze(ctx, request)
	}

	if err != nil {
		logger.SugaredLogger.Errorf("[多Agent系统] 分析失败: %v", err)
		return nil, err
	}

	// 4. 转换为现有报告格式
	report := m.convertToStockPickReport(request, result)

	logger.SugaredLogger.Infof("[多Agent系统] 分析完成: %s", stockName)
	return report, nil
}

// convertToStockPickReport 将多 Agent 分析结果转换为现有报告格式
func (m *MultiAgentStockApi) convertToStockPickReport(request orchestrator.StockAnalysisRequest, result *orchestrator.AnalysisResult) *models.StockPickReport {
	report := &models.StockPickReport{
		UserQuery:    request.Question,
		QuerySummary: request.StockName + " - " + request.Question,
		Result:       result.Report,
		Status:       "completed",
		AIConfigID:   uint(m.aiConfigID),
		AIModel:      "多Agent系统",
	}

	// 如果有增强版的日志，也添加到结果中
	if len(result.Logs) > 0 {
		report.Result += "\n\n---\n## 执行日志\n"
		for _, log := range result.Logs {
			report.Result += fmt.Sprintf("\n- [%s] [%s] %s",
				log.Timestamp.Format("2006-01-02 15:04:05"),
				log.AgentID,
				log.Message,
			)
		}
	}

	// 填充推荐股票列表（示例）
	report.Recommendations = []models.RecommendationItem{
		{
			Rank:            1,
			StockCode:       request.StockCode,
			StockName:       request.StockName,
			Reason:          result.Summary,
			RiskLevel:       m.getRiskLevel(result.Risk),
			Score:           result.Confidence * 100,
			TradeSuggestion: m.getTradeSuggestion(result),
		},
	}

	// 从分析结果中提取更多信息
	if result.Technical != nil {
		report.MarketAnalysis = fmt.Sprintf("技术面信号: %s, 趋势: %s",
			result.Technical.Signal, result.Technical.Trend)
	}

	if result.Fundamental != nil {
		report.FilterLogic = fmt.Sprintf("基本面评分: %.0f/100", result.Fundamental.OverallScore)
	}

	report.TotalScanned = 1
	report.CandidatesCount = 1

	return report
}

// getRiskLevel 获取风险等级
func (m *MultiAgentStockApi) getRiskLevel(risk *orchestrator.RiskAssessment) string {
	if risk == nil {
		return "medium"
	}
	switch risk.RiskLevel {
	case "低":
		return "low"
	case "中":
		return "medium"
	case "高":
		return "high"
	default:
		return "medium"
	}
}

// getTradeSuggestion 获取买卖建议
func (m *MultiAgentStockApi) getTradeSuggestion(result *orchestrator.AnalysisResult) string {
	if result == nil || result.Technical == nil {
		return "观望"
	}
	switch result.Technical.Signal {
	case "看多":
		return "买入"
	case "看空":
		return "卖出"
	default:
		return "观望"
	}
}
