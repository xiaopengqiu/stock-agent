package data

import (
	"go-stock/backend/models"
	"testing"
	"time"
)

func TestPositionService_AddPosition(t *testing.T) {
	service := NewPositionService()

	pos := &models.Position{
		StockCode:    "000001.SZ",
		StockName:    "平安银行",
		Quantity:     100,
		BuyPrice:     10.50,
		BuyDate:      time.Now(),
		CurrentPrice: 11.20,
		Notes:        "测试持仓",
		IsActive:     true,
	}

	err := service.AddPosition(pos)
	if err != nil {
		t.Errorf("AddPosition failed: %v", err)
	}

	if pos.ID == 0 {
		t.Error("Expected position ID to be set")
	}

	if pos.MarketValue != 1120.0 {
		t.Errorf("Expected MarketValue 1120, got %v", pos.MarketValue)
	}

	if pos.ProfitLoss != 70.0 {
		t.Errorf("Expected ProfitLoss 70, got %v", pos.ProfitLoss)
	}
}

func TestPositionService_GetPositions(t *testing.T) {
	service := NewPositionService()

	positions, err := service.GetPositions()
	if err != nil {
		t.Errorf("GetPositions failed: %v", err)
	}

	if len(positions) == 0 {
		t.Log("No positions found, this might be expected")
	}
}

func TestPositionService_GetPositionSummary(t *testing.T) {
	service := NewPositionService()

	totalMV, totalPL, count, err := service.GetPositionSummary()
	if err != nil {
		t.Errorf("GetPositionSummary failed: %v", err)
	}

	t.Logf("Summary: MV=%v, PL=%v, Count=%d", totalMV, totalPL, count)
}

func TestPosition_CalculateFields(t *testing.T) {
	pos := &models.Position{
		Quantity:     100,
		BuyPrice:     10.0,
		CurrentPrice: 12.0,
	}

	pos.CalculateFields()

	if pos.MarketValue != 1200.0 {
		t.Errorf("Expected MarketValue 1200, got %v", pos.MarketValue)
	}

	if pos.ProfitLoss != 200.0 {
		t.Errorf("Expected ProfitLoss 200, got %v", pos.ProfitLoss)
	}

	expectedPct := 20.0
	if pos.ProfitLossPct != expectedPct {
		t.Errorf("Expected ProfitLossPct %v, got %v", expectedPct, pos.ProfitLossPct)
	}
}
