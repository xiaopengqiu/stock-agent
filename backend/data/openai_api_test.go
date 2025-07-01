package data

import (
	"context"
	"go-stock/backend/db"
	"testing"
)

func TestNewDeepSeekOpenAiConfig(t *testing.T) {
	db.Init("../../data/stock.db")

	var tools []Tool
	tools = append(tools, Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        "SearchStockByIndicators",
			Description: "通过解析自然语言，形成选股指标或策略，返回符合指标或策略的股票列表",
			Parameters: FunctionParameters{
				Type: "object",
				Properties: map[string]any{
					"words": map[string]any{
						"type":        "string",
						"description": "选股指标或策略的自然语言",
					},
				},
				Required: []string{"words"},
			},
		},
	})

	ai := NewDeepSeekOpenAi(context.TODO())
	//res := ai.NewChatStream("长电科技", "sh600584", "长电科技分析和总结", nil)
	res := ai.NewSummaryStockNewsStreamWithTools("总结市场资讯，发掘潜力标的/行业/板块/概念，控制风险，最后按风险登记生成指标选股策略汇总表，每个策略中的指标分号分隔，写成一行", nil, tools)

	for {
		select {
		case msg := <-res:
			t.Log(msg)
		}
	}
}

func TestGetTopNewsList(t *testing.T) {
	news := GetTopNewsList(30)
	t.Log(news)
}

func TestSearchGuShiTongStockInfo(t *testing.T) {
	db.Init("../../data/stock.db")
	SearchGuShiTongStockInfo("hk01810", 60)
	SearchGuShiTongStockInfo("sh600745", 60)
	SearchGuShiTongStockInfo("gb_goog", 60)

}
