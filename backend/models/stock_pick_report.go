package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

// StockPickReport AI荐股报告
type StockPickReport struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// 用户需求
	UserQuery    string `gorm:"type:text;not null" json:"user_query"` // 用户输入的选股需求
	QuerySummary string `gorm:"type:text" json:"query_summary"`       // 需求摘要

	// 分析结果
	Result          string `gorm:"type:text" json:"result"`
	MarketAnalysis  string `gorm:"type:text" json:"market_analysis"` // 市场环境分析
	FilterLogic     string `gorm:"type:text" json:"filter_logic"`    // 筛选逻辑说明
	TotalScanned    int    `json:"total_scanned"`                    // 扫描股票总数
	CandidatesCount int    `json:"candidates_count"`                 // 候选股票数

	// 推荐股票列表
	Recommendations []RecommendationItem `gorm:"type:text;serializer:json" json:"recommendations"`

	// 使用的工具列表 (JSON)
	ToolsUsed string `gorm:"type:text" json:"tools_used"`

	// AI配置
	AIConfigID uint   `json:"ai_config_id"`
	AIModel    string `json:"ai_model"`

	// 状态
	Status string `gorm:"type:varchar(20);default:'completed'" json:"status"` // 'processing' | 'completed' | 'failed'
	Error  string `gorm:"type:text" json:"error"`
}

// RecommendationItem 推荐股票项
type RecommendationItem struct {
	Rank         int     `json:"rank"`          // 排名
	StockCode    string  `json:"stock_code"`    // 股票代码
	StockName    string  `json:"stock_name"`    // 股票名称
	CurrentPrice float64 `json:"current_price"` // 现价
	PriceChange  float64 `json:"price_change"`  // 涨跌幅
	Volume       int64   `json:"volume"`        // 成交量
	MarketValue  float64 `json:"market_value"`  // 市值

	// 分析内容
	TechnicalAnalysis   string  `json:"technical_analysis"`    // 技术面分析
	FundamentalAnalysis string  `json:"fundamental_analysis"`  // 基本面分析
	Reason              string  `json:"reason"`                // 推荐理由
	TargetPrice         float64 `json:"target_price"`          // 目标价位
	TargetChangePercent float64 `json:"target_change_percent"` // 目标涨幅
	RiskLevel           string  `json:"risk_level"`            // 风险等级: 'low' | 'medium' | 'high'
	RiskTips            string  `json:"risk_tips"`             // 风险提示
	Score               float64 `json:"score"`                 // 综合评分 (0-100)
	TradeSuggestion     string  `json:"trade_suggestion"`      // 买卖建议: '买入' | '卖出' | '观望'

	// 新增字段 - 买卖点建议
	RecommendedPrice float64   `gorm:"type:decimal(10,2)" json:"recommended_price"` // 推荐时价
	PreviousClose    float64   `gorm:"type:decimal(10,2)" json:"previous_close"`    // 昨收价
	BuyPriceRange    string    `gorm:"type:varchar(50)" json:"buy_price_range"`     // AI建议买入价区间
	StopLossPrice    float64   `gorm:"type:decimal(10,2)" json:"stop_loss_price"`   // AI建议止损价
	SectorConcept    string    `gorm:"type:varchar(200)" json:"sector_concept"`     // 板块概念
	Remarks          string    `gorm:"type:text" json:"remarks"`                    // 备注
	RecommendedAt    time.Time `gorm:"index" json:"recommended_at"`                 // 推荐时间

	// 关注状态
	IsFollowed bool `json:"is_followed"`

	// 技术指标
	MACD  string  `json:"macd"`  // MACD指标
	KDJ   string  `json:"kdj"`   // KDJ指标
	RSI   float64 `json:"rsi"`   // RSI指标
	Trend string  `json:"trend"` // 趋势: 'up' | 'down' | 'sideways'

	// 基本面指标
	PE            float64 `json:"pe"`             // 市盈率
	PB            float64 `json:"pb"`             // 市净率
	ROE           float64 `json:"roe"`            // 净资产收益率
	RevenueGrowth float64 `json:"revenue_growth"` // 营收增长率
	ProfitGrowth  float64 `json:"profit_growth"`  // 利润增长率
}

type RecommendationItems struct {
	Items []RecommendationItem
	Raw   string // 兼容旧 text
}

func (r RecommendationItems) Value() (driver.Value, error) {
	if len(r.Items) > 0 {
		return json.Marshal(r.Items)
	}
	return r.Raw, nil
}

func (r *RecommendationItems) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type for RecommendationItems")
	}

	str := string(bytes)

	// 尝试 JSON 解析
	if strings.HasPrefix(strings.TrimSpace(str), "[") ||
		strings.HasPrefix(strings.TrimSpace(str), "{") {

		var items []RecommendationItem
		if err := json.Unmarshal(bytes, &items); err == nil {
			r.Items = items
			return nil
		}
	}

	// 走到这里说明是旧 text 数据
	r.Raw = str
	return nil
}

// ToolUsage 工具使用记录
type ToolUsage struct {
	ToolName string `json:"tool_name"` // 工具名称
	Status   string `json:"status"`    // 'idle' | 'running' | 'success' | 'failed'
	CallTime string `json:"call_time"` // 调用时间
	Duration string `json:"duration"`  // 耗时
	Result   string `json:"result"`    // 工具调用结果摘要
}

// StockPickReportItem 荐股报告列表项（用于前端展示）
type StockPickReportItem struct {
	ID             uint      `json:"ID"`
	CreatedAt      time.Time `json:"CreatedAt"`
	QuerySummary   string    `json:"QuerySummary"`
	RecommendCount int       `json:"RecommendCount"` // 推荐股票数量
	Status         string    `json:"Status"`
}

// StockPickReportsResponse 荐股报告列表响应（用于Wails绑定）
type StockPickReportsResponse struct {
	Items []StockPickReportItem `json:"items"`
	Total int64                 `json:"total"`
}

func (StockPickReport) TableName() string {
	return "stock_pick_reports"
}
