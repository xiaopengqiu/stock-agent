 # AI荐股功能设计文档

## 1. 功能概述

AI荐股功能集成在市场行情栏目下，新增一个"AI荐股"标签页。用户可以通过自然语言对话方式描述选股需求，AI利用现有工具获取市场数据，经过两阶段分析（规则筛选+AI深度分析），最终生成推荐股票的完整报告和简洁列表视图。

**核心特性：**
- 对话式交互，自然语言表达选股需求
- 两阶段分析：规则快速筛选 + AI深度分析
- 完整分析报告和简洁列表双视图
- 推荐股票一键关注功能
- 历史记录存储和查看
- PDF/Markdown 导出功能
- 可调用工具列表展示

---

## 2. 整体架构

### 2.1 系统架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                           前端层 (Vue 3)                         │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  ai-stock-pick.vue (主组件)                             │   │
│  │  - 对话区域 (消息列表 + 输入框)                          │   │
│  │  - 推荐结果展示 (完整报告 / 简洁列表)                    │   │
│  │  - 工具列表展示                                          │   │
│  │  - 历史记录抽屉                                          │   │
│  │  - 导出功能                                              │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                         Wails IPC
                              │
┌─────────────────────────────────────────────────────────────────┐
│                           后端层 (Go)                            │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  app.go (Wails 绑定)                                      │   │
│  │  - AIStockPickChat() - 对话式荐股                        │   │
│  │  - GetStockPickReports() - 获取报告列表                  │   │
│  │  - GetStockPickReport() - 获取单个报告                  │   │
│  │  - ExportStockPickReport() - 导出报告                   │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  stock_pick_service.go (业务服务)                        │   │
│  │  - ProcessStockPick() - 处理荐股流程                    │   │
│  │  - SaveReport() - 保存报告                               │   │
│  │  - GetReports() - 查询报告                              │   │
│  │  - ExportToPDF() - PDF导出                              │   │
│  │  - ExportToMarkdown() - Markdown导出                    │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                                  │
│  ┌─────────────────────────────────────────────────────────┘   │
│  │  stock_pick_report.go (数据模型)                          │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  agent/agent.go (AI Agent)                               │   │
│  │  - 加载 Skill Prompt 模板                                │   │
│  │  - 调用工具获取数据                                       │   │
│  │  - 生成分析结果                                          │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                       数据存储层                                 │
│  - SQLite 数据库 (stock_pick_reports 表)                       │
│  - Skill Prompt 文件 (data/skills/ai-stock-pick.md)            │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 数据流

```
用户输入 → 前端验证 → 调用后端API
                    ↓
            加载Skill Prompt模板
                    ↓
            构建完整Prompt上下文
                    ↓
            调用AI Agent
                    ↓
    ┌──────────────────────────────┐
    │  第一阶段：规则快速筛选      │
    │  - 调用ChoiceStockByIndicators│
    │  - 调用QueryBKDictInfo       │
    │  - 筛选出50-100只候选股票    │
    └──────────────────────────────┘
                    ↓
    ┌──────────────────────────────┐
    │  第二阶段：AI深度分析        │
    │  - 调用QueryStockPriceInfo   │
    │  - 调用QueryStockKLine       │
    │  - 调用GetFinancialReport    │
    │  - 对前20-30只候选进行深度分析│

```

---

## 3. 前端组件设计

### 3.1 组件结构

**文件位置:** `frontend/src/components/ai-stock-pick.vue`

**布局设计:**
```
┌────────────────────────────────────────────────────────────────────────┐
│  [AI荐股标题]                   [历史记录▼] [导出▼]  [查看模式: ●完整 □简洁]  │
├────────────────────────────────────┬──────────────────────────────────────┤
│  对话区域                          │  推荐结果展示区域                    │
│  ┌──────────────────────────────┐ │  ┌──────────────────────────────────┐ │
│  │ 🔧 可用工具:                  │ │  │                                  │ │
│  │ [选股] [股价] [K线] [财报]   │ │  │  ┌──────────────────────────┐  │ │
│  │                              │ │  │  │ 推荐股票列表 (3只)       │  │ │
│  │ 💬 消息列表:                 │ │  │  ├──────────────────────────┤  │ │
│  │ ┌────────────────────────┐   │ │  │  │ 1. 科技龙头A (推荐)      │  │ │
│  │ │ AI: 请告诉我您的选股  │   │ │  │  │   代码: 000001            │  │ │
│  │ │     需求...           │   │ │  │  │   理由: 资金大幅流入     │  │ │
│  │ │                        │   │ │  │  │   目标价: ↑8%            │  │ │
│  │ │ User: 推荐今日资金    │   │ │  │  │   [一键关注]             │  │ │
│  │ │       流入大的科技股  │   │ │  │  ├──────────────────────────┤  │ │
│  │ │                        │   │ │  │  │ 2. 芯片制造B (关注)      │  │ │
│  │ │ AI: 正在分析市场...   │   │ │  │  │   [一键关注]             │  │ │
│  │ │     [流式输出]        │   │ │  │  └──────────────────────────┘  │ │
│  │ └────────────────────────┘   │ │  └──────────────────────────────────┘ │
│  │                              │ │  ┌──────────────────────────────────┐ │
│  │                              │ │  │ 市场分析摘要                      │ │
│  │                              │ │  │ - 今日大盘指数上涨1.2%           │ │
│  │                              │ │  │ - 科技板块资金净流入50亿         │ │
│  │                              │ │  │ - 推荐策略: 龙头+成长性          │ │
│  │                              │ │  └──────────────────────────────────┘ │
│  └──────────────────────────────┘ │                                      │
│  ┌──────────────────────────────┐ │  (完整报告视图: 使用折叠面板展示    │
│  │ [输入用户选股需求...] [发送] │ │   每只股票的详细分析)               │
│  └──────────────────────────────┘ │                                      │
└────────────────────────────────────┴──────────────────────────────────────┘
```

### 3.2 核心功能模块

#### 3.2.1 对话模块
- **消息列表渲染:** 使用 `n-scrollbar + v-for` 渲染消息
- **用户消息:** 右对齐，蓝底白字
- **AI消息:** 左对齐，支持Markdown渲染
- **流式输出:** 实时更新AI响应内容
- **输入框:** `n-input` 组件，支持多行输入，Enter发送，Shift+Enter换行

#### 3.2.2 推荐结果模块
- **完整报告视图:**
  - 使用 `n-collapse` 折叠面板展示推荐股票
  - 每个面板包含：股票基本信息、技术面分析、基本面分析、风险提示、推荐理由
- **简洁列表视图:**
  - 使用 `n-data-table` 展示表格
  - 列：排名、代码、名称、现价、涨跌幅、推荐理由、目标涨幅、评分、操作
  - 支持按评分、涨幅排序
  - 操作列：[关注] 按钮

#### 3.2.3 工具列表展示
- **位置:** 对话区域顶部
- **展示方式:** 使用 `n-tag` 组件
- **状态指示:**
  - 灰色：未调用
  - 蓝色：调用中
  - 绿色：调用成功
  - 红色：调用失败

#### 3.2.4 一键关注功能
- **实现:** 调用现有的 `window.go.main.Follow(stockCode)` 方法
- **交互:**
  - 点击关注按钮，调用API
  - 成功后按钮变为"已关注"，禁用点击
  - 失败显示错误提示

#### 3.2.5 历史记录功能
- **触发:** 点击"历史记录"按钮
- **UI:** 使用 `n-drawer` 侧边栏，从右侧滑出
- **内容:**
  - 历史荐股记录列表
  - 每条记录：时间、需求摘要、推荐数量、查看按钮
- **交互:** 点击查看按钮，加载历史报告到当前视图

#### 3.2.6 导出功能
- **触发:** 点击"导出"按钮，弹出选择菜单
- **选项:** [导出为PDF] [导出为Markdown]
- **实现:** 调用后端导出API，下载文件

### 3.3 组件状态

```javascript
const state = reactive({
  // 对话相关
  messages: [
    {
      id: 1,
      role: 'assistant', // 'user' | 'assistant'
      content: '请告诉我您的选股需求，例如：\n- 推荐今日资金流入大的科技股\n- 寻找市盈率低于20且业绩增长的银行股\n- 推荐近期创新高的新能源龙头股',
      timestamp: Date.now(),
      tools: [] // 本次对话使用的工具
    }
  ],
  inputText: '',
  loading: false,

  // 分析相关
  analyzing: false,
  toolsUsed: [], // 已使用的工具列表
  toolStatus: {}, // 工具状态: { toolName: 'idle' | 'running' | 'success' | 'failed' }

  // 推荐结果相关
  recommendations: [], // 推荐股票列表
  viewMode: 'full', // 'full' | 'simple'
  marketAnalysis: {}, // 市场分析摘要
  reportId: null, // 当前报告ID

  // 历史记录
  historyVisible: false,
  historyList: [], // 历史荐股记录列表
})
```

### 3.4 组件方法

```javascript
// 发送消息
async function sendMessage() {
  // 1. 添加用户消息
  // 2. 调用后端 API
  // 3. 流式接收AI响应
}

// 流式接收AI响应
async function streamResponse() {
  // 使用 Wails Events 接收流式输出
  window.wails.events.on('ai-stock-pick-response', handleStream)
}

// 一键关注
async function followStock(stockCode) {
  await window.go.main.Follow(stockCode)
  // 更新UI状态
}

// 加载历史记录
async function loadHistory() {
  const reports = await window.go.main.GetStockPickReports()
  state.historyList = reports
}

// 查看历史报告
async function viewHistoryReport(reportId) {
  const report = await window.go.main.GetStockPickReport(reportId)
  // 更新UI
}

// 导出报告
async function exportReport(format) {
  // format: 'pdf' | 'markdown'
  await window.go.main.ExportStockPickReport(state.reportId, format)
}
```

---

## 4. 后端设计

### 4.1 数据模型

**文件位置:** `backend/models/stock_pick_report.go`

```go
package models

import "time"

// StockPickReport AI荐股报告
type StockPickReport struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`

    // 用户需求
    UserQuery   string `gorm:"type:text;not null" json:"user_query"`   // 用户输入的选股需求
    QuerySummary string `gorm:"type:text" json:"query_summary"`          // 需求摘要

    // 分析结果
    MarketAnalysis string `gorm:"type:text" json:"market_analysis"`     // 市场环境分析
    FilterLogic   string `gorm:"type:text" json:"filter_logic"`          // 筛选逻辑说明
    TotalScanned  int    `json:"total_scanned"`                          // 扫描股票总数
    CandidatesCount int `json:"candidates_count"`                       // 候选股票数

    // 推荐股票列表 (JSON)
    Recommendations string `gorm:"type:text;not null" json:"recommendations"`

    // 使用的工具列表 (JSON)
    ToolsUsed string `gorm:"type:text" json:"tools_used"`

    // AI配置
    AIConfigID uint `json:"ai_config_id"`
    AIModel    string `json:"ai_model"`

    // 状态
    Status string `gorm:"type:varchar(20);default:'completed'" json:"status"` // 'processing' | 'completed' | 'failed'
    Error  string `gorm:"type:text" json:"error"`
}

// RecommendationItem 推荐股票项
type RecommendationItem struct {
    Rank          int     `json:"rank"`           // 排名
    StockCode     string  `json:"stock_code"`     // 股票代码
    StockName     string  `json:"stock_name"`     // 股票名称
    CurrentPrice  float64 `json:"current_price"`  // 现价
    PriceChange   float64 `json:"price_change"`   // 涨跌幅
    Volume        int64   `json:"volume"`         // 成交量
    MarketValue   float64 `json:"market_value"`   // 市值

    // 分析内容
    TechnicalAnalysis string `json:"technical_analysis"` // 技术面分析
    FundamentalAnalysis string `json:"fundamental_analysis"` // 基本面分析
    Reason            string `json:"reason"`            // 推荐理由
    TargetPrice       float64 `json:"target_price"`      // 目标价位
    TargetChangePercent float64 `json:"target_change_percent"` // 目标涨幅
    RiskLevel         string `json:"risk_level"`         // 风险等级
    RiskTips          string `json:"risk_tips"`          // 风险提示
    Score             float64 `json:"score"`              // 综合评分 (0-100)

    // 关注状态
    IsFollowed bool `json:"is_followed"`
}

// ToolUsage 工具使用记录
type ToolUsage struct {
    ToolName string `json:"tool_name"` // 工具名称
    Status   string `json:"status"`   // 'idle' | 'running' | 'success' | 'failed'
    CallTime string `json:"call_time"` // 调用时间
    Duration string `json:"duration"`  // 耗时
}

func (StockPickReport) TableName() string {
    return "stock_pick_reports"
}
```

### 4.2 服务层

**文件位置:** `backend/data/stock_pick_service.go`

```go
package data

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    "your-project/backend/agent"
    "your-project/backend/models"
    "your-project/backend/db"
)

type StockPickService struct {
    db *gorm.DB
}

func NewStockPickService() *StockPickService {
    return &StockPickService{db: db.GetDB()}
}

// StockPickRequest 荐股请求
type StockPickRequest struct {
    UserQuery   string `json:"user_query"`
    AIConfigID  uint   `json:"ai_config_id"`
}

// StockPickResponse 荐股响应
type StockPickResponse struct {
    ReportID    uint `json:"report_id"`
    StreamID    string `json:"stream_id"`
}

// ProcessStockPick 处理荐股流程
func (s *StockPickService) ProcessStockPick(ctx context.Context, req StockPickRequest, eventHandler func(eventType string, data interface{})) (*StockPickResponse, error) {
    // 1. 创建报告记录
    report := &models.StockPickReport{
        UserQuery:   req.UserQuery,
        QuerySummary: generateQuerySummary(req.UserQuery),
        AIConfigID:  req.AIConfigID,
        Status:      "processing",
    }
    if err := s.db.Create(report).Error; err != nil {
        return nil, err
    }

    // 2. 获取AI配置
    aiConfig, err := s.getAIConfig(req.AIConfigID)
    if err != nil {
        return nil, err
    }

    // 3. 加载Skill Prompt模板
    skillPrompt, err := s.loadSkillPrompt()
    if err != nil {
        return nil, err
    }

    // 4. 构建完整Prompt
    fullPrompt := s.buildPrompt(skillPrompt, req.UserQuery)

    // 5. 发送开始事件
    eventHandler("start", map[string]interface{}{
        "report_id": report.ID,
        "message":   "开始分析市场数据...",
    })

    // 6. 调用AI Agent
    aiAgent := agent.GetStockAiAgent(ctx, aiConfig)

    // 流式处理响应
    streamID := fmt.Sprintf("stock-pick-%d", report.ID)
    err = aiAgent.Stream(ctx, fullPrompt, func(chunk string) {
        eventHandler("stream", map[string]interface{}{
            "content": chunk,
        })
    })

    // 7. 解析AI响应，提取结构化数据
    // TODO: 需要Agent返回结构化数据

    // 8. 更新报告状态
    report.Status = "completed"
    s.db.Save(report)

    return &StockPickResponse{
        ReportID: report.ID,
        StreamID: streamID,
    }, nil
}

// SaveReport 保存荐股报告
func (s *StockPickService) SaveReport(report *models.StockPickReport) error {
    return s.db.Save(report).Error
}

// GetReports 获取荐股报告列表
func (s *StockPickService) GetReports(offset, limit int) ([]models.StockPickReport, int64, error) {
    var reports []models.StockPickReport
    var total int64

    s.db.Model(&models.StockPickReport{}).Count(&total)
    err := s.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&reports).Error

    return reports, total, err
}

// GetReport 获取单个报告
func (s *StockPickService) GetReport(id uint) (*models.StockPickReport, error) {
    var report models.StockPickReport
    err := s.db.First(&report, id).Error
    return &report, err
}

// loadSkillPrompt 加载Skill Prompt模板
func (s *StockPickService) loadSkillPrompt() (string, error) {
    // 从doc/skills/ai-stock-pick.md读取
    // TODO: 实现文件读取
    return "", nil
}

// buildPrompt 构建完整Prompt
func (s *StockPickService) buildPrompt(skillPrompt, userQuery string) string {
    // 将Skill Prompt模板和用户查询组合
    return fmt.Sprintf("%s\n\n用户选股需求：%s", skillPrompt, userQuery)
}

// generateQuerySummary 生成需求摘要
func generateQuerySummary(query string) string {
    // 截取前50个字符
    if len(query) > 50 {
        return query[:50] + "..."
    }
    return query
}

// ExportToPDF 导出为PDF
func (s *StockPickService) ExportToPDF(reportID uint, outputPath string) error {
    // TODO: 实现PDF导出
    return nil
}

// ExportToMarkdown 导出为Markdown
func (s *StockPickService) ExportToMarkdown(reportID uint, outputPath string) error {
    // TODO: 实现Markdown导出
    return nil
}
```

### 4.3 Wails绑定方法

**文件位置:** `app.go`

在 `App` 结构体中添加以下方法：

```go
// AIStockPickChat 对话式AI荐股
func (a *App) AIStockPickChat(userQuery string, aiConfigID uint) (string, error) {
    ctx := context.Background()

    // 创建荐股请求
    req := data.StockPickRequest{
        UserQuery:  userQuery,
        AIConfigID: aiConfigID,
    }

    // 流式事件处理器
    eventHandler := func(eventType string, data interface{}) {
        runtime.EventsEmit(ctx, "ai-stock-pick-response", map[string]interface{}{
            "type": eventType,
            "data": data,
        })
    }

    // 处理荐股流程
    response, err := a.stockPickService.ProcessStockPick(ctx, req, eventHandler)
    if err != nil {
        return "", err
    }

    return fmt.Sprintf("%d", response.ReportID), nil
}

// GetStockPickReports 获取荐股报告列表
func (a *App) GetStockPickReports(offset, limit int) ([]data.StockPickReportItem, int64, error) {
    reports, total, err := a.stockPickService.GetReports(offset, limit)
    // 转换为响应DTO
    // TODO: 实现转换
    return nil, total, err
}

// GetStockPickReport 获取单个荐股报告
func (a *App) GetStockPickReport(reportID uint) (*models.StockPickReport, error) {
    return a.stockPickService.GetReport(reportID)
}

// ExportStockPickReport 导出荐股报告
func (a *App) ExportStockPickReport(reportID uint, format string) (string, error) {
    ctx := context.Background()

    // 生成输出文件名
    timestamp := time.Now().Format("20060102-150405")
    var outputPath string

    if format == "pdf" {
        outputPath = fmt.Sprintf("stock-pick-report-%d-%s.pdf", reportID, timestamp)
        err := a.stockPickService.ExportToPDF(reportID, outputPath)
        if err != nil {
            return "", err
        }
    } else if format == "markdown" {
        outputPath = fmt.Sprintf("stock-pick-report-%d-%s.md", reportID, timestamp)
        err := a.stockPickService.ExportToMarkdown(reportID, outputPath)
        if err != nil {
            return "", err
        }
    }

    // 发送事件通知前端下载
    runtime.EventsEmit(ctx, "ai-stock-pick-export-ready", map[string]interface{}{
        "report_id":    reportID,
        "file_path":    outputPath,
        "format":       format,
    })

    return outputPath, nil
}

// GetAIConfig 获取AI配置
func (a *App) getAIConfig(configID uint) (data.AIConfig, error) {
    // TODO: 从数据库获取AI配置
    return data.AIConfig{}, nil
}
```

---

## 5. AI Skill Prompt设计

**文件位置:** `doc/skills/ai-stock-pick.md`

该文件包含完整的AI荐股分析Prompt模板，指导AI如何进行股票分析和推荐。

### 5.1 Prompt结构

```
# 角色定义
你是一位专业的证券分析师和AI选股专家...

# 可用工具列表
以下是你可以调用的工具：
1. ChoiceStockByIndicators - 根据自然语言筛选股票
2. QueryStockPriceInfo - 批量获取实时股价数据
3. QueryStockKLine - 获取股票K线数据
4. GetFinancialReport - 查询股票财务报表数据
5. QueryBKDictInfo - 获取板块/行业信息
...

# 分析流程
第一阶段：市场环境分析
- 分析当前大盘走势
- 分析热门板块资金流向
- 识别市场热点和风格特征

第二阶段：候选股票筛选
- 根据用户需求调用选股工具
- 初步筛选出50-100只候选股票

第三阶段：深度分析
- 对前20-30只候选股票进行深度分析
- 分析技术面（K线、趋势、量价关系）
- 分析基本面（财务指标、业绩增长）
- 综合评分和排序

第四阶段：生成推荐报告
- 筛选出3-5只最具潜力的股票
- 生成完整的分析报告
- 提供买入建议和风险提示

# 输出格式
请按照以下JSON格式输出分析结果：
{
  "market_analysis": "市场环境分析...",
  "filter_logic": "筛选逻辑说明...",
  "total_scanned": 1000,
  "candidates_count": 50,
  "recommendations": [
    {
      "rank": 1,
      "stock_code": "000001",
      "stock_name": "平安银行",
      ...
    }
  ]
}
```

### 5.2 Prompt内容要点

1. **角色设定:** 专业证券分析师，具有丰富的A股、港股、美股分析经验
2. **分析框架:** 宏观-行业-个股三层分析法
3. **技术分析:** 趋势、形态、量价、指标（MACD、KDJ、RSI）
4. **基本面分析:** PE、PB、ROE、营收增长、利润增长、现金流
5. **风险控制:** 仓位建议、止损点位、风险提示
6. **逻辑严谨:** 每个推荐都要有充分的理由支撑

---

## 6. 数据库设计

### 6.1 表结构

```sql
CREATE TABLE stock_pick_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,

    -- 用户需求
    user_query TEXT NOT NULL,
    query_summary TEXT,

    -- 分析结果
    market_analysis TEXT,
    filter_logic TEXT,
    total_scanned INTEGER,
    candidates_count INTEGER,

    -- 推荐股票列表 (JSON)
    recommendations TEXT NOT NULL,

    -- 使用的工具列表 (JSON)
    tools_used TEXT,

    -- AI配置
    ai_config_id INTEGER,
    ai_model TEXT,

    -- 状态
    status VARCHAR(20) DEFAULT 'completed',
    error TEXT
);

CREATE INDEX idx_stock_pick_created ON stock_pick_reports(created_at DESC);
CREATE INDEX idx_stock_pick_status ON stock_pick_reports(status);
CREATE INDEX idx_stock_pick_ai_config ON stock_pick_reports(ai_config_id);
```

---

## 7. 错误处理

### 7.1 前端错误处理
- AI调用失败：显示错误提示，允许重试
- 工具调用失败：记录工具状态，继续其他工具
- 数据解析失败：显示原始AI响应，提示格式问题
- 关注失败：显示错误提示，允许重试

### 7.2 后端错误处理
- AI配置不存在：返回默认配置或提示用户配置
- 工具调用超时：设置超时时间，超时后返回部分结果
- 数据库操作失败：记录日志，返回友好错误信息
- 导出失败：检查文件权限，返回详细错误信息

---

## 8. 性能优化

### 8.1 前端优化
- 虚拟滚动：推荐列表使用虚拟滚动
- 懒加载：历史记录分页加载
- 缓存：缓存的报告数据避免重复请求
- 防抖：输入框防抖，避免频繁调用

### 8.2 后端优化
- 异步处理：荐股流程异步执行，通过事件通知前端
- 缓存策略：市场数据缓存5分钟
- 并行处理：多只股票并行分析
- 分页查询：历史记录分页返回

---

## 9. 安全考虑

### 9.1 输入验证
- 前端验证：用户输入长度限制，防止过长输入
- 后端验证：参数合法性检查，防止注入攻击

### 9.2 数据安全
- 敏感信息过滤：不暴露用户个人信息
- SQL注入防护：使用参数化查询，ORM自动防护

---

## 10. 测试计划

### 10.1 单元测试
- StockPickService方法测试
- 数据模型测试
- Prompt加载测试

### 10.2 集成测试
- AI Agent调用测试
- 工具调用测试
- 流式输出测试

### 10.3 端到端测试
- 完整荐股流程测试
- 历史记录功能测试
- 导出功能测试
- 关注功能测试

---

## 11. 实施步骤

1. **创建数据模型** - `stock_pick_report.go`
2. **创建服务层** - `stock_pick_service.go`
3. **添加Wails绑定** - `app.go`
4. **创建Skill Prompt** - `doc/skills/ai-stock-pick.md`
5. **创建前端组件** - `ai-stock-pick.vue`
6. **集成到市场行情** - `market.vue`
7. **测试和调试**
8. **文档更新**

---

## 12. 后续优化方向

1. **个性化推荐:** 基于用户历史行为推荐
2. **回测功能:** 验证历史推荐的准确性
3. **多模型对比:** 支持多个AI模型同时分析
4. **实时更新:** 推荐股票实时价格监控
5. **社交分享:** 分享推荐报告到社区
