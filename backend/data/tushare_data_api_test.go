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
	tushareApi := NewTushareApi(GetSettingConfig())
	res := tushareApi.GetDaily("00927.HK", "20250101", "20250217", 30)
	t.Log(res)

}

func TestGetUSDaily(t *testing.T) {
	db.Init("../../data/stock.db")
	tushareApi := NewTushareApi(GetSettingConfig())

	res := tushareApi.GetDaily("gb_AAPL", "20250101", "20250217", 30)
	t.Log(res)

	//

}
