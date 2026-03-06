package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go-stock/backend/agent/tools"
	"go-stock/backend/data"
	"go-stock/backend/logger"
)

// ============================================
// Phase 3: 增强和优化 - 重试机制、详细日志、Agent交互
// ============================================

// EnhancedOrchestrator 增强版协调器
type EnhancedOrchestrator struct {
	Orchestrator
	logs        []AgentLogEntry
	logsMu      sync.Mutex
	messages    []AgentMessage
	messagesMu  sync.Mutex
	toolCalls   []ToolCall
	toolCallsMu sync.Mutex
	retryConfig *RetryConfig
	cache       *AgentCache
	useCache    bool
}

// NewEnhancedOrchestrator 创建增强版协调器
func NewEnhancedOrchestrator(aiConfigID int) *EnhancedOrchestrator {
	return &EnhancedOrchestrator{
		Orchestrator: *NewOrchestrator(aiConfigID),
		logs:         make([]AgentLogEntry, 0),
		messages:     make([]AgentMessage, 0),
		toolCalls:    make([]ToolCall, 0),
		retryConfig: &RetryConfig{
			MaxRetries:    3,
			InitialDelay:  1 * time.Second,
			MaxDelay:      10 * time.Second,
			BackoffFactor: 2.0,
		},
		cache:    GetGlobalCache(),
		useCache: true,
	}
}

// SetUseCache 设置是否使用缓存
func (e *EnhancedOrchestrator) SetUseCache(useCache bool) {
	e.useCache = useCache
	e.AddLog("orchestrator", "info", fmt.Sprintf("缓存已%s", map[bool]string{true: "启用", false: "禁用"}[useCache]), nil)
}

// SetCache 设置自定义缓存
func (e *EnhancedOrchestrator) SetCache(cache *AgentCache) {
	e.cache = cache
	e.AddLog("orchestrator", "info", "已设置自定义缓存", nil)
}

// GetCache 获取缓存
func (e *EnhancedOrchestrator) GetCache() *AgentCache {
	return e.cache
}

// SetRetryConfig 设置重试配置
func (e *EnhancedOrchestrator) SetRetryConfig(config *RetryConfig) {
	e.retryConfig = config
}

// AddLog 添加日志
func (e *EnhancedOrchestrator) AddLog(agentID, level, message string, data map[string]interface{}) {
	e.logsMu.Lock()
	defer e.logsMu.Unlock()

	logEntry := AgentLogEntry{
		AgentID:   agentID,
		Level:     level,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
	e.logs = append(e.logs, logEntry)

	// 同时输出到 logger
	switch level {
	case "info":
		logger.SugaredLogger.Infof("[%s] %s", agentID, message)
	case "warn":
		logger.SugaredLogger.Warnf("[%s] %s", agentID, message)
	case "error":
		logger.SugaredLogger.Errorf("[%s] %s", agentID, message)
	case "debug":
		logger.SugaredLogger.Debugf("[%s] %s", agentID, message)
	}
}

// SendMessage Agent 间发送消息
func (e *EnhancedOrchestrator) SendMessage(from, to, msgType string, content interface{}) {
	e.messagesMu.Lock()
	defer e.messagesMu.Unlock()

	message := AgentMessage{
		From:      from,
		To:        to,
		Type:      msgType,
		Content:   content,
		Timestamp: time.Now(),
	}
	e.messages = append(e.messages, message)

	e.AddLog(from, "info", fmt.Sprintf("发送消息到 %s: %s", to, msgType), map[string]interface{}{
		"to":      to,
		"type":    msgType,
		"content": content,
	})
}

// GetMessagesFor 获取发给指定 Agent 的消息
func (e *EnhancedOrchestrator) GetMessagesFor(to string) []AgentMessage {
	e.messagesMu.Lock()
	defer e.messagesMu.Unlock()

	result := make([]AgentMessage, 0)
	for _, msg := range e.messages {
		if msg.To == to {
			result = append(result, msg)
		}
	}
	return result
}

// RecordToolCall 记录工具调用
func (e *EnhancedOrchestrator) RecordToolCall(call *ToolCall) {
	e.toolCallsMu.Lock()
	defer e.toolCallsMu.Unlock()

	e.toolCalls = append(e.toolCalls, *call)

	e.AddLog("tool", "info", fmt.Sprintf("工具调用: %s", call.ToolName), map[string]interface{}{
		"toolName": call.ToolName,
		"success":  call.Success,
		"duration": call.Duration.String(),
	})
}

// ExecuteWithRetry 带重试的执行
func (e *EnhancedOrchestrator) ExecuteWithRetry(
	ctx context.Context,
	agentID string,
	fn func() (interface{}, error),
) (interface{}, error) {
	var lastErr error
	var result interface{}

	for attempt := 0; attempt <= e.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// 计算延迟
			delay := e.calculateRetryDelay(attempt)
			e.AddLog(agentID, "info", fmt.Sprintf("重试 %d/%d，等待 %s...", attempt, e.retryConfig.MaxRetries, delay), nil)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		startTime := time.Now()
		result, lastErr = fn()
		duration := time.Since(startTime)

		if lastErr == nil {
			e.AddLog(agentID, "info", fmt.Sprintf("执行成功，耗时 %s", duration), nil)
			return result, nil
		}

		e.AddLog(agentID, "warn", fmt.Sprintf("执行失败 (尝试 %d/%d): %v", attempt+1, e.retryConfig.MaxRetries+1, lastErr), nil)
	}

	return nil, fmt.Errorf("执行失败，已重试 %d 次: %w", e.retryConfig.MaxRetries, lastErr)
}

// calculateRetryDelay 计算重试延迟
func (e *EnhancedOrchestrator) calculateRetryDelay(attempt int) time.Duration {
	delay := e.retryConfig.InitialDelay
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * e.retryConfig.BackoffFactor)
		if delay > e.retryConfig.MaxDelay {
			delay = e.retryConfig.MaxDelay
			break
		}
	}
	return delay
}

// AnalyzeWithEnhancements 增强版分析（带重试和日志）
func (e *EnhancedOrchestrator) AnalyzeWithEnhancements(ctx context.Context, request StockAnalysisRequest) (*AnalysisResult, error) {
	e.AddLog("orchestrator", "info", "开始增强版股票分析", map[string]interface{}{
		"stockCode": request.StockCode,
		"stockName": request.StockName,
	})

	// 获取 AI 配置
	settingConfig := data.GetSettingConfig()
	aiConfig, ok := loFind(settingConfig.AiConfigs, func(item *data.AIConfig) bool {
		return uint(e.aiConfigID) == item.ID
	})
	if !ok {
		e.AddLog("orchestrator", "error", "未找到 AI 配置", nil)
		return nil, fmt.Errorf("未找到 AI 配置: %d", e.aiConfigID)
	}
	e.config = aiConfig

	// 创建任务列表
	tasks := e.createEnhancedTasks(request)

	// 初始化结果
	result := &AnalysisResult{
		Confidence: 0.8,
		Logs:       make([]AgentLogEntry, 0),
	}

	// 并行执行技术面、基本面、市场消息分析
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstError error

	// 1. 技术面分析（并行，带重试）
	wg.Add(1)
	go func() {
		defer wg.Done()
		e.AddLog("orchestrator", "info", "启动技术面分析...", nil)

		agent := NewTechnicalAgent()
		technicalResult, err := e.ExecuteWithRetry(ctx, "technical", func() (interface{}, error) {
			return agent.Execute(ctx, request, tasks[0])
		})

		if err != nil {
			e.AddLog("orchestrator", "error", fmt.Sprintf("技术面分析失败: %v", err), nil)
			if firstError == nil {
				firstError = err
			}
			return
		}

		mu.Lock()
		result.Technical = technicalResult.(*TechnicalAnalysis)
		mu.Unlock()

		e.AddLog("orchestrator", "info", "技术面分析完成", map[string]interface{}{
			"signal":     result.Technical.Signal,
			"confidence": result.Technical.Confidence,
		})
	}()

	// 2. 基本面分析（并行，带重试）
	wg.Add(1)
	go func() {
		defer wg.Done()
		e.AddLog("orchestrator", "info", "启动基本面分析...", nil)

		agent := NewFundamentalAgent()
		fundamentalResult, err := e.ExecuteWithRetry(ctx, "fundamental", func() (interface{}, error) {
			return agent.Execute(ctx, request, tasks[1])
		})

		if err != nil {
			e.AddLog("orchestrator", "error", fmt.Sprintf("基本面分析失败: %v", err), nil)
			if firstError == nil {
				firstError = err
			}
			return
		}

		mu.Lock()
		result.Fundamental = fundamentalResult.(*FundamentalAnalysis)
		mu.Unlock()

		e.AddLog("orchestrator", "info", "基本面分析完成", map[string]interface{}{
			"overallScore": result.Fundamental.OverallScore,
		})
	}()

	// 3. 市场消息分析（并行，带重试）
	wg.Add(1)
	go func() {
		defer wg.Done()
		e.AddLog("orchestrator", "info", "启动市场消息分析...", nil)

		agent := NewNewsAgent()
		newsResult, err := e.ExecuteWithRetry(ctx, "news", func() (interface{}, error) {
			return agent.Execute(ctx, request, tasks[2])
		})

		if err != nil {
			e.AddLog("orchestrator", "error", fmt.Sprintf("市场消息分析失败: %v", err), nil)
			if firstError == nil {
				firstError = err
			}
			return
		}

		mu.Lock()
		result.MarketNews = newsResult.(*MarketNewsAnalysis)
		mu.Unlock()

		e.AddLog("orchestrator", "info", "市场消息分析完成", map[string]interface{}{
			"sentiment": result.MarketNews.Sentiment,
		})
	}()

	// 等待并行任务完成
	e.AddLog("orchestrator", "info", "等待并行任务完成...", nil)
	wg.Wait()

	// 检查是否有错误
	if firstError != nil {
		e.AddLog("orchestrator", "warn", fmt.Sprintf("部分分析任务失败: %v, 但继续执行后续任务", firstError), nil)
	}

	// Agent 间消息交互 - 技术面向基本面发送数据
	if result.Technical != nil && result.Fundamental != nil {
		e.SendMessage("technical", "fundamental", "notify", map[string]interface{}{
			"signal":     result.Technical.Signal,
			"confidence": result.Technical.Confidence,
		})
	}

	// 4. 风险评估（串行，带重试）
	e.AddLog("orchestrator", "info", "执行风险评估...", nil)
	riskAgent := NewRiskAgent()
	riskResult, err := e.ExecuteWithRetry(ctx, "risk", func() (interface{}, error) {
		return riskAgent.Execute(ctx, request, tasks[3], result)
	})

	if err != nil {
		e.AddLog("orchestrator", "error", fmt.Sprintf("风险评估失败: %v", err), nil)
	} else {
		result.Risk = riskResult.(*RiskAssessment)
		e.AddLog("orchestrator", "info", "风险评估完成", map[string]interface{}{
			"riskLevel": result.Risk.RiskLevel,
		})
	}

	// 5. 报告生成（串行，带重试）
	e.AddLog("orchestrator", "info", "生成最终报告...", nil)
	reporterAgent := NewReporterAgent(e.config.ModelName)
	report, err := e.ExecuteWithRetry(ctx, "reporter", func() (interface{}, error) {
		return reporterAgent.Execute(ctx, request, tasks[4], result)
	})

	if err != nil {
		e.AddLog("orchestrator", "error", fmt.Sprintf("报告生成失败: %v", err), nil)
	} else {
		result.Report = report.(string)
		e.AddLog("orchestrator", "info", "报告生成完成", nil)
	}

	// 生成综合总结
	result.Summary = e.generateSummary(result)

	// 复制日志到结果
	e.logsMu.Lock()
	result.Logs = append([]AgentLogEntry{}, e.logs...)
	e.logsMu.Unlock()

	e.AddLog("orchestrator", "info", "股票分析完成", map[string]interface{}{
		"stockName":  request.StockName,
		"confidence": result.Confidence,
		"totalLogs":  len(result.Logs),
	})

	return result, nil
}

// createEnhancedTasks 创建增强版任务列表
func (e *EnhancedOrchestrator) createEnhancedTasks(request StockAnalysisRequest) []*AgentTask {
	now := time.Now()
	return []*AgentTask{
		{
			ID:            "technical-1",
			Type:          "technical",
			Status:        "pending",
			Request:       request,
			CreatedAt:     now,
			MaxRetries:    e.retryConfig.MaxRetries,
			RetryInterval: e.retryConfig.InitialDelay,
		},
		{
			ID:            "fundamental-1",
			Type:          "fundamental",
			Status:        "pending",
			Request:       request,
			CreatedAt:     now,
			MaxRetries:    e.retryConfig.MaxRetries,
			RetryInterval: e.retryConfig.InitialDelay,
		},
		{
			ID:            "news-1",
			Type:          "news",
			Status:        "pending",
			Request:       request,
			CreatedAt:     now,
			MaxRetries:    e.retryConfig.MaxRetries,
			RetryInterval: e.retryConfig.InitialDelay,
		},
		{
			ID:            "risk-1",
			Type:          "risk",
			Status:        "pending",
			Request:       request,
			CreatedAt:     now,
			MaxRetries:    e.retryConfig.MaxRetries,
			RetryInterval: e.retryConfig.InitialDelay,
		},
		{
			ID:            "reporter-1",
			Type:          "reporter",
			Status:        "pending",
			Request:       request,
			CreatedAt:     now,
			MaxRetries:    e.retryConfig.MaxRetries,
			RetryInterval: e.retryConfig.InitialDelay,
		},
	}
}

// GetLogs 获取所有日志
func (e *EnhancedOrchestrator) GetLogs() []AgentLogEntry {
	e.logsMu.Lock()
	defer e.logsMu.Unlock()
	return append([]AgentLogEntry{}, e.logs...)
}

// GetToolCalls 获取所有工具调用记录
func (e *EnhancedOrchestrator) GetToolCalls() []ToolCall {
	e.toolCallsMu.Lock()
	defer e.toolCallsMu.Unlock()
	return append([]ToolCall{}, e.toolCalls...)
}

// GetMessages 获取所有消息
func (e *EnhancedOrchestrator) GetMessages() []AgentMessage {
	e.messagesMu.Lock()
	defer e.messagesMu.Unlock()
	return append([]AgentMessage{}, e.messages...)
}

// CallToolWithRetry 带重试的工具调用
func (e *EnhancedOrchestrator) CallToolWithRetry(ctx context.Context, toolName string, params interface{}) (interface{}, error) {
	call := &ToolCall{
		ToolName:   toolName,
		Parameters: params,
		CalledAt:   time.Now(),
	}

	result, err := e.ExecuteWithRetry(ctx, "tool-"+toolName, func() (interface{}, error) {
		// 这里根据工具名称调用对应的工具
		switch toolName {
		case "QueryStockKLine":
			argsJSON, _ := json.Marshal(params)
			return tools.GetStockKLineTool().InvokableRun(ctx, string(argsJSON))
		case "GetFinancialReport":
			argsJSON, _ := json.Marshal(params)
			return tools.GetFinancialReportTool().InvokableRun(ctx, string(argsJSON))
		case "QueryMarketNews":
			argsJSON, _ := json.Marshal(params)
			return tools.GetQueryMarketNewsTool().InvokableRun(ctx, string(argsJSON))
		case "QueryShareholderCount":
			argsJSON, _ := json.Marshal(params)
			return tools.GetQueryShareholderCountTool().InvokableRun(ctx, string(argsJSON))
		default:
			return nil, fmt.Errorf("未知工具: %s", toolName)
		}
	})

	call.Duration = time.Since(call.CalledAt)
	call.Result = result
	if err != nil {
		call.Error = err.Error()
		call.Success = false
	} else {
		call.Success = true
	}

	e.RecordToolCall(call)
	return result, err
}
