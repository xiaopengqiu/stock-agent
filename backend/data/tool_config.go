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
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // "builtin" | "mcp" | "http"
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config"`
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
				Name:        "ChoiceStockByIndicators",
				Type:        "builtin",
				Description: "自然语言选股工具，支持用自然语言描述选股条件，如技术指标（MACD、KDJ、RSI、BOLL）、财务指标（PE、净利润增长率）、市场数据（换手率、量比、涨停板）等。可以同时查询多只股票的详细数据，是进行股票筛选和分析的核心工具。",
				Enabled:     true,
				Config:      map[string]interface{}{},
			},
			{
				Name:        "QueryStockKLine",
				Type:        "builtin",
				Description: "获取股票K线数据，支持A股、港股、美股。可指定天数获取日K数据，包含开盘价、收盘价、最高价、最低价、成交量等关键信息。K线数据是技术分析的基础，用于判断趋势、支撑阻力位和买卖信号。",
				Enabled:     true,
				Config:      map[string]interface{}{},
			},
			{
				Name:        "QueryInteractiveAnswerData",
				Type:        "builtin",
				Description: "获取投资者与上市公司互动问答数据，反映当前投资者关注的热点问题。可以通过关键词搜索相关问题，了解市场关注点和公司最新动态，是挖掘潜在投资机会和风险的重要信息来源。",
				Enabled:     true,
				Config:      map[string]interface{}{},
			},
			{
				Name:        "GetFinancialReport",
				Type:        "builtin",
				Description: "查询股票财务报表数据，包括资产负债表、利润表、现金流量表等核心财务信息。财务数据是基本面分析的基础，用于评估公司盈利能力、偿债能力、运营效率和成长潜力。",
				Enabled:     true,
				Config:      map[string]interface{}{},
			},
			// ========== 高优先级工具 ==========
			{
				Name:        "QueryStockCodeInfo",
				Type:        "builtin",
				Description: "查询股票/指数信息，支持通过名称、代码、拼音或拼音首字母搜索。返回股票名称、代码、交易所等基本信息，是进行其他股票查询的前置工具，帮助快速定位目标股票。",
				Enabled:     true,
				Config:      map[string]interface{}{},
			},
			{
				Name:        "QueryStockPriceInfo",
				Type:        "builtin",
				Description: "批量获取实时股价数据，支持多只股票同时查询。返回最新价格、涨跌幅、成交量、成交额等实时交易数据，是监控市场动态和持仓股票表现的必备工具。",
				Enabled:     true,
				Config:      map[string]interface{}{},
			},
			{
				Name:        "QueryMarketNews",
				Type:        "builtin",
				Description: "获取国内外市场资讯、财经日历、重要事件。内容包括财联社电报、全球新闻、外媒资讯等，帮助及时了解市场动态、政策变化和重大事件，把握市场脉搏。",
				Enabled:     true,
				Config:      map[string]interface{}{},
			},
			// ========== 中优先级工具 ==========
			{
				Name:        "QueryStockNewsTool",
				Type:        "builtin",
				Description: "按关键词搜索相关市场资讯和新闻，支持多个关键词组合搜索。可针对特定股票、板块或主题进行精准信息检索，是深入了解特定标的和热点的重要工具。",
				Enabled:     true,
				Config:      map[string]interface{}{},
			},
			{
				Name:        "GetIndustryResearchReport",
				Type:        "builtin",
				Description: "获取行业/板块研究报告，提供行业发展趋势、竞争格局、政策分析等深度内容。行业研究是自上而下投资分析的重要环节，帮助理解行业周期和投资机会。",
				Enabled:     true,
				Config:      map[string]interface{}{},
			},
			// ========== 低优先级工具（数据/字典类）==========
			{
				Name:        "QueryEconomicData",
				Type:        "builtin",
				Description: "查询宏观经济数据，包括GDP（国内生产总值）、CPI（居民消费价格指数）、PPI（工业品出厂价格指数）、PMI（采购经理人指数）。宏观数据是判断经济周期和政策方向的重要依据。",
				Enabled:     false,
				Config:      map[string]interface{}{},
			},
			{
				Name:        "QueryBKDictInfo",
				Type:        "builtin",
				Description: "获取所有板块/行业的名称和代码字典，提供完整的板块分类体系。是进行行业分析和板块轮动研究的基础参考工具，帮助快速了解市场板块结构。",
				Enabled:     false,
				Config:      map[string]interface{}{},
			},
			{
				Name:        "QueryHKStockPrice",
				Type:        "builtin",
				Description: "查询港股实时价格数据，支持港股通标的和主要港股。港股价格对A+H股两地上市的科技公司、医药股、金融股有重要参考价值，可用于跨市场比较分析。",
				Enabled:     true,
				Config:      map[string]interface{}{},
			},
			{
				Name:        "QueryShareholderCount",
				Type:        "builtin",
				Description: "查询股票的股东人数变化和筹码集中度，是判断主力吸筹或出货的重要指标。股东人数减少通常表示筹码集中（主力吸筹），股东人数增加表示筹码分散（主力出货）。",
				Enabled:     true,
				Config:      map[string]interface{}{},
			},
			{
				Name:        "GetNewsList2",
				Type:        "builtin",
				Description: "获取新闻列表（会自动刷新最新电报数据），支持按来源筛选和数量限制。提供最新的市场快讯和资讯，帮助第一时间掌握市场动态和突发消息。",
				Enabled:     true,
				Config:      map[string]interface{}{},
			},
			{
				Name:        "GetTelegraphListWithPaging",
				Type:        "builtin",
				Description: "分页获取电报列表，支持按来源筛选。可灵活控制页码和每页数量，适合回溯历史资讯和进行特定时段的信息搜集。",
				Enabled:     true,
				Config:      map[string]interface{}{},
			},
			{
				Name:        "TradingViewNewsDetail",
				Type:        "builtin",
				Description: "获取TradingView新闻的详细内容，提供国际市场视角和专业分析。TradingView是全球知名的交易平台，其新闻涵盖全球主要市场和资产类别。",
				Enabled:     true,
				Config:      map[string]interface{}{},
			},
			{
				Name:        "GetSecuritiesCompanyOpinionContent",
				Type:        "builtin",
				Description: "获取券商观点的详细内容，包括机构研报、投资策略、个股评级等。券商观点是专业机构投资者的研究成果，对投资决策有重要参考价值。",
				Enabled:     true,
				Config:      map[string]interface{}{},
			},
			{
				Name:        "GetNews24HoursList",
				Type:        "builtin",
				Description: "获取过去24小时内的新闻列表，支持按来源筛选和数量限制，自动去重。聚焦最新资讯，帮助快速回顾一天内的重要市场事件和新闻热点。",
				Enabled:     true,
				Config:      map[string]interface{}{},
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
