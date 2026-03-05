package tools

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"go-stock/backend/data"
	"strings"
)

// @Author StockAgent
// @Date 2026/3/3
// @Desc 港股价格查询工具
//-----------------------------------------------------------------------------------

func GetQueryHKStockPriceTool() tool.InvokableTool {
	return &ToolQueryHKStockPrice{}
}

type ToolQueryHKStockPrice struct{}

func (t ToolQueryHKStockPrice) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "QueryHKStockPrice",
		Desc: "查询港股实时价格数据，支持港股通标的和主要港股。港股价格对A+H股两地上市的科技公司、医药股、金融股有重要参考价值。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"stockCodes": {
				Type:     "string",
				Desc:     "港股代码，多个用逗号分隔。支持格式：00700、0700.HK、hk00700。港股通标的如：腾讯(00700)、美团(03690)、小米(01810)等",
				Required: true,
			},
		}),
	}, nil
}

func (t ToolQueryHKStockPrice) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	parms := map[string]any{}
	err := json.Unmarshal([]byte(argumentsInJSON), &parms)
	if err != nil {
		return "", err
	}

	stockCodes := strings.Split(parms["stockCodes"].(string), ",")
	var codes []string
	for _, code := range stockCodes {
		// 统一转换为hk开头的格式
		hkCode := normalizeHKStockCode(code)
		codes = append(codes, hkCode)
	}

	// 调用数据API获取港股实时数据
	stockDataApi := data.StockDataApi{}
	realTimeData, err := stockDataApi.GetStockCodeRealTimeData(codes...)
	if err != nil {
		return "", err
	}

	// 添加港股特有的分析字段
	hkAnalysis := t.enhanceHKStockData(realTimeData)

	marshal, err := json.Marshal(hkAnalysis)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

// normalizeHKStockCode 将各种格式的港股代码统一为hk开头
func normalizeHKStockCode(code string) string {
	code = strings.TrimSpace(strings.ToUpper(code))

	// 去除.HK后缀
	if strings.HasSuffix(code, ".HK") {
		code = strings.TrimSuffix(code, ".HK")
	}

	// 去除hk前缀
	if strings.HasPrefix(code, "HK") {
		code = code[2:]
	}

	// 补足5位代码
	for len(code) < 5 {
		code = "0" + code
	}

	return "hk" + code
}

// enhanceHKStockData 增强港股数据，添加AH股溢价分析等
func (t ToolQueryHKStockPrice) enhanceHKStockData(data interface{}) interface{} {
	// TODO: 添加港股特有的分析字段
	// 1. AH股溢价率（如果是A+H股）
	// 2. 港股通资金流向
	// 3. 港股特有技术指标
	return data
}
