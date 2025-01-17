package data

import (
	"testing"
)

func TestNewDeepSeekOpenAiConfig(t *testing.T) {
	ai := NewDeepSeekOpenAi()
	ai.NewChat("闻泰科技")

}
