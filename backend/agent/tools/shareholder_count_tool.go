package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"go-stock/backend/data"
)

// @Author StockAgent
// @Date 2026/3/3
// @Desc 股东人数查询工具
//-----------------------------------------------------------------------------------

func GetQueryShareholderCountTool() tool.InvokableTool {
	return &ToolQueryShareholderCount{}
}

type ToolQueryShareholderCount struct{}

func (t ToolQueryShareholderCount) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "QueryShareholderCount",
		Desc: "查询股票的股东人数变化和筹码集中度，是判断主力吸筹或出货的重要指标。股东人数减少通常表示筹码集中（主力吸筹），股东人数增加表示筹码分散（主力出货）。对判断股票主力动向和支撑压力位有重要参考价值。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"stockCode": {
				Type:     "string",
				Desc:     "股票代码，支持格式：sh600000、600000、000001。例如：平安银行(sz000001)、贵州茅台(sh600519)",
				Required: true,
			},
			"quarters": {
				Type:     "integer",
				Desc:     "查询最近多少个季度的数据，默认4个季度。建议取值范围：2-8个季度",
				Required: false,
			},
		}),
	}, nil
}

func (t ToolQueryShareholderCount) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	parms := map[string]any{}
	err := json.Unmarshal([]byte(argumentsInJSON), &parms)
	if err != nil {
		return "", err
	}

	stockCode := GetStockCode(parms["stockCode"].(string))

	quarters := 4
	if q, ok := parms["quarters"]; ok && q != nil {
		if qFloat, ok := q.(float64); ok {
			quarters = int(qFloat)
		}
	}

	// 调用数据API获取股东人数数据
	tushareApi := data.NewTushareApi(data.GetSettingConfig())
	shareholderData, err := tushareApi.GetShareholderCount(stockCode, quarters)
	if err != nil {
		// 如果tushare获取失败，尝试从其他数据源获取
		shareholderData, err = t.getShareholderDataFromAlternativeSource(stockCode, quarters)
		if err != nil {
			return "", fmt.Errorf("获取股东人数数据失败: %v", err)
		}
	}

	// 分析筹码集中度变化趋势
	analysis := t.analyzeShareholderTrend(shareholderData)

	// 构建返回结果
	result := map[string]interface{}{
		"stock_code":        stockCode,
		"data":              shareholderData,
		"trend_analysis":    analysis,
		"quarters_analyzed": quarters,
	}

	marshal, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

// getShareholderDataFromAlternativeSource 从备用数据源获取股东人数
func (t ToolQueryShareholderCount) getShareholderDataFromAlternativeSource(stockCode string, quarters int) (*data.ShareholderCountData, error) {
	// TODO: 实现从其他数据源获取股东人数的逻辑
	// 可以考虑的数据源：
	// 1. 东方财富
	// 2. 同花顺
	// 3. 雪球
	// 4. 新浪财经
	return nil, fmt.Errorf("备用数据源暂未实现")
}

// analyzeShareholderTrend 分析股东人数变化趋势
func (t ToolQueryShareholderCount) analyzeShareholderTrend(data *data.ShareholderCountData) map[string]interface{} {
	analysis := map[string]interface{}{
		"trend_direction":     "unknown",
		"concentration_level": "unknown",
		"signal_strength":     "neutral",
		"interpretation":      "",
	}

	if data == nil || len(data.QuarterlyData) < 2 {
		analysis["interpretation"] = "数据不足，无法分析趋势"
		return analysis
	}

	// 计算最新季度与上季度的变化
	latest := data.QuarterlyData[0]
	previous := data.QuarterlyData[1]

	changePercent := 0.0
	if previous.Count > 0 {
		changePercent = float64(latest.Count-previous.Count) / float64(previous.Count) * 100
	}

	// 判断趋势
	if changePercent < -5 {
		analysis["trend_direction"] = "decreasing"
		analysis["concentration_level"] = "increasing"
		analysis["signal_strength"] = "bullish"
		analysis["interpretation"] = fmt.Sprintf("股东人数减少%.2f%%，筹码趋于集中，可能为主力吸筹信号， bullish", -changePercent)
	} else if changePercent > 5 {
		analysis["trend_direction"] = "increasing"
		analysis["concentration_level"] = "decreasing"
		analysis["signal_strength"] = "bearish"
		analysis["interpretation"] = fmt.Sprintf("股东人数增加%.2f%%，筹码趋于分散，可能为主力出货信号， bearish", changePercent)
	} else {
		analysis["trend_direction"] = "stable"
		analysis["concentration_level"] = "stable"
		analysis["signal_strength"] = "neutral"
		analysis["interpretation"] = fmt.Sprintf("股东人数变化%.2f%%，筹码分布相对稳定", changePercent)
	}

	// 计算多季度趋势
	if len(data.QuarterlyData) >= 4 {
		totalChange := 0.0
		for i := 0; i < len(data.QuarterlyData)-1; i++ {
			if data.QuarterlyData[i+1].Count > 0 {
				change := float64(data.QuarterlyData[i].Count-data.QuarterlyData[i+1].Count) / float64(data.QuarterlyData[i+1].Count) * 100
				totalChange += change
			}
		}
		avgChange := totalChange / float64(len(data.QuarterlyData)-1)
		analysis["multi_quarter_trend"] = map[string]interface{}{
			"quarters_analyzed": len(data.QuarterlyData),
			"average_change":    avgChange,
			"trend_description": t.describeMultiQuarterTrend(avgChange),
		}
	}

	return analysis
}

// describeMultiQuarterTrend 描述多季度趋势
func (t ToolQueryShareholderCount) describeMultiQuarterTrend(avgChange float64) string {
	if avgChange < -3 {
		return "多个季度持续集中，主力深度吸筹，关注突破机会"
	} else if avgChange < -1 {
		return "筹码逐步集中，可能有资金在悄悄布局"
	} else if avgChange < 1 {
		return "筹码分布相对稳定，观望为主"
	} else if avgChange < 3 {
		return "筹码有所分散，注意主力出货风险"
	} else {
		return "筹码持续分散，主力明显出货，谨慎对待"
	}
}

// ShareholderCountData 定义在 data 包中
type ShareholderCountData struct {
	StockCode     string
	StockName     string
	QuarterlyData []QuarterlyShareholderInfo
}

type QuarterlyShareholderInfo struct {
	Quarter   string  // 季度，如 "2025Q3"
	Count     int     // 股东人数
	AvgShares float64 // 人均持股
	Change    int     // 较上季度变化
	ChangePct float64 // 变化百分比
}
