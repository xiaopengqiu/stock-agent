//go:build windows

package main

import (
	"context"
	"github.com/coocood/freecache"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/mathutil"
	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"time"
)

// App struct
type App struct {
	ctx   context.Context
	cache *freecache.Cache
}

// NewApp creates a new App application struct
func NewApp() *App {
	cacheSize := 512 * 1024
	cache := freecache.NewCache(cacheSize)
	return &App{
		cache: cache,
	}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx

	// 创建系统托盘
	go systray.Run(func() {
		onReady(a)
	}, func() {
		onExit(a)
	})

}

// domReady is called after front-end resources have been loaded
func (a *App) domReady(ctx context.Context) {
	// Add your action here

	//定时更新数据
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			runtime.WindowSetTitle(ctx, "go-stock "+time.Now().Format("2006-01-02 15:04"))
		}
	}()

	//定时更新数据
	go func() {
		ticker := time.NewTicker(time.Second * 2)
		defer ticker.Stop()
		for range ticker.C {
			if isTradingTime(time.Now()) {
				MonitorStockPrices(a)
			}
		}
	}()
}

// isTradingDay 判断是否是交易日
func isTradingDay(date time.Time) bool {
	weekday := date.Weekday()
	// 判断是否是周末
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}
	// 这里可以添加具体的节假日判断逻辑
	// 例如：判断是否是春节、国庆节等
	return true
}

// isTradingTime 判断是否是交易时间
func isTradingTime(date time.Time) bool {
	if !isTradingDay(date) {
		return false
	}

	hour, minute, _ := date.Clock()

	// 判断是否在9:15到11:30之间
	if (hour == 9 && minute >= 15) || (hour == 10) || (hour == 11 && minute <= 30) {
		return true
	}

	// 判断是否在13:00到15:00之间
	if (hour == 13) || (hour == 14) || (hour == 15 && minute <= 0) {
		return true
	}

	return false
}

func MonitorStockPrices(a *App) {
	dest := &[]data.FollowedStock{}
	db.Dao.Model(&data.FollowedStock{}).Find(dest)
	for _, item := range *dest {
		follow := item
		stockCode := follow.StockCode
		go func() {
			stockData, err := data.NewStockDataApi().GetStockCodeRealTimeData(stockCode)
			if err != nil {
				logger.SugaredLogger.Errorf("get stock code real time data error:%s", err.Error())
				return
			}
			price, err := convertor.ToFloat(stockData.Price)
			if err != nil {
				return
			}
			stockData.PrePrice = follow.Price
			if follow.Price != price {
				runtime.EventsEmit(a.ctx, "stock_price", stockData)
				go db.Dao.Model(follow).Where("stock_code = ?", stockCode).Updates(map[string]interface{}{
					"price": stockData.Price,
				})
			}
		}()
	}
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {

	dialog, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:         runtime.QuestionDialog,
		Title:        "go-stock",
		Message:      "确定关闭吗？",
		Buttons:      []string{"确定"},
		Icon:         icon,
		CancelButton: "取消",
	})

	if err != nil {
		logger.SugaredLogger.Errorf("dialog error:%s", err.Error())
		return false
	}
	logger.SugaredLogger.Debugf("dialog:%s", dialog)
	if dialog == "No" {
		return true
	}
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Perform your teardown here
	systray.Quit()
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) *data.StockInfo {
	stockInfo, _ := data.NewStockDataApi().GetStockCodeRealTimeData(name)
	return stockInfo
}

func (a *App) Follow(stockCode string) string {
	return data.NewStockDataApi().Follow(stockCode)
}

func (a *App) UnFollow(stockCode string) string {
	return data.NewStockDataApi().UnFollow(stockCode)
}

func (a *App) GetFollowList() []data.FollowedStock {
	return data.NewStockDataApi().GetFollowList()
}

func (a *App) GetStockList(key string) []data.StockBasic {
	return data.NewStockDataApi().GetStockList(key)
}

func (a *App) SetCostPriceAndVolume(stockCode string, price float64, volume int64) string {
	return data.NewStockDataApi().SetCostPriceAndVolume(price, volume, stockCode)
}

func (a *App) SetAlarmChangePercent(val, alarmPrice float64, stockCode string) string {
	return data.NewStockDataApi().SetAlarmChangePercent(val, alarmPrice, stockCode)
}
func (a *App) SetStockSort(sort int64, stockCode string) {
	data.NewStockDataApi().SetStockSort(sort, stockCode)
}
func (a *App) SendDingDingMessage(message string, stockCode string) string {
	ttl, _ := a.cache.TTL([]byte(stockCode))
	logger.SugaredLogger.Infof("stockCode %s ttl:%d", stockCode, ttl)
	if ttl > 0 {
		return ""
	}
	err := a.cache.Set([]byte(stockCode), []byte("1"), 60*5)
	if err != nil {
		logger.SugaredLogger.Errorf("set cache error:%s", err.Error())
		return ""
	}
	return data.NewDingDingAPI().SendDingDingMessage(message)
}

// SendDingDingMessageByType msgType 报警类型: 1 涨跌报警;2 股价报警 3 成本价报警
func (a *App) SendDingDingMessageByType(message string, stockCode string, msgType int) string {
	ttl, _ := a.cache.TTL([]byte(stockCode))
	logger.SugaredLogger.Infof("stockCode %s ttl:%d", stockCode, ttl)
	if ttl > 0 {
		return ""
	}
	err := a.cache.Set([]byte(stockCode), []byte("1"), getMsgTypeTTL(msgType))
	if err != nil {
		logger.SugaredLogger.Errorf("set cache error:%s", err.Error())
		return ""
	}
	stockInfo := &data.StockInfo{}
	db.Dao.Model(stockInfo).Where("code = ?", stockCode).First(stockInfo)
	if !data.NewAlertWindowsApi("go-stock消息通知", getMsgTypeName(msgType), GenNotificationMsg(stockInfo), "").SendNotification() {
		return data.NewDingDingAPI().SendDingDingMessage(message)
	}
	return "发送系统消息成功"
}

func GenNotificationMsg(stockInfo *data.StockInfo) string {
	Price, err := convertor.ToFloat(stockInfo.Price)
	if err != nil {
		Price = 0
	}
	PreClose, err := convertor.ToFloat(stockInfo.PreClose)
	if err != nil {
		PreClose = 0
	}
	var RF float64
	if PreClose > 0 {
		RF = mathutil.RoundToFloat(((Price-PreClose)/PreClose)*100, 2)
	}

	return "[" + stockInfo.Name + "] " + stockInfo.Price + " " + convertor.ToString(RF) + "% " + stockInfo.Date + " " + stockInfo.Time
}

// msgType : 1 涨跌报警(5分钟);2 股价报警(30分钟) 3 成本价报警(30分钟)
func getMsgTypeTTL(msgType int) int {
	switch msgType {
	case 1:
		return 60 * 5
	case 2:
		return 60 * 30
	case 3:
		return 60 * 30
	default:
		return 60 * 5
	}
}

func getMsgTypeName(msgType int) string {
	switch msgType {
	case 1:
		return "涨跌报警"
	case 2:
		return "股价报警"
	case 3:
		return "成本价报警"
	default:
		return "未知类型"
	}
}

func onExit(a *App) {
	// 清理操作
	logger.SugaredLogger.Infof("onExit")
	runtime.Quit(a.ctx)
}

func onReady(a *App) {

	// 初始化操作
	logger.SugaredLogger.Infof("onReady")
	systray.SetIcon(icon2)
	systray.SetTitle("go-stock")
	systray.SetTooltip("go-stock 股票行情实时获取")

	// 创建菜单项
	show := systray.AddMenuItem("显示", "显示应用程序")
	//hide := systray.AddMenuItem("隐藏", "隐藏应用程序")
	systray.AddSeparator()
	mQuitOrig := systray.AddMenuItem("退出", "退出应用程序")

	// 监听菜单项点击事件
	go func() {
		for {
			select {
			case <-mQuitOrig.ClickedCh:
				logger.SugaredLogger.Infof("退出应用程序")
				runtime.Quit(a.ctx)
				//systray.Quit()
			case <-show.ClickedCh:
				logger.SugaredLogger.Infof("显示应用程序")
				runtime.WindowShow(a.ctx)
				//runtime.WindowShow(a.ctx)
				//case <-hide.ClickedCh:
				//	logger.SugaredLogger.Infof("隐藏应用程序")
				//	runtime.Hide(a.ctx)

			}
		}
	}()
}
