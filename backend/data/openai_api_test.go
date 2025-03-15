package data

import (
	"context"
	"go-stock/backend/db"
	"testing"
)

func TestNewDeepSeekOpenAiConfig(t *testing.T) {
	db.Init("../../data/stock.db")
	ai := NewDeepSeekOpenAi(context.TODO())
	res := ai.NewChatStream("上海贝岭", "sh600171", "分析以上股票资金流入信息，找出适合买入的股票，给出具体操作建议")
	for {
		select {
		case msg := <-res:
			t.Log(msg)
		}
	}
}

func TestGetTopNewsList(t *testing.T) {
	GetTopNewsList(30)
}

func TestSearchGuShiTongStockInfo(t *testing.T) {
	SearchGuShiTongStockInfo("hk01810", 60)
	SearchGuShiTongStockInfo("sh600745", 60)
	SearchGuShiTongStockInfo("gb_goog", 60)

}
