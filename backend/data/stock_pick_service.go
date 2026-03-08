package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	"go-stock/backend/toolexec"
	"os"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
)

// ToolProvider 工具提供者接口，用于从外部获取工具列表
type ToolProvider interface {
	GetTools() []Tool
}

// StockPickService AI荐股服务
type StockPickService struct {
	ctx           context.Context
	AiTools       []Tool
	AiInvokeTools map[string]tool.InvokableTool
}

func NewStockPickServiceWithInvokeTool(ctx context.Context, tools []Tool, invokeTools map[string]tool.InvokableTool) *StockPickService {
	return &StockPickService{ctx: ctx, AiTools: tools, AiInvokeTools: invokeTools}
}

// SetToolProvider 设置工具提供者
func (s *StockPickService) SetToolProvider(tools []Tool) {
	s.AiTools = tools
}

// StockPickRequest 荐股请求
type StockPickRequest struct {
	UserQuery  string `json:"user_query"`
	AIConfigID uint   `json:"ai_config_id"`
}

// StockPickResponse 荐股响应
type StockPickResponse struct {
	ReportID uint   `json:"report_id"`
	StreamID string `json:"stream_id"`
}

// ProcessStockPick 处理荐股流程
func (s *StockPickService) ProcessStockPick(req StockPickRequest, eventHandler func(eventType string, data interface{})) (*StockPickResponse, error) {
	// 1. 创建报告记录
	report := &models.StockPickReport{
		UserQuery:       req.UserQuery,
		QuerySummary:    generateQuerySummary(req.UserQuery),
		AIConfigID:      req.AIConfigID,
		Status:          "processing",
		Recommendations: []models.RecommendationItem{}, // 初始化为空切片，避免NULL约束错误
	}
	if err := db.Dao.Create(report).Error; err != nil {
		logger.SugaredLogger.Errorf("创建荐股报告失败: %v", err)
		return nil, err
	}

	// 2. 获取AI配置
	settingConfig := GetSettingConfig()
	var aiConfig *AIConfig
	if req.AIConfigID > 0 {
		aiConfig, _ = lo.Find(settingConfig.AiConfigs, func(item *AIConfig) bool {
			return req.AIConfigID == item.ID
		})
	}
	if aiConfig == nil && len(settingConfig.AiConfigs) > 0 {
		aiConfig = settingConfig.AiConfigs[0]
	}
	if aiConfig == nil {
		report.Status = "failed"
		report.Error = "未配置AI服务"
		db.Dao.Save(report)
		return nil, errors.New("未配置AI服务")
	}

	// 3. 加载Skill Prompt模板
	skillPrompt, err := s.loadSkillPrompt()
	if err != nil {
		logger.SugaredLogger.Errorf("加载Skill Prompt失败: %v", err)
		report.Status = "failed"
		report.Error = err.Error()
		db.Dao.Save(report)
		return nil, err
	}

	// 4. 创建OpenAi实例
	openAI := &OpenAi{
		ctx:          s.ctx,
		BaseUrl:      aiConfig.BaseUrl,
		ApiKey:       aiConfig.ApiKey,
		Model:        aiConfig.ModelName,
		MaxTokens:    aiConfig.MaxTokens,
		Temperature:  aiConfig.Temperature,
		TimeOut:      aiConfig.TimeOut,
		Prompt:       skillPrompt,
		CrawlTimeOut: settingConfig.CrawlTimeOut,
		KDays:        settingConfig.KDays,
		BrowserPath:  settingConfig.BrowserPath,
	}

	//如果超时时间未设置，默认为300秒
	if openAI.TimeOut <= 0 {
		openAI.TimeOut = 300
	}
	if openAI.CrawlTimeOut <= 0 {
		openAI.CrawlTimeOut = 60
	}
	if openAI.KDays < 30 {
		openAI.KDays = 120
	}

	// 5. 发送开始事件
	if eventHandler != nil {
		eventHandler("start", map[string]interface{}{
			"report_id": report.ID,
			"message":   "开始分析市场数据...",
		})
	}

	// 6. 构建消息列表
	msg := s.buildMessages(skillPrompt, req.UserQuery)

	// 7. 调用AI进行荐股分析
	ch := make(chan map[string]any, 512)
	doneChan := make(chan bool, 1) // 用于同步解析完成

	// 创建工具执行器
	executor := toolexec.NewToolExecutor(s.AiInvokeTools)

	go func() {
		defer close(ch)
		AskAiWithTools(openAI, nil, msg, ch, req.UserQuery, s.AiTools, executor)
	}()

	// 8. 处理流式响应 - 在同一个 goroutine 中解析推荐数据
	go func() {
		var fullResponse strings.Builder
		for data := range ch {
			if eventHandler != nil {
				if strData, ok := data["content"].(string); ok && strData != "" {
					fullResponse.WriteString(strData)
					eventHandler("stream", data)
				}
			}
		}

		// 保存工具调用结果
		if toolResultsJSON, err := executor.GetResultsJSON(); err == nil {
			report.ToolCallResults = toolResultsJSON
			logger.SugaredLogger.Infof("保存工具调用结果，长度: %d", len(toolResultsJSON))
		} else {
			logger.SugaredLogger.Warnf("序列化工具调用结果失败: %v", err)
		}

		// AI响应完成后，解析推荐结果并更新数据库
		logger.SugaredLogger.Infof("AI分析完成，响应长度: %d", fullResponse.Len())
		if fullResponse.Len() > 0 {
			// 传递工具调用结果给解析函数
			if err := s.parseAndUpdateRecommendations(report, fullResponse.String(), executor.GetResults()); err != nil {
				logger.SugaredLogger.Errorf("解析和更新推荐数据失败: %v", err)
				// 即使解析失败，也要设置完成状态，避免报告一直处于 processing 状态
				report.Error = err.Error()
			}
		} else {
			logger.SugaredLogger.Warn("AI响应为空，设置状态为失败")
			report.Status = "failed"
			report.Error = "AI响应为空"
		}

		// 9. 更新报告状态和AI模型信息
		report.Status = "completed"
		report.AIModel = aiConfig.ModelName
		if err := db.Dao.Save(report).Error; err != nil {
			logger.SugaredLogger.Errorf("更新荐股报告状态失败: %v", err)
		}

		logger.SugaredLogger.Infof("荐股报告 %d 处理完成", report.ID)
		doneChan <- true
	}()

	return &StockPickResponse{
		ReportID: report.ID,
		StreamID: fmt.Sprintf("stock-pick-%d", report.ID),
	}, nil
}

// buildMessages 构建消息列表
func (s *StockPickService) buildMessages(skillPrompt, userQuery string) []map[string]interface{} {
	msg := []map[string]interface{}{
		{
			"role":    "system",
			"content": skillPrompt,
		},
	}

	// 添加当前时间
	msg = append(msg, map[string]interface{}{
		"role":    "user",
		"content": "当前时间",
	})
	msg = append(msg, map[string]interface{}{
		"role":    "assistant",
		"content": "当前本地时间是:" + time.Now().Format("2006-01-02 15:04:05"),
	})

	// 添加市场指数行情
	msg = append(msg, map[string]interface{}{
		"role":    "user",
		"content": "当前市场指数行情",
	})
	msg = append(msg, map[string]interface{}{
		"role":    "assistant",
		"content": "当前市场指数行情情况如下：\n" + s.getMarketIndexInfo(),
	})

	// 添加用户查询
	msg = append(msg, map[string]interface{}{
		"role":    "user",
		"content": userQuery,
	})

	return msg
}

// getMarketIndexInfo 获取市场指数信息
func (s *StockPickService) getMarketIndexInfo() string {
	var sb strings.Builder
	sb.WriteString(GetZSInfo("上证指数", "sh000001", 30) + "\n")
	sb.WriteString(GetZSInfo("深证成指", "sz399001", 30) + "\n")
	sb.WriteString(GetZSInfo("创业板指数", "sz399006", 30) + "\n")
	sb.WriteString(GetZSInfo("科创50", "sh000688", 30) + "\n")
	sb.WriteString(GetZSInfo("沪深300指数", "sh000300", 30) + "\n")
	return sb.String()
}

// SaveReport 保存荐股报告
func (s *StockPickService) SaveReport(report *models.StockPickReport) error {
	return db.Dao.Save(report).Error
}

// GetReports 获取荐股报告列表
func (s *StockPickService) GetReports(offset, limit int) ([]models.StockPickReport, int64, error) {
	// 初始化为空数组，避免确保不返回 nil
	reports := make([]models.StockPickReport, 0)
	var total int64

	logger.SugaredLogger.Infof("GetReports 调用: offset=%d, limit=%d", offset, limit)

	// 先获取总数
	if err := db.Dao.Model(&models.StockPickReport{}).Count(&total).Error; err != nil {
		logger.SugaredLogger.Errorf("获取报告总数失败: %v", err)
		return reports, total, err
	}
	logger.SugaredLogger.Infof("报告总数: %d", total)

	// 查询分页数据
	err := db.Dao.Order("created_at DESC").Offset(offset).Limit(limit).Find(&reports).Error
	if err != nil {
		logger.SugaredLogger.Errorf("查询报告列表失败: %v", err)
		return reports, total, err
	}

	logger.SugaredLogger.Infof("查询到 %d 条报告", len(reports))
	for i, r := range reports {
		logger.SugaredLogger.Debugf("报告[%d]: ID=%d, QuerySummary=%s, Status=%s", i, r.ID, r.QuerySummary, r.Status)
	}

	return reports, total, err
}

// GetReport 获取单个报告
func (s *StockPickService) GetReport(id uint) (*models.StockPickReport, error) {
	var report models.StockPickReport
	logger.SugaredLogger.Infof("GetReport: 查询报告ID=%d, 表名=%s", id, report.TableName())

	// 使用调试日志查看实际执行的SQL
	result := db.Dao.Where("id = ?", id).First(&report)
	logger.SugaredLogger.Debugf("GetReport: 查询结果 rows_affected=%d, error=%v", result.RowsAffected, result.Error)

	if result.Error != nil {
		logger.SugaredLogger.Errorf("GetReport: 查询失败 ID=%d, error=%v", id, result.Error)

		// 如果是记录未找到错误，尝试查找包括软删除在内的所有记录
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.SugaredLogger.Warnf("GetReport: 记录未找到，尝试查询包括软删除的记录 ID=%d", id)

			// 使用 Unscoped 查询所有记录（包括软删除的）
			var deletedReport models.StockPickReport
			deletedResult := db.Dao.Unscoped().Where("id = ?", id).First(&deletedReport)
			if deletedResult.Error == nil {
				logger.SugaredLogger.Warnf("GetReport: 发现软删除记录 ID=%d, deleted_at=%v", id, deletedReport.DeletedAt)
				return nil, fmt.Errorf("记录ID %d 已被软删除", id)
			} else {
				logger.SugaredLogger.Errorf("GetReport: 包括软删除也找不到记录 ID=%d, error=%v", id, deletedResult.Error)
			}

			// 检查表中的所有记录数量
			var totalCount int64
			db.Dao.Model(&models.StockPickReport{}).Count(&totalCount)
			logger.SugaredLogger.Infof("GetReport: 表中总记录数=%d", totalCount)

			// 检查包括软删除的总数
			var totalWithDeleted int64
			db.Dao.Unscoped().Model(&models.StockPickReport{}).Count(&totalWithDeleted)
			logger.SugaredLogger.Infof("GetReport: 表中总记录数(包括软删除)=%d", totalWithDeleted)
		}

		return nil, result.Error
	}

	logger.SugaredLogger.Infof("GetReport: 查询成功 ID=%d, QuerySummary=%s", report.ID, report.QuerySummary)
	return &report, nil
}

// DeleteReport 删除报告
func (s *StockPickService) DeleteReport(id uint) error {
	return db.Dao.Delete(&models.StockPickReport{}, id).Error
}

// loadSkillPrompt 加载Skill Prompt模板
func (s *StockPickService) loadSkillPrompt() (string, error) {
	// 尝试从data/skills/ai-stock-pick.md读取
	skillPath := "data/skills/ai-stock-pick.md"

	// 读取文件内容
	bytes, err := os.ReadFile(skillPath)
	if err != nil {
		logger.SugaredLogger.Warnf("读取Skill Prompt文件失败，使用默认Prompt: %v", err)
		return s.getDefaultPrompt(), nil
	}

	prompt := string(bytes)
	if strings.TrimSpace(prompt) == "" {
		return s.getDefaultPrompt(), nil
	}

	return prompt, nil
}

// getDefaultPrompt 获取默认Prompt
func (s *StockPickService) getDefaultPrompt() string {
	return `【角色设定】
你是一位拥有20年实战经验的顶级证券分析师和AI选股专家，精通技术分析、基本面分析、市场心理学和量化交易。擅长发现成长股、捕捉行业轮动机会，在牛熊市中都能保持稳定收益。你的风格是价值投资与技术择时相结合，注重风险控制。

【核心功能】
市场分析维度：
- 宏观经济（GDP/CPI/货币政策）
- 行业景气度（产业链/政策红利/技术革新）
- 资金流向（主力资金/北向资金/融资余额）

个股三维诊断：
- 基本面：PE/PB/ROE/现金流/护城河
- 技术面：K线形态/均线系统/量价关系/指标背离
- 资金面：主力动向/北向资金/融资余额/大宗交易

【可用数据分析工具】
你必须充分利用以下工具进行深入分析，按重要性排序：

1. ChoiceStockByIndicators - 【首选】根据自然语言筛选股票，获取技术指标（MACD、KDJ、RSI、BOLL、均线等）和财务数据
   - 使用示例："股票名称;MACD,KDJ,RSI,BOLL,5日均线,10日均线,30日均线,60日均线,成交量"
   - 可以获取：技术指标、PE、PB、ROE、净利润增长率等核心数据

2. QueryStockKLine - 获取K线数据进行技术分析
   - 获取最近60-90天的日K线数据
   - 分析：趋势形态、支撑压力位、量价配合、均线系统

3. GetFinancialReport - 获取财务报表进行基本面分析
   - 获取资产负债表、利润表、现金流量表
   - 分析：盈利能力、偿债能力、成长性、现金流状况

4. QueryShareholderCount - 分析股东人数和筹码集中度
   - 获取最近4-8个季度的股东人数变化
   - 判断：主力吸筹/出货、筹码集中/分散

5. QueryStockNewsTool - 获取相关新闻和资讯
   - 搜索股票相关的最新市场资讯
   - 分析：舆情热度、利好利空消息

6. GetIndustryResearchReport - 获取行业研究报告（如适用）
   - 获取股票所属行业的研报
   - 分析：行业景气度、竞争格局、发展趋势

7. QueryEconomicData - 获取宏观经济数据（如需要）
   - GDP、CPI、PPI、PMI等宏观指标
   - 分析：宏观经济环境对市场的影响

【工具使用策略】
筛选阶段：
- 优先使用 ChoiceStockByIndicators 获取批量股票的综合数据
- 从筛选结果中选择5-10只最符合要求的股票进行深度分析

深度分析阶段（对每只候选股票）：
1. 必用：QueryStockKLine - 技术面分析
2. 必用：GetFinancialReport - 基本面分析
3. 必用：QueryShareholderCount - 筹码分析
4. 可选：QueryStockNewsTool - 舆情分析
5. 可选：GetIndustryResearchReport - 行业分析

【工作流程】
第一阶段：市场环境分析
- 分析当前大盘走势
- 分析热门板块资金流向
- 识别市场热点和风格特征

第二阶段：候选股票筛选
- 使用 ChoiceStockByIndicators 筛选候选股票
- 初步筛选出20-30只候选股票

第三阶段：深度分析（关键！必须调用工具）
对每只候选股票，按顺序执行：
1. 调用 QueryStockKLine 获取K线数据，分析技术面
2. 调用 GetFinancialReport 获取财报，分析基本面
3. 调用 QueryShareholderCount 分析筹码结构
4. 必要时调用 QueryStockNewsTool 了解最新资讯
5. 必要时调用 GetIndustryResearchReport 了解行业情况

第四阶段：综合评分与推荐
- 整合所有工具返回的数据
- 从技术面、基本面、资金面、舆情面综合评分
- 筛选出3-5只最具潜力的股票
- 生成完整的分析报告

【输出要求】
请按照以下格式输出推荐股票列表：

## 推荐股票

### 1. [股票名称] ([股票代码])

**推荐理由**：
[详细的推荐理由，说明为什么推荐这只股票]

**板块概念**：
[股票所属板块或行业概念]

- **当前价格**：XX.XX
- **涨跌幅**：XX.XX%
- **目标价位**：XX.XX
- **目标涨幅**：XX%
- **综合评分**：XX/100
- **买卖建议**：[买入/观望/卖出]

**技术面分析**：
[结合K线数据的详细分析，包括趋势、支撑位、压力位、指标信号等]

**基本面分析**：
[结合财报的详细分析，包括PE、PB、ROE、业绩增长等]

**风险提示**：
[具体的风险提示内容]

---

### 2. [股票名称] ([股票代码])
[同上格式...]

【数据分析要求】
- 技术面分析必须基于实际K线数据，提到具体的价格点位、均线数值、成交量变化
- 基本面分析必须引用具体的财务指标数值，如PE、PB、ROE、净利润增长率等
- 筹码分析要说明股东人数变化趋势和筹码集中程度
- 所有分析都要标注数据来源（如"根据K线数据分析"、"根据财务报告"等）

【注意事项】
- 严格遵守监管要求，不做收益承诺
- 区分投资建议与市场观点
- 重要数据标注来源及更新时间
- 根据用户认知水平调整专业术语密度
- 提供风险提示，控制仓位建议
- 必须实际调用工具获取数据，不要凭空捏造分析
- 如果某个工具调用失败，可以跳过该维度，但要说明"该维度数据暂不可用"
`
}

// generateQuerySummary 生成需求摘要
func generateQuerySummary(query string) string {
	// 截取前50个字符
	if len(query) > 50 {
		return query[:50] + "..."
	}
	return query
}

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

	// 使用已解析的推荐数据
	recommendations := report.Recommendations
	if len(recommendations) > 0 {
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
			if rec.TradeSuggestion != "" {
				sb.WriteString(fmt.Sprintf("- **买卖建议**: %s\n", rec.TradeSuggestion))
			}

			if rec.Reason != "" {
				sb.WriteString(fmt.Sprintf("\n**推荐理由**:\n%s\n\n", rec.Reason))
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
		// 没有推荐数据
		sb.WriteString("暂无推荐股票")
	}

	sb.WriteString("---\n\n")

	// 添加 AI 完整分析（Result 字段）
	if report.Result != "" {
		sb.WriteString("## AI 完整分析\n\n")
		sb.WriteString("> 以下为 AI 生成的原始分析内容：\n\n")
		sb.WriteString(report.Result)
		sb.WriteString("\n\n")
		sb.WriteString("---\n\n")
	}

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

// UpdateReportWithRecommendations 更新报告推荐数据
func (s *StockPickService) UpdateReportWithRecommendations(reportID uint, marketAnalysis, filterLogic string, recommendations []models.RecommendationItem) error {
	report, err := s.GetReport(reportID)
	if err != nil {
		return err
	}

	report.MarketAnalysis = marketAnalysis
	report.FilterLogic = filterLogic
	report.TotalScanned = len(recommendations) * 10 // 估算值
	report.CandidatesCount = len(recommendations)
	report.Recommendations = recommendations

	return db.Dao.Save(report).Error
}

// GetRecommendations 获取报告的推荐列表
func (s *StockPickService) GetRecommendations(reportID uint) ([]models.RecommendationItem, error) {
	report, err := s.GetReport(reportID)
	if err != nil {
		return nil, err
	}

	// 直接返回已解析的推荐数据
	if len(report.Recommendations) > 0 {
		return report.Recommendations, nil
	}

	// 如果没有推荐数据，尝试从Result字段解析
	logger.SugaredLogger.Infof("没有推荐数据，尝试从markdown解析推荐数据")
	return s.parseRecommendationsFromMarkdown(report), nil
}

// UpdateRecommendationFollowStatus 更新关注状态
func (s *StockPickService) UpdateRecommendationFollowStatus(reportID uint, stockCode string, isFollowed bool) error {
	report, err := s.GetReport(reportID)
	if err != nil {
		return err
	}

	// 直接使用已解析的推荐数据
	recommendations := report.Recommendations
	if len(recommendations) == 0 {
		// 如果没有推荐数据，从markdown解析
		logger.SugaredLogger.Infof("没有推荐数据，从markdown更新关注状态")
		recommendations = s.parseRecommendationsFromMarkdown(report)
	}

	// 更新关注状态
	for i := range recommendations {
		if recommendations[i].StockCode == stockCode {
			recommendations[i].IsFollowed = isFollowed
			break
		}
	}

	// 保存更新后的推荐数据
	report.Recommendations = recommendations
	return db.Dao.Save(report).Error
}

// GetFollowedStockCodes 获取已关注的股票代码
func (s *StockPickService) GetFollowedStockCodes() ([]string, error) {
	var codes []string
	followedStocks := NewStockDataApi().GetFollowList(0)
	if followedStocks != nil {
		for _, stock := range *followedStocks {
			codes = append(codes, stock.StockCode)
		}
	}
	return codes, nil
}

// CheckStockFollowed 检查股票是否已关注
func (s *StockPickService) CheckStockFollowed(stockCode string) bool {
	_, err := NewStockDataApi().GetFollowedStockByStockCode(stockCode)
	return err == nil
}

// ParseRecommendationsFromAIContent 从AI内容解析推荐股票
func (s *StockPickService) ParseRecommendationsFromAIContent(content string) ([]models.RecommendationItem, error) {
	var recommendations []models.RecommendationItem

	// 这里可以添加解析逻辑，从AI输出中提取结构化的推荐数据
	// 简化版：如果AI输出格式不规范，返回空列表

	// TODO: 实现更复杂的解析逻辑
	// 可以使用正则表达式或自然语言处理来提取股票代码和相关信息

	return recommendations, nil
}

// GetMarketData 获取市场数据（供AI分析使用）
func (s *StockPickService) GetMarketData() map[string]interface{} {
	data := make(map[string]interface{})

	// 获取主要指数
	data["market_indices"] = map[string]string{
		"sh000001": GetZSInfo("上证指数", "sh000001", 30),
		"sz399001": GetZSInfo("深证成指", "sz399001", 30),
		"sz399006": GetZSInfo("创业板指数", "sz399006", 30),
		"sh000688": GetZSInfo("科创50", "sh000688", 30),
	}

	// 获取宏观经济数据
	gdp := NewMarketNewsApi().GetGDP()
	if len(gdp.GDPResult.Data) > 0 {
		data["gdp"] = gdp.GDPResult.Data[0]
	}

	cpi := NewMarketNewsApi().GetCPI()
	if len(cpi.CPIResult.Data) > 0 {
		data["cpi"] = cpi.CPIResult.Data[0]
	}

	return data
}

// ValidateAIConfig 验证AI配置
func (s *StockPickService) ValidateAIConfig(aiConfigID uint) error {
	settingConfig := GetSettingConfig()
	if len(settingConfig.AiConfigs) == 0 {
		return errors.New("未配置AI服务")
	}

	if aiConfigID > 0 {
		found := lo.ContainsBy(settingConfig.AiConfigs, func(item *AIConfig) bool {
			return item.ID == aiConfigID
		})
		if !found {
			return errors.New("指定的AI配置不存在")
		}
	}

	return nil
}

// FormatMarkdown 格式化Markdown输出
func (s *StockPickService) FormatMarkdown(content string) string {
	// 将换行符转换为Markdown格式
	lines := strings.Split(content, "\n")
	var sb strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			// 如果行以数字和点开头，视为列表项
			if len(trimmed) > 0 && trimmed[0] >= '1' && trimmed[0] <= '9' && len(trimmed) > 1 && trimmed[1] == '.' {
				sb.WriteString(trimmed + "\n")
			} else {
				sb.WriteString(trimmed + "\n")
			}
		}
	}

	return sb.String()
}

// GetStockInfoByCode 根据股票代码获取股票信息
func (s *StockPickService) GetStockInfoByCode(stockCode string) (string, string, error) {
	stockData, err := NewStockDataApi().GetStockCodeRealTimeData(stockCode)
	if err != nil || len(*stockData) == 0 {
		return "", "", errors.New("获取股票信息失败")
	}

	stock := (*stockData)[0]
	stockName := RemoveAllNonDigitChar(stockCode)

	return stockName, stock.Price, nil
}

// FormatStockRecommendation 格式化股票推荐信息
func (s *StockPickService) FormatStockRecommendation(rec models.RecommendationItem) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("### %d. %s (%s)\n", rec.Rank, rec.StockName, rec.StockCode))
	sb.WriteString(fmt.Sprintf("- **当前价格**: %.2f\n", rec.CurrentPrice))
	sb.WriteString(fmt.Sprintf("- **涨跌幅**: %.2f%%\n", rec.PriceChange))

	if rec.TargetPrice > 0 {
		sb.WriteString(fmt.Sprintf("- **目标价位**: %.2f\n", rec.TargetPrice))
	}

	if rec.TargetChangePercent > 0 {
		sb.WriteString(fmt.Sprintf("- **目标涨幅**: %.2f%%\n", rec.TargetChangePercent))
	}

	if rec.Score > 0 {
		sb.WriteString(fmt.Sprintf("- **综合评分**: %.1f/100\n", rec.Score))
	}

	if rec.Reason != "" {
		sb.WriteString(fmt.Sprintf("- **推荐理由**: %s\n", rec.Reason))
	}

	if rec.TechnicalAnalysis != "" {
		sb.WriteString(fmt.Sprintf("- **技术面**: %s\n", rec.TechnicalAnalysis))
	}

	if rec.FundamentalAnalysis != "" {
		sb.WriteString(fmt.Sprintf("- **基本面**: %s\n", rec.FundamentalAnalysis))
	}

	if rec.RiskTips != "" {
		sb.WriteString(fmt.Sprintf("- **风险提示**: %s\n", rec.RiskTips))
	}

	sb.WriteString("\n")

	return sb.String()
}

// GetStockPickStats 获取荐股统计信息
func (s *StockPickService) GetStockPickStats() map[string]interface{} {
	var total, completed, failed int64

	db.Dao.Model(&models.StockPickReport{}).Count(&total)
	db.Dao.Model(&models.StockPickReport{}).Where("status = ?", "completed").Count(&completed)
	db.Dao.Model(&models.StockPickReport{}).Where("status = ?", "failed").Count(&failed)

	return map[string]interface{}{
		"total":     total,
		"completed": completed,
		"failed":    failed,
		"success_rate": func() float64 {
			if total == 0 {
				return 0
			}
			return float64(completed) / float64(total) * 100
		}(),
	}
}

// SendToolCallEvent 发送工具调用事件
func (s *StockPickService) SendToolCallEvent(toolName, status string) {
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "ai-stock-pick-tool", map[string]interface{}{
			"tool_name": toolName,
			"status":    status,
			"timestamp": time.Now().Unix(),
		})
	}
}

// StreamRecommendationUpdate 流式更新推荐结果
func (s *StockPickService) StreamRecommendationUpdate(reportID uint, content string) {
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "ai-stock-pick-update", map[string]interface{}{
			"report_id": reportID,
			"content":   content,
			"timestamp": time.Now().Unix(),
		})
	}
}

// ClearOldReports 清理旧的报告（保留最近30天）
func (s *StockPickService) ClearOldReports() error {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	result := db.Dao.Where("created_at < ?", thirtyDaysAgo).Delete(&models.StockPickReport{})
	if result.Error != nil {
		return result.Error
	}
	logger.SugaredLogger.Infof("清理了%d条旧的荐股报告", result.RowsAffected)
	return nil
}

// parseAndUpdateRecommendations 解析AI响应并更新到数据库
func (s *StockPickService) parseAndUpdateRecommendations(report *models.StockPickReport, content string, toolResults *models.ToolCallResultsCollection) error {
	// 确保 report 不为 nil
	if report == nil {
		return errors.New("report 参数为 nil")
	}

	logger.SugaredLogger.Infof("开始解析AI响应，报告ID: %d", report.ID)

	// 添加defer用于捕获panic
	defer func() {
		if r := recover(); r != nil {
			logger.SugaredLogger.Errorf("parseAndUpdateRecommendations 发生panic: %v", r)
		}
	}()

	// 检查输入参数
	if content == "" {
		logger.SugaredLogger.Warnf("报告ID %d 的AI响应内容为空", report.ID)
		// 即使内容为空，也保存报告状态为已完成，但设置错误信息
		report.Recommendations = nil
		report.Error = "AI响应内容为空"
		if err := db.Dao.Save(report).Error; err != nil {
			logger.SugaredLogger.Errorf("保存荐股报告失败: %v", err)
			return err
		}
		return nil
	}

	report.Result = content

	// 解析市场环境分析和筛选逻辑
	logger.SugaredLogger.Debugf("开始提取市场分析和筛选逻辑")
	marketAnalysis, filterLogic := extractMarketAndFilterInfo(content)
	logger.SugaredLogger.Debugf("市场分析长度: %d, 筛选逻辑长度: %d", len(marketAnalysis), len(filterLogic))

	// 解析推荐股票列表
	logger.SugaredLogger.Debugf("开始解析推荐股票列表")
	recommendations := s.parseRecommendationsFromContent(content)
	logger.SugaredLogger.Infof("解析到 %d 个推荐股票", len(recommendations))

	// 更新报告数据
	report.MarketAnalysis = marketAnalysis
	report.FilterLogic = filterLogic
	report.TotalScanned = len(recommendations) * 10 // 估算扫描数量
	report.CandidatesCount = len(recommendations)
	report.Recommendations = recommendations

	// 使用AI报告解析器解析结构化分析，传入工具调用结果
	logger.SugaredLogger.Infof("开始使用AI报告解析器解析结构化分析")
	parser := NewAIReportParser()
	parser.ParseBatchWithToolResults(content, report.Recommendations, toolResults)
	logger.SugaredLogger.Infof("AI报告解析器解析完成")

	// 保存到数据库
	logger.SugaredLogger.Debugf("开始保存到数据库")
	if err := db.Dao.Save(report).Error; err != nil {
		logger.SugaredLogger.Errorf("保存荐股报告失败: %v", err)
		return err
	}

	logger.SugaredLogger.Infof("荐股报告更新成功，推荐数量: %d", len(recommendations))

	// 发送事件到前端
	if s.ctx != nil {
		logger.SugaredLogger.Debugf("发送事件到前端")
		runtime.EventsEmit(s.ctx, "ai-stock-pick-update", map[string]interface{}{
			"report_id":        report.ID,
			"recommendations":  recommendations,
			"market_analysis":  marketAnalysis,
			"filter_logic":     filterLogic,
			"total_scanned":    report.TotalScanned,
			"candidates_count": report.CandidatesCount,
			"status":           "completed",
			"timestamp":        time.Now().Unix(),
		})
	}

	return nil
}

// safeTruncate 安全截取字符串，防止panic
func safeTruncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// extractMarketAndFilterInfo 从内容中提取市场分析和筛选逻辑
func extractMarketAndFilterInfo(content string) (marketAnalysis, filterLogic string) {
	lines := strings.Split(content, "\n")

	var marketBuilder strings.Builder
	var filterBuilder strings.Builder
	var inMarketSection bool
	var inFilterSection bool
	var inRecommendationSection bool

	for _, line := range lines {
		lowerLine := strings.ToLower(strings.TrimSpace(line))

		// 检测章节开始
		if strings.Contains(lowerLine, "市场环境分析") || strings.Contains(lowerLine, "市场分析") {
			inMarketSection = true
			inFilterSection = false
			continue
		}
		if strings.Contains(lowerLine, "筛选逻辑") || strings.Contains(lowerLine, "筛选条件") {
			inFilterSection = true
			inMarketSection = false
			continue
		}
		if strings.Contains(lowerLine, "推荐股票") || strings.Contains(lowerLine, "推荐列表") {
			inRecommendationSection = true
			inMarketSection = false
			inFilterSection = false
			continue
		}

		// 如果进入推荐章节，停止提取
		if inRecommendationSection {
			break
		}

		// 收集内容
		if inMarketSection && strings.TrimSpace(line) != "" {
			marketBuilder.WriteString(line + "\n")
		}
		if inFilterSection && strings.TrimSpace(line) != "" {
			filterBuilder.WriteString(line + "\n")
		}
	}

	return strings.TrimSpace(marketBuilder.String()), strings.TrimSpace(filterBuilder.String())
}

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

	logger.SugaredLogger.Debugf("提取到的JSON字符串: %s", safeTruncate(jsonStr, 200))

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
					Rank:                len(recommendations) + 1,
					StockCode:           getString(recMap, "stock_code"),
					StockName:           getString(recMap, "stock_name"),
					CurrentPrice:        getFloat(recMap, "current_price"),
					PriceChange:         getFloat(recMap, "price_change"),
					TargetPrice:         getFloat(recMap, "target_price"),
					TargetChangePercent: getFloat(recMap, "target_change_percent"),
					Reason:              getString(recMap, "reason"),
					TechnicalAnalysis:   getString(recMap, "technical_analysis"),
					FundamentalAnalysis: getString(recMap, "fundamental_analysis"),
					RiskTips:            getString(recMap, "risk_tips"),
					RiskLevel:           getString(recMap, "risk_level"),
					Score:               getFloat(recMap, "score"),
					IsFollowed:          false, // 后续更新
				}
				recommendations = append(recommendations, item)
			}
		}
		return recommendations, true
	}

	return nil, false
}

// getString 从map中获取字符串值
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// getFloat 从map中获取浮点数值
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

// parseRecommendationsFromContent 从markdown内容解析推荐股票
func (s *StockPickService) parseRecommendationsFromContent(content string) []models.RecommendationItem {
	logger.SugaredLogger.Infof("开始解析markdown推荐内容，内容长度: %d", len(content))

	var recommendations []models.RecommendationItem

	lines := strings.Split(content, "\n")
	logger.SugaredLogger.Debugf("内容分割为 %d 行", len(lines))

	var currentRec *models.RecommendationItem
	var inRecommendationSection bool
	var currentSection string // 当前正在收集的内容类型: "reason", "technical", "fundamental", "risk"
	var contentBuilder strings.Builder

	for lineIdx, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		lowerLine := strings.ToLower(trimmedLine)

		// 检测推荐章节开始
		if strings.Contains(lowerLine, "## 推荐股票") || strings.Contains(lowerLine, "## 推荐列表") {
			inRecommendationSection = true
			logger.SugaredLogger.Debugf("在行 %d 检测到推荐章节开始", lineIdx)
			continue
		}

		// 检测推荐章节结束
		// 注意：不将"---"作为章节结束标志，因为它是股票之间的分隔符
		if inRecommendationSection && strings.HasPrefix(lowerLine, "## ") && !strings.Contains(lowerLine, "推荐股票") && !strings.Contains(lowerLine, "推荐列表") {
			// 遇到其他二级标题，结束推荐章节
			inRecommendationSection = false
		}

		if !inRecommendationSection {
			continue
		}

		// 解析股票标题行：### 1. 中芯国际 (sh688981)
		if strings.HasPrefix(trimmedLine, "###") && strings.Contains(trimmedLine, "(") && strings.Contains(trimmedLine, ")") {
			// 保存上一个推荐的收集内容
			if currentRec != nil {
				s.finalizeRecContent(currentRec, currentSection, &contentBuilder)
				recommendations = append(recommendations, *currentRec)
				logger.SugaredLogger.Debugf("添加推荐项，当前总数: %d", len(recommendations))
			}

			// 重置收集器
			currentSection = ""
			contentBuilder.Reset()

			// 解析新的推荐股票
			if rec, err := s.parseStockTitle(trimmedLine, len(recommendations)+1); err == nil {
				currentRec = rec
				logger.SugaredLogger.Debugf("解析到推荐项 %d: 代码=%s, 名称=%s", rec.Rank, rec.StockCode, rec.StockName)
			} else {
				logger.SugaredLogger.Warnf("解析股票标题失败: %v, 行内容: %s", err, trimmedLine)
				currentRec = nil
			}
			continue
		}

		if currentRec == nil {
			continue
		}

		// 检查是否是分隔符
		if trimmedLine == "---" || trimmedLine == "***" {
			continue
		}

		// 检测新的内容区块标题（粗体标题）
		isSectionHeader := false
		if (strings.HasPrefix(trimmedLine, "**") && strings.HasSuffix(trimmedLine, "**")) ||
			(strings.HasPrefix(trimmedLine, "**") && strings.Contains(trimmedLine, "：")) ||
			(strings.HasPrefix(trimmedLine, "**") && strings.Contains(trimmedLine, ":")) {

			// 先保存上一个区块的内容
			s.finalizeRecContent(currentRec, currentSection, &contentBuilder)
			contentBuilder.Reset()

			// 检测新区块类型
			sectionTitle := strings.ToLower(trimmedLine)
			if strings.Contains(sectionTitle, "推荐理由") {
				currentSection = "reason"
				isSectionHeader = true
			} else if strings.Contains(sectionTitle, "技术面") {
				currentSection = "technical"
				isSectionHeader = true
			} else if strings.Contains(sectionTitle, "基本面") {
				currentSection = "fundamental"
				isSectionHeader = true
			} else if strings.Contains(sectionTitle, "风险") {
				currentSection = "risk"
				isSectionHeader = true
			} else if strings.Contains(sectionTitle, "板块") || strings.Contains(sectionTitle, "行业") || strings.Contains(sectionTitle, "概念") {
				currentSection = "sector"
				isSectionHeader = true
			}

			// 尝试用单行解析方式也解析一下
			s.parseDetailLine(trimmedLine, currentRec)
		}

		if isSectionHeader {
			continue
		}

		// 如果有当前正在收集的区块，继续收集内容
		if currentSection != "" && trimmedLine != "" {
			if contentBuilder.Len() > 0 {
				contentBuilder.WriteString("\n")
			}
			contentBuilder.WriteString(trimmedLine)
		} else {
			// 没有指定区块，尝试单行解析
			s.parseDetailLine(trimmedLine, currentRec)
		}
	}

	// 保存最后一个推荐
	if currentRec != nil {
		s.finalizeRecContent(currentRec, currentSection, &contentBuilder)
		recommendations = append(recommendations, *currentRec)
		logger.SugaredLogger.Debugf("保存最后一个推荐项，总数: %d", len(recommendations))
	}

	// 更新关注状态
	for i := range recommendations {
		recommendations[i].IsFollowed = s.CheckStockFollowed(recommendations[i].StockCode)
	}

	logger.SugaredLogger.Infof("解析完成，共 %d 个推荐项", len(recommendations))
	return recommendations
}

// finalizeRecContent 完成内容收集并填充到推荐项
func (s *StockPickService) finalizeRecContent(rec *models.RecommendationItem, section string, builder *strings.Builder) {
	content := strings.TrimSpace(builder.String())
	if content == "" {
		return
	}

	switch section {
	case "reason":
		if rec.Reason == "" || len(content) > len(rec.Reason) {
			rec.Reason = content
		}
	case "technical":
		if rec.TechnicalAnalysis == "" || len(content) > len(rec.TechnicalAnalysis) {
			rec.TechnicalAnalysis = content
		}
	case "fundamental":
		if rec.FundamentalAnalysis == "" || len(content) > len(rec.FundamentalAnalysis) {
			rec.FundamentalAnalysis = content
		}
	case "risk":
		if rec.RiskTips == "" || len(content) > len(rec.RiskTips) {
			rec.RiskTips = content
		}
	case "sector":
		if rec.SectorConcept == "" || len(content) > len(rec.SectorConcept) {
			rec.SectorConcept = content
		}
	}
}

// parseRecommendationsFromMarkdown 从报告对象解析推荐股票（用于历史报告）
func (s *StockPickService) parseRecommendationsFromMarkdown(report *models.StockPickReport) []models.RecommendationItem {
	// 构建完整的markdown内容
	var content strings.Builder

	if report.MarketAnalysis != "" {
		content.WriteString("## 市场环境分析\n\n")
		content.WriteString(report.MarketAnalysis)
		content.WriteString("\n\n")
	}

	if report.FilterLogic != "" {
		content.WriteString("## 筛选逻辑\n\n")
		content.WriteString(report.FilterLogic)
		content.WriteString("\n\n")
	}

	// 如果Result字段包含markdown格式的推荐，使用它
	if strings.Contains(report.Result, "###") || strings.Contains(report.Result, "##") {
		return s.parseRecommendationsFromContent(report.Result)
	}

	// 没有可解析的推荐数据，返回空列表
	return []models.RecommendationItem{}
}

// parseStockTitle 解析股票标题：### 1. 中芯国际 (sh688981)
func (s *StockPickService) parseStockTitle(line string, rank int) (*models.RecommendationItem, error) {
	logger.SugaredLogger.Debugf("解析股票标题: %s, rank: %d", line, rank)

	// 去除 ### 和可能的排名前缀
	cleanLine := strings.TrimSpace(strings.TrimPrefix(line, "###"))
	cleanLine = strings.TrimSpace(strings.TrimPrefix(cleanLine, fmt.Sprintf("%d.", rank)))
	cleanLine = strings.TrimSpace(cleanLine)

	logger.SugaredLogger.Debugf("清理后的标题: %s", cleanLine)

	// 提取股票代码和名称：格式 "中芯国际 (sh688981)"
	codeStartIdx := strings.LastIndex(cleanLine, "(")
	codeEndIdx := strings.LastIndex(cleanLine, ")")

	if codeStartIdx == -1 || codeEndIdx == -1 || codeStartIdx >= codeEndIdx {
		logger.SugaredLogger.Warnf("股票标题格式不匹配: %s", line)
		return nil, fmt.Errorf("无法解析股票标题格式: %s", line)
	}

	stockCode := strings.TrimSpace(cleanLine[codeStartIdx+1 : codeEndIdx])
	stockName := strings.TrimSpace(cleanLine[:codeStartIdx])

	logger.SugaredLogger.Debugf("提取到股票代码: %s, 股票名称: %s", stockCode, stockName)

	if !isValidStockCode(stockCode) {
		logger.SugaredLogger.Warnf("无效的股票代码: %s", stockCode)
		return nil, fmt.Errorf("无效的股票代码格式: %s", stockCode)
	}

	// 创建推荐项
	rec := &models.RecommendationItem{
		Rank:       rank,
		StockCode:  strings.ToUpper(stockCode),
		StockName:  stockName,
		IsFollowed: s.CheckStockFollowed(stockCode),
	}

	// 获取股票基础信息（包含板块信息）
	if stockBasic := s.getStockBasicInfo(stockCode); stockBasic != nil {
		rec.SectorConcept = stockBasic.Industry
		if stockBasic.Industry == "" && stockBasic.Area != "" {
			rec.SectorConcept = stockBasic.Area
		}
		logger.SugaredLogger.Debugf("获取到股票板块信息: %s", rec.SectorConcept)
	}

	// 尝试获取实时价格
	if stockInfo, err := NewStockDataApi().GetStockCodeRealTimeData(stockCode); err == nil && stockInfo != nil && len(*stockInfo) > 0 {
		stock := (*stockInfo)[0]
		if price, err := parsePrice(stock.Price); err == nil {
			rec.CurrentPrice = price
		}
		rec.PriceChange = stock.ChangePercent
		logger.SugaredLogger.Debugf("获取到实时价格: %s, 涨跌幅: %s", stock.Price, stock.ChangePercent)
	} else {
		logger.SugaredLogger.Warnf("获取股票 %s 实时价格失败: %v", stockCode, err)
	}

	return rec, nil
}

// parseRecommendationsFromText 从文本格式解析推荐股票（原有逻辑）
func (s *StockPickService) parseRecommendationsFromText(content string) []models.RecommendationItem {
	logger.SugaredLogger.Infof("使用文本解析推荐内容，内容长度: %d", len(content))

	var recommendations []models.RecommendationItem

	lines := strings.Split(content, "\n")
	logger.SugaredLogger.Debugf("内容分割为 %d 行", len(lines))

	var currentRec *models.RecommendationItem
	var detailBuilder strings.Builder
	inRecommendationSection := false

	for lineIdx, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		lowerLine := strings.ToLower(trimmedLine)

		// 检测推荐章节开始
		if strings.Contains(lowerLine, "推荐股票") || strings.Contains(lowerLine, "推荐列表") {
			inRecommendationSection = true
			logger.SugaredLogger.Debugf("在行 %d 检测到推荐章节开始", lineIdx)
			continue
		}

		if !inRecommendationSection {
			continue
		}

		// 尝试解析推荐行: 格式 "1. [股票代码] [股票名称] - [推荐理由]"
		if matches := parseRecommendationLine(trimmedLine); matches != nil {
			// 保存上一个推荐
			if currentRec != nil && detailBuilder.Len() > 0 {
				currentRec.Reason = strings.TrimSpace(detailBuilder.String())
				recommendations = append(recommendations, *currentRec)
				logger.SugaredLogger.Debugf("添加推荐项，当前总数: %d", len(recommendations))
			}

			// 创建新推荐
			rank := len(recommendations) + 1
			stockCode := matches["stock_code"]
			stockName := matches["stock_name"]
			reason := matches["reason"]

			logger.SugaredLogger.Debugf("解析到推荐项 %d: 代码=%s, 名称=%s", rank, stockCode, stockName)

			currentRec = &models.RecommendationItem{
				Rank:       rank,
				StockCode:  stockCode,
				StockName:  stockName,
				Reason:     reason,
				IsFollowed: s.CheckStockFollowed(stockCode),
			}

			// 尝试获取实时价格
			logger.SugaredLogger.Debugf("尝试获取股票 %s 的实时数据", stockCode)
			if stockInfo, err := NewStockDataApi().GetStockCodeRealTimeData(stockCode); err != nil {
				logger.SugaredLogger.Warnf("获取股票 %s 实时数据失败: %v", stockCode, err)
			} else if stockInfo == nil {
				logger.SugaredLogger.Warnf("获取股票 %s 实时数据返回nil", stockCode)
			} else if len(*stockInfo) == 0 {
				logger.SugaredLogger.Warnf("获取股票 %s 实时数据返回空数组", stockCode)
			} else {
				stock := (*stockInfo)[0]
				logger.SugaredLogger.Debugf("获取到股票数据: Price=%s, ChangePercent=%s", stock.Price, stock.ChangePercent)
				if price, err := parsePrice(stock.Price); err == nil {
					currentRec.CurrentPrice = price
				} else {
					logger.SugaredLogger.Warnf("解析股票 %s 价格失败: %v, 原始值: %s", stockCode, err, stock.Price)
				}
				currentRec.PriceChange = stock.ChangePercent
			}

			detailBuilder.Reset()
		} else if currentRec != nil {
			// 解析详细信息
			s.parseDetailLine(trimmedLine, currentRec)
		}
	}

	// 保存最后一个推荐
	if currentRec != nil {
		if detailBuilder.Len() > 0 {
			currentRec.Reason = strings.TrimSpace(detailBuilder.String())
		}
		recommendations = append(recommendations, *currentRec)
	}

	logger.SugaredLogger.Infof("解析完成，共 %d 个推荐项", len(recommendations))
	return recommendations
}

// parseRecommendationLine 解析推荐行
func parseRecommendationLine(line string) map[string]string {
	// 格式: "1. [股票代码] [股票名称] - [推荐理由]" 或 "1. 股票代码 股票名称 - 推荐理由"
	matches := make(map[string]string)

	// 首先去除排名前缀 (如 "1. ")
	cleanLine := strings.TrimSpace(line)
	if strings.Contains(cleanLine, ".") {
		dotIdx := strings.Index(cleanLine, ".")
		if dotIdx < 3 { // 排名前缀通常是 "1." 或 "12."
			cleanLine = strings.TrimSpace(cleanLine[dotIdx+1:])
		}
	}

	// 解析股票代码 - 支持 [股票代码] 或 股票代码 格式
	var stockCode string
	var remaining string
	if strings.HasPrefix(cleanLine, "[") && strings.Contains(cleanLine, "]") {
		endIdx := strings.Index(cleanLine, "]")
		stockCode = strings.TrimSpace(cleanLine[1:endIdx])
		remaining = strings.TrimSpace(cleanLine[endIdx+1:])
	} else {
		// 找到第一个空格位置
		spaceIdx := strings.Index(cleanLine, " ")
		if spaceIdx > 0 {
			stockCode = strings.TrimSpace(cleanLine[:spaceIdx])
			remaining = strings.TrimSpace(cleanLine[spaceIdx+1:])
		}
	}

	if stockCode == "" {
		return nil
	}

	matches["stock_code"] = strings.ToUpper(stockCode)

	// 解析股票名称和理由
	if strings.Contains(remaining, "-") {
		sepIdx := strings.Index(remaining, "-")
		matches["stock_name"] = strings.TrimSpace(remaining[:sepIdx])
		matches["reason"] = strings.TrimSpace(remaining[sepIdx+1:])
	} else {
		matches["stock_name"] = strings.TrimSpace(remaining)
	}

	if len(matches["stock_code"]) > 0 && len(matches["stock_name"]) > 0 {
		return matches
	}
	return nil
}

// isValidStockCode 验证股票代码格式
func isValidStockCode(code string) bool {
	//code = strings.ToLower(code)
	//return strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "sz") ||
	//	strings.HasPrefix(code, "hk") || strings.HasPrefix(code, "us")
	return true
}

// parseDetailLine 解析详细信息行
func (s *StockPickService) parseDetailLine(line string, rec *models.RecommendationItem) {
	lowerLine := strings.ToLower(line)
	trimmedLine := strings.TrimSpace(line)

	// 检查是否是标签行（如 "- 技术面分析：..."）
	isLabeledLine := strings.HasPrefix(trimmedLine, "-") || strings.HasPrefix(trimmedLine, "•") || strings.HasPrefix(trimmedLine, "*")
	if isLabeledLine {
		trimmedLine = strings.TrimSpace(strings.TrimLeft(trimmedLine, "-•*"))
		lowerLine = strings.ToLower(trimmedLine)
	}

	if strings.Contains(lowerLine, "当前价格") || strings.Contains(lowerLine, "现价") {
		if price, err := parsePriceFromLine(trimmedLine); err == nil {
			rec.CurrentPrice = price
		}
	} else if strings.Contains(lowerLine, "涨跌幅") {
		if change, err := parsePriceFromLine(trimmedLine); err == nil {
			rec.PriceChange = change
		}
	} else if strings.Contains(lowerLine, "目标价位") || strings.Contains(lowerLine, "目标价") {
		if price, err := parsePriceFromLine(trimmedLine); err == nil {
			rec.TargetPrice = price
		}
	} else if strings.Contains(lowerLine, "上涨空间") || strings.Contains(lowerLine, "目标涨幅") {
		if change, err := parsePriceFromLine(trimmedLine); err == nil {
			rec.TargetChangePercent = change
		}
	} else if strings.Contains(lowerLine, "综合评分") || strings.Contains(lowerLine, "评分") {
		if score, err := parsePriceFromLine(trimmedLine); err == nil {
			rec.Score = score
		}
	} else if strings.Contains(lowerLine, "买卖建议") {
		rec.TradeSuggestion = parseTextValue(trimmedLine)
	} else if strings.Contains(lowerLine, "板块") || strings.Contains(lowerLine, "行业") || strings.Contains(lowerLine, "概念") {
		// 解析板块/行业/概念信息
		if value := parseTextValue(trimmedLine); value != "" {
			// 如果当前还没有板块信息，或者这个更详细，则更新
			if rec.SectorConcept == "" || len(value) > len(rec.SectorConcept) {
				rec.SectorConcept = value
			}
		}
	} else if strings.Contains(lowerLine, "推荐理由") {
		// 推荐理由通常是多行的，这里先收集第一行
		if value := parseTextValue(trimmedLine); value != "" {
			rec.Reason = value
		}
	} else if strings.Contains(lowerLine, "技术面分析") || strings.Contains(lowerLine, "技术面") {
		if value := parseTextValue(trimmedLine); value != "" {
			rec.TechnicalAnalysis = value
		}
	} else if strings.Contains(lowerLine, "基本面分析") || strings.Contains(lowerLine, "基本面") {
		if value := parseTextValue(trimmedLine); value != "" {
			rec.FundamentalAnalysis = value
		}
	} else if strings.Contains(lowerLine, "风险提示") || strings.Contains(lowerLine, "风险分析") {
		if value := parseTextValue(trimmedLine); value != "" {
			rec.RiskTips = value
		}
	}
}

// parsePrice 从字符串解析价格
func parsePrice(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(strings.TrimSpace(s), "%f", &result)
	return result, err
}

// parsePriceFromLine 从行中解析价格
func parsePriceFromLine(line string) (float64, error) {
	// 提取冒号或中文冒号后的内容
	if idx := strings.Index(line, ":"); idx >= 0 {
		return parsePrice(line[idx+1:])
	}
	if idx := strings.Index(line, "："); idx >= 0 {
		return parsePrice(line[idx+1:])
	}
	return 0, fmt.Errorf("无法解析价格")
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

// getStockBasicInfo 获取股票基础信息
func (s *StockPickService) getStockBasicInfo(stockCode string) *StockBasic {
	var stock StockBasic

	// 处理股票代码格式转换
	tsCode := convertToTsCode(stockCode)
	symbol := extractSymbol(stockCode)

	logger.SugaredLogger.Debugf("查询股票基础信息: 原始代码=%s, TS代码=%s, Symbol=%s", stockCode, tsCode, symbol)

	// 尝试多种代码格式查询
	err := db.Dao.Where("ts_code = ? OR symbol = ?", tsCode, symbol).First(&stock).Error
	if err != nil {
		logger.SugaredLogger.Debugf("获取股票 %s 基础信息失败: %v", stockCode, err)
		// 尝试只查询symbol
		err = db.Dao.Where("symbol = ?", symbol).First(&stock).Error
		if err != nil {
			logger.SugaredLogger.Debugf("通过symbol查询也失败: %v", err)
			return nil
		}
	}

	logger.SugaredLogger.Debugf("获取到股票基础信息: 代码=%s, 名称=%s, 行业=%s, 地域=%s",
		stock.TsCode, stock.Name, stock.Industry, stock.Area)

	return &stock
}

// convertToTsCode 转换为Tushare格式
// 例如: "600519" -> "600519.SH", "sh600519" -> "600519.SH"
func convertToTsCode(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))

	// 已经是TS格式
	if strings.Contains(code, ".") {
		return code
	}

	// sh前缀
	if strings.HasPrefix(code, "SH") {
		return strings.TrimPrefix(code, "SH") + ".SH"
	}

	// sz前缀
	if strings.HasPrefix(code, "SZ") {
		return strings.TrimPrefix(code, "SZ") + ".SZ"
	}

	// hk前缀
	if strings.HasPrefix(code, "HK") {
		return strings.TrimPrefix(code, "HK") + ".HK"
	}

	// us前缀
	if strings.HasPrefix(code, "US") {
		return strings.TrimPrefix(code, "US") + ".US"
	}

	// 根据代码长度判断
	if len(code) == 6 {
		if strings.HasPrefix(code, "6") {
			return code + ".SH"
		}
		return code + ".SZ"
	}

	return code
}

// extractSymbol 提取纯数字代码
func extractSymbol(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))

	// 移除交易所后缀
	if idx := strings.Index(code, "."); idx != -1 {
		code = code[:idx]
	}

	// 移除sh/sz/hk/us前缀
	code = strings.TrimPrefix(code, "SH")
	code = strings.TrimPrefix(code, "SZ")
	code = strings.TrimPrefix(code, "HK")
	code = strings.TrimPrefix(code, "US")

	return code
}
