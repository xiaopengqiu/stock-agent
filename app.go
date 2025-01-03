package main

import (
	"context"
	"github.com/coocood/freecache"
	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go-stock/backend/data"
	"go-stock/backend/logger"
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
	//ticker := time.NewTicker(time.Second)
	//defer ticker.Stop()
	////定时更新数据
	//go func() {
	//	for range ticker.C {
	//		runtime.WindowSetTitle(ctx, "go-stock "+time.Now().Format("2006-01-02 15:04:05"))
	//	}
	//}()
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

func (a *App) SetAlarmChangePercent(val float64, stockCode string) string {
	return data.NewStockDataApi().SetAlarmChangePercent(val, stockCode)
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
	systray.SetTooltip("这是一个简单的系统托盘示例")
	// 创建菜单项
	mQuitOrig := systray.AddMenuItem("退出", "退出应用程序")
	show := systray.AddMenuItem("显示", "显示应用程序")
	hide := systray.AddMenuItem("隐藏应用程序", "隐藏应用程序")

	// 监听菜单项点击事件
	go func() {
		for {
			select {
			case <-mQuitOrig.ClickedCh:
				logger.SugaredLogger.Infof("退出应用程序")
				runtime.Quit(a.ctx)
				systray.Quit()
			case <-show.ClickedCh:
				logger.SugaredLogger.Infof("显示应用程序")
				runtime.Show(a.ctx)
				//runtime.WindowShow(a.ctx)
			case <-hide.ClickedCh:
				logger.SugaredLogger.Infof("隐藏应用程序")
				runtime.Hide(a.ctx)

			}
		}
	}()
}
