package data

import (
	"fmt"
	"go-stock/backend/db"
	"go-stock/backend/models"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type PositionService struct {
	dao *gorm.DB
}

func NewPositionService() *PositionService {
	return &PositionService{dao: db.Dao}
}

// AddPosition 添加新持仓
func (s *PositionService) AddPosition(pos *models.Position) error {
	pos.CalculateFields()
	return s.dao.Create(pos).Error
}

// UpdatePosition 更新持仓
func (s *PositionService) UpdatePosition(id uint, pos *models.Position) error {
	// 先获取现有持仓记录，保留 CurrentPrice 等字段
	var existingPos models.Position
	if err := s.dao.First(&existingPos, id).Error; err != nil {
		return err
	}

	// 更新允许修改的字段
	if pos.StockCode != "" {
		existingPos.StockCode = pos.StockCode
	}
	if pos.StockName != "" {
		existingPos.StockName = pos.StockName
	}
	if pos.Quantity > 0 {
		existingPos.Quantity = pos.Quantity
	}
	if pos.BuyPrice > 0 {
		existingPos.BuyPrice = pos.BuyPrice
	}
	if !pos.BuyDate.IsZero() {
		existingPos.BuyDate = pos.BuyDate
	}
	existingPos.Notes = pos.Notes

	// 重新计算字段（使用保留的 CurrentPrice）
	existingPos.CalculateFields()

	// 保存更新
	return s.dao.Save(&existingPos).Error
}

// DeletePosition 删除持仓（软删除）
func (s *PositionService) DeletePosition(id uint) error {
	return s.dao.Delete(&models.Position{}, id).Error
}

// GetPositions 获取所有活跃持仓
func (s *PositionService) GetPositions() ([]*models.Position, error) {
	var positions []*models.Position
	err := s.dao.Where("is_active = ?", true).Order("created_at DESC").Find(&positions).Error
	return positions, err
}

// GetPositionByID 根据ID获取持仓
func (s *PositionService) GetPositionByID(id uint) (*models.Position, error) {
	var pos models.Position
	err := s.dao.First(&pos, id).Error
	if err != nil {
		return nil, err
	}
	return &pos, nil
}

// RefreshPositions 更新所有持仓的当前价格和计算字段
func (s *PositionService) RefreshPositions() error {
	positions, err := s.GetPositions()
	if err != nil {
		return err
	}

	if len(positions) == 0 {
		return nil
	}

	// 收集所有股票代码
	codes := make([]string, 0, len(positions))
	for _, pos := range positions {
		codes = append(codes, pos.StockCode)
	}

	// 批量获取实时价格
	stockApi := NewStockDataApi()
	stockInfos, err := stockApi.GetStockCodeRealTimeData(codes...)
	if err != nil {
		return err
	}

	// 创建价格映射
	priceMap := make(map[string]float64)
	if stockInfos != nil {
		for _, info := range *stockInfos {
			if price, err := strconv.ParseFloat(info.Price, 64); err == nil {
				priceMap[info.Code] = price
			}
		}
	}

	// 更新每个持仓
	for _, pos := range positions {
		if price, ok := priceMap[pos.StockCode]; ok {
			pos.CurrentPrice = price
			pos.CalculateFields()
			s.dao.Save(pos)
		}
	}

	return nil
}

// GetPositionSummary 获取持仓汇总统计
func (s *PositionService) GetPositionSummary() (totalMarketValue, totalProfitLoss float64, positionCount int, err error) {
	positions, err := s.GetPositions()
	if err != nil {
		return 0, 0, 0, err
	}

	positionCount = len(positions)
	for _, pos := range positions {
		totalMarketValue += pos.MarketValue
		totalProfitLoss += pos.ProfitLoss
	}

	return totalMarketValue, totalProfitLoss, positionCount, nil
}

// AddPositionAnalysis 添加持仓分析结果
func (s *PositionService) AddPositionAnalysis(analysis *models.PositionAnalysis) error {
	return s.dao.Create(analysis).Error
}

// GetPositionAnalysis 获取持仓的最新分析结果
func (s *PositionService) GetPositionAnalysis(positionID uint) (*models.PositionAnalysis, error) {
	var analysis models.PositionAnalysis
	err := s.dao.Where("position_id = ?", positionID).Order("created_at DESC").First(&analysis).Error
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}

// GetAllPositionAnalyses 获取持仓的所有分析历史
func (s *PositionService) GetAllPositionAnalyses(positionID uint) ([]*models.PositionAnalysis, error) {
	var analyses []*models.PositionAnalysis
	err := s.dao.Where("position_id = ?", positionID).Order("created_at DESC").Find(&analyses).Error
	return analyses, err
}

// AddFromRecommendation 从荐股报告添加持仓
func (s *PositionService) AddFromRecommendation(stockCode, stockName string, buyPrice float64, quantity int, notes string) (*models.Position, error) {
	pos := &models.Position{
		StockCode:    stockCode,
		StockName:    stockName,
		Quantity:     quantity,
		BuyPrice:     buyPrice,
		BuyDate:      time.Now(),
		CurrentPrice: buyPrice, // 初始时现价等于买入价
		Notes:        notes,
		IsActive:     true,
	}
	pos.CalculateFields()

	err := s.AddPosition(pos)
	if err != nil {
		return nil, err
	}
	return pos, nil
}

// SyncFollowedToPositions 从自选股同步持仓（将设置了成本价和数量的自选股转入持仓）
func (s *PositionService) SyncFollowedToPositions() (string, error) {
	// 获取所有自选股
	var followedStocks []FollowedStock
	err := s.dao.Model(&FollowedStock{}).Where("is_del = ?", 0).Find(&followedStocks).Error
	if err != nil {
		return "", err
	}

	syncedCount := 0
	for _, fs := range followedStocks {
		// 只同步设置了成本价和数量的自选股
		if fs.CostPrice <= 0 || fs.Volume <= 0 {
			continue
		}

		// 检查该股票是否已在持仓中
		var existingCount int64
		s.dao.Model(&models.Position{}).Where("stock_code = ? AND is_active = ?", fs.StockCode, true).Count(&existingCount)
		if existingCount > 0 {
			continue // 已存在则跳过
		}

		// 创建持仓记录
		pos := &models.Position{
			StockCode:    fs.StockCode,
			StockName:    fs.Name,
			Quantity:     int(fs.Volume),
			BuyPrice:     fs.CostPrice,
			BuyDate:      time.Now(),
			CurrentPrice: fs.Price,
			Notes:        "从自选股同步",
			IsActive:     true,
		}
		pos.CalculateFields()

		if err := s.AddPosition(pos); err != nil {
			continue
		}
		syncedCount++
	}

	if syncedCount == 0 {
		return "没有可同步的持仓，请在自选股中先设置成本价和持股数量", nil
	}
	return fmt.Sprintf("成功同步 %d 只持仓", syncedCount), nil
}
