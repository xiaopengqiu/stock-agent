package main

import (
	"embed"
	"encoding/json"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go-stock/backend/data"
	"go-stock/backend/db"
	"log"
	"os"
	goruntime "runtime"
	"time"
)

//go:embed frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

//go:embed build/app.ico
var icon2 []byte

//go:embed build/stock_basic.json
var stocksBin []byte

//go:generate cp -R ./data ./build/bin
func main() {
	checkDir("data")
	db.Init("")
	db.Dao.AutoMigrate(&data.StockInfo{})
	db.Dao.AutoMigrate(&data.StockBasic{})
	db.Dao.AutoMigrate(&data.FollowedStock{})
	db.Dao.AutoMigrate(&data.IndexBasic{})

	if stocksBin != nil && len(stocksBin) > 0 {
		go initStockData()
	}
	updateBasicInfo()

	// Create an instance of the app structure
	app := NewApp()
	AppMenu := menu.NewMenu()
	FileMenu := AppMenu.AddSubmenu("设置")
	FileMenu.AddText("显示搜索框", keys.CmdOrCtrl("s"), func(callbackData *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "showSearch", 1)
	})
	FileMenu.AddText("隐藏搜索框", keys.CmdOrCtrl("d"), func(callbackData *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "showSearch", 0)
	})
	FileMenu.AddText("刷新数据", keys.CmdOrCtrl("r"), func(callbackData *menu.CallbackData) {
		//runtime.EventsEmit(app.ctx, "refresh", "setting-"+time.Now().Format("2006-01-02 15:04:05"))
		runtime.EventsEmit(app.ctx, "refreshFollowList", "refresh-"+time.Now().Format("2006-01-02 15:04:05"))
	})
	FileMenu.AddSeparator()
	FileMenu.AddText("窗口全屏", keys.CmdOrCtrl("f"), func(callback *menu.CallbackData) {
		runtime.WindowFullscreen(app.ctx)
	})
	FileMenu.AddText("窗口还原", keys.Key("Esc"), func(callback *menu.CallbackData) {
		runtime.WindowUnfullscreen(app.ctx)
	})

	if goruntime.GOOS == "windows" {
		FileMenu.AddText("隐藏到托盘区", keys.CmdOrCtrl("h"), func(_ *menu.CallbackData) {
			runtime.WindowHide(app.ctx)
		})
	}

	//FileMenu.AddText("退出", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
	//	runtime.Quit(app.ctx)
	//})

	// Create application with options
	err := wails.Run(&options.App{
		Title:             "go-stock",
		Width:             1366,
		Height:            920,
		MinWidth:          1024,
		MinHeight:         768,
		MaxWidth:          1280,
		MaxHeight:         960,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         true,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		Assets:            assets,
		Menu:              AppMenu,
		Logger:            nil,
		LogLevel:          logger.DEBUG,
		OnStartup:         app.startup,
		OnDomReady:        app.domReady,
		OnBeforeClose:     app.beforeClose,
		OnShutdown:        app.shutdown,
		WindowStartState:  options.Normal,
		Bind: []interface{}{
			app,
		},
		// Windows platform specific options
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			// DisableFramelessWindowDecorations: false,
			WebviewUserDataPath: "",
		},
		// Mac platform specific options
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "go-stock",
				Message: "",
				Icon:    icon,
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

}

func updateBasicInfo() {
	//更新基本信息
	go data.NewStockDataApi().GetStockBaseInfo()
	go data.NewStockDataApi().GetIndexBasic()
}

func initStockData() {
	var count int64
	db.Dao.Model(&data.StockBasic{}).Count(&count)
	if count > 0 {
		return
	}
	logger.NewDefaultLogger().Info("init stock data")
	res := &data.TushareStockBasicResponse{}
	err := json.Unmarshal(stocksBin, res)
	if err != nil {
		logger.NewDefaultLogger().Error(err.Error())
		return
	}
	for _, item := range res.Data.Items {
		stock := &data.StockBasic{}
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
		db.Dao.Model(&data.StockBasic{}).FirstOrCreate(stock, &data.StockBasic{TsCode: stock.TsCode}).Updates(stock)
	}
}

func checkDir(dir string) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		os.Mkdir(dir, os.ModePerm)
		logger.NewDefaultLogger().Info("create dir: " + dir)
	}
}
