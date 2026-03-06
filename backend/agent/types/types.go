package types

import (
	"time"
)

// StockAnalysisRequest 用户选股分析请求
type StockAnalysisRequest struct {
	StockCode   string `json:"stockCode"`   // 股票代码
	StockName   string `json:"stockName"`   // 股票名称
	Question    string `json:"question"`    // 用户问题
	RiskLevel   string `json:"riskLevel"`   // 风险偏好 (保守/稳健/激进)
	TimeHorizon string `json:"timeHorizon"` // 投资周期 (短线/中线/长线)
}

// TechnicalAnalysis 技术面分析结果
type TechnicalAnalysis struct {
	Trend      string                     `json:"trend"`      // 趋势判断 (上涨/下跌/震荡)
	Indicators map[string]IndicatorResult `json:"indicators"` // 技术指标结果
	Support    []float64                  `json:"support"`    // 支撑位
	Resistance []float64                  `json:"resistance"` // 阻力位
	Signal     string                     `json:"signal"`     // 技术信号 (看多/看空/中性)
	Confidence float64                    `json:"confidence"` // 置信度 (0-1)
	RawData    string                     `json:"rawData"`    // 原始数据（K线等）
}

// IndicatorResult 技术指标结果
type IndicatorResult struct {
	Name   string      `json:"name"`   // 指标名称
	Value  interface{} `json:"value"`  // 指标值
	Signal string      `json:"signal"` // 指标信号 (看多/看空/中性)
}

// FundamentalAnalysis 基本面分析结果
type FundamentalAnalysis struct {
	FinancialMetrics map[string]float64 `json:"financialMetrics"` // 财务指标
	Valuation        ValuationResult    `json:"valuation"`        // 估值结果
	Growth           GrowthAnalysis     `json:"growth"`           // 成长分析
	PeersComparison  []PeerComparison   `json:"peersComparison"`  // 同业对比
	OverallScore     float64            `json:"overallScore"`     // 基本面评分 (0-100)
	RawData          string             `json:"rawData"`          // 原始数据（财务报告等）
}

// ValuationResult 估值结果
type ValuationResult struct {
	PE         float64 `json:"pe"`         // 市盈率
	PB         float64 `json:"pb"`         // 市净率
	PS         float64 `json:"ps"`         // 市销率
	Valuation  string  `json:"valuation"`  // 估值判断 (低估/合理/高估)
	Confidence float64 `json:"confidence"` // 置信度
}

// GrowthAnalysis 成长分析
type GrowthAnalysis struct {
	RevenueGrowth float64 `json:"revenueGrowth"` // 营收增长率
	ProfitGrowth  float64 `json:"profitGrowth"`  // 利润增长率
	ROE           float64 `json:"roe"`           // 净资产收益率
	GrowthTrend   string  `json:"growthTrend"`   // 成长趋势 (加速/稳定/减速)
}

// PeerComparison 同业对比
type PeerComparison struct {
	StockCode string  `json:"stockCode"` // 股票代码
	StockName string  `json:"stockName"` // 股票名称
	PE        float64 `json:"pe"`        // 市盈率
	PB        float64 `json:"pb"`        // 市净率
	ROE       float64 `json:"roe"`       // 净资产收益率
	Rank      int     `json:"rank"`      // 排名
}

// MarketNewsAnalysis 市场消息分析结果
type MarketNewsAnalysis struct {
	RecentNews []NewsItem    `json:"recentNews"` // 近期新闻
	Sentiment  string        `json:"sentiment"`  // 整体情绪 (乐观/悲观/中性)
	KeyEvents  []EventImpact `json:"keyEvents"`  // 关键事件
	HotTopics  []string      `json:"hotTopics"`  // 热门话题
	RawData    string        `json:"rawData"`    // 原始数据（新闻等）
}

// NewsItem 新闻项
type NewsItem struct {
	Title     string    `json:"title"`     // 新闻标题
	Content   string    `json:"content"`   // 新闻内容
	Source    string    `json:"source"`    // 来源
	Time      time.Time `json:"time"`      // 发布时间
	Sentiment string    `json:"sentiment"` // 情感倾向 (正面/负面/中性)
}

// EventImpact 事件影响
type EventImpact struct {
	Event      string  `json:"event"`      // 事件描述
	Impact     string  `json:"impact"`     // 影响分析
	Severity   string  `json:"severity"`   // 严重程度 (高/中/低)
	Confidence float64 `json:"confidence"` // 置信度
}

// RiskAssessment 风险评估结果
type RiskAssessment struct {
	RiskLevel      string  `json:"riskLevel"`      // 风险等级 (低/中/高)
	PositionAdvice string  `json:"positionAdvice"` // 仓位建议
	StopLoss       float64 `json:"stopLoss"`       // 止损价
	TakeProfit     float64 `json:"takeProfit"`     // 止盈价
	MaxDrawdown    float64 `json:"maxDrawdown"`    // 预估最大回撤
}

// AnalysisResult 综合分析结果
type AnalysisResult struct {
	Summary     string               `json:"summary"`     // 综合总结
	Technical   *TechnicalAnalysis   `json:"technical"`   // 技术面分析
	Fundamental *FundamentalAnalysis `json:"fundamental"` // 基本面分析
	MarketNews  *MarketNewsAnalysis  `json:"marketNews"`  // 市场消息分析
	Risk        *RiskAssessment      `json:"risk"`        // 风险评估
	Report      string               `json:"report"`      // 完整报告
	Confidence  float64              `json:"confidence"`  // 置信度 (0-1)
}

// AgentTask Agent 任务定义
type AgentTask struct {
	ID         string      `json:"id"`         // 任务ID
	Type       string      `json:"type"`       // 任务类型 (technical/fundamental/news/risk/report)
	Status     string      `json:"status"`     // 状态 (pending/running/completed/failed)
	Request    interface{} `json:"request"`    // 任务请求
	Result     interface{} `json:"result"`     // 任务结果
	Error      string      `json:"error"`      // 错误信息
	CreatedAt  time.Time   `json:"createdAt"`  // 创建时间
	StartedAt  *time.Time  `json:"startedAt"`  // 开始时间
	FinishedAt *time.Time  `json:"finishedAt"` // 完成时间
}

// AgentConfig Agent 配置
type AgentConfig struct {
	Name         string  `json:"name"`         // Agent 名称
	Description  string  `json:"description"`  // 描述
	SystemPrompt string  `json:"systemPrompt"` // 系统提示词
	Temperature  float64 `json:"temperature"`  // 温度参数
	MaxTokens    int     `json:"maxTokens"`    // 最大 Token 数
}
