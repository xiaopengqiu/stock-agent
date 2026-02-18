package data

import (
	"fmt"
	"go-stock/backend/logger"
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
				Description: "根据自然语言筛选股票，返回自然语言选股条件要求的股票所有相关数据。输入股票名称可以获取当前股票最新的股价交易数据和基础财务指标信息，多个股票名称使用,分隔。",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]any{
						"words": map[string]any{
							"type": "string",
							"description": "选股自然语言。" +
								"例1：创新药,半导体;PE<30;净利润增长率>50%。 " +
								"例2：上证指数,科创50。 " +
								"例3：长电科技,上海贝岭。" +
								"例4：长电科技,上海贝岭;KDJ,MACD,RSI,BOLL,主力净流入/流出" +
								"例5：换手率大于3%小于25%.量比1以上. 10日内有过涨停.股价处于峰值的二分之一以下.流通股本<100亿.当日和连续四日净流入;股价在20日均线以上.分时图股价在均线之上.热门板块下涨幅领先的A股. 当日量能20000手以上.沪深个股.近一年市盈率波动小于150%.MACD金叉;不要ST股及不要退市股，非北交所，每股收益>0。" +
								"例6：沪深主板.流通市值小于100亿.市值大于10亿.60分钟dif大于dea.60分钟skdj指标k值大于d值.skdj指标k值小于90.换手率大于3%.成交额大于1亿元.量比大于2.涨幅大于2%小于7%.股价大于5小于50.创业板.10日均线大于20日均线;不要ST股及不要退市股;不要北交所;不要科创板;不要创业板。" +
								"例7：股价在20日线上，一月之内涨停次数>=1，量比大于1，换手率大于3%，流通市值大于 50亿小于200亿。" +
								"例8：基本条件：前期有爆量，回调到 10 日线，当日是缩量阴线，均线趋势向上。;优选条件：一月之内涨停次数>=1",
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
				Description: "获取股票日K线数据。",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]any{
						"days": map[string]any{
							"type":        "string",
							"description": "日K数据条数",
						},
						"stockCode": map[string]any{
							"type":        "string",
							"description": "股票代码（A股：sh,sz开头;港股hk开头,美股：us开头）",
						},
					},
					Required: []string{"days", "stockCode"},
				},
			},
		},
		"InteractiveAnswer": {
			Type: "function",
			Function: ToolFunction{
				Name:        "InteractiveAnswer",
				Description: "获取投资者与上市公司互动问答的数据,反映当前投资者关注的热点问题",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]any{
						"page": map[string]any{
							"type":        "string",
							"description": "分页号",
						},
						"pageSize": map[string]any{
							"type":        "string",
							"description": "分页大小",
						},
						"keyWord": map[string]any{
							"type":        "string",
							"description": "搜索关键词（可输入股票名称或者当前热门板块/行业/概念/标的/事件等）",
						},
					},
					Required: []string{"page", "pageSize"},
				},
			},
		},
		"GetStockResearchReport": {
			Type: "function",
			Function: ToolFunction{
				Name:        "GetStockResearchReport",
				Description: "获取股票的分析/研究报告",
				Parameters: FunctionParameters{
					Type: "object",
					Properties: map[string]any{
						"stockCode": map[string]any{
							"type":        "string",
							"description": "股票代码",
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
	// TODO: 实现MCP 工具集成
	// 目前先返回空实现，后续完善
	logger.SugaredLogger.Warnf("MCP 工具支持尚未实现: %s", item.Name)
	return Tool{}, fmt.Errorf("MCP 工具支持尚未实现")
}

// createHTTPTool 创建 HTTP 工具
func createHTTPTool(item ToolItem) (Tool, error) {
	// 解析 HTTP 工具配置
	url, ok := item.Config["url"].(string)
	if !ok || url == "" {
		return Tool{}, fmt.Errorf("HTTP 工具需要配置 url")
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
