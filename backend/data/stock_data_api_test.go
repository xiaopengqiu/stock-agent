package data

import (
	"encoding/json"
	"github.com/duke-git/lancet/v2/convertor"
	"go-stock/backend/db"
	"io/ioutil"
	"strings"
	"testing"
)

// @Author spark
// @Date 2024/12/10 9:55
// @Desc
//-----------------------------------------------------------------------------------

func TestNewStockDataApi(t *testing.T) {
	stockDataApi := NewStockDataApi()
	t.Log(stockDataApi.GetStockCodeRealTimeData("sh600859"))
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
	t.Log(stockDataApi.GetFollowList())

}
