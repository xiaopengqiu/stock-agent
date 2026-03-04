package adaptive

import (
	"encoding/json"
	"fmt"
	"go-stock/backend/logger"
	"sync"
	"time"
)

// FailureType 失败类型
type FailureType string

const (
	FailureTypeTimeout    FailureType = "timeout"    // 超时
	FailureTypeException  FailureType = "exception"  // 异常
	FailureTypeValidation FailureType = "validation" // 验证失败
	FailureTypeResource   FailureType = "resource"   // 资源不足
	FailureTypeDependency FailureType = "dependency" // 依赖失败
	FailureTypeLogic      FailureType = "logic"      // 逻辑错误
	FailureTypeUnknown    FailureType = "unknown"    // 未知
)

// FailureRecord 失败记录
type FailureRecord struct {
	ID            string                 `json:"id"`
	WorkflowID    string                 `json:"workflowId"`
	StepID        string                 `json:"stepId"`
	AgentName     string                 `json:"agentName"`
	FailureType   FailureType            `json:"failureType"`
	ErrorMessage  string                 `json:"errorMessage"`
	StackTrace    string                 `json:"stackTrace,omitempty"`
	Context       map[string]interface{} `json:"context"`
	Input         interface{}            `json:"input"`
	AttemptNumber int                    `json:"attemptNumber"`
	Timestamp     time.Time              `json:"timestamp"`
}

// AdaptationStrategy 自适应策略
type AdaptationStrategy struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	ApplicableTypes []FailureType       `json:"applicableTypes"`
	Conditions      []StrategyCondition `json:"conditions"`
	Actions         []StrategyAction    `json:"actions"`
	Priority        int                 `json:"priority"`
	Enabled         bool                `json:"enabled"`
	SuccessRate     float64             `json:"successRate"`
	UsageCount      int                 `json:"usageCount"`
}

// StrategyCondition 策略条件
type StrategyCondition struct {
	Type     string      `json:"type"`     // field, pattern, metric
	Field    string      `json:"field"`    // 字段名
	Operator string      `json:"operator"` // eq, ne, gt, lt, contains, regex
	Value    interface{} `json:"value"`
}

// StrategyAction 策略动作
type StrategyAction struct {
	Type       string                 `json:"type"`       // retry, adjust, switch, skip, notify
	Target     string                 `json:"target"`     // 目标对象
	Parameters map[string]interface{} `json:"parameters"` // 动作参数
}

// AdaptiveEngine 自适应引擎
type AdaptiveEngine struct {
	failureHistory []FailureRecord
	strategies     []AdaptationStrategy
	learningModel  *LearningModel
	mutex          sync.RWMutex
	config         *AdaptiveConfig
}

// AdaptiveConfig 自适应配置
type AdaptiveConfig struct {
	MaxHistorySize         int           `json:"maxHistorySize"`
	LearningRate           float64       `json:"learningRate"`
	StrategyUpdateInterval time.Duration `json:"strategyUpdateInterval"`
	MinConfidence          float64       `json:"minConfidence"`
}

// LearningModel 学习模型
type LearningModel struct {
	FailurePatterns   map[string]PatternStats    `json:"failurePatterns"`
	StrategyOutcomes  map[string]StrategyOutcome `json:"strategyOutcomes"`
	ContextEmbeddings map[string][]float64       `json:"contextEmbeddings"`
}

// PatternStats 模式统计
type PatternStats struct {
	Pattern     string  `json:"pattern"`
	Count       int     `json:"count"`
	SuccessRate float64 `json:"successRate"`
	AvgTime     float64 `json:"avgTime"`
}

// StrategyOutcome 策略结果
type StrategyOutcome struct {
	StrategyID    string  `json:"strategyId"`
	SuccessCount  int     `json:"successCount"`
	FailureCount  int     `json:"failureCount"`
	AvgResolution float64 `json:"avgResolution"`
}

// NewAdaptiveEngine 创建自适应引擎
func NewAdaptiveEngine(config *AdaptiveConfig) *AdaptiveEngine {
	if config == nil {
		config = &AdaptiveConfig{
			MaxHistorySize:         1000,
			LearningRate:           0.1,
			StrategyUpdateInterval: time.Hour,
			MinConfidence:          0.7,
		}
	}

	return &AdaptiveEngine{
		failureHistory: make([]FailureRecord, 0),
		strategies:     make([]AdaptationStrategy, 0),
		learningModel: &LearningModel{
			FailurePatterns:   make(map[string]PatternStats),
			StrategyOutcomes:  make(map[string]StrategyOutcome),
			ContextEmbeddings: make(map[string][]float64),
		},
		config: config,
	}
}

// RecordFailure 记录失败
func (e *AdaptiveEngine) RecordFailure(record FailureRecord) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// 添加时间戳
	record.Timestamp = time.Now()

	// 添加到历史记录
	e.failureHistory = append(e.failureHistory, record)

	// 限制历史记录大小
	if len(e.failureHistory) > e.config.MaxHistorySize {
		e.failureHistory = e.failureHistory[1:]
	}

	// 更新学习模型
	e.updateLearningModel(record)

	logger.SugaredLogger.Infof("记录失败: %s, 类型: %s, 代理: %s",
		record.ID, record.FailureType, record.AgentName)

	return nil
}

// updateLearningModel 更新学习模型
func (e *AdaptiveEngine) updateLearningModel(record FailureRecord) {
	// 提取失败模式
	pattern := e.extractPattern(record)

	// 更新模式统计
	if stats, ok := e.learningModel.FailurePatterns[pattern]; ok {
		stats.Count++
		e.learningModel.FailurePatterns[pattern] = stats
	} else {
		e.learningModel.FailurePatterns[pattern] = PatternStats{
			Pattern: pattern,
			Count:   1,
		}
	}

	// 更新上下文嵌入 (简化版)
	contextKey := fmt.Sprintf("%s:%s", record.WorkflowID, record.StepID)
	e.learningModel.ContextEmbeddings[contextKey] = e.computeEmbedding(record)
}

// extractPattern 提取失败模式
func (e *AdaptiveEngine) extractPattern(record FailureRecord) string {
	// 基于失败类型和错误消息提取关键模式
	return fmt.Sprintf("%s:%s", record.FailureType, e.summarizeError(record.ErrorMessage))
}

// summarizeError 总结错误信息
func (e *AdaptiveEngine) summarizeError(errorMessage string) string {
	// 提取错误的关键特征
	if len(errorMessage) > 100 {
		return errorMessage[:100] + "..."
	}
	return errorMessage
}

// computeEmbedding 计算上下文嵌入 (简化版)
func (e *AdaptiveEngine) computeEmbedding(record FailureRecord) []float64 {
	// 简化的嵌入计算
	embedding := make([]float64, 10)

	// 基于失败类型编码
	switch record.FailureType {
	case FailureTypeTimeout:
		embedding[0] = 1.0
	case FailureTypeException:
		embedding[1] = 1.0
	case FailureTypeValidation:
		embedding[2] = 1.0
	}

	// 基于时间编码 (小时)
	hour := float64(record.Timestamp.Hour())
	embedding[3] = hour / 24.0

	return embedding
}

// RegisterStrategy 注册自适应策略
func (e *AdaptiveEngine) RegisterStrategy(strategy AdaptationStrategy) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// 验证策略
	if strategy.ID == "" {
		return fmt.Errorf("strategy ID is required")
	}

	// 添加到策略列表
	e.strategies = append(e.strategies, strategy)

	// 按优先级排序
	e.sortStrategies()

	logger.SugaredLogger.Infof("注册自适应策略: %s, 优先级: %d", strategy.Name, strategy.Priority)

	return nil
}

// sortStrategies 按优先级排序策略
func (e *AdaptiveEngine) sortStrategies() {
	// 简单的冒泡排序
	n := len(e.strategies)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if e.strategies[j].Priority < e.strategies[j+1].Priority {
				e.strategies[j], e.strategies[j+1] = e.strategies[j+1], e.strategies[j]
			}
		}
	}
}

// FindMatchingStrategy 查找匹配的自适应策略
func (e *AdaptiveEngine) FindMatchingStrategy(record FailureRecord) (*AdaptationStrategy, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	// 遍历所有策略，找到第一个匹配的
	for i := range e.strategies {
		strategy := &e.strategies[i]

		// 检查策略是否启用
		if !strategy.Enabled {
			continue
		}

		// 检查失败类型是否匹配
		if !e.matchesFailureType(strategy, record.FailureType) {
			continue
		}

		// 检查条件是否满足
		if !e.matchesConditions(strategy, record) {
			continue
		}

		// 找到匹配的策略
		return strategy, nil
	}

	return nil, fmt.Errorf("no matching strategy found for failure type: %s", record.FailureType)
}

// matchesFailureType 检查失败类型是否匹配
func (e *AdaptiveEngine) matchesFailureType(strategy *AdaptationStrategy, failureType FailureType) bool {
	// 如果没有指定适用类型，则认为适用所有类型
	if len(strategy.ApplicableTypes) == 0 {
		return true
	}

	// 检查失败类型是否在适用类型列表中
	for _, applicableType := range strategy.ApplicableTypes {
		if applicableType == failureType {
			return true
		}
	}

	return false
}

// matchesConditions 检查条件是否满足
func (e *AdaptiveEngine) matchesConditions(strategy *AdaptationStrategy, record FailureRecord) bool {
	// 如果没有条件，则认为满足
	if len(strategy.Conditions) == 0 {
		return true
	}

	// 检查所有条件（目前简化为AND关系）
	for _, condition := range strategy.Conditions {
		if !e.evaluateCondition(condition, record) {
			return false
		}
	}

	return true
}

// evaluateCondition 评估单个条件
func (e *AdaptiveEngine) evaluateCondition(condition StrategyCondition, record FailureRecord) bool {
	// 获取字段值
	var fieldValue interface{}
	switch condition.Field {
	case "failureType":
		fieldValue = string(record.FailureType)
	case "agentName":
		fieldValue = record.AgentName
	case "workflowId":
		fieldValue = record.WorkflowID
	case "stepId":
		fieldValue = record.StepID
	case "attemptNumber":
		fieldValue = record.AttemptNumber
	case "errorMessage":
		fieldValue = record.ErrorMessage
	default:
		// 尝试从上下文中获取
		if record.Context != nil {
			fieldValue = record.Context[condition.Field]
		}
	}

	// 比较值
	switch condition.Operator {
	case "eq":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", condition.Value)
	case "ne":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", condition.Value)
	case "gt":
		return compareValues(fieldValue, condition.Value) > 0
	case "lt":
		return compareValues(fieldValue, condition.Value) < 0
	case "contains":
		return contains(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", condition.Value))
	case "regex":
		return matchesRegex(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", condition.Value))
	default:
		return false
	}
}

// compareValues 比较两个值 (简化版)
func compareValues(a, b interface{}) int {
	// 尝试转换为float64进行比较
	var aFloat, bFloat float64

	switch v := a.(type) {
	case int:
		aFloat = float64(v)
	case int64:
		aFloat = float64(v)
	case float64:
		aFloat = v
	default:
		return 0
	}

	switch v := b.(type) {
	case int:
		bFloat = float64(v)
	case int64:
		bFloat = float64(v)
	case float64:
		bFloat = v
	default:
		return 0
	}

	if aFloat < bFloat {
		return -1
	} else if aFloat > bFloat {
		return 1
	}
	return 0
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0)))
}

// containsAt 检查在指定位置是否包含子串
func containsAt(s, substr string, start int) bool {
	if start+len(substr) > len(s) {
		return false
	}
	return s[start:start+len(substr)] == substr
}

// matchesRegex 检查字符串是否匹配正则表达式 (简化版，仅支持简单的通配符)
func matchesRegex(s, pattern string) bool {
	// 简化实现：支持 * 和 ? 通配符
	// 实际应用中应该使用标准库的正则表达式
	if pattern == "*" {
		return true
	}
	if pattern == s {
		return true
	}
	// 这里可以实现更复杂的通配符匹配逻辑
	return false
}

// ExecuteAdaptation 执行自适应调整
func (e *AdaptiveEngine) ExecuteAdaptation(record FailureRecord, strategy *AdaptationStrategy) (*AdaptationResult, error) {
	logger.SugaredLogger.Infof("执行自适应调整: %s", strategy.Name)

	result := &AdaptationResult{
		OriginalRecord: record,
		StrategyUsed:   strategy,
		Actions:        make([]ExecutedAction, 0),
		Timestamp:      time.Now(),
	}

	// 执行策略中的每个动作
	for _, action := range strategy.Actions {
		executedAction, err := e.executeAction(action, record, result)
		if err != nil {
			logger.SugaredLogger.Errorf("执行动作失败: %v", err)
			result.Success = false
			result.Error = err
			return result, err
		}
		result.Actions = append(result.Actions, *executedAction)
	}

	// 更新策略成功率
	e.updateStrategySuccessRate(strategy, true)

	result.Success = true
	logger.SugaredLogger.Infof("自适应调整执行成功: %s", strategy.Name)

	return result, nil
}

// AdaptationResult 自适应结果
type AdaptationResult struct {
	ID             string              `json:"id"`
	OriginalRecord FailureRecord       `json:"originalRecord"`
	StrategyUsed   *AdaptationStrategy `json:"strategyUsed"`
	Actions        []ExecutedAction    `json:"actions"`
	Success        bool                `json:"success"`
	Error          error               `json:"error,omitempty"`
	Timestamp      time.Time           `json:"timestamp"`
}

// ExecutedAction 执行的动作
type ExecutedAction struct {
	Type       string                 `json:"type"`
	Target     string                 `json:"target"`
	Parameters map[string]interface{} `json:"parameters"`
	Result     interface{}            `json:"result"`
	Success    bool                   `json:"success"`
	Error      string                 `json:"error,omitempty"`
	Duration   int64                  `json:"duration"` // 毫秒
}

// executeAction 执行单个动作
func (e *AdaptiveEngine) executeAction(action StrategyAction, record FailureRecord, result *AdaptationResult) (*ExecutedAction, error) {
	executed := &ExecutedAction{
		Type:       action.Type,
		Target:     action.Target,
		Parameters: action.Parameters,
		Success:    false,
	}

	startTime := time.Now()

	// 根据动作类型执行不同的逻辑
	switch action.Type {
	case "retry":
		err := e.executeRetry(action, record)
		if err != nil {
			executed.Error = err.Error()
			return executed, err
		}
		executed.Result = map[string]interface{}{"retried": true}

	case "adjust":
		err := e.executeAdjust(action, record)
		if err != nil {
			executed.Error = err.Error()
			return executed, err
		}
		executed.Result = map[string]interface{}{"adjusted": true}

	case "switch":
		err := e.executeSwitch(action, record)
		if err != nil {
			executed.Error = err.Error()
			return executed, err
		}
		executed.Result = map[string]interface{}{"switched": true}

	case "skip":
		// 跳过当前步骤
		executed.Result = map[string]interface{}{"skipped": true}

	case "notify":
		err := e.executeNotify(action, record)
		if err != nil {
			executed.Error = err.Error()
			return executed, err
		}
		executed.Result = map[string]interface{}{"notified": true}

	default:
		return executed, fmt.Errorf("unknown action type: %s", action.Type)
	}

	executed.Success = true
	executed.Duration = time.Since(startTime).Milliseconds()

	return executed, nil
}

// 各种动作的执行方法

func (e *AdaptiveEngine) executeRetry(action StrategyAction, record FailureRecord) error {
	// 实现重试逻辑
	maxRetries, ok := action.Parameters["maxRetries"].(int)
	if !ok {
		maxRetries = 3
	}

	retryDelay, ok := action.Parameters["retryDelay"].(int)
	if !ok {
		retryDelay = 1000
	}

	logger.SugaredLogger.Infof("执行重试动作: maxRetries=%d, retryDelay=%d", maxRetries, retryDelay)

	// 这里可以实现实际的重试逻辑
	return nil
}

func (e *AdaptiveEngine) executeAdjust(action StrategyAction, record FailureRecord) error {
	// 实现调整逻辑
	adjustmentType, ok := action.Parameters["type"].(string)
	if !ok {
		adjustmentType = "general"
	}

	logger.SugaredLogger.Infof("执行调整动作: type=%s", adjustmentType)

	// 根据调整类型执行不同的调整
	switch adjustmentType {
	case "timeout":
		// 调整超时时间
		newTimeout := action.Parameters["timeout"].(int)
		logger.SugaredLogger.Infof("调整超时时间为: %d", newTimeout)

	case "resources":
		// 调整资源分配
		resourceType := action.Parameters["resourceType"].(string)
		resourceValue := action.Parameters["resourceValue"].(int)
		logger.SugaredLogger.Infof("调整资源 %s 为: %d", resourceType, resourceValue)
	}

	return nil
}

func (e *AdaptiveEngine) executeSwitch(action StrategyAction, record FailureRecord) error {
	// 实现切换逻辑
	target, ok := action.Parameters["target"].(string)
	if !ok {
		return fmt.Errorf("switch action requires 'target' parameter")
	}

	reason, _ := action.Parameters["reason"].(string)

	logger.SugaredLogger.Infof("执行切换动作: target=%s, reason=%s", target, reason)

	// 这里可以实现实际的切换逻辑，比如切换到备用代理、备用资源等

	return nil
}

func (e *AdaptiveEngine) executeNotify(action StrategyAction, record FailureRecord) error {
	// 实现通知逻辑
	channel, ok := action.Parameters["channel"].(string)
	if !ok {
		channel = "log" // 默认使用日志
	}

	message, _ := action.Parameters["message"].(string)
	if message == "" {
		message = fmt.Sprintf("Failure detected: %s", record.FailureType)
	}

	logger.SugaredLogger.Infof("执行通知动作: channel=%s, message=%s", channel, message)

	// 根据通知渠道执行不同的通知逻辑
	switch channel {
	case "log":
		logger.SugaredLogger.Warnf("[NOTIFICATION] %s - Workflow: %s, Step: %s, Error: %s",
			message, record.WorkflowID, record.StepID, record.ErrorMessage)

	case "webhook":
		webhookURL, ok := action.Parameters["webhookUrl"].(string)
		if ok && webhookURL != "" {
			// 发送webhook通知
			e.sendWebhookNotification(webhookURL, record, message)
		}

	case "email":
		// 发送邮件通知
		// 这里可以集成邮件服务

	case "dingtalk":
		// 发送钉钉通知
		// 这里可以集成钉钉机器人
	}

	return nil
}

// sendWebhookNotification 发送webhook通知
func (e *AdaptiveEngine) sendWebhookNotification(webhookURL string, record FailureRecord, message string) {
	// 构建webhook payload
	payload := map[string]interface{}{
		"message":      message,
		"failureType":  record.FailureType,
		"workflowId":   record.WorkflowID,
		"stepId":       record.StepID,
		"agentName":    record.AgentName,
		"errorMessage": record.ErrorMessage,
		"timestamp":    record.Timestamp.Format(time.RFC3339),
	}

	// 发送HTTP POST请求
	jsonData, _ := json.Marshal(payload)

	logger.SugaredLogger.Infof("发送webhook通知到: %s, payload: %s", webhookURL, string(jsonData))

	// 这里可以实现实际的HTTP发送逻辑
	// 可以使用 http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
}

// updateStrategySuccessRate 更新策略成功率
func (e *AdaptiveEngine) updateStrategySuccessRate(strategy *AdaptationStrategy, success bool) {
	// 更新策略使用次数
	strategy.UsageCount++

	// 获取或创建策略结果记录
	outcome, ok := e.learningModel.StrategyOutcomes[strategy.ID]
	if !ok {
		outcome = StrategyOutcome{
			StrategyID: strategy.ID,
		}
	}

	// 更新成功/失败计数
	if success {
		outcome.SuccessCount++
	} else {
		outcome.FailureCount++
	}

	// 计算成功率
	total := outcome.SuccessCount + outcome.FailureCount
	if total > 0 {
		strategy.SuccessRate = float64(outcome.SuccessCount) / float64(total)
	}

	// 保存策略结果
	e.learningModel.StrategyOutcomes[strategy.ID] = outcome

	logger.SugaredLogger.Infof("更新策略成功率: %s, 成功率: %.2f%%, 使用次数: %d",
		strategy.Name, strategy.SuccessRate*100, strategy.UsageCount)
}

// GetLearningStatistics 获取学习统计
func (e *AdaptiveEngine) GetLearningStatistics() map[string]interface{} {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	stats := map[string]interface{}{
		"totalFailures":    len(e.failureHistory),
		"totalStrategies":  len(e.strategies),
		"failurePatterns":  len(e.learningModel.FailurePatterns),
		"strategyOutcomes": len(e.learningModel.StrategyOutcomes),
	}

	// 计算整体成功率
	var totalSuccess, totalUsage int
	for _, strategy := range e.strategies {
		totalSuccess += int(strategy.SuccessRate * float64(strategy.UsageCount))
		totalUsage += strategy.UsageCount
	}

	if totalUsage > 0 {
		stats["overallSuccessRate"] = float64(totalSuccess) / float64(totalUsage)
	}

	return stats
}

// ExportLearningModel 导出学习模型
func (e *AdaptiveEngine) ExportLearningModel() ([]byte, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	model := map[string]interface{}{
		"failurePatterns":   e.learningModel.FailurePatterns,
		"strategyOutcomes":  e.learningModel.StrategyOutcomes,
		"contextEmbeddings": e.learningModel.ContextEmbeddings,
		"exportTime":        time.Now().Format(time.RFC3339),
	}

	return json.MarshalIndent(model, "", "  ")
}

// ImportLearningModel 导入学习模型
func (e *AdaptiveEngine) ImportLearningModel(data []byte) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	var model struct {
		FailurePatterns   map[string]PatternStats    `json:"failurePatterns"`
		StrategyOutcomes  map[string]StrategyOutcome `json:"strategyOutcomes"`
		ContextEmbeddings map[string][]float64       `json:"contextEmbeddings"`
	}

	if err := json.Unmarshal(data, &model); err != nil {
		return fmt.Errorf("failed to unmarshal learning model: %w", err)
	}

	// 合并学习模型（不覆盖已有数据）
	for k, v := range model.FailurePatterns {
		if _, exists := e.learningModel.FailurePatterns[k]; !exists {
			e.learningModel.FailurePatterns[k] = v
		}
	}

	for k, v := range model.StrategyOutcomes {
		if _, exists := e.learningModel.StrategyOutcomes[k]; !exists {
			e.learningModel.StrategyOutcomes[k] = v
		}
	}

	for k, v := range model.ContextEmbeddings {
		if _, exists := e.learningModel.ContextEmbeddings[k]; !exists {
			e.learningModel.ContextEmbeddings[k] = v
		}
	}

	logger.SugaredLogger.Info("成功导入学习模型")
	return nil
}

// ClearHistory 清除历史记录
func (e *AdaptiveEngine) ClearHistory() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.failureHistory = make([]FailureRecord, 0)
	logger.SugaredLogger.Info("已清除失败历史记录")
}

// ResetLearningModel 重置学习模型
func (e *AdaptiveEngine) ResetLearningModel() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.learningModel = &LearningModel{
		FailurePatterns:   make(map[string]PatternStats),
		StrategyOutcomes:  make(map[string]StrategyOutcome),
		ContextEmbeddings: make(map[string][]float64),
	}
	logger.SugaredLogger.Info("已重置学习模型")
}

// init 初始化（如果有需要在包初始化时执行的逻辑）
func init() {
	// 这里可以添加包初始化逻辑
	logger.SugaredLogger.Info("adaptive workflow package initialized")
}
