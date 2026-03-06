package orchestrator

import (
	"sync"
	"time"
)

// ============================================
// Phase 4: 缓存机制
// 提升多Agent系统性能
// ============================================

// CacheEntry 缓存条目
type CacheEntry struct {
	Value      interface{}
	Expiration time.Time
}

// AgentCache Agent级别的缓存系统
type AgentCache struct {
	cache map[string]CacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

// NewAgentCache 创建新的Agent缓存
func NewAgentCache(ttl time.Duration) *AgentCache {
	cache := &AgentCache{
		cache: make(map[string]CacheEntry),
		ttl:   ttl,
	}
	// 启动定期清理过期缓存
	go cache.startCleanup()
	return cache
}

// Set 设置缓存
func (c *AgentCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = CacheEntry{
		Value:      value,
		Expiration: time.Now().Add(c.ttl),
	}
}

// Get 获取缓存
func (c *AgentCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(entry.Expiration) {
		// 过期了，删除并返回不存在
		go c.Delete(key)
		return nil, false
	}

	return entry.Value, true
}

// Delete 删除缓存
func (c *AgentCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, key)
}

// Clear 清空所有缓存
func (c *AgentCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]CacheEntry)
}

// Size 获取缓存大小
func (c *AgentCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// startCleanup 定期清理过期缓存
func (c *AgentCache) startCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup 清理过期缓存
func (c *AgentCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.cache {
		if now.After(entry.Expiration) {
			delete(c.cache, key)
		}
	}
}

// ============================================
// 常用缓存键生成器
// ============================================

// GenerateStockCacheKey 生成股票数据缓存键
func GenerateStockCacheKey(prefix, stockCode string) string {
	return prefix + ":" + stockCode
}

// GenerateKLineCacheKey 生成K线数据缓存键
func GenerateKLineCacheKey(stockCode, days string) string {
	return "kline:" + stockCode + ":" + days
}

// GenerateFinancialReportCacheKey 生成财务报告缓存键
func GenerateFinancialReportCacheKey(stockCode string) string {
	return "financial:" + stockCode
}

// GenerateNewsCacheKey 生成新闻缓存键
func GenerateNewsCacheKey(stockCode string) string {
	return "news:" + stockCode
}

// ============================================
// 全局缓存实例
// ============================================

var (
	globalCache *AgentCache
	cacheOnce   sync.Once
)

// GetGlobalCache 获取全局缓存实例
func GetGlobalCache() *AgentCache {
	cacheOnce.Do(func() {
		// 默认TTL: 30分钟
		globalCache = NewAgentCache(30 * time.Minute)
	})
	return globalCache
}

// SetGlobalCacheTTL 设置全局缓存TTL
func SetGlobalCacheTTL(ttl time.Duration) {
	if globalCache != nil {
		globalCache.ttl = ttl
	}
}
