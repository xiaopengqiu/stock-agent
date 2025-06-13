package main

import (
	"go-stock/backend/data"
	"go-stock/backend/models"
)

// @Author spark
// @Date 2025/6/8 20:45
// @Desc
//-----------------------------------------------------------------------------------

func (a *App) LongTigerRank(date string) *[]models.LongTigerRankData {
	return data.NewMarketNewsApi().LongTiger(date)
}

func (a *App) StockResearchReport() []any {
	return data.NewMarketNewsApi().StockResearchReport(7)
}
