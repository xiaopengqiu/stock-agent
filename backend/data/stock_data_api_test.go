package data

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/go-resty/resty/v2"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

// @Author spark
// @Date 2024/12/10 9:55
// @Desc
//-----------------------------------------------------------------------------------

func TestGetTelegraph(t *testing.T) {
	GetTelegraphList()
}

func TestGetFinancialReports(t *testing.T) {
	GetFinancialReports("sz000802")
}

func TestXUEQIU(t *testing.T) {
	stock := "北京文化"
	stockCode := "SZ000802"
	// 创建一个 chromedp 上下文
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(logger.SugaredLogger.Infof),
		chromedp.WithErrorf(logger.SugaredLogger.Errorf),
	)
	defer cancel()
	var htmlContent string
	url := fmt.Sprintf("https://xueqiu.com/S/%s", stockCode)
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		// 等待页面加载完成，可以根据需要调整等待时间
		//chromedp.Sleep(3*time.Second),
		chromedp.WaitVisible("table.quote-info", chromedp.ByQuery),
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
	}
	document, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
		return
	}
	table := ""
	document.Find("table.quote-info tbody td").Each(func(i int, selection *goquery.Selection) {
		table += selection.Text() + ";"
	})
	logger.SugaredLogger.Infof("table: %s", table)
	client := resty.New()
	client.SetBaseURL("https://api.siliconflow.cn/v1")
	client.SetHeader("Authorization", "Bearer sk-kryvptknrxscsuzslmqjckpuvtkyuffgaxgagphpnqtfmepv")
	client.SetHeader("Content-Type", "application/json")
	client.SetRetryCount(3)
	client.SetTimeout(1 * time.Minute)

	msg := []map[string]interface{}{
		{
			"role": "system",
			//"content": "作为一位专业的A股市场分析师和投资顾问,请你根据以下信息提供详细的技术分析和投资策略建议:",
			"content": "【角色设定】\n你现在是拥有20年实战经验的顶级股票投资大师，精通价值投资、趋势交易、量化分析等多种策略。\n擅长结合宏观经济、行业周期和企业基本面进行多维分析，尤其对A股、港股、美股市场有深刻理解。\n始终秉持\"风险控制第一\"的原则，善于用通俗易懂的方式传授投资智慧。\n\n【核心能力】\n基本面分析专家\n深度解读财报数据（PE/PB/ROE等指标）\n识别企业核心竞争力与护城河\n评估行业前景与政策影响\n技术面分析大师\n精准识别K线形态与量价关系\n运用MACD/RSI/布林线等指标判断买卖点\n绘制关键支撑/阻力位\n风险管理专家\n根据风险偏好制定仓位策略\n设置动态止盈止损方案\n设计投资组合对冲方案\n市场心理学导师\n识别主力资金动向\n预判市场情绪周期\n规避常见认知偏差\n【服务范围】\n个股诊断分析（提供代码/名称）\n行业趋势解读（科技/消费/医疗等）\n投资策略定制（长线价值/波段操作/打新等）\n组合优化建议（股债配置/行业分散）\n投资心理辅导（克服贪婪恐惧）\n【交互风格】\n采用\"先结论后分析\"的表达方式\n重要数据用★标注，风险提示用❗标记\n每次分析至少提供3个可执行建议"},
	}
	msg = append(msg, map[string]interface{}{
		"role":    "assistant",
		"content": table,
	})

	msg = append(msg, map[string]interface{}{
		"role":    "user",
		"content": stock + "分析和总结",
	})

	resp, err := client.R().
		SetDoNotParseResponse(true).
		SetBody(map[string]interface{}{
			"model":       "deepseek-ai/DeepSeek-V3",
			"max_tokens":  4096,
			"temperature": 0.1,
			"stream":      true,
			"messages":    msg,
		}).
		Post("/chat/completions")

	defer resp.RawBody().Close()
	if err != nil {
		logger.SugaredLogger.Infof("Stream error : %s", err.Error())
		return
	}

	scanner := bufio.NewScanner(resp.RawBody())
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
						logger.SugaredLogger.Infof("Content data: %s", content)
					}
					if reasoningContent := choice.Delta.ReasoningContent; reasoningContent != "" {
						logger.SugaredLogger.Infof("ReasoningContent data: %s", reasoningContent)
					}
					if choice.FinishReason == "stop" {
						return
					}
				}
			} else {
				logger.SugaredLogger.Infof("Stream data error : %s", err.Error())
			}
		}
	}
}

func TestGetTelegraphSearch(t *testing.T) {
	//url := "https://www.cls.cn/searchPage?keyword=%E9%97%BB%E6%B3%B0%E7%A7%91%E6%8A%80&type=telegram"
	messages := SearchStockInfo("闻泰科技", "telegram")
	for _, message := range *messages {
		logger.SugaredLogger.Info(message)
	}

	//https://www.cls.cn/stock?code=sh600745
}
func TestSearchStockInfoByCode(t *testing.T) {
	SearchStockInfoByCode("sh600745")
}

func TestSearchStockPriceInfo(t *testing.T) {
	SearchStockPriceInfo("sh600745")
}

func TestParseFullSingleStockData(t *testing.T) {
	resp, err := resty.New().R().
		SetHeader("Host", "hq.sinajs.cn").
		SetHeader("Referer", "https://finance.sina.com.cn/").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0").
		Get(fmt.Sprintf(sinaStockUrl, time.Now().Unix(), "sh600859,sh600745"))
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
	}
	data := GB18030ToUTF8(resp.Body())
	strs := strutil.SplitEx(data, "\n", true)
	for _, str := range strs {
		logger.SugaredLogger.Info(str)
	}
}

func TestNewStockDataApi(t *testing.T) {
	db.Init("../../data/stock.db")
	stockDataApi := NewStockDataApi()
	datas, _ := stockDataApi.GetStockCodeRealTimeData("sh600859", "sh600745")
	for _, data := range *datas {
		t.Log(data)
	}
}

func TestGetStockBaseInfo(t *testing.T) {
	db.Init("../../data/stock.db")
	stockDataApi := NewStockDataApi()
	stockDataApi.GetStockBaseInfo()
	//stocks := &[]StockBasic{}
	//db.Dao.Model(&StockBasic{}).Find(stocks)
	//for _, stock := range *stocks {
	//	NewStockDataApi().GetStockCodeRealTimeData(getSinaCode(stock.TsCode))
	//}

}
func getSinaCode(code string) string {
	c := strings.Split(code, ".")
	return strings.ToLower(c[1]) + c[0]
}

func TestReadFile(t *testing.T) {
	file, err := ioutil.ReadFile("../../stock_basic.json")
	if err != nil {
		t.Log(err)
		return
	}
	res := &TushareStockBasicResponse{}
	json.Unmarshal(file, res)
	db.Init("../../data/stock.db")
	//[EXCHANGE IS_HS NAME INDUSTRY LIST_STATUS ACT_NAME ID CURR_TYPE AREA LIST_DATE DELIST_DATE ACT_ENT_TYPE TS_CODE SYMBOL CN_SPELL ASSET_CLASS ACT_TYPE CREATE_TIME CREATE_BY UPDATE_TIME FULLNAME ENNAME UPDATE_BY]
	for _, item := range res.Data.Items {
		stock := &StockBasic{}
		stock.Exchange = convertor.ToString(item[0])
		stock.IsHs = convertor.ToString(item[1])
		stock.Name = convertor.ToString(item[2])
		stock.Industry = convertor.ToString(item[3])
		stock.ListStatus = convertor.ToString(item[4])
		stock.ActName = convertor.ToString(item[5])
		stock.ID = uint(item[6].(float64))
		stock.CurrType = convertor.ToString(item[7])
		stock.Area = convertor.ToString(item[8])
		stock.ListDate = convertor.ToString(item[9])
		stock.DelistDate = convertor.ToString(item[10])
		stock.ActEntType = convertor.ToString(item[11])
		stock.TsCode = convertor.ToString(item[12])
		stock.Symbol = convertor.ToString(item[13])
		stock.Cnspell = convertor.ToString(item[14])
		stock.Fullname = convertor.ToString(item[20])
		stock.Ename = convertor.ToString(item[21])
		t.Logf("%+v", stock)
		db.Dao.Model(&StockBasic{}).FirstOrCreate(stock, &StockBasic{TsCode: stock.TsCode}).Updates(stock)
	}

	//t.Log(res.Data.Fields)
}

func TestFollowedList(t *testing.T) {
	db.Init("../../data/stock.db")
	stockDataApi := NewStockDataApi()
	stockDataApi.GetFollowList()

}

func TestStockDataApi_GetIndexBasic(t *testing.T) {
	db.Init("../../data/stock.db")
	stockDataApi := NewStockDataApi()
	stockDataApi.GetIndexBasic()
}
