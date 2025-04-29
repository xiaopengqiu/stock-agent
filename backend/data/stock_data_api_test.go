package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/random"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/go-resty/resty/v2"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"
	"time"
)

// @Author spark
// @Date 2024/12/10 9:55
// @Desc
//-----------------------------------------------------------------------------------

func TestGetTelegraph(t *testing.T) {
	db.Init("../../data/stock.db")

	//telegraphs := GetTelegraphList(30)
	//for _, telegraph := range *telegraphs {
	//	logger.SugaredLogger.Info(telegraph)
	//}
	list := NewMarketNewsApi().GetNewTelegraph(30)
	for _, telegraph := range *list {
		logger.SugaredLogger.Infof("telegraph:%+v", telegraph)
	}
}

func TestGetFinancialReports(t *testing.T) {
	db.Init("../../data/stock.db")
	//GetFinancialReports("sz000802", 30)
	//GetFinancialReports("hk00927", 30)
	//GetFinancialReports("gb_aapl", 30)
	GetFinancialReportsByXUEQIU("sz000802", 30)
	GetFinancialReportsByXUEQIU("gb_aapl", 30)
	GetFinancialReportsByXUEQIU("hk00927", 30)

}

func TestGetTelegraphSearch(t *testing.T) {
	//url := "https://www.cls.cn/searchPage?keyword=%E9%97%BB%E6%B3%B0%E7%A7%91%E6%8A%80&type=telegram"
	messages := SearchStockInfo("谷歌", "telegram", 30)
	for _, message := range *messages {
		logger.SugaredLogger.Info(message)
	}

	//https://www.cls.cn/stock?code=sh600745
}
func TestSearchStockInfoByCode(t *testing.T) {
	SearchStockInfoByCode("sh600745")
}

func TestSearchStockPriceInfo(t *testing.T) {
	db.Init("../../data/stock.db")
	//SearchStockPriceInfo("中信证券", "hk06030", 30)
	//SearchStockPriceInfo("上海贝岭", "sh600171", 30)
	SearchStockPriceInfo("苹果公司", "gb_aapl", 30)
	//SearchStockPriceInfo("微创光电", "bj430198", 30)
	getZSInfo("创业板指数", "sz399006", 30)
	//getZSInfo("上证综合指数", "sh000001", 30)
	//getZSInfo("沪深300指数", "sh000300", 30)

}

func TestGetKLineData(t *testing.T) {
	db.Init("../../data/stock.db")
	k := NewStockDataApi().GetKLineData("sh600171", "240", 30)
	//for _, kline := range *k {
	//	logger.SugaredLogger.Infof("%+#v", kline)
	//}
	jsonData, _ := json.Marshal(*k)
	markdownTable, err := JSONToMarkdownTable(jsonData)
	if err != nil {
		logger.SugaredLogger.Errorf("json.Marshal error:%s", err.Error())
	}
	logger.SugaredLogger.Infof("markdownTable:\n%s", markdownTable)

}
func TestGetHK_KLineData(t *testing.T) {
	db.Init("../../data/stock.db")
	k := NewStockDataApi().GetHK_KLineData("hk01810", "day", 1)
	jsonData, _ := json.Marshal(*k)
	markdownTable, err := JSONToMarkdownTable(jsonData)
	if err != nil {
		logger.SugaredLogger.Errorf("json.Marshal error:%s", err.Error())
	}
	logger.SugaredLogger.Infof("markdownTable:\n%s", markdownTable)

}

func TestGetHKStockInfo(t *testing.T) {
	db.Init("../../data/stock.db")
	//NewStockDataApi().GetHKStockInfo(200)
	//NewStockDataApi().GetSinaHKStockInfo()
	//m:105,m:106,m:107  //美股
	//m:128+t:3,m:128+t:4,m:128+t:1,m:128+t:2 //港股
	for i := 1; i <= 592; i++ {
		NewStockDataApi().getDCStockInfo("us", i, 20)
		time.Sleep(time.Duration(random.RandInt(1, 3)) * time.Second)
	}
}

func TestParseTxStockData(t *testing.T) {
	str := "v_r_hk09660=\"100~地平线机器人-W~09660~6.340~5.690~5.800~210980204.0~0~0~6.340~0~0~0~0~0~0~0~0~0~6.340~0~0~0~0~0~0~0~0~0~210980204.0~2025/04/29\n14:14:52~0.650~11.42~6.450~5.710~6.340~210980204.0~1295585259.040~0~33.03~~0~0~13.01~702.2123~836.8986~HORIZONROBOT-W~0.00~10.380~3.320~1.00~-53.74~0~0~0~0~0~33.03~6.50~1.90~600~76.11~19.85~GP~19.70~11.51~0.63~-17.23~46.76~13200293682.00~11075904412.00~33.03~0.000~6.141~58.90~HKD~1~30\";"
	info, _ := ParseTxStockData(str)
	logger.SugaredLogger.Infof("%+#v", info)
}

func TestGetRealTimeStockPriceInfo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	text, texttime := GetRealTimeStockPriceInfo(ctx, "sh600171")
	logger.SugaredLogger.Infof("res:%s,%s", text, texttime)

	text, texttime = GetRealTimeStockPriceInfo(ctx, "sh600438")
	logger.SugaredLogger.Infof("res:%s,%s", text, texttime)

	texttime = strings.ReplaceAll(texttime, "）", "")
	texttime = strings.ReplaceAll(texttime, "（", "")
	parts := strings.Split(texttime, " ")
	logger.SugaredLogger.Infof("parts:%+v", parts)

	//去除中文字符
	// 正则表达式匹配中文字符
	re := regexp.MustCompile(`\p{Han}+`)
	texttime = re.ReplaceAllString(texttime, "")

	logger.SugaredLogger.Infof("texttime:%s", texttime)
	location, err := time.ParseInLocation("2006-01-02 15:04:05", texttime, time.Local)
	if err != nil {
		return
	}
	logger.SugaredLogger.Infof("location:%s", location.Format("2006-01-02 15:04:05"))
}

func TestParseFullSingleStockData(t *testing.T) {
	resp, err := resty.New().R().
		SetHeader("Host", "hq.sinajs.cn").
		SetHeader("Referer", "https://finance.sina.com.cn/").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0").
		Get(fmt.Sprintf(sinaStockUrl, time.Now().Unix(), "sh600584,sz000938,hk01810,hk00856,gb_aapl,gb_tsla,sb873721,bj430300"))
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
	}
	data := GB18030ToUTF8(resp.Body())
	strs := strutil.SplitEx(data, "\n", true)
	for _, str := range strs {
		logger.SugaredLogger.Info(str)
		stockData, err := ParseFullSingleStockData(str)
		if err != nil {
			return
		}
		logger.SugaredLogger.Infof("%+#v", stockData)
	}

	result, er := ParseFullSingleStockData("var hq_str_gb_tsla = \"特斯拉,268.8472,-5.55,2025-03-04 22:52:56,-15.8028,270.9300,278.2800,268.1000,488.5400,138.8030,23618295,88214389,864751599149,2.23,120.550000,0.00,0.00,0.00,0.00,3216517037,61,0.0000,0.00,0.00,,Mar 04 09:52AM EST,284.6500,0,1,2025,6458502467.0000,0.0000,0.0000,0.0000,0.0000,284.6500\";")
	if er != nil {
		logger.SugaredLogger.Error(er.Error())
	}
	logger.SugaredLogger.Infof("%+#v", result)
}

func TestNewStockDataApi(t *testing.T) {
	db.Init("../../data/stock.db")
	stockDataApi := NewStockDataApi()
	datas, _ := stockDataApi.GetStockCodeRealTimeData("sh600859", "sh600745", "gb_tsla", "hk09660", "hk00700")
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
	stockDataApi.GetFollowList(1)

}

func TestStockDataApi_GetIndexBasic(t *testing.T) {
	db.Init("../../data/stock.db")
	stockDataApi := NewStockDataApi()
	stockDataApi.GetIndexBasic()
}
