package main

import (
	"testing"
	"time"
)

// @Author spark
// @Date 2025/2/24 9:35
// @Desc
// -----------------------------------------------------------------------------------
func TestIsHKTradingTime(t *testing.T) {
	f := IsHKTradingTime(time.Now())
	t.Log(f)
}

func TestIsUSTradingTime(t *testing.T) {
	t.Log(IsUSTradingTime(time.Now()))
}
