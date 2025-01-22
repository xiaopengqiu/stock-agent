package data

import (
	"go-stock/backend/db"
	"testing"
)

func TestNewDeepSeekOpenAiConfig(t *testing.T) {
	db.Init("../../data/stock.db")
	ai := NewDeepSeekOpenAi()
	res := ai.NewChatStream("闻泰科技", "sh600745")
	for {
		select {
		case msg := <-res:
			if msg == "" {
				return
			}
			t.Log(msg)
		}
	}
}
