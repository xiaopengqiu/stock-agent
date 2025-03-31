package data

import (
	"context"
	"go-stock/backend/db"
	"testing"
)

func TestNewDeepSeekOpenAiConfig(t *testing.T) {
	db.Init("../../data/stock.db")
	ai := NewDeepSeekOpenAi(context.TODO())
	res := ai.NewChatStream("上海贝岭", "sh600171", "上海贝岭分析和总结", nil)
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
	//db.Init("../../data/stock.db")
	SearchGuShiTongStockInfo("hk01810", 60)
	SearchGuShiTongStockInfo("sh600745", 60)
	SearchGuShiTongStockInfo("gb_goog", 60)

}
