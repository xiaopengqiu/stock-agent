# AI荐股功能优化计划

## 需求概述

本次优化针对AI荐股功能的4个问题：
1. 历史记录加载报错：`TypeError: Cannot read properties of null (reading '0')`
2. 完整报告页面无法上下滚动，导致内容无法浏览
3. 简洁列表无法正确解析报告中提到的股票信息
4. 导出Markdown格式要求支持指定保存目录

---

## 问题分析

### 问题1：历史记录加载报错

**问题定位**：
- 文件：`frontend/src/components/ai-stock-pick-fixed.vue`
- 行号：503-504
- 代码：`historyList.value = result[0]`

**根本原因**：
后端`GetStockPickReports`方法返回3个值：`(items, total, error)`，但前端只取第一个元素。当没有历史记录时，`result[0]`可能为`null`或`undefined`。

**修复方案**：
1. 修改前端代码，正确处理返回值
2. 添加空值检查
3. 改进错误提示

---

### 问题2：完整报告页面无法滚动

**问题定位**：
- 文件：`frontend/src/components/ai-stock-pick-fixed.vue`
- 行号：112-127
- 代码：`<div style="height: 100%; overflow-y: auto;">`

**根本原因**：
1. `MdPreview`组件可能设置了固定高度
2. 父容器的高度设置可能不正确
3. CSS flex布局可能影响了滚动行为

**修复方案**：
1. 调整MdPreview组件容器的高度计算
2. 确保滚动区域有明确的高度约束
3. 使用CSS flexbox优化布局
4. 添加`max-height`和`overflow`属性

---

### 问题3：股票信息解析不准确

**问题定位**：
- 文件：`backend/data/stock_pick_service.go`
- 方法：`parseRecommendationsFromContent` (886-974行)
- 文件：`data/skills/ai-stock-pick.md`

**根本原因**：
1. 解析器期望固定格式：`1. [股票代码] [股票名称] - [推荐理由]`
2. AI输出的格式不确定，可能导致解析失败
3. Prompt没有强制要求使用JSON或特定格式

**修复方案**：
1. 修改`ai-stock-pick.md`中的Prompt，要求AI返回结构化数据
2. 在Prompt中添加JSON格式的输出要求
3. 改进解析逻辑，支持更多格式
4. 添加格式验证和错误恢复机制

---

### 问题4：导出Markdown文件保存路径问题

**问题定位**：
- 文件：`backend/data/stock_pick_service.go`
- 方法：`ExportToMarkdown` (444-479行)

**根本原因**：
1. 方法只返回文件名，没有实际保存文件
2. 没有文件写入逻辑
3. 没有路径选择功能

**修复方案**：
1. 在Go中实现文件保存逻辑
2. 添加`runtime.DialogSaveFile`支持用户选择保存路径
3. 确保文件内容正确写入
4. 添加错误处理和用户反馈

---

## 实施计划

### Phase 1: 历史记录加载修复

**前端修改** (`frontend/src/components/ai-stock-pick-fixed.vue`)

```javascript
// 修改 loadHistory 函数
async function loadHistory() {
  loadingHistory.value = true
  try {
    const result = await GetStockPickReports(0, 20)
    // GetStockPickReports 返回 [items, total, error]
    // 正确处理返回值
    const items = result[0] || []
    const total = result[1] || 0

    historyList.value = items

    if (items.length === 0) {
      message.info('暂无历史记录')
    }
  } catch (error) {
    console.error('加载历史记录失败:', error)
    message.error('加载历史记录失败: ' + error)
  } finally {
    loadingHistory.value = false
  }
}
```

**测试验证**：
1. 无历史记录时不应报错
2. 有历史记录时正确显示
3. 错误处理正常工作

---

### Phase 2: 完整报告页面滚动修复

**前端样式修改** (`frontend/src/components/ai-stock-pick-fixed.vue`)

```vue
<!-- 修改推荐结果区域的容器 -->
<div style="height: 100%; display: flex; flex-direction: column;">
  <!-- ... -->

  <!-- 右侧推荐结果区域 -->
  <n-col :span="12" style="height: 100%; display: flex; flex-direction: column;">
    <n-card title="推荐结果" style="height: 100%; display: flex; flex-direction: column;">
      <!-- header -->

      <!-- 内容区域 - 确保可以滚动 -->
      <div style="flex: 1; overflow-y: auto; min-height: 0;">
        <!-- 完整报告视图 -->
        <div v-if="viewMode === 'full'" class="full-report-container">
          <MdPreview :modelValue="fullReport" :theme="darkTheme ? 'dark' : 'light'"/>
        </div>

        <!-- 简洁列表视图 -->
        <div v-else>
          <n-data-table
            :columns="columns"
            :data="simpleRecommendations"
            :pagination="pagination"
            size="small"
          />
        </div>
      </div>
    </n-card>
  </n-col>
</div>
```

**CSS样式添加**：

```css
<style scoped>
.n-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.n-card > :deep(.n-card__content) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.full-report-container {
  overflow-y: auto;
  max-height: 100%;
}

.full-report-container :deep(.md-editor-preview) {
  overflow: visible;
  height: auto;
}
</style>
```

---

### Phase 3: 股票信息解析优化

**Prompt修改** (`data/skills/ai-stock-pick.md`)

在"输出要求"部分添加JSON格式要求：

```markdown
## 辅助输出格式

为了确保推荐股票信息能被正确解析，请在最后添加以下JSON格式的推荐数据：

```json
{
  "recommendations": [
    {
      "stock_code": "sh600000",
      "stock_name": "浦发银行",
      "current_price": 10.50,
      "price_change": 2.35,
      "reason": "稳健的基本面和良好的分红政策",
      "target_price": 12.00,
      "target_change_percent": 14.29,
      "technical_analysis": "股价突破60日均线...",
      "fundamental_analysis": "当前PE 8.5倍，估值偏低...",
      "risk_level": "low",
      "risk_tips": "注意宏观政策变化",
      "score": 85
    }
  ]
}
```

**重要说明**：
- stock_code: 必须包含sh/sz/hk/us前缀
- current_price: 当前价格（数字）
- price_change: 涨跌幅（数字，百分比）
- target_change_percent: 目标涨幅（数字，百分比）
- score: 综合评分（0-100）
```

**解析逻辑增强** (`backend/data/stock_pick_service.go`)

```go
// extractJSONFromContent 从AI响应中提取JSON数据
func extractJSONFromContent(content string) ([]models.RecommendationItem, bool) {
    // 查找JSON块
    jsonStart := strings.Index(content, "```json")
    if jsonStart == -1 {
        jsonStart = strings.Index(content, "{")
    }
    if jsonStart == -1 {
        return nil, false
    }

    jsonEnd := strings.LastIndex(content, "```")
    if jsonEnd == -1 || jsonEnd < jsonStart {
        jsonEnd = strings.LastIndex(content, "}")
    }
    if jsonEnd == -1 {
        return nil, false
    }

    // 提取JSON字符串
    jsonStr := content[jsonStart:jsonEnd]
    jsonStr = strings.TrimPrefix(jsonStr, "```json")
    jsonStr = strings.TrimPrefix(jsonStr, "```")
    jsonStr = strings.TrimSpace(jsonStr)

    // 解析JSON
    var result map[string]interface{}
    if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
        logger.SugaredLogger.Warnf("解析JSON失败: %v", err)
        return nil, false
    }

    // 转换为RecommendationItem
    if recs, ok := result["recommendations"].([]interface{}); ok {
        var recommendations []models.RecommendationItem
        for _, rec := range recs {
            if recMap, ok := rec.(map[string]interface{}); ok {
                item := models.RecommendationItem{
                    StockCode: getString(recMap, "stock_code"),
                    StockName: getString(recMap, "stock_name"),
                    CurrentPrice: getFloat(recMap, "current_price"),
                    PriceChange: getFloat(recMap, "price_change"),
                    TargetPrice: getFloat(recMap, "target_price"),
                    TargetChangePercent: getFloat(recMap, "target_change_percent"),
                    Reason: getString(recMap, "reason"),
                    TechnicalAnalysis: getString(recMap, "technical_analysis"),
                    FundamentalAnalysis: getString(recMap, "fundamental_analysis"),
                    RiskTips: getString(recMap, "risk_tips"),
                    RiskLevel: getString(recMap, "risk_level"),
                    Score: getFloat(recMap, "score"),
                    IsFollowed: false, // 后续更新
                }
                recommendations = append(recommendations, item)
            }
        }
        return recommendations, true
    }

    return nil, false
}

// 辅助函数
func getString(m map[string]interface{}, key string) string {
    if val, ok := m[key]; ok {
        if s, ok := val.(string); ok {
            return s
        }
    }
    return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
    if val, ok := m[key]; ok {
        switch v := val.(type) {
        case float64:
            return v
        case float32:
            return float64(v)
        case int:
            return float64(v)
        case int64:
            return float64(v)
        }
    }
    return 0
}
```

**修改parseRecommendationsFromContent方法**：

```go
func (s *StockPickService) parseRecommendationsFromContent(content string) []models.RecommendationItem {
    logger.SugaredLogger.Infof("开始解析推荐内容，内容长度: %d", len(content))

    // 优先尝试解析JSON格式
    if recs, ok := extractJSONFromContent(content); ok {
        logger.SugaredLogger.Infof("JSON解析成功，推荐数量: %d", len(recs))

        // 更新关注状态
        for i := range recs {
            recs[i].IsFollowed = s.CheckStockFollowed(recs[i].StockCode)
        }
        return recs
    }

    // JSON解析失败，使用原有的文本解析逻辑
    logger.SugaredLogger.Warn("JSON解析失败，使用文本解析")
    return s.parseRecommendationsFromText(content)
}

// parseRecommendationsFromText 从文本格式解析推荐股票
func (s *StockPickService) parseRecommendationsFromText(content string) []models.RecommendationItem {
    // 原有的parseRecommendationsFromContent逻辑
    // ...（保持不变）
}
```

---

### Phase 4: 导出Markdown文件保存

**后端修改** (`app.go`)

```go
// ExportStockPickReport exports a report to markdown file
func (a *App) ExportStockPickReport(reportID uint, format string) (string, error) {
    service := data.NewStockPickService(a.ctx)

    // 生成Markdown内容
    content, err := service.ExportToMarkdownContent(reportID)
    if err != nil {
        return "", err
    }

    // 获取报告信息用于生成默认文件名
    report, err := service.GetReport(reportID)
    if err != nil {
        return "", err
    }

    // 生成默认文件名
    timestamp := time.Now().Format("20060102-150405")
    defaultFileName := fmt.Sprintf("stock-pick-report-%d-%s.md", reportID, timestamp)

    // 打开保存文件对话框
    selection, err := runtime.DialogSaveFile(a.ctx, runtime.SaveDialogOptions{
        Title:           "保存荐股报告",
        DefaultFilename:  defaultFileName,
        Filters: []runtime.FileFilter{
            {
                DisplayName: "Markdown Files (*.md)",
                Pattern:     "*.md",
            },
            {
                DisplayName: "Text Files (*.txt)",
                Pattern:     "*.txt",
            },
            {
                DisplayName: "All Files (*.*)",
                Pattern:     "*.*",
            },
        },
    })

    if err != nil {
        return "", fmt.Errorf("打开保存对话框失败: %w", err)
    }

    if selection == "" {
        return "", errors.New("用户取消了保存")
    }

    // 写入文件
    if err := os.WriteFile(selection, []byte(content), 0644); err != nil {
        return "", fmt.Errorf("写入文件失败: %w", err)
    }

    return fmt.Sprintf("已导出报告: %s", selection), nil
}
```

**服务层修改** (`backend/data/stock_pick_service.go`)

```go
// ExportToMarkdownContent 生成Markdown内容
func (s *StockPickService) ExportToMarkdownContent(reportID uint) (string, error) {
    report, err := s.GetReport(reportID)
    if err != nil {
        return "", err
    }

    var sb strings.Builder
    sb.WriteString("# AI荐股报告\n\n")
    sb.WriteString(fmt.Sprintf("**生成时间**: %s\n\n", report.CreatedAt.Format("2006-01-02 15:04:05")))
    sb.WriteString(fmt.Sprintf("**选股需求**: %s\n\n", report.UserQuery))

    if report.MarketAnalysis != "" {
        sb.WriteString("## 市场环境分析\n\n")
        sb.WriteString(report.MarketAnalysis)
        sb.WriteString("\n\n")
    }

    if report.FilterLogic != "" {
        sb.WriteString("## 筛选逻辑\n\n")
        sb.WriteString(report.FilterLogic)
        sb.WriteString("\n\n")
    }

    sb.WriteString("## 推荐股票\n\n")

    // 尝试解析并格式化推荐列表
    var recommendations []models.RecommendationItem
    if err := json.Unmarshal([]byte(report.Recommendations), &recommendations); err == nil {
        // 格式化推荐股票
        for i, rec := range recommendations {
            sb.WriteString(fmt.Sprintf("### %d. %s (%s)\n\n", i+1, rec.StockName, rec.StockCode))
            sb.WriteString(fmt.Sprintf("- **当前价格**: %.2f元\n", rec.CurrentPrice))
            sb.WriteString(fmt.Sprintf("- **涨跌幅**: %.2f%%\n", rec.PriceChange))

            if rec.TargetPrice > 0 {
                sb.WriteString(fmt.Sprintf("- **目标价位**: %.2f元\n", rec.TargetPrice))
            }
            if rec.TargetChangePercent > 0 {
                sb.WriteString(fmt.Sprintf("- **目标涨幅**: %.2f%%\n", rec.TargetChangePercent))
            }
            if rec.Score > 0 {
                sb.WriteString(fmt.Sprintf("- **综合评分**: %.1f/100\n", rec.Score))
            }

            if rec.Reason != "" {
                sb.WriteString(fmt.Sprintf("\n**推荐理由**:\n%s\n\n", rec rec.Reason))
            }
            if rec.TechnicalAnalysis != "" {
                sb.WriteString(fmt.Sprintf("**技术面分析**:\n%s\n\n", rec.TechnicalAnalysis))
            }
            if rec.FundamentalAnalysis != "" {
                sb.WriteString(fmt.Sprintf("**基本面分析**:\n%s\n\n", rec.FundamentalAnalysis))
            }
            if rec.RiskTips != "" {
                sb.WriteString(fmt.Sprintf("**风险提示**:\n%s\n\n", rec.RiskTips))
            }

            sb.WriteString("---\n\n")
        }
    } else {
        // 解析失败，直接使用原始内容
        sb.WriteString(report.Recommendations)
    }

    sb.WriteString("---\n\n")
    sb.WriteString("*本报告由AI智能分析生成，仅供参考，不构成投资建议。股市有风险，投资需谨慎。*\n")

    return sb.String(), nil
}

// ExportToMarkdown 保留原方法用于兼容
func (s *StockPickService) ExportToMarkdown(reportID uint) (string, error) {
    _, err := s.ExportToMarkdownContent(reportID)
    if err != nil {
        return "", err
    }
    timestamp := time.Now().Format("20060102-150405")
    fileName := fmt.Sprintf("stock-pick-report-%d-%s.md", reportID, timestamp)
    return fileName, nil
}
```

---

## 依赖关系

```
Phase 1 (历史记录修复)
    ├── 无依赖

Phase 2 (滚动修复)
    ├── 无依赖

Phase 3 (解析优化)
    ├── 依赖: 无

Phase 4 (导出优化)
    ├── 依赖: Phase 3 (解析优化，因为需要格式化推荐列表)
```

---

## 风险评估

| 风险项 | 级别 | 缓解措施 |
|--------|------|----------|
| 历史记录修复可能破坏现有数据展示 | 低 | 添加向后兼容检查 |
| 滚动修复可能影响其他页面布局 | 中 | 使用scoped CSS限制影响范围 |
| JSON解析可能被旧AI响应失败 | 低 | 保留文本解析作为fallback |
| 导出功能依赖Wails运行时 | 中 | 添加错误处理和降级方案 |

---

## 测试计划

### Phase 1 测试
- [ ] 无历史记录时不报错
- [ ] 有历史记录时正确显示
- [ ] 分页功能正常（如已实现）

### Phase 2 测试
- [ ] 长报告内容可以上下滚动
- [ ] 滚动条正常显示
- [ ] 响应式布局在不同窗口尺寸下正常

### Phase 3 测试
- [ ] 新的AI响应能正确解析JSON格式
- [ ] 旧的AI响应仍能通过文本解析
- [ ] 简洁列表显示正确的股票信息

### Phase 4 测试
- [ ] 导出对话框正常弹出
- [ ] 可以选择保存路径
- [ ] 文件正确保存到指定位置
- [ ] 文件内容格式正确

---

## 实施顺序建议

由于各阶段相互独立，建议按以下顺序实施：

1. **Phase 1**: 快速修复，立即见效
2. **Phase 2**: UI改进，提升用户体验
3. **Phase 3**: 核心功能优化，需要测试AI响应
4. **Phase 4**: 功能增强，依赖Phase 3

---

## 完成标准

- [ ] 所有4个问题已修复
- [ ] 单元测试通过
- [ ] 手动测试验证功能正常
- [ ] 无新的bug引入
- [ ] 代码符合Go/Vue编码规范
