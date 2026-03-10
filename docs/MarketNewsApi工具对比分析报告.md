# MarketNewsApi 工具对比分析报告

**分析时间:** 2026-03-05  
**对比项目:** temp-go-stock vs stock-agent (当前项目)**

---

## 一、函数清单对比

| 函数名 | temp-go-stock | stock-agent (我们项目 | 差异 |
|---------|----------------|-------------------|------|
| `TelegraphList` | ✅ | ✅ | 相同 |
| `GetNewTelegraph` | ✅ | ✅ | 相同 |
| `GetNewsList` | ✅ | ✅ | 相同 |
| `GetNewsList2` | ✅ | ❌ | temp 独有 |
| `GetTelegraphList` | ✅ | ✅ | 相同 |
| `GetTelegraphListWithPaging` | ✅ | ❌ | temp 独有 |
| `GetSinaNews` | ✅ | ✅ | 相同 |
| `GlobalStockIndexes` | ✅ | ✅ | 相同 |
| `GetIndustryRank` | ✅ | ✅ | 相同 |
| `GetIndustryMoneyRankSina` | ✅ | ✅ | 相同 |
| `GetMoneyRankSina` | ✅ | ✅ | 相同 |
| `GetStockMoneyTrendByDay` | ✅ | ✅ | 相同 |
| `TopStocksRankingList` | ✅ | ✅ | 相同 |
| `LongTiger` | ✅ | ✅ | 相同 |
| `IndustryResearchReport` | ✅ | ✅ | 相同 |
| `StockResearchReport` | ✅ | ✅ | 相同 |
| `StockNotice` | ✅ | ✅ | 相同 |
| `EMDictCode` | ✅ | ✅ | 相同 |
| `TradingViewNews` | ✅ | ✅ | 相同 |
| `TradingViewNewsDetail` | ✅ | ❌ | temp 独有 |
| `XUEQIUHotStock` | ✅ | ✅ | 相同 |
| `HotEvent` | ✅ | ✅ | 相同 |
| `HotTopic` | ✅ | ✅ | 相同 |
| `InvestCalendar` | ✅ | ✅ | 相同 |
| `ClsCalendar` | ✅ | ✅ | 相同 |
| `GetGDP` | ✅ | ✅ | 相同 |
| `GetCPI` | ✅ | ✅ | 相同 |
| `GetPPI` | ✅ | ✅ | 相同 |
| `GetPMI` | ✅ | ✅ | 相同 |
| `GetIndustryReportInfo` | ✅ | ✅ | 相同 |
| `GetSecuritiesCompanyOpinionContent` | ✅ | ❌ | temp 独有 |
| `ReutersNew` | ✅ | ✅ | 相同 |
| `InteractiveAnswer` | ✅ | ✅ | 相同 |
| `CailianpressWeb` | ✅ | ✅ | 相同 |
| `GetNews24HoursList` | ✅ | ❌ | temp 独有 |

---

## 二、temp-go-stock 独有函数详解

### 2.1 `GetNewsList2` - 新闻列表获取 v2

**用途：** 另一个版本的新闻列表获取

**在项目中需要实现？** 建议：❌ 没有

---

### 2.2 `GetTelegraphListWithPaging` - 分页获取

**用途：** 带分页参数的电报列表

**在项目中需要实现？** 建议：❌ 可选（如果有分页需求才需要

---

### 2.3 `TradingViewNewsDetail` - TradingView 新闻详情

**用途：** 获取 TradingView 单条新闻的详细内容

**在项目中需要实现？** 建议：✅ 建议添加（对新闻详情很有用）

---

### 2.4 `GetSecuritiesCompanyOpinionContent` - 券商观点内容

**用途：** 获取券商机构的观点和分析内容

**在项目中需要实现？** 建议：✅ 强烈建议添加（专业机构观点对 AI 分析很有价值）

---

### 2.5 `GetNews24HoursList` - 24小时新闻列表

**用途：** 获取最近24小时的新闻列表

**在项目中需要实现？** 建议：✅ 建议添加（实时性强）

---

## 三、我们项目中可以借鉴的优化点

| 优先级 | 优化项 | 说明 |
|--------|--------|
| 🔴 **高** | `TradingViewNewsDetail` | 新闻详情能提供更详细的新闻内容 |
| 🔴 **高** | `GetSecuritiesCompanyOpinionContent` | 增加券商观点，提升AI分析专业度 |
| 🔴 **高** | `GetNews24HoursList` | 24小时新闻实时性更强 |
| 🟡 **中** | `GetNewsList2` | 备用新闻源 |
| 🟢 **低** | `GetTelegraphListWithPaging` | 分页功能（按需） |

---

## 四、推荐实施优先级

### Phase 1（高优先级 - 立即添加

1. **`GetSecuritiesCompanyOpinionContent` - 券商观点内容
   - 价值：提供专业机构分析，大幅提升 AI 荐股的专业性
   - 实现难度：中等

2. **`TradingViewNewsDetail` - TradingView 新闻详情
   - 价值：获取新闻详情，让 AI 能看到更详细的新闻内容
   - 实现难度：低

3. **`GetNews24HoursList` - 24小时新闻
   - 价值：24小时实时新闻
   - 实现难度：低

---

### Phase 2（中优先级 - 按需添加）

4. **`GetNewsList2` - 备用新闻源
   - 价值：增加数据源冗余，避免单一源故障
   - 实现难度：低

5. **`GetTelegraphListWithPaging` - 分页功能
   - 价值：大数据量分页查询
   - 实现难度：低

---

## 五、总结

**temp-go-stock 项目相比我们项目，**多了5个函数**，这些函数可以增强了数据来源和功能。

**建议优先添加的3个高价值函数**：
1. `GetSecuritiesCompanyOpinionContent` - 券商观点（🔴 高）
2. `TradingViewNewsDetail` - 新闻详情（🔴 高）
3. `GetNews24HoursList` - 24小时新闻（🔴 高）

**这3个函数可以显著提升AI分析能力。**
