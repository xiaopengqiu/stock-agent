package data

import (
	"go-stock/backend/db"
	"testing"
)

func TestNewDeepSeekOpenAiConfig(t *testing.T) {
	db.Init("../../data/stock.db")
	ai := NewDeepSeekOpenAi()
	res := ai.NewChatStream("北京文化", "sz000802")
	for {
		select {
		case msg := <-res:
			if msg == "" {
				continue
			}
			t.Log(msg)
		}
	}
}
