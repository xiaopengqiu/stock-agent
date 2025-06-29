package data

import (
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"testing"
)

func TestSearchStock(t *testing.T) {
	db.Init("../../data/stock.db")

	res := NewSearchStockApi("换手率连续5日大于5%;科技行业").SearchStock()
	data := res["data"].(map[string]any)
	result := data["result"].(map[string]any)
	dataList := result["dataList"].([]any)
	for _, v := range dataList {
		logger.SugaredLogger.Infof("v:%+v", v)
	}
	columns := result["columns"].([]any)
	for _, v := range columns {
		logger.SugaredLogger.Infof("v:%+v", v)
	}

}
