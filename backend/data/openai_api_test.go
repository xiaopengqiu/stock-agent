package data

import (
	"context"
	"go-stock/backend/db"
	"testing"
)

func TestNewDeepSeekOpenAiConfig(t *testing.T) {
	db.Init("../../data/stock.db")
	ai := NewDeepSeekOpenAi(context.TODO())
	res := ai.NewChatStream("北京文化", "sz000802", "")
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

}
