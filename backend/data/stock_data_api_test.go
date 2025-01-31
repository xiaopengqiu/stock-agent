package data

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
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
	url := "https://www.cls.cn/telegraph"
	response, err := resty.New().R().
		SetHeader("Referer", "https://www.cls.cn/").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36 Edg/117.0.2045.60").
		Get(fmt.Sprintf(url))
	if err != nil {
		return
	}
	logger.SugaredLogger.Info(string(response.Body()))
	document, err := goquery.NewDocumentFromReader(strings.NewReader(string(response.Body())))
	if err != nil {
		return
	}
	document.Find("div.telegraph-content-box").Each(func(i int, selection *goquery.Selection) {
		text := selection.Text()
		if strings.Contains(text, "【公告】") {
			logger.SugaredLogger.Info(text)
		}
	})
}

func TestGetTelegraphSearch(t *testing.T) {
	//url := "https://www.cls.cn/searchPage?keyword=%E9%97%BB%E6%B3%B0%E7%A7%91%E6%8A%80&type=telegram"
	messages := SearchStockInfo("闻泰科技", "telegram")
	for _, message := range *messages {
		logger.SugaredLogger.Info(message)
	}

	//https://www.cls.cn/stock?code=sh600745
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
