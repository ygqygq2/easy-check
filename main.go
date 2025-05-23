package main

import (
	"context"
	"easy-check/internal/checker"
	"easy-check/internal/constants"
	"easy-check/internal/db"
	"easy-check/internal/initializer"
	"easy-check/internal/logger"
	"easy-check/internal/machineid"
	"embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var version string // 通过 ldflags 注入
//go:embed all:frontend/dist
var assets embed.FS
var wailsContext *context.Context

func main() {
	if version == "" {
		version = "dev" // 默认值
	}
	fmt.Printf("easy-check ui version: %s\n", version)

	// 如果是重启操作，稍微延迟一下以确保上一个实例已完全退出
	if os.Getenv("EASY_CHECK_RESTART") == "true" {
		time.Sleep(1 * time.Second)
	}

	machineID, err := machineid.GetMachineID()
	if err != nil {
		fmt.Printf("Failed to get machine ID, %v\n", err)
		os.Exit(1)
	}
	// 初始化配置和通知器
	fmt.Println("Initializing application context...")
	appCtx, err := initializer.Initialize(machineID, version)
	if err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}
	defer appCtx.Logger.Close()
	fmt.Println("Application context initialized successfully")

	// 启动后台任务
	go func() {
		// fmt.Println("Starting background task...")
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
		Logger:           &logger.WailsLogger{Logger: appCtx.Logger},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               machineID,
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	secondInstanceArgs := secondInstanceData.Args

	println("user opened second instance", strings.Join(secondInstanceData.Args, ","))
	println("user opened second from", secondInstanceData.WorkingDirectory)
	runtime.WindowUnminimise(*wailsContext)
	runtime.Show(*wailsContext)
	go runtime.EventsEmit(*wailsContext, "launchArgs", secondInstanceArgs)
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
