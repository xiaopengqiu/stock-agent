package models

import (
	"time"

	"gorm.io/gorm"
)

// Position 持仓记录
type Position struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// 股票信息
	StockCode    string    `gorm:"index;not null;size:20" json:"stock_code"`    // 股票代码
	StockName    string    `gorm:"size:100" json:"stock_name"`                    // 股票名称

	// 持仓信息
	Quantity     int       `gorm:"default:0" json:"quantity"`                   // 持股数量
	BuyPrice     float64   `gorm:"type:decimal(10,3)" json:"buy_price"`         // 买入价位
	BuyDate      time.Time `json:"buy_date"`                                    // 买入日期
	CurrentPrice float64   `gorm:"type:decimal(10,3)" json:"current_price"`     // 当前价格（实时更新）

	// 计算字段
	ProfitLoss   float64   `gorm:"type:decimal(12,2)" json:"profit_loss"`       // 盈亏金额
	ProfitLossPct float64  `gorm:"type:decimal(8,3)" json:"profit_loss_pct"`    // 盈亏比例（保留3位小数）
	MarketValue  float64   `gorm:"type:decimal(14,2)" json:"market_value"`      // 持仓市值

	// 备注
	Notes        string    `gorm:"type:text" json:"notes"`                      // 备注

	// 状态
	IsActive     bool      `gorm:"default:true;index" json:"is_active"`          // 是否活跃
}

// PositionAnalysis 持仓分析结果（AI生成）
type PositionAnalysis struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	PositionID   uint      `gorm:"index;not null" json:"position_id"`            // 关联持仓ID

	// AI分析结果
	OverallAdvice string    `gorm:"type:varchar(50)" json:"overall_advice"`        // 总体建议：持有/加仓/减仓/清仓
	Confidence    float64   `gorm:"type:decimal(5,2)" json:"confidence"`          // 建议置信度 (0-1)

	// 价格建议
	SuggestedBuyPrice  *float64 `gorm:"type:decimal(10,2)" json:"suggested_buy_price"`   // 补仓价位
	SuggestedSellPrice *float64 `gorm:"type:decimal(10,2)" json:"suggested_sell_price"`  // 止盈价位
	StopLossPrice      *float64 `gorm:"type:decimal(10,2)" json:"stop_loss_price"`       // 止损价位

	// 详细分析
	TechnicalAnalysis   string `gorm:"type:text" json:"technical_analysis"`     // 技术面分析
	FundamentalAnalysis string `gorm:"type:text" json:"fundamental_analysis"`   // 基本面分析
	RiskAnalysis        string `gorm:"type:text" json:"risk_analysis"`          // 风险分析

	// 原始AI响应
	RawResponse string `gorm:"type:text" json:"raw_response"`
}

// TableName 设置表名
func (Position) TableName() string {
	return "positions"
}

// TableName 设置表名
func (PositionAnalysis) TableName() string {
	return "position_analyses"
}

// PortfolioAnalysis 整体仓位分析结果（AI生成）
type PortfolioAnalysis struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// 整体分析
	OverallAssessment string `gorm:"type:text" json:"overall_assessment"`    // 整体评估

	// 仓位分布分析
	AllocationAnalysis string `gorm:"type:text" json:"allocation_analysis"`  // 仓位分布分析

	// 风险评估
	RiskAssessment string `gorm:"type:text" json:"risk_assessment"`          // 风险评估

	// 调整建议
	AdjustmentSuggestions string `gorm:"type:text" json:"adjustment_suggestions"`  // 具体调整建议

	// 原始AI响应
	RawResponse string `gorm:"type:text" json:"raw_response"`
}

// TableName 设置表名
func (PortfolioAnalysis) TableName() string {
	return "portfolio_analyses"
}

// CalculateFields 计算持仓的盈亏等字段
func (p *Position) CalculateFields() {
	if p.CurrentPrice > 0 && p.Quantity > 0 {
		// 持仓市值 = 当前价格 * 数量
		p.MarketValue = p.CurrentPrice * float64(p.Quantity)

		// 盈亏金额 = (当前价格 - 买入价格) * 数量
		if p.BuyPrice > 0 {
			p.ProfitLoss = (p.CurrentPrice - p.BuyPrice) * float64(p.Quantity)

			// 盈亏比例 = (当前价格 - 买入价格) / 买入价格 * 100
			p.ProfitLossPct = ((p.CurrentPrice - p.BuyPrice) / p.BuyPrice) * 100
		}
	}
}
