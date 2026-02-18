# 动态工具加载系统设计文档

## 概述

将 `getStockPickTools` 工具获取功能改造成读取配置模版的方式加载已有工具，确保工具调用可以实现动态修改。配置文件存储在 `data/tools` 目录下，支持加载所有内置函数、MCP 工具和 HTTP 接口调用。

## 需求分析

### 核心需求

1. **配置化工具管理**：将硬编码的工具列表替换为配置文件加载方式
2. **多类型工具支持**：支持内置函数、MCP 工具和 HTTP 接口工具
3. **动态修改**：工具配置可通过配置文件修改，无需重启应用
4. **向后兼容**：保持与现有工具调用系统的兼容性

### 成功标准

- 所有现有工具功能正常工作
- 工具配置可通过配置文件修改
- 支持多种工具类型（内置函数、MCP、HTTP）
- 配置加载过程稳定可靠
- 工具调用性能无明显下降
- 配置无效时有合理的降级处理

## 架构分析

### 当前工具系统架构

```
App (app.go)
├── AddTools() - 硬编码工具列表
├── AiTools []data.Tool - 工具存储
└── StockPickService.getStockPickTools() - 硬编码工具列表

backend/data/
├── openai_api.go - Tool 结构体定义
└── stock_pick_service.go - getStockPickTools() 实现

backend/agent/
├── registry/
│   ├── builtins.go - 内置工具注册
│   └── registry.go - 工具注册表（内置+MCP）
└── tools/ - 工具实现（bk_dict_tool.go 等）
```

### 现有工具类型

1. **内置函数工具**（agent/tools/ 目录下）：
   - QueryBKDictInfo
   - QueryEconomicData
   - QueryStockPriceInfo
   - QueryStockCodeInfo
   - QueryMarketNews
   - ChoiceStockBy
Indicators
   - QueryStockKLine
   - QueryInteractiveAnswerData
   - GetFinancialReport
   - QueryStockNewsTool
   - GetIndustryResearchReport

2. **MCP 工具**（通过 MCP 服务器加载）：
   - 配置文件：data/mcp_config.json
   - 支持动态连接到 MCP 服务器

3. **HTTP 接口工具**（待实现）：
   - 通过 HTTP 接口提供的工具

### 数据结构分析

```go
// 工具基本结构（data/openai_api.go）
type Tool struct {
    Type     string       `json:"type"`
    Function ToolFunction `json:"function"`
}

type ToolFunction struct {
    Name        string             `json:"name"`
    Description string             `json:"description"`
    Parameters  FunctionParameters `json:"parameters"`
}

type FunctionParameters struct {
    Type       string         `json:"type"`
    Properties map[string]any `json:"properties"`
    Required   []string       `json:"required"`
}
```

## 设计方案

### 配置文件格式设计

#### 目录结构

```
data/
└── tools/
    ├── config.json - 主配置文件（工具列表）
    └── definitions/ - 工具详细定义（可选）
```

#### 主配置文件格式 (data/tools/config.json)

```json
{
  "version": "1.0",
  "tools": [
    {
      "name": "SearchStockByIndicators",
      "type": "builtin",
      "enabled": true,
      "config": {}
    },
    {
      "name": "GetStockK
Line",
      "type": "builtin",
      "enabled": true,
      "config": {}
    },
    {
      "name": "InteractiveAnswer",
      "type": "builtin",
      "enabled": true,
      "config": {}
    },
    {
      "name":. "GetStockResearchReport",
      "type": "builtin",
      "enabled": true,
      "config": {}
    },
    {
      "name": "mcp-tool-1",
      "type": "mcp",
      "enabled": true,
      "config": {
        "server": "default"
      }
    },
    {
      "name": "http-tool-1",
      "type": "http",
      "enabled": true,
      "config": {
        "url": "http://api.example.com/tool1",
        "method": "POST",
        "headers": {}
      }
    }
  ]
}
```

### Go 数据结构设计

```go
// Tool 配置结构
type ToolConfig struct {
    Version string     `json:"version"`
    Tools   []ToolItem `json:"tools"`
}

type ToolItem struct {
    Name     string                 `json:"name"`
    Type     string                 `json:"type"` // "builtin" | "mcp" | "http"
    Enabled  bool                   `json:"enabled"`
    Config   map[string]interface{} `json:"config"`
}
```

### 核心组件设计

#### 1. 工具配置管理器

**文件：backend/data/tool_config.go**

```go
package data

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sync"
    "time"
    "go-stock/backend/logger"
)

var (
    configCache *ToolConfig
    cacheLock   sync.RWMutex
    lastLoaded  time.Time
    cacheTTL    = 30 * time.Second // 30 秒缓存
)

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
    return &ToolConfig{
        Version: "1.0",
        Tools: []ToolItem{
            {
                Name:     "SearchStockByIndicators",
                Type:     "builtin",
                Enabled:  true,
                Config:   map[string]interface{}{},
            },
            {
                Name:     "GetStockKLine",
                Type:     "builtin",
                Enabled:  true,
                Config:   map[string]interface{}{},
            },
            {
                Name:     "InteractiveAnswer",
                Type:     "builtin",
                Enabled:  true,
                Config:   map[string]interface{}{},
            },
            {
                Name:     "GetStockResearchReport",
                Type:     "builtin",
                Enabled:  true,
                Config:   map[string]interface{}{},
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
```

#### 2. 工具工厂

**文件：backend/data/tool_factory.go**

```go
package data

import (
    "fmt"
    "context"
    "go-stock/backend/agent"
)

// CreateTool 根据配置创建工具
func CreateTool(item ToolItem) (Tool, error) {
    switch item.Type {
    case "builtin":
        return createBuiltinTool(item)
    case "mcp":
        return createMCPTool(item)
    case "http":
        return createHTTPTool(item)
    default:
        return Tool{}, fmt.Errorf("未知工具类型: %s", item.Type)
    }
}

// createBuiltinTool 创建内置工具
func createBuiltinTool(item ToolItem) (Tool, error) {
    toolsMap := getBuiltinToolsMap()
    tool, exists := toolsMap[item.Name]
    if !exists {
        return Tool{}, fmt.Errorf("未知的内置工具: %s", item.Name)
    }
    return tool, nil
}

// getBuiltinToolsMap 获取内置工具映射
func getBuiltinToolsMap() map[string]Tool {
    return map[string]Tool{
        "SearchStockByIndicators": {
            Type: "function",
            Function: ToolFunction{
                Name:        "SearchStockByIndicators",
                Description: "根据自然语言筛选股票",
                Parameters: FunctionParameters{
                    Type: "object",
                    Properties: map[string]any{
                        "words": map[string]string{
                            "type":        "string",
                            "description": "选股自然语言描述，例如：涨停、涨幅大于5%、科技股等",
                        },
                    },
                    Required: []string{"words"},
                },
            },
        },
        "GetStockKLine": {
            Type: "function",
            Function: ToolFunction{
                Name:        "GetStockKLine",
                Description: "获取股票K线数据",
                Parameters: FunctionParameters{
                    Type: "object",
                    Properties: map[string]any{
                        "stockCode": map[string]string{
                            "type":        "string",
                            "description": "股票代码，如：sh000001",
                        },
                        "days": map[string]string{
                            "type":        "string",
                            "description": "获取多少天的K线数据，默认90天",
                        },
                    },
                    Required: []string{"stockCode"},
                },
            },
        },
        "InteractiveAnswer": {
            Type: "function",
            Function: ToolFunction{
                Name:        "InteractiveAnswer",
                Description: "获取投资者与上市公司互动问答数据",
                Parameters: FunctionParameters{
                    Type: "object",
                    Properties: map[string]any{
                        "page": map[string]string{
                            "type":        "string",
                            "description": "页码，默认1",
                        },
                        "pageSize": map[string]string{
                            "type":        "string",
                            "description": "每页数量，默认50",
                        },
                        "keyWord": map[string]string{
                            "type":        "string",
                            "description": "关键词",
                        },
                    },
                    Required: []string{},
                },
            },
        },
        "GetStockResearchReport": {
            Type: "function",
            Function: ToolFunction{
                Name:        "GetStockResearchReport",
                Description: "获取个股研报",
                Parameters: FunctionParameters{
                    Type: "object",
                    Properties: map[string]any{
                        "stockCode": map[string]string{
                            "type":        "string",
                            "description": "股票代码，如：sh000001",
                        },
                    },
                    Required: []string{"stockCode"},
                },
            },
        },
    }
}

// createMCPTool 创建 MCP 工具
func createMCPTool(item ToolItem) (Tool, error) {
    // 获取工具注册表
    toolReg := agent.GetToolRegistry(context.Background())
    if toolReg == nil {
        return Tool{}, fmt.Errorf("工具注册表未初始化")
    }

    allTools := toolReg.GetAllTools()

    // 查找 MCP 工具
    for _, t := range allTools {
        info, err := t.Info(context.Background())
        if err != nil {
            continue
        }
        if info.Name == item.Name {
            // 将 tool.InvokableTool 转换为 data.Tool 格式
            return convertMCPToolToDataTool(info), nil
        }
    }

    return Tool{}, fmt.Errorf("MCP 工具未找到: %s", item.Name)
}

// convertMCPToolToDataTool 将 MCP 工具转换为 data.Tool 格式
func convertMCPToolToDataTool(info *schema.ToolInfo) Tool {
    // 简化版转换，只提取基本信息
    return Tool{
        Type: "function",
        Function: ToolFunction{
            Name:        info.Name,
            Description: info.Desc,
            Parameters: FunctionParameters{
                Type:       "object",
                Properties: map[string]any{}, // 简化处理
                Required:   []string{},
            },
        },
    }
}

// createHTTPTool 创建 HTTP 工具
func createHTTPTool(item ToolItem) (Tool, error) {
    // 解析 HTTP 工具配置
    url, ok := item.Config["url"].(string)
    if !ok || url == "" {
        return Tool{}, fmt.Errorf("HTTP 工具需要配置 url")
    }

    method, ok := item.Config["method"].(string)
    if !ok {
        method = "POST"
    }

    // 从配置中解析参数定义
    var properties map[string]any = map[string]any{}
    if params, ok := item.Config["parameters"].(map[string]any); ok {
        properties = params
    }

    description, _ := item.Config["description"].(string)
    if description == "" {
        description = "HTTP 接口工具"
    }

    // 构建工具定义
    return Tool{
        Type: "function",
        Function: ToolFunction{
            Name:        item.Name,
            Description: description,
            Parameters: FunctionParameters{
                Type:       "object",
                Properties: properties,
                Required:   getRequiredParams(item.Config),
            },
        },
    }, nil
}

// getRequiredParams 获取必需参数
func getRequiredParams(config map[string]any) []string {
    if required, ok := config["required"].([]string); ok {
        return required
    }
    return []string{}
}
```

#### 3. 修改 StockPickService

**文件：backend/data/stock_pick_service.go**

```go
// getStockPickTools 获取荐股工具列表（从配置加载）
func (s *StockPickService) getStockPickTools() []Tool {
    var tools []Tool

    // 加载工具配置
    config, err := LoadToolConfig()
    if err != nil {
        logger.SugaredLogger.Errorf("加载工具配置失败: %v", err)
        return s.getDefaultTools() // 使用默认配置
    }

    // 根据配置创建工具
    for _, toolItem := range config.Tools {
        if !toolItem.Enabled {
            continue
        }

        tool, err := CreateTool(toolItem)
        if err != nil {
            logger.SugaredLogger.Errorf("创建工具失败 %s: %v", toolItem.Name, err)
            continue
        }

        tools = append(tools, tool)
    }

    // 如果配置中没有工具，返回默认工具
    if len(tools) == 0 {
        logger.SugaredLogger.Warnf("工具配置为空，使用默认工具")
        return s.getDefaultTools()
    }

    return tools
}

// getDefaultTools 获取默认工具列表（保持向后兼容）
func (s *StockPickService) getDefaultTools() []Tool {
    return []Tool{
        {
            Type: "function",
            Function: ToolFunction{
                Name:        "SearchStockByIndicators",
                Description: "根据自然语言筛选股票",
                Parameters: FunctionParameters{
                    Type: "object",
                    Properties: map[string]any{
                        "words": map[string]string{
                            "type":        "string",
                            "description": "选股自然语言描述，例如：涨停、涨幅大于5%、科技股等",
                        },
                    },
                    Required: []string{"words"},
                },
            },
        },
        {
            Type: "function",
            Function: ToolFunction{
                Name:        "GetStockKLine",
                Description: "获取股票K线数据",
                Parameters: FunctionParameters{
                    Type: "object",
                    Properties: map[string]any{
                        "stockCode": map[string]string{
                            "type":        "string",
                            "description": "股票代码，如：sh000001",
                        },
                        "days": map[string]string{
                            "type":        "string",
                            "description": "获取多少天的K线数据，默认90天",
                        },
                    },
                    Required: []string{"stockCode"},
                },
            },
        },
        {
            Type: "function",
            Function: ToolFunction{
                Name:        "InteractiveAnswer",
                Description: "获取投资者与上市公司互动问答数据",
                Parameters: FunctionParameters{
                    Type: "object",
                    Properties: map[string]any{
                        "page": map[string]string{
                            "type":        "string",
                            "description": "页码，默认1",
                        },
                        "pageSize": map[string]string{
                            "type":        "string",
                            "description": "每页数量，默认50",
                        },
                        "keyWord": map[string]string{
                            "type":        "string",
                            "description": "关键词",
                        },
                    },
                    Required: []string{},
                },
            },
        },
        {
            Type: "function",
            Function: ToolFunction{
                Name:        "GetStockResearchReport",
                Description: "获取个股研报",
                Parameters: FunctionParameters{
                    Type: "object",
                    Properties: map[string]any{
                        "stockCode": map[string]string{
                            "type":        "string",
                            "description": "股票代码，如：sh000001",
                        },
                    },
                    Required: []string{"stockCode"},
                },
            },
        },
    }
}
```

#### 4. 修改 App 初始化

**文件：app.go**

```go
func NewApp() *App {
    cacheSize := 512 * 1024
    cache := freecache.NewCache(cacheSize)
    c := cron.New(cron.WithSeconds())
    c.Start()

    // 从配置加载工具
    var tools []data.Tool
    config, err := data.LoadToolConfig()
    if err != nil {
        logger.SugaredLogger.Errorf("加载工具配置失败: %v", err)
        tools = AddTools([]data.Tool{}) // 降级到硬编码
    } else {
        tools = loadToolsFromConfig(config)
    }

    return &App{
        ctx:               nil,
        cache:             cache,
        cron:              c,
        cronEntrys:        make(map[string]cron.EntryID),
        AiTools:           tools,
        SponsorInfo:       map[string]any{},
        PromptTemplateSvc: data.NewPromptTemplateApi(),
    }
}

func loadToolsFromConfig(config *data.ToolConfig) []data.Tool {
    var tools []data.Tool

    for _, toolItem := range config.Tools {
        if !toolItem.Enabled {
            continue
        }

        tool, err := data.CreateTool(toolItem)
        if err != nil {
            logger.SugaredLogger.Errorf("创建工具失败 %s: %v", toolItem.Name, err)
            continue
        }

        tools = append(tools, tool)
    }

    return tools
}
```

#### 5. 添加工具管理 API

**文件：app.go**

```go
// GetToolConfig 获取工具配置
func (a *App) GetToolConfig() string {
    config, err := data.LoadToolConfig()
    if err != nil {
        logger.SugaredLogger.Errorf("获取工具配置失败: %v", err)
        return "{}"
    }

    bytes, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        logger.SugaredLogger.Errorf("序列化工具配置失败: %v", err)
        return "{}"
    }

    return string(bytes)
}

// SetToolConfig 更新工具配置
func (a *App) SetToolConfig(configStr string) string {
    var config data.ToolConfig
    if err := json.Unmarshal([]byte(configStr), &config); err != nil {
        logger.SugaredLogger.Errorf("解析工具配置失败: %v", err)
        return "无效的配置格式"
    }

    if err := data.SaveToolConfig(&config); err != nil {
        logger.SugaredLogger.Errorf("保存工具配置失败: %v", err)
        return "保存配置失败"
    }

    // 重新加载工具
    a.AiTools = loadToolsFromConfig(&config)

    logger.SugaredLogger.Infof("工具配置已更新，加载了 %d 个工具", len(a.AiTools))
    return "工具配置已更新"
}

// ReloadTools 重新加载工具
func (a *App) ReloadTools() string {
    config, err := data.LoadToolConfig()
    if err != nil {
        logger.SugaredLogger.Errorf("重新加载工具配置失败: %v", err)
        return "重新加载失败"
    }

    a.AiTools = loadToolsFromConfig(config)
    logger.SugaredLogger.Infof("工具已重新加载，共 %d 个工具", len(a.AiTools))

    return fmt.Sprintf("已重新加载 %d 个工具", len(a.AiTools))
}
```

## 风险评估

### 风险 1：配置文件格式错误

**影响程度：** 中

**缓解措施：**
- 加载配置时验证格式
- 配置无效时使用默认配置
- 详细的错误日志记录

### 风险 2：工具类型不支持

**影响程度：** 低

**缓解措施：**
- 对未知工具类型提供警告
- 忽略不支持的工具类型
- 保持与现有系统兼容性

### 风险 3：HTTP 工具调用失败

**影响程度：** 中

**缓解措施：**
- HTTP 请求超时处理
- 错误状态码处理
- 重试机制

### 风险 4：配置更新后工具未生效

**影响程度：** 低

**缓解措施：**
- 更新配置后立即重新加载工具
- 提供配置更新确认信息
- 工具加载失败时回退到默认配置

## 实现阶段分解

### 阶段 1：基础架构
1. 创建 `data/tools/config.json` 配置文件
2. 创建 `backend/data/tool_config.go` 配置管理类
3. 创建 `backend/data/tool_factory.go` 工具工厂

### 阶段 2：配置加载
1. 修改 `StockPickService.getStockPickTools()` 使用配置加载
2. 修改 `App.NewApp()` 使用配置加载工具
3. 添加配置加载和验证逻辑

### 阶段 3：多类型工具支持
1. 实现内置工具加载
2. 实现 MCP 工具加载
3. 实现 HTTP 工具加载

### 阶段 4：API 和 UI
1. 添加 `GetToolConfig` API
2. 添加 `SetToolConfig` API
3. 添加 `ReloadTools` API
4. 创建前端工具管理 UI 组件

### 阶段 5：测试和优化
1. 编写单元测试
2. 性能测试
3. 集成测试
4. 错误处理优化

## 估算复杂度

- **后端开发：** 6-8 小时
- **前端开发：** 4-5 小时
- **测试：** 3-4 小时
- **文档：** 1-2 小时

**总计：** 14-19 小时

**复杂度：** 中等

## 下一步

等待用户确认后开始实现。
