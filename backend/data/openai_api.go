package data

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/go-resty/resty/v2"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	"strings"
	"sync"
	"time"
)

// @Author spark
// @Date 2025/1/16 13:19
// @Desc
// -----------------------------------------------------------------------------------
type OpenAi struct {
	ctx              context.Context
	BaseUrl          string  `json:"base_url"`
	ApiKey           string  `json:"api_key"`
	Model            string  `json:"model"`
	MaxTokens        int     `json:"max_tokens"`
	Temperature      float64 `json:"temperature"`
	Prompt           string  `json:"prompt"`
	TimeOut          int     `json:"time_out"`
	QuestionTemplate string  `json:"question_template"`
	CrawlTimeOut     int64   `json:"crawl_time_out"`
}

func NewDeepSeekOpenAi(ctx context.Context) *OpenAi {
	config := getConfig()
	if config.OpenAiEnable {
		if config.OpenAiApiTimeOut <= 0 {
			config.OpenAiApiTimeOut = 60 * 5
		}
		if config.CrawlTimeOut <= 0 {
			config.CrawlTimeOut = 60
		}
	}
	return &OpenAi{
		ctx:              ctx,
		BaseUrl:          config.OpenAiBaseUrl,
		ApiKey:           config.OpenAiApiKey,
		Model:            config.OpenAiModelName,
		MaxTokens:        config.OpenAiMaxTokens,
		Temperature:      config.OpenAiTemperature,
		Prompt:           config.Prompt,
		TimeOut:          config.OpenAiApiTimeOut,
		QuestionTemplate: config.QuestionTemplate,
		CrawlTimeOut:     config.CrawlTimeOut,
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

func (o OpenAi) NewChatStream(stock, stockCode string) <-chan string {
	ch := make(chan string, 512)

	defer func() {
		if err := recover(); err != nil {
			logger.SugaredLogger.Error("NewChatStream panic", err)
		}
	}()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.SugaredLogger.Errorf("NewChatStream goroutine  panic :%s", err)
				logger.SugaredLogger.Errorf("NewChatStream goroutine  panic  stock:%s stockCode:%s", stock, stockCode)
				logger.SugaredLogger.Errorf("NewChatStream goroutine  panic  config:%v", o)
			}
		}()
		defer close(ch)
		msg := []map[string]interface{}{
			{
				"role": "system",
				//"content": "作为一位专业的A股市场分析师和投资顾问,请你根据以下信息提供详细的技术分析和投资策略建议:",
				//"content": "【角色设定】\n你是一位拥有20年实战经验的顶级股票分析师，精通技术分析、基本面分析、市场心理学和量化交易。擅长发现成长股、捕捉行业轮动机会，在牛熊市中都能保持稳定收益。你的风格是价值投资与技术择时相结合，注重风险控制。\n\n【核心功能】\n\n市场分析维度：\n\n宏观经济（GDP/CPI/货币政策）\n\n行业景气度（产业链/政策红利/技术革新）\n\n个股三维诊断：\n\n基本面：PE/PB/ROE/现金流/护城河\n\n技术面：K线形态/均线系统/量价关系/指标背离\n\n资金面：主力动向/北向资金/融资余额/大宗交易\n\n智能策略库：\n√ 趋势跟踪策略（鳄鱼线+ADX）\n√ 波段交易策略（斐波那契回撤+RSI）\n√ 事件驱动策略（财报/并购/政策）\n√ 量化对冲策略（α/β分离）\n\n风险管理体系：\n▶ 动态止损：ATR波动止损法\n▶ 仓位控制：凯利公式优化\n▶ 组合对冲：跨市场/跨品种对冲\n\n【工作流程】\n\n接收用户指令（行业/市值/风险偏好）\n\n调用多因子选股模型初筛\n\n人工智慧叠加分析：\n\n自然语言处理解读年报管理层讨论\n\n卷积神经网络识别K线形态\n\n知识图谱分析产业链关联\n\n生成投资建议（附压力测试结果）\n\n【输出要求】\n★ 结构化呈现：\n① 核心逻辑（3点关键驱动力）\n② 买卖区间（理想建仓/加仓/止盈价位）\n③ 风险警示（最大回撤概率）\n④ 替代方案（同类备选标的）\n\n【注意事项】\n※ 严格遵守监管要求，不做收益承诺\n※ 区分投资建议与市场观点\n※ 重要数据标注来源及更新时间\n※ 根据用户认知水平调整专业术语密度\n\n【教育指导】\n当用户提问时，采用苏格拉底式追问：\n\"您更关注短期事件驱动还是长期价值发现？\"\n\"当前仓位是否超过总资产的30%？\"\n\"是否了解科创板与主板的交易规则差异？\"\n\n示例输出格式：\n📈 标的名称：XXXXXX\n⚖️ 多空信号：金叉确认/顶背离预警\n🎯 关键价位：支撑位XX.XX/压力位XX.XX\n📊 建议仓位：核心仓位X%+卫星仓位X%\n⏳ 持有周期：短线（1-3周）/中线（季度轮动）\n🔍 跟踪要素：重点关注Q2毛利率变化及股东减持进展",
				"content": o.Prompt,
			},
		}

		replaceTemplates := map[string]string{
			"{{stockName}}": RemoveAllBlankChar(stock),
			"{{stockCode}}": RemoveAllBlankChar(stockCode),
		}

		followedStock := &FollowedStock{
			StockCode: stockCode,
		}
		db.Dao.Model(&followedStock).Where("stock_code = ?", stockCode).First(followedStock)
		if followedStock.CostPrice > 0 {
			replaceTemplates["{{costPrice}}"] = fmt.Sprintf("%.2f", followedStock.CostPrice)
		}

		question := strutil.ReplaceWithMap(o.QuestionTemplate, replaceTemplates)
		logger.SugaredLogger.Infof("NewChatStream stock:%s stockCode:%s", stock, stockCode)
		logger.SugaredLogger.Infof("Prompt：%s", o.Prompt)
		logger.SugaredLogger.Infof("User Prompt config:%v", o.QuestionTemplate)
		logger.SugaredLogger.Infof("User question:%s", question)

		wg := &sync.WaitGroup{}
		wg.Add(5)
		go func() {
			defer wg.Done()
			messages := SearchStockPriceInfo(stockCode, o.CrawlTimeOut)
			if messages == nil || len(*messages) == 0 {
				logger.SugaredLogger.Error("获取股票价格失败")
				ch <- "***❗获取股票价格失败,分析结果可能不准确***<hr>"
				go runtime.EventsEmit(o.ctx, "warnMsg", "❗获取股票价格失败,分析结果可能不准确")
				return
			}
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
			messages := GetFinancialReports(stockCode, o.CrawlTimeOut)
			if messages == nil || len(*messages) == 0 {
				logger.SugaredLogger.Error("获取股票财报失败")
				ch <- "***❗获取股票财报失败,分析结果可能不准确***<hr>"
				go runtime.EventsEmit(o.ctx, "warnMsg", "❗获取股票财报失败,分析结果可能不准确")
				return
			}
			for _, message := range *messages {
				msg = append(msg, map[string]interface{}{
					"role":    "assistant",
					"content": stock + message,
				})
			}
		}()

		go func() {
			defer wg.Done()
			messages := GetTelegraphList(o.CrawlTimeOut)
			if messages == nil || len(*messages) == 0 {
				logger.SugaredLogger.Error("获取市场资讯失败")
				//ch <- "***❗获取市场资讯失败,分析结果可能不准确***<hr>"
				//go runtime.EventsEmit(o.ctx, "warnMsg", "❗获取市场资讯失败,分析结果可能不准确")
				return
			}
			for _, message := range *messages {
				msg = append(msg, map[string]interface{}{
					"role":    "assistant",
					"content": message,
				})
			}
			messages = GetTopNewsList(o.CrawlTimeOut)
			if messages == nil || len(*messages) == 0 {
				logger.SugaredLogger.Error("获取新闻资讯失败")
				//ch <- "***❗获取新闻资讯失败,分析结果可能不准确***<hr>"
				//go runtime.EventsEmit(o.ctx, "warnMsg", "❗获取新闻资讯失败,分析结果可能不准确")
				return
			}
			for _, message := range *messages {
				msg = append(msg, map[string]interface{}{
					"role":    "assistant",
					"content": message,
				})
			}
		}()

		//go func() {
		//	defer wg.Done()
		//	messages := SearchStockInfo(stock, "depth", o.CrawlTimeOut)
		//	if messages == nil || len(*messages) == 0 {
		//		logger.SugaredLogger.Error("获取股票资讯失败")
		//		//ch <- "***❗获取股票资讯失败,分析结果可能不准确***<hr>"
		//		//go runtime.EventsEmit(o.ctx, "warnMsg", "❗获取股票资讯失败,分析结果可能不准确")
		//		return
		//	}
		//	for _, message := range *messages {
		//		msg = append(msg, map[string]interface{}{
		//			"role":    "assistant",
		//			"content": message,
		//		})
		//	}
		//}()
		go func() {
			defer wg.Done()
			messages := SearchStockInfo(stock, "telegram", o.CrawlTimeOut)
			if messages == nil || len(*messages) == 0 {
				logger.SugaredLogger.Error("获取股票电报资讯失败")
				//ch <- "***❗获取股票电报资讯失败,分析结果可能不准确***<hr>"
				//go runtime.EventsEmit(o.ctx, "warnMsg", "❗获取股票电报资讯失败,分析结果可能不准确")
				return
			}
			for _, message := range *messages {
				msg = append(msg, map[string]interface{}{
					"role":    "assistant",
					"content": message,
				})
			}
		}()

		go func() {
			defer wg.Done()
			messages := SearchGuShiTongStockInfo(stockCode, o.CrawlTimeOut)
			if messages == nil || len(*messages) == 0 {
				logger.SugaredLogger.Error("获取股势通资讯失败")
				//ch <- "***❗获取股势通资讯失败,分析结果可能不准确***<hr>"
				//go runtime.EventsEmit(o.ctx, "warnMsg", "❗获取股势通资讯失败,分析结果可能不准确")
				return
			}
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
			"content": question,
		})
		client := resty.New()
		client.SetBaseURL(o.BaseUrl)
		client.SetHeader("Authorization", "Bearer "+o.ApiKey)
		client.SetHeader("Content-Type", "application/json")
		//client.SetRetryCount(3)
		if o.TimeOut <= 0 {
			o.TimeOut = 300
		}
		client.SetTimeout(time.Duration(o.TimeOut) * time.Second)
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

		body := resp.RawBody()
		defer body.Close()
		if err != nil {
			logger.SugaredLogger.Infof("Stream error : %s", err.Error())
			ch <- err.Error()
			return
		}

		scanner := bufio.NewScanner(body)
		for scanner.Scan() {
			line := scanner.Text()
			logger.SugaredLogger.Infof("Received data: %s", line)
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					return
				}

				var streamResponse struct {
					Choices []struct {
						Delta struct {
							Content          string `json:"content"`
							ReasoningContent string `json:"reasoning_content"`
						} `json:"delta"`
						FinishReason string `json:"finish_reason"`
					} `json:"choices"`
				}

				if err := json.Unmarshal([]byte(data), &streamResponse); err == nil {
					for _, choice := range streamResponse.Choices {
						if content := choice.Delta.Content; content != "" {
							ch <- content
							logger.SugaredLogger.Infof("Content data: %s", content)
						}
						if reasoningContent := choice.Delta.ReasoningContent; reasoningContent != "" {
							ch <- reasoningContent
							logger.SugaredLogger.Infof("ReasoningContent data: %s", reasoningContent)
						}
						if choice.FinishReason == "stop" {
							return
						}
					}
				} else {
					if err != nil {
						logger.SugaredLogger.Infof("Stream data error : %s", err.Error())
						ch <- err.Error()
					} else {
						logger.SugaredLogger.Infof("Stream data error : %s", data)
						ch <- data
					}
				}
			} else {
				ch <- line
			}

		}
	}()
	return ch
}

func SearchGuShiTongStockInfo(stock string, crawlTimeOut int64) *[]string {
	crawlerAPI := CrawlerApi{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(crawlTimeOut)*time.Second)
	defer cancel()

	crawlerAPI = crawlerAPI.NewCrawler(ctx, CrawlerBaseInfo{
		Name:    "百度股市通",
		BaseUrl: "https://gushitong.baidu.com",
		Headers: map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/133.0.0.0"},
	})
	url := "https://gushitong.baidu.com/stock/ab-" + RemoveAllNonDigitChar(stock)
	logger.SugaredLogger.Infof("SearchGuShiTongStockInfo搜索股票-%s: %s", stock, url)
	actions := []chromedp.Action{
		chromedp.Navigate(url),
		chromedp.WaitVisible("div.cos-tab"),
		chromedp.Click("div.cos-tab:nth-child(5)", chromedp.ByQuery),
		chromedp.ScrollIntoView("div.body-box"),
		chromedp.WaitVisible("div.body-col"),
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil),
		chromedp.Sleep(1 * time.Second),
	}
	htmlContent, success := crawlerAPI.GetHtmlWithActions(&actions, true)
	var messages []string
	if success {
		document, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
		if err != nil {
			logger.SugaredLogger.Error(err.Error())
			return &[]string{}
		}
		document.Find("div.finance-hover,div.list-date").Each(func(i int, selection *goquery.Selection) {
			text := strutil.RemoveWhiteSpace(selection.Text(), false)
			messages = append(messages, text)
			logger.SugaredLogger.Infof("SearchGuShiTongStockInfo搜索到消息-%s: %s", "", text)
		})
		logger.SugaredLogger.Infof("messages:%d", len(messages))
	}
	return &messages
}

func GetFinancialReports(stockCode string, crawlTimeOut int64) *[]string {
	// 创建一个 chromedp 上下文
	timeoutCtx, timeoutCtxCancel := context.WithTimeout(context.Background(), time.Duration(crawlTimeOut)*time.Second)
	defer timeoutCtxCancel()
	var ctx context.Context
	var cancel context.CancelFunc
	path, e := checkBrowserOnWindows()
	logger.SugaredLogger.Infof("GetFinancialReports path:%s", path)

	if e {
		pctx, pcancel := chromedp.NewExecAllocator(
			timeoutCtx,
			chromedp.ExecPath(path),
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-javascript", false),
			chromedp.Flag("disable-gpu", true),
			chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/133.0.0.0"),
			chromedp.Flag("disable-background-networking", true),
			chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
			chromedp.Flag("disable-background-timer-throttling", true),
			chromedp.Flag("disable-backgrounding-occluded-windows", true),
			chromedp.Flag("disable-breakpad", true),
			chromedp.Flag("disable-client-side-phishing-detection", true),
			chromedp.Flag("disable-default-apps", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("disable-extensions", true),
			chromedp.Flag("disable-features", "site-per-process,Translate,BlinkGenPropertyTrees"),
			chromedp.Flag("disable-hang-monitor", true),
			chromedp.Flag("disable-ipc-flooding-protection", true),
			chromedp.Flag("disable-popup-blocking", true),
			chromedp.Flag("disable-prompt-on-repost", true),
			chromedp.Flag("disable-renderer-backgrounding", true),
			chromedp.Flag("disable-sync", true),
			chromedp.Flag("force-color-profile", "srgb"),
			chromedp.Flag("metrics-recording-only", true),
			chromedp.Flag("safebrowsing-disable-auto-update", true),
			chromedp.Flag("enable-automation", true),
			chromedp.Flag("password-store", "basic"),
			chromedp.Flag("use-mock-keychain", true),
		)
		defer pcancel()
		ctx, cancel = chromedp.NewContext(
			pctx,
			chromedp.WithLogf(logger.SugaredLogger.Infof),
			chromedp.WithErrorf(logger.SugaredLogger.Errorf),
		)
	} else {
		ctx, cancel = chromedp.NewContext(
			timeoutCtx,
			chromedp.WithLogf(logger.SugaredLogger.Infof),
			chromedp.WithErrorf(logger.SugaredLogger.Errorf),
		)
	}
	defer cancel()
	var htmlContent string
	url := fmt.Sprintf("https://xueqiu.com/snowman/S/%s/detail#/ZYCWZB", stockCode)
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		// 等待页面加载完成，可以根据需要调整等待时间
		chromedp.WaitVisible("table.table", chromedp.ByQuery),
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
	}
	document, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
		return &[]string{}
	}
	var messages []string
	document.Find("table tr").Each(func(i int, selection *goquery.Selection) {
		tr := ""
		selection.Find("th,td").Each(func(i int, selection *goquery.Selection) {
			ret := selection.Find("p").First().Text()
			if ret == "" {
				ret = selection.Text()
			}
			text := strutil.RemoveNonPrintable(ret)
			tr += text + " "
		})
		logger.SugaredLogger.Infof("%s", tr+" \n")
		messages = append(messages, tr+" \n")
	})
	return &messages
}

func (o OpenAi) NewCommonChatStream(stock, stockCode, apiURL, apiKey, Model string) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		client := resty.New()
		client.SetHeader("Authorization", "Bearer "+apiKey)
		client.SetHeader("Content-Type", "application/json")
		client.SetRetryCount(3)

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
			messages := SearchStockPriceInfo(stockCode, o.CrawlTimeOut)
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

func GetTelegraphList(crawlTimeOut int64) *[]string {
	url := "https://www.cls.cn/telegraph"
	response, err := resty.New().SetTimeout(time.Duration(crawlTimeOut)*time.Second).R().
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
		logger.SugaredLogger.Info(selection.Text())
		telegraph = append(telegraph, selection.Text())
	})
	return &telegraph
}

func GetTopNewsList(crawlTimeOut int64) *[]string {
	url := "https://www.cls.cn"
	response, err := resty.New().SetTimeout(time.Duration(crawlTimeOut)*time.Second).R().
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
	document.Find("div.home-article-title a,div.home-article-rec a").Each(func(i int, selection *goquery.Selection) {
		logger.SugaredLogger.Info(selection.Text())
		telegraph = append(telegraph, selection.Text())
	})
	return &telegraph
}

func (o OpenAi) SaveAIResponseResult(stockCode, stockName, result string) {
	db.Dao.Create(&models.AIResponseResult{
		StockCode: stockCode,
		StockName: stockName,
		ModelName: o.Model,
		Content:   result,
	})
}

func (o OpenAi) GetAIResponseResult(stock string) *models.AIResponseResult {
	var result models.AIResponseResult
	db.Dao.Where("stock_code = ?", stock).Order("id desc").Limit(1).First(&result)
	return &result
}
