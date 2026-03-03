# AI荐股报告解析最佳实践设计

## 设计文档信息

- **设计日期**: 2026-02-25
- **版本**: 1.0
- **作者**: Claude
- **状态**: 已批准

---

## 1. 问题背景

### 1.1 当前问题

当前AI荐股功能的实现方式存在以下问题：

1. 在prompt中要求大模型同时返回markdown报告和JSON数据
2. JSON数据会显示在报告中，影响用户体验
3. 解析逻辑复杂，需要处理两种格式的输出
4. 用户看到的报告中包含技术性的JSON格式数据

### 1.2 设计目标

- 大模型只返回纯净的markdown格式荐股报告
- 报告遵循固定的标题格式
- 通过解析报告的行格式来提取推荐的股票信息
- 提取的信息包括：股票代码、股票名称、推荐理由、买卖建议等
- 提升用户体验，让报告更加专业和易读

---

## 2. Prompt设计

### 2.1 文件位置

`data/skills/ai-stock-pick.md`

### 2.2 核心修改

**删除内容**：
- 要求大模型返回JSON格式数据的部分
- 任何提及"structured output"、"JSON schema"等的内容

**新增内容**：

```markdown
## 输出要求

### 完整报告格式（严格遵循）

请严格按照以下格式输出推荐报告，**无需添加额外的JSON或其他格式数据**：

```
# AI荐股报告

## 市场环境分析

[分析当前大盘走势、热点板块、资金流向、风险水平]

## 筛选逻辑

[说明选股的筛选条件和标准]

## 推荐股票

### 1. [股票名称] ([股票代码])
**当前价格**: XX.XX元
**涨跌幅**: ±XX.XX%
**目标价位**: XX.XX元
**上涨空间**: ±XX.XX%
**综合评分**: XX/100
**买卖建议**: [买入/卖出/观望]

**推荐理由**:
[3点核心驱动力]

**技术面分析**:
[趋势分析、K线形态、技术指标]

**基本面分析**:
[估值水平、业绩表现、增长潜力]

**风险提示**:
[主要风险点和应对策略]

---

### 2. [股票名称] ([股票代码])
[同上格式]

---

## 投资建议

**建议仓位**: XX%（根据市场环境）
**持有周期**: 短线/中线/长线
**跟踪要素**: [需要重点关注的指标]
```

### 报告结构说明

1. **标题层级严格要求**：
   - 一级标题：`# AI荐股报告`（固定）
   - 二级标题：`## 市场环境分析`、`## 筛选逻辑`、`## 推荐股票`、`## 投资建议`（固定顺序）
   - 三级标题：`### 1. [股票名称] ([股票代码])`（每个推荐股票）

2. **推荐股票部分格式**：
   - 每个推荐股票必须包含：股票代码、股票名称、当前价格、涨跌幅、目标价位、上涨空间、综合评分、买卖建议、推荐理由、技术面分析、基本面分析、风险提示

3. **注意事项**：
   - 不要在报告任何位置添加JSON格式数据
   - 不要添加额外的代码块或格式化标记
   - 确保所有数值信息清晰明确
```

### 2.3 输出示例

```markdown
# AI荐股报告

## 市场环境分析

今日A股市场震荡上行，上证指数收涨1.25%。科技板块表现强势，半导体、AI概念领涨两市。北向资金全天净流入85亿元，显示外资对A股的信心增强。当前市场风险水平中等，建议适度控制仓位。

## 筛选逻辑

1. 行业选择：聚焦半导体、AI、新能源等科技成长板块
2. 技术面：筛选突破重要压力位、趋势向上的个股
3. 基本面：优选业绩增长确定性强、估值合理的优质标的
4. 资金面：关注近期有主力资金大幅流入的个股

## 推荐股票

### 1. 中芯国际 (sh688981)
**当前价格**: 45.67元
**涨跌幅**: +3.25%
**目标价位**: 52.00元
**上涨空间**: +13.9%
**综合评分**: 85/100
**买卖建议**: 买入

**推荐理由**:
1. 行业龙头，国产替代加速受益者
2. 资金大幅流入，主力抢筹迹象明显
3. 技术面突破60日均线压力位，趋势向好

**技术面分析**:
股价突破60日均线，MACD金叉向上，RSI指标65，处于强势区域。成交量放大，换手率健康，显示资金积极参与。

**基本面分析**:
当前PE 45倍，略高于行业平均，但考虑到业绩增长预期和国产替代逻辑，估值相对合理。Q3业绩超预期，同比增长25%。

**风险提示**:
- 半导体行业竞争加剧，国产替代进程可能慢于预期
- 技术迭代风险，先进制程研发投入较大
- 建议控制仓位，分批建仓，设置止损位

---

### 2. 宁德时代 (sz300750)
**当前价格**: 185.32元
**涨跌幅**: +2.18%
**目标价位**: 205.00元
**上涨空间**: +10.6%
**综合评分**: 88/100
**买卖建议**: 买入

**推荐理由**:
1. 全球动力电池龙头，市场地位稳固
2. 海外订单持续落地，增长确定性高
3. 储能业务快速增长，第二增长曲线清晰

**技术面分析**:
股价在年线附近获得支撑，近期反弹势头强劲。MACD底部金叉，KDJ指标金叉向上，短期有望继续走强。

**基本面分析**:
当前PE 28倍，低于历史中位数水平，具备安全边际。储能业务占比提升至30%，成为新的增长引擎。

**风险提示**:
- 原材料价格波动影响毛利率
- 海外市场拓展面临地缘政治风险
- 建议关注原材料价格变化和海外订单进展

---

### 3. 比亚迪 (sz002594)
**当前价格**: 238.56元
**涨跌幅**: +1.85%
**目标价位**: 255.00元
**上涨空间**: +6.9%
**综合评分**: 82/100
**买卖建议**: 买入

**推荐理由**:
1. 新能源汽车销量持续领跑
2. 刀片电池技术领先，具备成本优势
3. 全球化战略稳步推进

**技术面分析**:
股价在半年线上方运行，中期趋势向上。RSIser指标55，处于健康区域。成交温和放大，显示资金逐步介入。

**基本面分析**:
当前PE 35倍，估值处于合理区间。2023年销量有望突破300万辆，市场份额进一步提升。

**风险提示**:
- 行业竞争加剧，价格战持续
- 下游需求不及预期风险
- 建议关注月度销量数据和行业竞争格局

---

## 投资建议

**建议仓位**: 60%（考虑到当前市场环境，建议适度参与）
**持有周期**: 中线（3-6个月）
**跟踪要素**:
1. 大盘走势和北向资金流向
2. 科技板块整体表现
3. 个股业绩公告和重大事项
4. 主力资金动向
```

---

## 3. 数据结构设计

### 3.1 文件位置

`backend/models/stock_pick_report.go`

### 3.2 扩展数据模型

为 `RecommendationItem` 结构体添加 `TradeSuggestion` 字段：

```go
// RecommendationItem 推荐股票项
type RecommendationItem struct {
	Rank               int     `json:"rank"`          // 排名
	StockCode          string  `json:"stock_code"`    // 股票代码
	StockName          string  `json:"stock_name"`    // 股票名称
	CurrentPrice       float64 `json:"current_price"` // 现价
	PriceChange        float64 `json:"price_change"`  // 涨跌幅
	Volume             int64   `json:"volume"`        // 成交量
	MarketValue        float64 `json:"market_value"`  // 市值

	// 分析内容
	TechnicalAnalysis   string  `json:"technical_analysis"`    // 技术面分析
	FundamentalAnalysis string  `json:"fundamental_analysis"`  // 基本面分析
	Reason              string  `json:"reason"`                // 推荐理由
	TargetPrice         float64 `json:"target_price"`          // 目标价位
	TargetChangePercent float64 `json:"target_change_percent"` // 目标涨幅
	RiskLevel           string  `json:"risk_level"`            // 风险等级: 'low' | 'medium' | 'high'
	RiskTips            string  `json:"risk_tips"`             // 风险提示
	Score               float64 `json:"score"`                 // 综合评分 (0-100)
	TradeSuggestion     string  `json:"trade_suggestion"`      // 买卖建议: '买入' | '卖出' | '观望'

	// 关注状态
	IsFollowed bool `json:"is_followed"`

	// 技术指标
	MACD  string  `json:"macd"`  // MACD指标
	KDJ   string  `json:"kdj"`   // KDJ指标
	RSI   float64 `json:"rsi"`   // RSI指标
	Trend string  `json:"trend"` // 趋势: 'up' | 'down' | 'sideways'

	// 基本面指标
	PE            float64 `json:"pe"`             // 市盈率
	PB            float64 `json:"pb"`             // 市净率
	ROE           float64 `json:"roe"`            // 净资产收益率
	RevenueGrowth float64 `json:"revenue_growth"` // 营收增长率
	ProfitGrowth  float64 `json:"profit_growth"`  // 利润增长率
}
```

---

## 4. 解析逻辑设计

### 4.1 文件位置

`backend/data/stock_pick_service.go`

### 4.2 核心解析函数

```go
// parseRecommendationsFromContent 从markdown内容解析推荐股票
func (s *StockPickService) parseRecommendationsFromContent(content string) []models.RecommendationItem {
	logger.SugaredLogger.Infof("开始解析markdown推荐内容，内容长度: %d", len(content))

	var recommendations []models.RecommendationItem

	lines := strings.Split(content, "\n")
	logger.SugaredLogger.Debugf("内容分割为 %d 行", len(lines))

	var currentRec *models.RecommendationItem
	var inRecommendationSection bool

	for lineIdx, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		lowerLine := strings.ToLower(trimmedLine)

		// 检测推荐章节开始
		if strings.Contains(lowerLine, "## 推荐股票") {
			inRecommendationSection = true
			logger.SugaredLogger.Debugf("在行 %d 检测到推荐章节开始", lineIdx)
			continue
		}

		// 检测推荐章节结束
		if inRecommendationSection && strings.Contains(lowerLine, "## 投资建议") {
			break
		}

		if !inRecommendationSection {
			continue
		}

		// 解析股票标题行：### 1. 中芯国际 (sh688981)
		if strings.HasPrefix(trimmedLine, "###") && strings.Contains(trimmedLine, "(") && strings.Contains(trimmedLine, ")") {
			// 保存上一个推荐
			if currentRec != nil {
				recommendations = append(recommendations, *currentRec)
				logger.SugaredLogger.Debugf("添加推荐项，当前总数: %d", len(recommendations))
			}

			// 解析新的推荐股票
			if rec, err := s.parseStockTitle(trimmedLine, len(recommendations)+1); err == nil {
				currentRec = rec
				logger.SugaredLogger.Debugf("解析到推荐项 %d: 代码=%s, 名称=%s", rec.Rank, rec.StockCode, rec.StockName)
			}
		} else if currentRec != nil {
			// 解析详细信息
			s.parseDetailLine(trimmedLine, currentRec)
		}
	}

	// 保存最后一个推荐
	if currentRec != nil {
		recommendations = append(recommendations, *currentRec)
	}

	// 更新关注状态
	for i := range recommendations {
		recommendations[i].IsFollowed = s.CheckStockFollowed(recommendations[i].StockCode)
	}

	logger.SugaredLogger.Infof("解析完成，共 %d 个推荐项", len(recommendations))
	return recommendations
}

// parseStockTitle 解析股票标题：### 1. 中芯国际 (sh688981)
func (s *StockPickService) parseStockTitle(line string, rank int) (*models.RecommendationItem, error) {
	// 去除 ### 和可能的排名前缀
	cleanLine := strings.TrimSpace(strings.TrimPrefix(line, "###"))
	cleanLine = strings.TrimSpace(strings.TrimPrefix(cleanLine, fmt.Sprintf("%d.", rank)))
	cleanLine = strings.TrimSpace(cleanLine)

	// 提取股票代码和名称：格式 "中芯国际 (sh688981)"
	codeStartIdx := strings.LastIndex(cleanLine, "(")
	codeEndIdx := strings.LastIndex(cleanLine, ")")

	if codeStartIdx == -1 || codeEndIdx == -1 || codeStartIdx >= codeEndIdx {
		return nil, fmt.Errorf("无法解析股票标题格式: %s", line)
	}

	stockCode := strings.TrimSpace(cleanLine[codeStartIdx+1 : codeEndIdx])
	stockName := strings.TrimSpace(cleanLine[:codeStartIdx])

	if !isValidStockCode(stockCode) {
		return nil, fmt.Errorf("无效的股票代码格式: %s", stockCode)
	}

	// 创建推荐项
	rec := &models.RecommendationItem{
		Rank:       rank,
		StockCode:  strings.ToUpper(stockCode),
		StockName:  stockName,
		IsFollowed: s.CheckStockFollowed(stockCode),
	}

	// 尝试获取实时价格
	if stockInfo, err := NewStockDataApi().GetStockCodeRealTimeData(stockCode); err == nil && stockInfo != nil && len(*stockInfo) > 0 {
		stock := (*stockInfo)[0]
		if price, err := parsePrice(stock.Price); err == nil {
			rec.CurrentPrice = price
		}
		rec.PriceChange = stock.ChangePercent
	}

	return rec, nil
}

// parseDetailLine 解析详细信息行
func (s *StockPickService) parseDetailLine(line string, rec *models.RecommendationItem) {
	lowerLine := strings.ToLower(line)

	if strings.Contains(lowerLine, "当前价格") {
		if price, err := parsePriceFromLine(line); err == nil {
			rec.CurrentPrice = price
		}
	} else if strings.Contains(lowerLine, "涨跌幅") {
		if change, err := parsePriceFromLine(line); err == nil {
			rec.PriceChange = change
		}
	} else if strings.Contains(lowerLine, "目标价位") || strings.Contains(lowerLine, "目标价") {
		if price, err := parsePriceFromLine(line); err == nil {
			rec.TargetPrice = price
		}
	} else if strings.Contains(lowerLine, "上涨空间") || strings.Contains(lowerLine, "目标涨幅") {
		if change, err := parsePriceFromLine(line); err == nil {
			rec.TargetChangePercent = change
		}
	} else if strings.Contains(lowerLine, "综合评分") {
		if score, err := parsePriceFromLine(line); err == nil {
			rec.Score = score
		}
	} else if strings.Contains(lowerLine, "买卖建议") {
		rec.TradeSuggestion = parseTextValue(line)
	}
}

// parseTextValue 解析文本值
func parseTextValue(line string) string {
	// 提取冒号后的内容
	if idx := strings.Index(line, ":"); idx >= 0 {
		return strings.TrimSpace(line[idx+1:])
	}
	if idx := strings.Index(line, "："); idx >= 0 {
		return strings.TrimSpace(line[idx+1:])
	}
	return line
}

// parsePriceFromLine 从行中解析价格
func parsePriceFromLine(line string) (float64, error) {
	// 提取冒号后的内容
	text := parseTextValue(line)
	// 移除单位符号
	text = strings.ReplaceAll(text, "元", "")
	text = strings.ReplaceAll(text, "%", "")
	return strconv.ParseFloat(strings.TrimSpace(text), 64)
}

// isValidStockCode 验证股票代码格式
func isValidStockCode(code string) bool {
	code = strings.ToLower(strings.TrimSpace(code))

	// 支持多种股票代码格式
	patterns := []string{
		`^sh[0-9]{6}$`,  // 上海A股
		`^sz[0-9]{6}$`,  // 深圳A股
		`^hk[0-9]{4}$`,  // 港股
		`^us[a-z0-9]+$`, // 美股
		`^[0-9]{6}$`,    // 纯数字代码
	}

	for _, pattern := range patterns {
		if match, _ := regexp.MatchString(pattern, code); match {
			return true
		}
	}

	return false
}
```

### 4.3 错误处理

```go
// parseRecommendationsFromContent 增强错误处理
func (s *StockPickService) parseRecommendationsFromContent(content string) []models.RecommendationItem {
	var recommendations []models.RecommendationItem

	defer func() {
		if err := recover(); err != nil {
			logger.SugaredLogger.Errorf("解析推荐内容时发生panic: %v", err)
			recommendations = []models.RecommendationItem{}
		}
	}()

	// 检查内容是否为空
	if strings.TrimSpace(content) == "" {
		logger.SugaredLogger.Warn("AI响应内容为空")
		return []models.RecommendationItem{}
	}

	// 检查是否包含推荐股票章节
	if !strings.Contains(strings.ToLower(content), "## 推荐股票") {
		logger.SugaredLogger.Warn("AI响应中未找到推荐股票章节")
		return []models.RecommendationItem{}
	}

	// ... 解析逻辑

	return recommendations
}
```

---

## 5. 实施步骤

### 阶段1：修改Prompt和数据模型

1. **更新 prompt 文件**
   - 文件：`data/skills/ai-stock-pick.md`
   - 删除JSON输出要求
   - 定义严格的markdown格式

2. **扩展数据模型**
   - 文件：`backend/models/stock_pick_report.go`
   - 添加 `TradeSuggestion` 字段

### 阶段2：重写解析逻辑

1. **重写解析函数**
   - 文件：`backend/data/stock_pick_service.go`
   - 重写 `parseRecommendationsFromContent` 函数
   - 删除JSON解析逻辑

2. **添加解析辅助函数**
   - `parseStockTitle` - 解析股票标题
   - `parseDetailLine` - 解析详细信息行
   - `parseTextValue` - 解析文本值
   - `parsePriceFromLine` - 解析价格值
   - `isValidStockCode` - 验证股票代码格式

### 阶段3：更新前端展示

1. **更新表格列**
   - 文件：`frontend/src/components/ai-stock-pick.vue`
   - 在简洁列表中添加买卖建议列

2. **更新格式化函数**
   - 修改 `formatReportToMarkdown` 函数
   - 支持新的报告格式

### 阶段4：测试验证

1. **功能测试**
   - 测试AI荐股流程
   - 验证解析逻辑正确

2. **边界测试**
   - 测试各种格式的报告
   - 测试解析稳定性

3. **性能测试**
   - 测试解析大量推荐股票

---

## 6. 测试策略

### 6.1 单元测试

**文件**：`backend/data/stock_pick_service_test.go`

```go
package data

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/go-stock/backend/models"
)

func TestParseRecommendationsFromContent(t *testing.T) {
	testContent := `
# AI荐股报告

## 市场环境分析

今日大盘震荡上涨。

## 筛选逻辑

筛选条件：科技股，资金流入。

## 推荐股票

### 1. 中芯国际 (sh688981)
**当前价格**: 45.67元
**涨跌幅**: +3.25%
**目标价位**: 52.00元
**上涨空间**: +13.9%
**综合评分**: 85/100
**买卖建议**: 买入

**推荐理由**:
1. 行业龙头，国产替代受益者
2. 资金大幅流入，主力抢筹迹象明显
3. 技术面突破压力位，趋势向好

**技术面分析**:
股价突破60日均线，MACD金叉向上，RSI指标65，处于强势区域。

**基本面分析**:
当前PE 45倍，略高于行业平均，但考虑到业绩增长预期，估值合理。

**风险提示**:
- 行业竞争加剧
- 技术迭代风险
- 建议控制仓位，分批建仓

---

### 2. 宁德时代 (sz300750)
**当前价格**: 185.32元
**涨跌幅**: +2.18%
**目标价位**: 205.00元
**上涨空间**: +10.6%
**综合评分**: 88/100
**买卖建议**: 买入

**推荐理由**:
1. 全球动力电池龙头
2. 海外订单持续落地
3. 储能业务快速增长

**技术面分析**:
股价在年线附近获得支撑，近期反弹势头强劲。

**基本面分析**:
当前PE 28倍，低于历史中位数水平。

**风险提示**:
- 原材料价格波动
- 海外市场拓展风险

---

## 投资建议

**建议仓位**: 60%
**持有周期**: 中线
**跟踪要素**: 大盘走势、板块表现
`

	service := &StockPickService{}
	recommendations := service.parseRecommendationsFromContent(testContent)

	assert.NotNil(t, recommendations)
	assert.Equal(t, 2, len(recommendations))

	// 验证第一只股票
	assert.Equal(t, "SH688981", strings.ToUpper(recommendations[0].StockCode))
	assert.Equal(t, "中芯国际", recommendations[0].StockName)
	assert.Equal(t, 1, recommendations[0].Rank)
	assert.Equal(t, 45.67, recommendations[0].CurrentPrice)
	assert.Equal(t, 3.25, recommendations[0].PriceChange)
	assert.Equal(t, 52.00, recommendations[0].TargetPrice)
	assert.Equal(t, 13.9, recommendations[0].TargetChangePercent)
	assert.Equal(t, 85.0, recommendations[0].Score)
	assert.Equal(t, "买入", recommendations[0].TradeSuggestion)

	// 验证第二只股票
	assert.Equal(t, "SZ300750", strings.ToUpper(recommendations[1].StockCode))
	assert.Equal(t, "宁德时代", recommendations[1].StockName)
	assert.Equal(t, 2, recommendations[1].Rank)
}

func TestParseStockTitle(t *testing.T) {
	service := &StockPickService{}

	tests := []struct {
		name      string
		line      string
		rank      int
		wantCode  string
		wantName  string
		wantError bool
	}{
		{
			name:     "正常格式",
			line:     "### 1. 中芯国际 (sh688981)",
			rank:     1,
			wantCode: "SH688981",
			wantName: "中芯国际",
		},
		{
			name:     "无前缀格式",
			line:     "### 中芯国际 (sh688981)",
			rank:     1,
			wantCode: "SH688981",
			wantName: "中芯国际",
		},
		{
			name:      "无括号格式",
			line:      "### 1. 中芯国际 sh688981",
			rank:      1,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, err := service.parseStockTitle(tt.line, tt.rank)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCode, strings.ToUpper(rec.StockCode))
				assert.Equal(t, tt.wantName, rec.StockName)
			}
		})
	}
}

func TestParseDetailLine(t *testing.T) {
	tests := []struct {
		name           string
		line           string
		wantPrice      float64
		wantChange      string
		wantSuggestion string
	}{
		{
			name:           "当前价格",
			line:           "**当前价格**: 45.67元",
			wantPrice:      45.67,
			wantSuggestion: "",
		},
		{
			name:           "涨跌幅",
			line:           "**涨跌幅**: +3.25%",
			wantChange:     "3.25%",
			wantSuggestion: "",
		},
		{
			name:           "买卖建议",
			line:           "**买卖建议**: 买入",
			wantSuggestion: "买入",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &models.RecommendationItem{}
			service := &StockPickService{}
			service.parseDetailLine(tt.line, rec)
			assert.Equal(t, tt.wantSuggestion, rec.TradeSuggestion)
		})
	}
}

func TestIsValidStockCode(t *testing.T) {
	tests := []struct {
		code    string
		wantValid bool
	}{
		{"sh688981", true},
		{"SZ300750", true},
		{"hk0700", true},
		{"hk00700", true},
		{"USaapl", true},
		{"AAPL", true},
		{"600000", true},
		{"invalid", false},
		{"sh123", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			assert.Equal(t, tt.wantValid, isValidStockCode(tt.code))
		})
	}
}
```

### 6.2 集成测试

1. 测试完整的AI荐股流程
2. 验证从AI响应到前端展示的完整链路
3. 测试不同AI模型的输出解析

### 6.3 边界测试

1. 空响应测试
2. 格式不完整的响应测试
3. 多只股票推荐测试
4. 特殊字符处理测试

---

## 7. 风险评估

| 风险 | 级别 | 影响 | 缓解措施 |
|------|------|------|----------|
| 大模型输出格式不规范 | 高 | 解析失败，无法提取推荐信息 | 容错解析、详细的日志记录、提供友好的错误提示 |
| 解析失败影响用户体验 | 中 | 用户看不到推荐列表 | 解析失败返回空列表，不中断报告展示流程 |
| 历史数据兼容性 | 低 | 旧报告无法正确解析 | 添加版本标记，支持新旧格式兼容 |
| 性能问题 | 低 | 大量推荐股票时解析慢 | 优化解析算法，必要时使用并行处理 |

---

## 8. 优化建议

### 8.1 性能优化

1. **并行解析**：对于包含大量推荐股票的报告，可以使用并发处理
2. **缓存机制**：对已解析的报告进行缓存，避免重复解析
3. **增量解析**：对于长报告，支持增量解析和流式处理

### 8.2 可扩展性

1. **版本控制**：记录prompt和解析逻辑的版本，支持向后兼容
2. **插件化**：将解析逻辑封装为独立模块，便于后续扩展
3. **配置化**：通过配置文件定义解析规则，无需修改代码

### 8.3 监控

1. **解析成功率**：监控解析成功率，及时发现格式问题
2. **解析耗时**：监控解析耗时，优化性能瓶颈
3. **错误日志**：记录详细的错误日志，便于问题排查

---

## 9. 总结

本设计文档通过以下方式解决了当前AI荐股功能中的问题：

1. **统一的输出格式**：定义了严格的markdown报告格式，消除了JSON数据混入报告的问题
2. **清晰的解析逻辑**：通过逐行解析提取股票信息，逻辑清晰易于维护
3. **完善的错误处理**：提供了全面的错误处理机制，确保系统稳定性
4. **良好的扩展性**：支持未来添加更多解析字段和格式

实施该方案后，用户将获得更加专业和易读的荐股报告，同时系统将具备更好的可维护性和扩展性。

---

## 10. 附录

### 10.1 相关文件清单

| 文件路径 | 修改类型 | 说明 |
|----------|----------|------|
| `data/skills/ai-stock-pick.md` | 修改 | 删除JSON输出要求，定义markdown格式 |
| `backend/models/stock_pick_report.go` | 修改 | 添加TradeSuggestion字段 |
| `backend/data/stock_pick_service.go` | 重写 | 重写解析逻辑 |
| `backend/data/stock_pick_service_test.go` | 新增 | 添加单元测试 |
| `frontend/src/components/ai-stock-pick.vue` | 修改 | 更新前端展示 |

### 10.2 参考资料

- [Wails官方文档](https://wails.io/docs)
- [Go语言正则表达式](https://pkg.go.dev/regexp)
- [Markdown规范](https://commonmark.org/)
- [Vue 3文档](https://vuejs.org/)
