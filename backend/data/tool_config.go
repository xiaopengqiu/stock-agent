package data

import (
	"encoding/json"
	"fmt"
	"go-stock/backend/logger"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	configCache *ToolConfig
	cacheLock   sync.RWMutex
	lastLoaded  time.Time
	cacheTTL    = 30 * time.Second // 30 秒缓存
)

// ToolConfig 工具配置结构
type ToolConfig struct {
	Version string     `json:"version"`
	Tools   []ToolItem `json:"tools"`
}

// ToolItem 工具配置项
type ToolItem struct {
	Name    string                 `json:"name"`
	Type    string                 `json:"type"` // "builtin" | "mcp" | "http"
	Enabled bool                   `json:"enabled"`
	Config  map[string]interface{} `json:"config"`
}

// LoadToolConfig 加载工具配置（带缓存）
func LoadToolConfig() (*ToolConfig, error) {
	cacheLock.RLock()

	// 检查缓存是否有效
	if configCache != nil && time.Since(lastLoaded) < cacheTTL {
		defer cacheLock.RUnlock()
		return configCache, nil
	}

	cacheLock.RUnlock()
	cacheLock.Lock()
	defer cacheLock.Unlock()

	// 重新加载配置
	configPath := "data/tools/config.json"

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.SugaredLogger.Warnf("工具配置文件不存在，使用默认配置: %s", configPath)
		configCache = getDefaultToolConfig()
		lastLoaded = time.Now()
		return configCache, nil
	}

	// 读取配置文件
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		logger.SugaredLogger.Errorf("读取工具配置文件失败: %v", err)
		if configCache == nil {
			configCache = getDefaultToolConfig()
		}
		lastLoaded = time.Now()
		return configCache, nil
	}

	// 解析配置
	var config ToolConfig
	if err := json.Unmarshal(bytes, &config); err != nil {
		logger.SugaredLogger.Errorf("解析工具配置文件失败: %v", err)
		if configCache == nil {
			configCache = getDefaultToolConfig()
		}
		lastLoaded = time.Now()
		return configCache, nil
	}

	configCache = &config
	lastLoaded = time.Now()
	return configCache, nil
}

// getDefaultToolConfig 获取默认工具配置
func getDefaultToolConfig() *ToolConfig {
	return GetDefaultToolConfig()
}

// GetDefaultToolConfig 获取默认工具配置（导出版本）
func GetDefaultToolConfig() *ToolConfig {
	return &ToolConfig{
		Version: "1.0",
		Tools: []ToolItem{
			{
				Name:    "SearchStockByIndicators",
				Type:    "builtin",
				Enabled: true,
				Config:  map[string]interface{}{},
			},
			{
				Name:    "GetStockKLine",
				Type:    "builtin",
				Enabled: true,
				Config:  map[string]interface{}{},
			},
			{
				Name:    "InteractiveAnswer",
				Type:    "builtin",
				Enabled: true,
				Config:  map[string]interface{}{},
			},
			{
				Name:    "GetStockResearchReport",
				Type:    "builtin",
				Enabled: true,
				Config:  map[string]interface{}{},
			},
			// ========== 高优先级工具 ==========
			{
				Name:    "SearchStockCode",
				Type:    "builtin",
				Enabled: true,
				Config:  map[string]interface{}{},
			},
			{
				Name:    "GetStockPriceInfo",
				Type:    "builtin",
				Enabled: true,
				Config:  map[string]interface{}{},
			},
			{
				Name:    "GetMarketNews",
				Type:    "builtin",
				Enabled: true,
				Config:  map[string]interface{}{},
			},
			// ========== 中优先级工具 ==========
			{
				Name:    "GetStockNews",
				Type:    "builtin",
				Enabled: true,
				Config:  map[string]interface{}{},
			},
			{
				Name:    "GetFinancialReports",
				Type:    "builtin",
				Enabled: true,
				Config:  map[string]interface{}{},
			},
			{
				Name:    "GetIndustryResearchReport",
				Type:    "builtin",
				Enabled: true,
				Config:  map[string]interface{}{},
			},
			// ========== 低优先级工具（数据/字典类）==========
			{
				Name:    "GetEconomicData",
				Type:    "builtin",
				Enabled: false,
				Config:  map[string]interface{}{},
			},
			{
				Name:    "GetBKDict",
				Type:    "builtin",
				Enabled: false,
				Config:  map[string]interface{}{},
			},
		},
	}
}

// validateToolConfig 验证工具配置
func validateToolConfig(config *ToolConfig) error {
	// 验证版本
	if config.Version == "" {
		return fmt.Errorf("配置缺少版本信息")
	}

	// 验证工具配置
	for i, toolItem := range config.Tools {
		if toolItem.Name == "" {
			return fmt.Errorf("工具 %d 缺少名称", i+1)
		}

		if toolItem.Type == "" {
			return fmt.Errorf("工具 %s 缺少类型", toolItem.Name)
		}

		if toolItem.Type != "builtin" && toolItem.Type != "mcp" && toolItem.Type != "http" {
			return fmt.Errorf("工具 %s 类型无效: %s", toolItem.Name, toolItem.Type)
		}

		if toolItem.Type == "http" {
			_, hasURL := toolItem.Config["url"]
			if !hasURL {
				return fmt.Errorf("HTTP 工具 %s 需要配置 url", toolItem.Name)
			}
		}
	}

	return nil
}

// SaveToolConfig 保存工具配置
func SaveToolConfig(config *ToolConfig) error {
	// 验证配置
	if err := validateToolConfig(config); err != nil {
		return err
	}

	configPath := "data/tools/config.json"

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	// 序列化配置
	bytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件
	if err := os.WriteFile(configPath, bytes, 0644); err != nil {
		return err
	}

	// 更新缓存
	cacheLock.Lock()
	configCache = config
	lastLoaded = time.Now()
	cacheLock.Unlock()

	return nil
}

// InvalidateToolConfigCache 清除工具配置缓存
func InvalidateToolConfigCache() {
	cacheLock.Lock()
	defer cacheLock.Unlock()
	configCache = nil
}
