package orchestrator

import (
	"context"
)

// TechnicalAnalyzer 技术分析接口
type TechnicalAnalyzer interface {
	Analyze(ctx context.Context, request StockAnalysisRequest) (*TechnicalAnalysis, error)
}

// FundamentalAnalyzer 基本面分析接口
type FundamentalAnalyzer interface {
	Analyze(ctx context.Context, request StockAnalysisRequest) (*FundamentalAnalysis, error)
}

// NewsAnalyzer 消息分析接口
type NewsAnalyzer interface {
	Analyze(ctx context.Context, request StockAnalysisRequest) (*MarketNewsAnalysis, error)
}

// RiskAssessor 风险评估接口
type RiskAssessor interface {
	Assess(ctx context.Context, request StockAnalysisRequest, partialResult *AnalysisResult) (*RiskAssessment, error)
}

// Reporter 报告生成接口
type Reporter interface {
	GenerateReport(ctx context.Context, request StockAnalysisRequest, result *AnalysisResult) (string, error)
}
