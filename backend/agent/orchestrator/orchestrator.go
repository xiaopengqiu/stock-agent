package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go-stock/backend/data"
	"go-stock/backend/logger"
)

// Orchestrator 主协调器 Agent
type Orchestrator struct {
	aiConfigID int
	config     *data.AIConfig
}

// NewOrchestrator 创建新的协调器
func NewOrchestrator(aiConfigID int) *Orchestrator {
	return &Orchestrator{
		aiConfigID: aiConfigID,
	}
}

// Analyze 执行完整的股票分析
func (o *Orchestrator) Analyze(ctx context.Context, request StockAnalysisRequest) (*AnalysisResult, error) {
	logger.SugaredLogger.Infof("开始分析股票: %s (%s)", request.StockName, request.StockCode)

	// 获取 AI 配置
	settingConfig := data.GetSettingConfig()
	aiConfig, ok := loFind(settingConfig.AiConfigs, func(item *data.AIConfig) bool {
		return uint(o.aiConfigID) == item.ID
	})
	if !ok {
		return nil, fmt.Errorf("未找到 AI 配置: %d", o.aiConfigID)
	}
	o.config = aiConfig

	// 创建任务列表
	tasks := o.createTasks(request)

	// 初始化结果
	result := &AnalysisResult{
		Confidence: 0.8,
	}

	// 并行执行技术面、基本面、市场消息分析
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstError error

	// 1. 技术面分析（并行）
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.SugaredLogger.Infof("执行技术面分析...")
		agent := NewTechnicalAgent()
		technicalResult, err := agent.Execute(ctx, request, tasks[0])
		if err != nil {
			logger.SugaredLogger.Errorf("技术面分析失败: %v", err)
			if firstError == nil {
				firstError = err
			}
			return
		}
		mu.Lock()
		result.Technical = technicalResult
		mu.Unlock()
		logger.SugaredLogger.Infof("技术面分析完成")
	}()

	// 2. 基本面分析（并行）
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.SugaredLogger.Infof("执行基本面分析...")
		agent := NewFundamentalAgent()
		fundamentalResult, err := agent.Execute(ctx, request, tasks[1])
		if err != nil {
			logger.SugaredLogger.Errorf("基本面分析失败: %v", err)
			if firstError == nil {
				firstError = err
			}
			return
		}
		mu.Lock()
		result.Fundamental = fundamentalResult
		mu.Unlock()
		logger.SugaredLogger.Infof("基本面分析完成")
	}()

	// 3. 市场消息分析（并行）
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.SugaredLogger.Infof("执行市场消息分析...")
		agent := NewNewsAgent()
		newsResult, err := agent.Execute(ctx, request, tasks[2])
		if err != nil {
			logger.SugaredLogger.Errorf("市场消息分析失败: %v", err)
			if firstError == nil {
				firstError = err
			}
			return
		}
		mu.Lock()
		result.MarketNews = newsResult
		mu.Unlock()
		logger.SugaredLogger.Infof("市场消息分析完成")
	}()

	// 等待并行任务完成
	logger.SugaredLogger.Infof("等待并行任务完成...")
	wg.Wait()

	// 检查是否有错误
	if firstError != nil {
		logger.SugaredLogger.Warnf("部分分析任务失败: %v, 但继续执行后续任务", firstError)
	}

	// 4. 风险评估（串行，依赖前面的结果）
	logger.SugaredLogger.Infof("执行风险评估...")
	riskAgent := NewRiskAgent()
	riskResult, err := riskAgent.Execute(ctx, request, tasks[3], result)
	if err != nil {
		logger.SugaredLogger.Errorf("风险评估失败: %v", err)
	} else {
		result.Risk = riskResult
	}

	// 5. 报告生成（串行，依赖前面的结果）
	logger.SugaredLogger.Infof("生成最终报告...")
	reporterAgent := NewReporterAgent(o.config.ModelName)
	report, err := reporterAgent.Execute(ctx, request, tasks[4], result)
	if err != nil {
		logger.SugaredLogger.Errorf("报告生成失败: %v", err)
	} else {
		result.Report = report
	}

	// 生成综合总结
	result.Summary = o.generateSummary(result)

	logger.SugaredLogger.Infof("股票分析完成: %s", request.StockName)
	return result, nil
}

// createTasks 创建任务列表
func (o *Orchestrator) createTasks(request StockAnalysisRequest) []*AgentTask {
	now := time.Now()
	return []*AgentTask{
		{
			ID:        "technical-1",
			Type:      "technical",
			Status:    "pending",
			Request:   request,
			CreatedAt: now,
		},
		{
			ID:        "fundamental-1",
			Type:      "fundamental",
			Status:    "pending",
			Request:   request,
			CreatedAt: now,
		},
		{
			ID:        "news-1",
			Type:      "news",
			Status:    "pending",
			Request:   request,
			CreatedAt: now,
		},
		{
			ID:        "risk-1",
			Type:      "risk",
			Status:    "pending",
			Request:   request,
			CreatedAt: now,
		},
		{
			ID:        "reporter-1",
			Type:      "reporter",
			Status:    "pending",
			Request:   request,
			CreatedAt: now,
		},
	}
}

// generateSummary 生成综合总结
func (o *Orchestrator) generateSummary(result *AnalysisResult) string {
	summary := fmt.Sprintf("【%s】综合分析报告\n\n", time.Now().Format("2006-01-02 15:04"))

	if result.Technical != nil {
		summary += fmt.Sprintf("📊 技术面: %s (置信度: %.0f%%)\n",
			result.Technical.Signal, result.Technical.Confidence*100)
	}

	if result.Fundamental != nil {
		summary += fmt.Sprintf("📈 基本面: 评分 %.0f/100\n", result.Fundamental.OverallScore)
	}

	if result.MarketNews != nil {
		summary += fmt.Sprintf("📰 市场情绪: %s\n", result.MarketNews.Sentiment)
	}

	if result.Risk != nil {
		summary += fmt.Sprintf("⚠️ 风险等级: %s\n", result.Risk.RiskLevel)
		if result.Risk.PositionAdvice != "" {
			summary += fmt.Sprintf("💡 仓位建议: %s\n", result.Risk.PositionAdvice)
		}
	}

	summary += fmt.Sprintf("\n🎯 综合置信度: %.0f%%", result.Confidence*100)
	return summary
}

// loFind 简化的查找函数
func loFind[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, item := range slice {
		if predicate(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// ToJSON 转换为 JSON
func (r *AnalysisResult) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
