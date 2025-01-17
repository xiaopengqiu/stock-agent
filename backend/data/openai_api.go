package data

import (
	"github.com/go-resty/resty/v2"
	"go-stock/backend/logger"
)

// @Author spark
// @Date 2025/1/16 13:19
// @Desc
// -----------------------------------------------------------------------------------
type OpenAi struct {
	BaseUrl     string  `json:"base_url"`
	ApiKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

func NewDeepSeekOpenAi() *OpenAi {
	config := getConfig()
	return &OpenAi{
		BaseUrl:     config.OpenAiBaseUrl,
		ApiKey:      config.OpenAiApiKey,
		Model:       config.OpenAiModelName,
		MaxTokens:   config.OpenAiMaxTokens,
		Temperature: config.OpenAiTemperature,
	}
}

type AiResponse struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens          int `json:"prompt_tokens"`
		CompletionTokens      int `json:"completion_tokens"`
		TotalTokens           int `json:"total_tokens"`
		PromptCacheHitTokens  int `json:"prompt_cache_hit_tokens"`
		PromptCacheMissTokens int `json:"prompt_cache_miss_tokens"`
	} `json:"usage"`
	SystemFingerprint string `json:"system_fingerprint"`
}

func (o OpenAi) NewChat(stock string) string {
	client := resty.New()
	client.SetBaseURL(o.BaseUrl)
	client.SetHeader("Authorization", "Bearer "+o.ApiKey)
	client.SetHeader("Content-Type", "application/json")

	res := &AiResponse{}
	_, err := client.R().
		SetResult(res).
		SetBody(map[string]interface{}{
			"model":       o.Model,
			"max_tokens":  o.MaxTokens,
			"temperature": o.Temperature,
			"messages": []map[string]interface{}{
				{
					"role": "system",
					"content": "作为一位专业的A股市场分析师和投资顾问,请你根据以下信息提供详细的技术分析和投资策略建议:" +
						"1. 市场背景:\n" +
						"- 当前A股市场整体走势(如:牛市、熊市、震荡市)\n " +
						"- 近期影响市场的主要宏观经济因素\n " +
						"- 市场情绪指标(如:融资融券余额、成交量变化)  " +
						"2. 技术指标分析: " +
						"- 当前股价水平" +
						"- 所在boll区间" +
						"- 上证指数的MA(移动平均线)、MACD、KDJ指标分析\n " +
						"- 行业板块轮动情况\n " +
						"- 近期市场热点和龙头股票的技术形态  " +
						"3. 风险评估:\n " +
						"- 当前市场主要风险因素\n " +
						"- 如何设置止损和止盈位\n " +
						"- 资金管理建议(如:仓位控制)  " +
						"4. 投资策略:\n " +
						"- 短期(1-2周)、中期(1-3月)和长期(3-6月)的市场预期\n " +
						"- 不同风险偏好投资者的策略建议\n " +
						"- 值得关注的行业板块和个股推荐(请给出2-3个具体例子,包括股票代码和名称)  " +
						"5. 技术面和基本面结合:\n " +
						"- 如何将技术分析与公司基本面分析相结合\n " +
						"- 识别高质量股票的关键指标  " +
						"请提供详细的分析和具体的操作建议,包括入场时机、持仓周期和退出策略。同时,请强调风险控制的重要性,并提醒投资者需要根据自身情况做出决策。  " +
						"你的分析和建议应当客观、全面,并基于当前可获得的市场数据。如果某些信息无法确定,请明确指出并解释原因。",
				},
				{
					"role":    "user",
					"content": "点评一下" + stock + ",以Markdown输出",
				},
			},
		}).
		Post("/chat/completions")
	if err != nil {
		return ""
	}
	logger.SugaredLogger.Infof("%v", res.Choices[0].Message.Content)
	return res.Choices[0].Message.Content
}
