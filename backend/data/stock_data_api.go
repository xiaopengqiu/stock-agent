package data

// @Author spark
// @Date 2024/12/10 9:21
// @Desc
//-----------------------------------------------------------------------------------
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/go-resty/resty/v2"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"io/ioutil"
	"strings"
	"time"
)

// http://hq.sinajs.cn/rn=1730966120830&list=sh600000,sh600859
const sina_stook_url = "http://hq.sinajs.cn/rn=%d&list=%s"
const tushare_api_url = "http://api.tushare.pro"
const TushareToken = "9125ec636217a99a3218a64fc63507e95205f2666590792923cbaedf"

type StockDataApi struct {
	client *resty.Client
}
type StockInfo struct {
	gorm.Model
	Date     string  `json:"日期" gorm:"index"`
	Time     string  `json:"时间" gorm:"index"`
	Code     string  `json:"股票代码" gorm:"index"`
	Name     string  `json:"股票名称" gorm:"index"`
	PrePrice float64 `json:"上次当前价格"`
	Price    string  `json:"当前价格"`
	Volume   string  `json:"成交的股票数"`
	Amount   string  `json:"成交金额"`
	Open     string  `json:"今日开盘价"`
	PreClose string  `json:"昨日收盘价"`
	High     string  `json:"今日最高价"`
	Low      string  `json:"今日最低价"`
	Bid      string  `json:"竞买价"`
	Ask      string  `json:"竞卖价"`
	B1P      string  `json:"买一报价"`
	B1V      string  `json:"买一申报"`
	B2P      string  `json:"买二报价"`
	B2V      string  `json:"买二申报"`
	B3P      string  `json:"买三报价"`
	B3V      string  `json:"买三申报"`
	B4P      string  `json:"买四报价"`
	B4V      string  `json:"买四申报"`
	B5P      string  `json:"买五报价"`
	B5V      string  `json:"买五申报"`
	A1P      string  `json:"卖一报价"`
	A1V      string  `json:"卖一申报"`
	A2P      string  `json:"卖二报价"`
	A2V      string  `json:"卖二申报"`
	A3P      string  `json:"卖三报价"`
	A3V      string  `json:"卖三申报"`
	A4P      string  `json:"卖四报价"`
	A4V      string  `json:"卖四申报"`
	A5P      string  `json:"卖五报价"`
	A5V      string  `json:"卖五申报"`

	//以下是字段值需二次计算
	ChangePercent     float64 `json:"changePercent"`     //涨跌幅
	ChangePrice       float64 `json:"changePrice"`       //涨跌额
	HighRate          float64 `json:"highRate"`          //最高涨跌
	LowRate           float64 `json:"lowRate"`           //最低涨跌
	CostPrice         float64 `json:"costPrice"`         //成本价
	CostVolume        int64   `json:"costVolume"`        //持仓数量
	Profit            float64 `json:"profit"`            //总盈亏率
	ProfitAmount      float64 `json:"profitAmount"`      //总盈亏金额
	ProfitAmountToday float64 `json:"profitAmountToday"` //今日盈亏金额

	Sort               int64   `json:"sort"` //排序
	AlarmChangePercent float64 `json:"alarmChangePercent"`
	AlarmPrice         float64 `json:"alarmPrice"`
}

func (receiver StockInfo) TableName() string {
	return "stock_info"
}

type TushareRequest struct {
	ApiName string `json:"api_name"`
	Token   string `json:"token"`
	Params  any    `json:"params"`
	Fields  string `json:"fields"`
}
type TushareResponse struct {
	RequestId string `json:"request_id"`
	Code      int    `json:"code"`
	Data      any    `json:"data"`
	Msg       string `json:"msg"`
}

/*
	字段	类型	说明
	ts_code	str	TS代码
	symbol	str	股票代码
	name	str	股票名称
	area	str	地域
	industry	str	所属行业
	fullname	str	股票全称
	enname	str	英文全称
	cnspell	str	拼音缩写
	market	str	市场类型
	exchange	str	交易所代码
	curr_type	str	交易货币
	list_status	str	上市状态 L上市 D退市 P暂停上市
	list_date	str	上市日期
	delist_date	str	退市日期
	is_hs	str	是否沪深港通标的，N否 H沪股通 S深股通
	act_name	str	实控人名称
	act_ent_type	str	实控人企业性质*/

type StockBasic struct {
	gorm.Model
	TsCode     string `json:"ts_code" gorm:"index"`
	Symbol     string `json:"symbol" gorm:"index"`
	Name       string `json:"name" gorm:"index"`
	Area       string `json:"area"`
	Industry   string `json:"industry" gorm:"index"`
	Fullname   string `json:"fullname"`
	Ename      string `json:"enname"`
	Cnspell    string `json:"cnspell"`
	Market     string `json:"market"`
	Exchange   string `json:"exchange"`
	CurrType   string `json:"curr_type"`
	ListStatus string `json:"list_status"`
	ListDate   string `json:"list_date"`
	DelistDate string `json:"delist_date"`
	IsHs       string `json:"is_hs"`
	ActName    string `json:"act_name"`
	ActEntType string `json:"act_ent_type"`
}

type FollowedStock struct {
	StockCode          string
	Name               string
	Volume             int64
	CostPrice          float64
	Price              float64
	PriceChange        float64
	ChangePercent      float64
	AlarmChangePercent float64
	AlarmPrice         float64
	Time               time.Time
	Sort               int64
	IsDel              soft_delete.DeletedAt `gorm:"softDelete:flag"`
}

func (receiver FollowedStock) TableName() string {
	return "followed_stock"
}

type TushareStockBasicResponse struct {
	TushareResponse
	Data StockBasicResponse `json:"data"`
}

type StockBasicResponse struct {
	Fields  []string `json:"fields"`
	Items   [][]any  `json:"items"`
	HasMore bool     `json:"has_more"`
	Count   int      `json:"count"`
}

func (receiver StockBasic) TableName() string {
	return "tushare_stock_basic"
}
func NewStockDataApi() *StockDataApi {
	return &StockDataApi{
		client: resty.New(),
	}
}

// GetIndexBasic 获取指数信息
func (receiver StockDataApi) GetIndexBasic() {
	res := &TushareStockBasicResponse{}
	fields := "ts_code,name,market,publisher,category,base_date,base_point,list_date,fullname,index_type,weight_rule,desc"
	_, err := receiver.client.R().
		SetHeader("content-type", "application/json").
		SetBody(&TushareRequest{
			ApiName: "index_basic",
			Token:   TushareToken,
			Params:  nil,
			Fields:  fields}).
		SetResult(res).
		Post(tushare_api_url)
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
		return
	}
	if res.Code != 0 {
		logger.SugaredLogger.Error(res.Msg)
		return
	}
	//ioutil.WriteFile("index_basic.json", resp.Body(), 0666)

	for _, item := range res.Data.Items {
		data := map[string]any{}
		for _, field := range strings.Split(fields, ",") {
			idx := slice.IndexOf(res.Data.Fields, field)
			if idx == -1 {
				continue
			}
			data[field] = item[idx]
		}
		index := &IndexBasic{}
		jsonData, _ := json.Marshal(data)
		err := json.Unmarshal(jsonData, index)
		if err != nil {
			continue
		}
		db.Dao.Model(&IndexBasic{}).FirstOrCreate(index, &IndexBasic{TsCode: index.TsCode}).Where("ts_code = ?", index.TsCode).Updates(index)
	}

}

// map转换为结构体

func (receiver StockDataApi) GetStockBaseInfo() {
	res := &TushareStockBasicResponse{}
	fields := "ts_code,symbol,name,area,industry,cnspell,market,list_date,act_name,act_ent_type,fullname,exchange,list_status,curr_type,enname,delist_date,is_hs"
	_, err := receiver.client.R().
		SetHeader("content-type", "application/json").
		SetBody(&TushareRequest{
			ApiName: "stock_basic",
			Token:   TushareToken,
			Params:  nil,
			Fields:  fields,
		}).
		SetResult(res).
		Post(tushare_api_url)
	//logger.SugaredLogger.Infof("GetStockBaseInfo %s", string(resp.Body()))
	//resp.Body()写入文件
	//ioutil.WriteFile("stock_basic.json", resp.Body(), 0666)
	//logger.SugaredLogger.Infof("GetStockBaseInfo %+v", res)
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
		return
	}
	if res.Code != 0 {
		logger.SugaredLogger.Error(res.Msg)
		return
	}
	for _, item := range res.Data.Items {
		stock := &StockBasic{}
		data := map[string]any{}
		for _, field := range strings.Split(fields, ",") {
			logger.SugaredLogger.Infof("field: %s", field)
			idx := slice.IndexOf(res.Data.Fields, field)
			if idx == -1 {
				continue
			}
			data[field] = item[idx]
		}
		jsonData, _ := json.Marshal(data)
		err := json.Unmarshal(jsonData, stock)
		if err != nil {
			continue
		}
		db.Dao.Model(&StockBasic{}).FirstOrCreate(stock, &StockBasic{TsCode: stock.TsCode}).Where("ts_code = ?", stock.TsCode).Updates(stock)
	}

}

func (receiver StockDataApi) GetStockCodeRealTimeData(StockCodes ...string) (*[]StockInfo, error) {
	resp, err := receiver.client.R().
		SetHeader("Host", "hq.sinajs.cn").
		SetHeader("Referer", "https://finance.sina.com.cn/").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0").
		Get(fmt.Sprintf(sina_stook_url, time.Now().Unix(), slice.Join(StockCodes, ",")))
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
		return &[]StockInfo{}, err
	}

	stockInfos := make([]StockInfo, 0)
	str := GB18030ToUTF8(resp.Body())
	dataStr := strutil.SplitEx(str, "\n", true)
	for _, data := range dataStr {
		//logger.SugaredLogger.Info(data)
		stockData, err := ParseFullSingleStockData(data)
		if err != nil {
			logger.SugaredLogger.Error(err.Error())
			continue
		}
		stockInfos = append(stockInfos, *stockData)

		go func() {
			var count int64
			db.Dao.Model(&StockInfo{}).Where("code = ?", stockData.Code).Count(&count)
			if count == 0 {
				db.Dao.Model(&StockInfo{}).Create(stockData)
			} else {
				db.Dao.Model(&StockInfo{}).Where("code = ?", stockData.Code).Updates(stockData)
			}
		}()

	}

	return &stockInfos, err
}

func (receiver StockDataApi) Follow(stockCode string) string {
	logger.SugaredLogger.Infof("Follow %s", stockCode)
	stockInfos, err := receiver.GetStockCodeRealTimeData(stockCode)
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
		return "关注失败"
	}
	stockInfo := (*stockInfos)[0]
	price, _ := convertor.ToFloat(stockInfo.Price)
	db.Dao.Model(&FollowedStock{}).FirstOrCreate(&FollowedStock{
		StockCode:          stockCode,
		Name:               stockInfo.Name,
		Price:              price,
		Time:               time.Now(),
		ChangePercent:      0,
		PriceChange:        0,
		Sort:               0,
		AlarmChangePercent: 3,
		AlarmPrice:         price + 1,
	}, &FollowedStock{StockCode: stockCode})
	return "关注成功"
}

func (receiver StockDataApi) UnFollow(stockCode string) string {
	db.Dao.Model(&FollowedStock{}).Where("stock_code = ?", stockCode).Delete(&FollowedStock{})
	return "取消关注成功"
}

func (receiver StockDataApi) SetCostPriceAndVolume(price float64, volume int64, stockCode string) string {
	err := db.Dao.Model(&FollowedStock{}).Where("stock_code = ?", stockCode).Update("cost_price", price).Update("volume", volume).Error
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
		return "设置失败"
	}
	return "设置成功"
}

func (receiver StockDataApi) SetAlarmChangePercent(val, alarmPrice float64, stockCode string) string {
	err := db.Dao.Model(&FollowedStock{}).Where("stock_code = ?", stockCode).Updates(&map[string]any{
		"alarm_change_percent": val,
		"alarm_price":          alarmPrice,
	}).Error
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
		return "设置失败"
	}
	return "设置成功"
}

func (receiver StockDataApi) SetStockSort(sort int64, stockCode string) {
	db.Dao.Model(&FollowedStock{}).Where("stock_code = ?", stockCode).Update("sort", sort)
}

func (receiver StockDataApi) GetFollowList() []FollowedStock {
	var result []FollowedStock
	db.Dao.Model(&FollowedStock{}).Order("sort asc,time desc").Find(&result)
	return result
}

func (receiver StockDataApi) GetStockList(key string) []StockBasic {
	var result []StockBasic
	db.Dao.Model(&StockBasic{}).Where("name like ? or ts_code like ?", "%"+key+"%", "%"+key+"%").Find(&result)
	var result2 []IndexBasic
	db.Dao.Model(&IndexBasic{}).Where("market in ?", []string{"SSE", "SZSE"}).Where("name like ? or ts_code like ?", "%"+key+"%", "%"+key+"%").Find(&result2)

	for _, item := range result2 {
		result = append(result, StockBasic{
			TsCode:   item.TsCode,
			Name:     item.Name,
			Fullname: item.FullName,
			Symbol:   item.Symbol,
			Market:   item.Market,
			ListDate: item.ListDate,
		})
	}

	return result
}

// GB18030 转换为 UTF8
func GB18030ToUTF8(bs []byte) string {
	reader := transform.NewReader(bytes.NewReader(bs), simplifiedchinese.GB18030.NewDecoder())
	d, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	return string(d)
}

func ParseFullSingleStockData(data string) (*StockInfo, error) {
	datas := strutil.SplitAndTrim(data, "=", "\"")
	if len(datas) < 2 {
		return nil, fmt.Errorf("invalid data format")
	}
	code := strings.Split(datas[0], "hq_str_")[1]
	result := make(map[string]string)
	parts := strutil.SplitAndTrim(datas[1], ",", "\"")
	//parts := strings.Split(data, ",")
	if len(parts) < 32 {
		return nil, fmt.Errorf("invalid data format")
	}
	/*
		0：”大秦铁路”，股票名字；
		1：”27.55″，今日开盘价；
		2：”27.25″，昨日收盘价；
		3：”26.91″，当前价格；
		4：”27.55″，今日最高价；
		5：”26.20″，今日最低价；
		6：”26.91″，竞买价，即“买一”报价；
		7：”26.92″，竞卖价，即“卖一”报价；
		8：”22114263″，成交的股票数，由于股票交易以一百股为基本单位，所以在使用时，通常把该值除以一百；
		9：”589824680″，成交金额，单位为“元”，为了一目了然，通常以“万元”为成交金额的单位，所以通常把该值除以一万；
		10：”4695″，“买一”申报4695股，即47手；
		11：”26.91″，“买一”报价；
		12：”57590″，“买二”
		13：”26.90″，“买二”
		14：”14700″，“买三”
		15：”26.89″，“买三”
		16：”14300″，“买四”
		17：”26.88″，“买四”
		18：”15100″，“买五”
		19：”26.87″，“买五”
		20：”3100″，“卖一”申报3100股，即31手；
		21：”26.92″，“卖一”报价
		(22, 23), (24, 25), (26,27), (28, 29)分别为“卖二”至“卖四的情况”
		30：”2008-01-11″，日期；
		31：”15:05:32″，时间；*/
	result["股票代码"] = code
	result["股票名称"] = parts[0]
	result["今日开盘价"] = parts[1]
	result["昨日收盘价"] = parts[2]
	result["当前价格"] = parts[3]
	result["今日最高价"] = parts[4]
	result["今日最低价"] = parts[5]
	result["竞买价"] = parts[6]
	result["竞卖价"] = parts[7]
	result["成交的股票数"] = parts[8]
	result["成交金额"] = parts[9]
	result["买一申报"] = parts[10]
	result["买一报价"] = parts[11]
	result["买二申报"] = parts[12]
	result["买二报价"] = parts[13]
	result["买三申报"] = parts[14]
	result["买三报价"] = parts[15]
	result["买四申报"] = parts[16]
	result["买四报价"] = parts[17]
	result["买五申报"] = parts[18]
	result["买五报价"] = parts[19]
	result["卖一申报"] = parts[20]
	result["卖一报价"] = parts[21]
	result["卖二申报"] = parts[22]
	result["卖二报价"] = parts[23]
	result["卖三申报"] = parts[24]
	result["卖三报价"] = parts[25]
	result["卖四申报"] = parts[26]
	result["卖四报价"] = parts[27]
	result["卖五申报"] = parts[28]
	result["卖五报价"] = parts[29]
	result["日期"] = parts[30]
	result["时间"] = parts[31]
	//logger.SugaredLogger.Infof("股票数据解析完成: %v", result)
	marshal, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	stockInfo := &StockInfo{}
	err = json.Unmarshal(marshal, &stockInfo)
	if err != nil {
		return nil, err
	}
	//logger.SugaredLogger.Infof("股票数据解析完成stockInfo: %+v", stockInfo)

	return stockInfo, nil
}

type IndexBasic struct {
	gorm.Model
	TsCode        string  `json:"ts_code" gorm:"index"`
	Symbol        string  `json:"symbol" gorm:"index"`
	Name          string  `json:"name" gorm:"index"`
	FullName      string  `json:"fullname"`
	IndexType     string  `json:"index_type"`
	IndexCategory string  `json:"category"`
	Market        string  `json:"market"`
	ListDate      string  `json:"list_date"`
	BaseDate      string  `json:"base_date"`
	BasePoint     float64 `json:"base_point"`
	Publisher     string  `json:"publisher"`
	WeightRule    string  `json:"weight_rule"`
	DESC          string  `json:"desc"`
}

func (IndexBasic) TableName() string {
	return "tushare_index_basic"
}

func SearchStockInfo(stock, msgType string) *[]string {
	// 创建一个 chromedp 上下文
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	defer cancel()
	var htmlContent string
	url := fmt.Sprintf("https://www.cls.cn/searchPage?keyword=%s&type=%s", stock, msgType)
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		// 等待页面加载完成，可以根据需要调整等待时间
		chromedp.Sleep(3*time.Second),
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
		return &[]string{}
	}
	document, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		logger.SugaredLogger.Error(err.Error())
		return &[]string{}
	}
	var messages []string
	document.Find("div.search-telegraph-list,a.search-content").Each(func(i int, selection *goquery.Selection) {
		text := strutil.RemoveNonPrintable(selection.Text())
		if strings.Contains(text, stock) {
			messages = append(messages, text)
			logger.SugaredLogger.Infof("搜索到消息: %s", text)
		}
	})
	return &messages
}
