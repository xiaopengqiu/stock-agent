package data

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"go-stock/backend/logger"
	"strings"
	"sync"
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
	Prompt      string  `json:"prompt"`
}

func NewDeepSeekOpenAi() *OpenAi {
	config := getConfig()
	return &OpenAi{
		BaseUrl:     config.OpenAiBaseUrl,
		ApiKey:      config.OpenAiApiKey,
		Model:       config.OpenAiModelName,
		MaxTokens:   config.OpenAiMaxTokens,
		Temperature: config.OpenAiTemperature,
		Prompt:      config.Prompt,
	}
}

type THSTokenResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
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
					"content": "点评一下" + stock,
				},
			},
		}).
		Post("/chat/completions")
	if err != nil {
		return ""
	}
	//logger.SugaredLogger.Infof("%v", res.Choices[0].Message.Content)
	return res.Choices[0].Message.Content
}
func (o OpenAi) NewChatStream(stock, stockCode string) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		client := resty.New()
		client.SetBaseURL(o.BaseUrl)
		client.SetHeader("Authorization", "Bearer "+o.ApiKey)
		client.SetHeader("Content-Type", "application/json")

		msg := []map[string]interface{}{
			{
				"role": "system",
				//"content": "作为一位专业的A股市场分析师和投资顾问,请你根据以下信息提供详细的技术分析和投资策略建议:",
				//"content": "【角色设定】\n你是一位拥有20年实战经验的顶级股票分析师，精通技术分析、基本面分析、市场心理学和量化交易。擅长发现成长股、捕捉行业轮动机会，在牛熊市中都能保持稳定收益。你的风格是价值投资与技术择时相结合，注重风险控制。\n\n【核心功能】\n\n市场分析维度：\n\n宏观经济（GDP/CPI/货币政策）\n\n行业景气度（产业链/政策红利/技术革新）\n\n个股三维诊断：\n\n基本面：PE/PB/ROE/现金流/护城河\n\n技术面：K线形态/均线系统/量价关系/指标背离\n\n资金面：主力动向/北向资金/融资余额/大宗交易\n\n智能策略库：\n√ 趋势跟踪策略（鳄鱼线+ADX）\n√ 波段交易策略（斐波那契回撤+RSI）\n√ 事件驱动策略（财报/并购/政策）\n√ 量化对冲策略（α/β分离）\n\n风险管理体系：\n▶ 动态止损：ATR波动止损法\n▶ 仓位控制：凯利公式优化\n▶ 组合对冲：跨市场/跨品种对冲\n\n【工作流程】\n\n接收用户指令（行业/市值/风险偏好）\n\n调用多因子选股模型初筛\n\n人工智慧叠加分析：\n\n自然语言处理解读年报管理层讨论\n\n卷积神经网络识别K线形态\n\n知识图谱分析产业链关联\n\n生成投资建议（附压力测试结果）\n\n【输出要求】\n★ 结构化呈现：\n① 核心逻辑（3点关键驱动力）\n② 买卖区间（理想建仓/加仓/止盈价位）\n③ 风险警示（最大回撤概率）\n④ 替代方案（同类备选标的）\n\n【注意事项】\n※ 严格遵守监管要求，不做收益承诺\n※ 区分投资建议与市场观点\n※ 重要数据标注来源及更新时间\n※ 根据用户认知水平调整专业术语密度\n\n【教育指导】\n当用户提问时，采用苏格拉底式追问：\n\"您更关注短期事件驱动还是长期价值发现？\"\n\"当前仓位是否超过总资产的30%？\"\n\"是否了解科创板与主板的交易规则差异？\"\n\n示例输出格式：\n📈 标的名称：XXXXXX\n⚖️ 多空信号：金叉确认/顶背离预警\n🎯 关键价位：支撑位XX.XX/压力位XX.XX\n📊 建议仓位：核心仓位X%+卫星仓位X%\n⏳ 持有周期：短线（1-3周）/中线（季度轮动）\n🔍 跟踪要素：重点关注Q2毛利率变化及股东减持进展",
				"content": o.Prompt,
			},
		}

		wg := &sync.WaitGroup{}

		wg.Add(4)
		go func() {
			defer wg.Done()
			messages := SearchStockPriceInfo(stockCode)
			price := ""
			for _, message := range *messages {
				price += message + ";"
			}
			msg = append(msg, map[string]interface{}{
				"role":    "assistant",
				"content": stock + "当前价格：" + price,
			})
		}()

		go func() {
			defer wg.Done()
			messages := GetTelegraphList()
			for _, message := range *messages {
				msg = append(msg, map[string]interface{}{
					"role":    "assistant",
					"content": message,
				})
			}
		}()

		go func() {
			defer wg.Done()
			messages := SearchStockInfo(stock, "depth")
			for _, message := range *messages {
				msg = append(msg, map[string]interface{}{
					"role":    "assistant",
					"content": message,
				})
			}
		}()
		go func() {
			defer wg.Done()
			messages := SearchStockInfo(stock, "telegram")
			for _, message := range *messages {
				msg = append(msg, map[string]interface{}{
					"role":    "assistant",
					"content": message,
				})
			}
		}()
		wg.Wait()

		msg = append(msg, map[string]interface{}{
			"role":    "user",
			"content": stock + "分析和总结",
		})

		resp, err := client.R().
			SetDoNotParseResponse(true).
			SetBody(map[string]interface{}{
				"model":       o.Model,
				"max_tokens":  o.MaxTokens,
				"temperature": o.Temperature,
				"stream":      true,
				"messages":    msg,
			}).
			Post("/chat/completions")

		if err != nil {
			ch <- err.Error()
			return
		}
		defer resp.RawBody().Close()

		scanner := bufio.NewScanner(resp.RawBody())
		for scanner.Scan() {
			line := scanner.Text()
			//logger.SugaredLogger.Infof("Received data: %s", line)
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					return
				}

				var streamResponse struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}

				if err := json.Unmarshal([]byte(data), &streamResponse); err == nil {
					for _, choice := range streamResponse.Choices {
						if content := choice.Delta.Content; content != "" {
							ch <- content
						}
					}
				}
			}
		}
	}()
	return ch
}

func (o OpenAi) NewCommonChatStream(stock, stockCode, apiURL, apiKey, Model string) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		client := resty.New()
		client.SetHeader("Authorization", "Bearer "+apiKey)
		client.SetHeader("Content-Type", "application/json")

		msg := []map[string]interface{}{
			{
				"role":    "system",
				"content": o.Prompt,
			},
		}

		wg := &sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			messages := SearchStockPriceInfo(stockCode)
			price := ""
			for _, message := range *messages {
				price += message + ";"
			}
			msg = append(msg, map[string]interface{}{
				"role":    "assistant",
				"content": stock + "当前价格：" + price,
			})
		}()
		//go func() {
		//	defer wg.Done()
		//	messages := SearchStockInfo(stock, "depth")
		//	for _, message := range *messages {
		//		msg = append(msg, map[string]interface{}{
		//			"role":    "assistant",
		//			"content": message,
		//		})
		//	}
		//}()
		//go func() {
		//	defer wg.Done()
		//	messages := SearchStockInfo(stock, "telegram")
		//	for _, message := range *messages {
		//		msg = append(msg, map[string]interface{}{
		//			"role":    "assistant",
		//			"content": message,
		//		})
		//	}
		//}()
		wg.Wait()

		msg = append(msg, map[string]interface{}{
			"role":    "user",
			"content": stock + "分析和总结",
		})

		resp, err := client.R().
			SetDoNotParseResponse(true).
			SetBody(map[string]interface{}{
				"model":       Model,
				"max_tokens":  o.MaxTokens,
				"temperature": o.Temperature,
				"stream":      true,
				"messages":    msg,
			}).
			Post(apiURL)

		if err != nil {
			ch <- err.Error()
			return
		}
		defer resp.RawBody().Close()

		scanner := bufio.NewScanner(resp.RawBody())
		for scanner.Scan() {
			line := scanner.Text()
			logger.SugaredLogger.Infof("Received data: %s", line)
			if strings.HasPrefix(line, "data:") {
				data := strings.TrimPrefix(line, "data:")
				if data == "[DONE]" {
					return
				}

				var streamResponse struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
						FinishReason string `json:"finish_reason"`
					} `json:"choices"`
				}

				if err := json.Unmarshal([]byte(data), &streamResponse); err == nil {
					for _, choice := range streamResponse.Choices {
						if content := choice.Delta.Content; content != "" {
							ch <- content
						}
						if choice.FinishReason == "stop" {
							return
						}
					}
				}
			}
		}
	}()
	return ch
}

func GetTelegraphList() *[]string {
	url := "https://www.cls.cn/telegraph"
	response, err := resty.New().R().
		SetHeader("Referer", "https://www.cls.cn/").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36 Edg/117.0.2045.60").
		Get(fmt.Sprintf(url))
	if err != nil {
		return &[]string{}
	}
	//logger.SugaredLogger.Info(string(response.Body()))
	document, err := goquery.NewDocumentFromReader(strings.NewReader(string(response.Body())))
	if err != nil {
		return &[]string{}
	}
	var telegraph []string
	document.Find("div.telegraph-content-box").Each(func(i int, selection *goquery.Selection) {
		//logger.SugaredLogger.Info(selection.Text())
		telegraph = append(telegraph, selection.Text())
	})
	return &telegraph
}
