package main

import (
	"easy-check/internal/checker"
	"easy-check/internal/constants"
	"easy-check/internal/db"
	"easy-check/internal/initializer"
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

var version string // 通过 ldflags 注入
//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if version == "" {
		version = "dev" // 默认值
	}
	fmt.Printf("easy-check ui version: %s\n", version)

	// 初始化配置和通知器
	fmt.Println("Initializing application context...")
	appCtx, err := initializer.Initialize(version)
	if err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}
	defer appCtx.Logger.Close()
	fmt.Println("Application context initialized successfully")

	// 启动后台任务
	go func() {
		fmt.Println("Starting background task...")
		runBackgroundTask(appCtx)
	}()

	// Create an instance of the app structure
	app := NewApp(appCtx)

	constInfo := constants.GetSharedConstants(appCtx)
	// Create application with options
	err = wails.Run(&options.App{
		Title:  constInfo.AppName,
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// runBackgroundTask 启动后台任务
func runBackgroundTask(appCtx *initializer.AppContext) {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Recovered from panic: %v", r)
			appCtx.Logger.Log(message, "error")
		}
	}()

	pinger := checker.NewPinger()
	alertStatusManager, err := db.NewAlertStatusManager(appCtx.DB.Instance, appCtx.Logger, appCtx.Config.Db)
	if err != nil {
		appCtx.Logger.Fatal("Failed to create AlertStatusManager", "error")
	}

	chk := checker.NewChecker(appCtx.Config, pinger, appCtx.Logger, alertStatusManager, appCtx.TSDB)

	for {
		chk.PingHosts()
		time.Sleep(time.Duration(appCtx.Config.Ping.Interval) * time.Second)
	}
}
