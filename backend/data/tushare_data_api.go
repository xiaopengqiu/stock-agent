package data

import (
	"fmt"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/go-resty/resty/v2"
	"go-stock/backend/logger"
	"strings"
	"time"
)

// @Author spark
// @Date 2025/2/17 12:33
// @Desc
//-----------------------------------------------------------------------------------

type TushareApi struct {
	client *resty.Client
	config *SettingConfig
}

func NewTushareApi(config *SettingConfig) *TushareApi {
	return &TushareApi{
		client: resty.New(),
		config: config,
	}
}

// GetDaily tushare A股日线行情
func (receiver TushareApi) GetDaily(tsCode, startDate, endDate string, crawlTimeOut int64) string {
	//logger.SugaredLogger.Debugf("tushare daily request: ts_code=%s, start_date=%s, end_date=%s", tsCode, startDate, endDate)
	fields := "ts_code,trade_date,open,high,low,close,pre_close,change,pct_chg,vol,amount"
	resp := &TushareStockBasicResponse{}
	stockType := getStockType(tsCode)
	tsCodeNEW := getTsCode(tsCode)
	//logger.SugaredLogger.Debugf("tushare daily request: %s,tsCode:%s,tsCodeNEW:%s", stockType, tsCode, tsCodeNEW)
	_, err := receiver.client.SetTimeout(time.Duration(crawlTimeOut)*time.Second).R().
		SetHeader("content-type", "application/json").
		SetBody(&TushareRequest{
			ApiName: stockType,
			Token:   receiver.config.TushareToken,
			Params: map[string]any{
				"ts_code":    tsCodeNEW,
				"start_date": startDate,
				"end_date":   endDate,
			},
			Fields: fields}).
		SetResult(resp).
		Post(tushareApiUrl)
	if err != nil {
		logger.SugaredLogger.Error(err)
		return ""
	}
	res := ""
	if resp.Data.Items != nil && len(resp.Data.Items) > 0 {
		fieldsStr := slice.JoinFunc(resp.Data.Fields, ",", func(s string) string {
			return "\"" + convertor.ToString(s) + "\""
		})
		res += fieldsStr + "\n"
		for _, item := range resp.Data.Items {
			//logger.SugaredLogger.Debugf("%s", slice.Join(item, ","))
			t := slice.JoinFunc(item, ",", func(s any) any {
				return "\"" + convertor.ToString(s) + "\""
			})
			res += t + "\n"
		}
	}
	//logger.SugaredLogger.Debugf("tushare response: %s", res)
	return res
}

func getTsCode(code string) any {
	if strutil.HasPrefixAny(code, []string{"US", "us", "gb_"}) {
		code = strings.Replace(code, "gb_", "", 1)
		code = strings.Replace(code, "us", "", 1)
		return code
	}
	return code
}

func getStockType(code string) string {
	if strutil.HasSuffixAny(code, []string{"SZ", "SH", "sh", "sz"}) {
		return "daily"
	}
	if strutil.HasSuffixAny(code, []string{"HK", "hk"}) {
		return "hk_daily"
	}
	if strutil.HasPrefixAny(code, []string{"US", "us", "gb_"}) {
		return "us_daily"
	}
	return ""
}

// ============================================
// 港股价格查询 API
// ============================================

// HKStockPriceInfo 港股价格信息
type HKStockPriceInfo struct {
	StockCode     string  `json:"stock_code"`     // 股票代码
	StockName     string  `json:"stock_name"`     // 股票名称
	CurrentPrice  float64 `json:"current_price"`  // 当前价格
	Change        float64 `json:"change"`         // 涨跌额
	ChangePercent float64 `json:"change_percent"` // 涨跌幅%
	OpenPrice     float64 `json:"open_price"`     // 开盘价
	HighPrice     float64 `json:"high_price"`     // 最高价
	LowPrice      float64 `json:"low_price"`      // 最低价
	PrevClose     float64 `json:"prev_close"`     // 昨收价
	Volume        int64   `json:"volume"`         // 成交量
	Turnover      float64 `json:"turnover"`       // 成交额
	UpdateTime    string  `json:"update_time"`    // 更新时间
}

// GetHKStockPrice 获取港股实时价格
// 支持港股通标的和主要港股
func (api TushareApi) GetHKStockPrice(stockCodes []string) ([]HKStockPriceInfo, error) {
	if len(stockCodes) == 0 {
		return nil, fmt.Errorf("股票代码不能为空")
	}

	// 转换代码格式为hk开头
	var hkCodes []string
	for _, code := range stockCodes {
		hkCode := normalizeHKStockCode(code)
		hkCodes = append(hkCodes, hkCode)
	}

	logger.SugaredLogger.Infof("查询港股价格: %v", hkCodes)

	// 使用现有的GetStockCodeRealTimeData方法获取数据
	stockDataApi := NewStockDataApi()
	realTimeData, err := stockDataApi.GetStockCodeRealTimeData(hkCodes...)
	if err != nil {
		logger.SugaredLogger.Errorf("获取港股价格失败: %v", err)
		return nil, err
	}

	if realTimeData == nil || len(*realTimeData) == 0 {
		return nil, fmt.Errorf("未获取到港股价格数据")
	}

	// 转换为HKStockPriceInfo格式
	var results []HKStockPriceInfo
	for _, stock := range *realTimeData {
		info := HKStockPriceInfo{
			StockCode:     stock.Code,
			StockName:     stock.Name,
			CurrentPrice:  parseFloat(stock.Price),
			Change:        stock.ChangePrice,
			ChangePercent: stock.ChangePercent,
			OpenPrice:     parseFloat(stock.Open),
			HighPrice:     parseFloat(stock.High),
			LowPrice:      parseFloat(stock.Low),
			PrevClose:     parseFloat(stock.PreClose),
			Volume:        parseInt64(stock.Volume),
			Turnover:      parseFloat(stock.Amount),
			UpdateTime:    stock.Time,
		}
		results = append(results, info)
	}

	return results, nil
}

// normalizeHKStockCode 将各种格式的港股代码统一为hk开头
func normalizeHKStockCode(code string) string {
	code = strings.TrimSpace(strings.ToUpper(code))

	// 去除.HK后缀
	if strings.HasSuffix(code, ".HK") {
		code = strings.TrimSuffix(code, ".HK")
	}

	// 去除HK前缀
	if strings.HasPrefix(code, "HK") {
		code = code[2:]
	}

	// 补足5位代码
	for len(code) < 5 {
		code = "0" + code
	}

	return "hk" + code
}

// parseFloat 安全解析float
func parseFloat(s string) float64 {
	var result float64
	fmt.Sscanf(s, "%f", &result)
	return result
}

// parseInt64 安全解析int64
func parseInt64(s string) int64 {
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}

// ============================================
// 股东人数查询 API
// ============================================

// ShareholderCountData 股东人数数据
type ShareholderCountData struct {
	StockCode     string                     `json:"stock_code"`
	StockName     string                     `json:"stock_name"`
	QuarterlyData []QuarterlyShareholderInfo `json:"quarterly_data"`
}

// QuarterlyShareholderInfo 季度股东人数信息
type QuarterlyShareholderInfo struct {
	Quarter   string  `json:"quarter"`    // 季度，如 "2025Q3"
	Count     int     `json:"count"`      // 股东人数
	AvgShares float64 `json:"avg_shares"` // 人均持股
	Change    int     `json:"change"`     // 较上季度变化
	ChangePct float64 `json:"change_pct"` // 变化百分比
}

// GetShareholderCount 获取股东人数数据
// quarters: 查询最近多少个季度的数据
func (api TushareApi) GetShareholderCount(stockCode string, quarters int) (*ShareholderCountData, error) {
	if quarters <= 0 {
		quarters = 4
	}
	if quarters > 20 {
		quarters = 20 // 最多查询20个季度（5年）
	}

	logger.SugaredLogger.Infof("查询股东人数: %s, 季度数: %d", stockCode, quarters)

	// 使用Tushare API获取股东人数数据
	result, err := api.queryShareholderCountFromTushare(stockCode, quarters)
	if err != nil {
		logger.SugaredLogger.Errorf("从Tushare获取股东人数失败: %v", err)
		// 尝试从其他数据源获取
		result, err = api.queryShareholderCountFromAlternative(stockCode, quarters)
		if err != nil {
			return nil, fmt.Errorf("获取股东人数数据失败: %v", err)
		}
	}

	return result, nil
}

// queryShareholderCountFromTushare 从Tushare查询股东人数
func (api TushareApi) queryShareholderCountFromTushare(stockCode string, quarters int) (*ShareholderCountData, error) {
	// Tushare的stk_holdernumber接口
	fields := "ts_code,end_date,holder_num,holder_nums"

	resp := &TushareStockBasicResponse{}
	tsCode := getTsCode(stockCode)

	_, err := api.client.SetTimeout(30*time.Second).R().
		SetHeader("content-type", "application/json").
		SetBody(&TushareRequest{
			ApiName: "stk_holdernumber",
			Token:   api.config.TushareToken,
			Params: map[string]any{
				"ts_code": tsCode,
			},
			Fields: fields}).
		SetResult(resp).
		Post(tushareApiUrl)

	if err != nil {
		return nil, err
	}

	// 解析返回数据
	if resp.Data.Items == nil || len(resp.Data.Items) == 0 {
		return nil, fmt.Errorf("未获取到股东人数数据")
	}

	// 构建ShareholderCountData
	result := &ShareholderCountData{
		StockCode:     stockCode,
		QuarterlyData: []QuarterlyShareholderInfo{},
	}

	// 解析Tushare返回的数据
	// 字段顺序：ts_code, end_date, holder_num, holder_nums
	for i, item := range resp.Data.Items {
		if i >= quarters {
			break
		}

		if len(item) < 3 {
			continue
		}

		quarter := convertor.ToString(item[1]) // end_date
		count := 0
		if item[2] != nil {
			count64, _ := convertor.ToInt(item[2]) // holder_num
			count = int(count64)
		}

		info := QuarterlyShareholderInfo{
			Quarter: quarter,
			Count:   count,
		}

		result.QuarterlyData = append(result.QuarterlyData, info)
	}

	// 计算变化率和趋势
	for i := 0; i < len(result.QuarterlyData); i++ {
		if i < len(result.QuarterlyData)-1 {
			current := result.QuarterlyData[i]
			next := result.QuarterlyData[i+1]

			result.QuarterlyData[i].Change = current.Count - next.Count
			if next.Count > 0 {
				result.QuarterlyData[i].ChangePct = float64(current.Count-next.Count) / float64(next.Count) * 100
			}
		}
	}

	return result, nil
}

// queryShareholderCountFromAlternative 从备用数据源查询股东人数
func (api TushareApi) queryShareholderCountFromAlternative(stockCode string, quarters int) (*ShareholderCountData, error) {
	// TODO: 实现从其他数据源获取股东人数
	// 可考虑的数据源：
	// 1. 东方财富
	// 2. 同花顺
	// 3. 新浪财经
	return nil, fmt.Errorf("备用数据源暂未实现")
}
