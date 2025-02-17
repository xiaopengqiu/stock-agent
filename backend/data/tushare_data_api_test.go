package data

import (
	"go-stock/backend/db"
	"testing"
)

// @Author spark
// @Date 2025/2/17 12:44
// @Desc
// -----------------------------------------------------------------------------------
func TestGetDaily(t *testing.T) {
	db.Init("../../data/stock.db")
	tushareApi := NewTushareApi(getConfig())
	res := tushareApi.GetDaily("000802.SZ", "20250101", "20250217", 30)
	t.Log(res)

}
