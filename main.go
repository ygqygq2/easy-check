package main

import (
	"easy-check/internal/assets"
	"easy-check/internal/checker"
	"easy-check/internal/config"
	"easy-check/internal/constants"
	"easy-check/internal/db"
	"easy-check/internal/initializer"
	"easy-check/internal/logger"
	"easy-check/internal/machineid"
	"easy-check/internal/notifier"
	"easy-check/internal/router"
	"easy-check/internal/services"
	"embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var version string // 通过 ldflags 注入
//go:embed all:frontend/dist
var frontendAssets embed.FS

var encryptionKey = [32]byte{
	0x57, 0xad, 0xf9, 0xaa, 0x2d, 0xf1, 0x53, 0x28,
	0x2e, 0x2c, 0x6f, 0x6f, 0xfd, 0xf8, 0xc7, 0x55,
	0x9a, 0x53, 0x92, 0xef, 0xed, 0x50, 0xec, 0x6b,
	0xc3, 0x4c, 0x09, 0x06, 0xc7, 0x9c, 0xa1, 0x4d,
}

func main() {
	enableSystray := false
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

	pinger := checker.NewPinger()
	alertStatusManager, err := db.NewAlertStatusManager(appCtx.DB.Instance, appCtx.Logger, appCtx.Config.Db)
	if err != nil {
		appCtx.Logger.Fatal("Failed to create AlertStatusManager", "error")
	}
	chk := checker.NewChecker(appCtx.Config, pinger, appCtx.Logger, alertStatusManager, appCtx.TSDB)

	// ========== 1. 启动配置文件热加载监听 ==========
	tickerControlChan := make(chan time.Duration)
	go config.WatchConfigFile(filepath.Join("configs", "config.yaml"), appCtx.Logger, func(newConfig *config.Config) {
		oldInterval := appCtx.Config.Interval
		oldPingInterval := appCtx.Config.Ping.Interval
		oldLogConfig := appCtx.Config.Log
		appCtx.Config = newConfig
		
		// 更新Checker的配置（线程安全）
		chk.UpdateConfig(newConfig)
		
		appCtx.Logger.Log("Configuration reloaded successfully", "info")
		// 日志配置热更新
		if oldLogConfig != appCtx.Config.Log {
			logConfig := logger.Config{
				File:         appCtx.Config.Log.File,
				MaxSize:      appCtx.Config.Log.MaxSize,
				MaxAge:       appCtx.Config.Log.MaxAge,
				MaxBackups:   appCtx.Config.Log.MaxBackups,
				Compress:     appCtx.Config.Log.Compress,
				ConsoleLevel: appCtx.Config.Log.ConsoleLevel,
				FileLevel:    appCtx.Config.Log.FileLevel,
			}
			appCtx.Logger.UpdateConfig(logConfig)
			appCtx.Logger.Log("Logger configuration updated", "info")
		}
		// interval 热更新
		var newIntervalToUse int
		if appCtx.Config.Ping.Interval > 0 {
			newIntervalToUse = appCtx.Config.Ping.Interval
		} else {
			newIntervalToUse = appCtx.Config.Interval
		}
		var oldIntervalInUse int
		if oldPingInterval > 0 {
			oldIntervalInUse = oldPingInterval
		} else {
			oldIntervalInUse = oldInterval
		}
		if oldIntervalInUse != newIntervalToUse {
			appCtx.Logger.Log(fmt.Sprintf("Interval changed from %d to %d seconds", oldIntervalInUse, newIntervalToUse), "info")
			tickerControlChan <- time.Duration(newIntervalToUse) * time.Second
		}
	})

	// ========== 2. 注册路由、HTTP服务、Wails窗口 ==========
	router.RegisterRoutes(appCtx)
	go func() {
		if err := http.ListenAndServe("127.0.0.1:32180", nil); err != nil {
			appCtx.Logger.Fatalf("Server failed: %v", err)
		}
	}()

	constInfo := constants.GetSharedConstants(appCtx)
	appService := services.NewAppService(appCtx)

	var window *application.WebviewWindow
	app := application.New(application.Options{
		Name:        constInfo.AppName,
		Description: "简单网络检测工具",
		Services: []application.Service{
			application.NewService(appService),
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID:      "com.ygqygq2.easy-check",
			EncryptionKey: encryptionKey,
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				if window != nil {
					window.EmitEvent("secondInstanceLaunched", data)
					window.Show()
					window.Focus()
				}
			},
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(frontendAssets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	window = app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Title:  constInfo.AppName,
		Width:  1024,
		Height: 768,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour:    application.NewRGB(27, 38, 54),
		URL:                 "/",
		MinimiseButtonState: application.ButtonEnabled,
		MaximiseButtonState: application.ButtonDisabled,
		CloseButtonState:    application.ButtonDisabled,
	})

	// 根据开关决定是否初始化系统托盘
	if enableSystray {
		sysTray := app.NewSystemTray()
		sysTray.SetLabel(constInfo.AppName)

		// 将窗口附加到系统托盘（图标点击可控制窗口显示/隐藏）
		sysTray.AttachWindow(window)
		sysTray.WindowOffset(100)
		sysTray.WindowDebounce(200 * time.Millisecond)
		logo, err := assets.GetAsset("images/logo.png")
		if err != nil {
			fmt.Printf("Failed to load logo asset: %v\n", err)
		}
		if runtime.GOOS == "darwin" {
			sysTray.SetTemplateIcon(logo)
		} else {
			sysTray.SetIcon(logo)
		}

		// 托盘菜单：打开窗口、退出
		menu := app.NewMenu()
		menu.Add("打开窗口").OnClick(func(_ *application.Context) {
			window.Show()
		})
		menu.Add("退出").OnClick(func(_ *application.Context) {
			app.Quit()
		})
		sysTray.SetMenu(menu)

		// Windows 平台下额外监听任务栏重建事件
		if runtime.GOOS == "windows" {
			go watchTaskbarCreated(sysTray, window, logo, menu)
		}
	}

	// ========== 3. 启动后台任务（ping 检查、告警消费者） ==========
	go func() {
		runBackgroundTask(appCtx, tickerControlChan, chk)
	}()

	err = app.Run()
	if err != nil {
		println("Error:", err.Error())
	}
}

// runBackgroundTask 启动后台任务，支持 interval 动态调整
func runBackgroundTask(appCtx *initializer.AppContext, tickerControlChan chan time.Duration, chk *checker.Checker) {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Recovered from panic: %v", r)
			appCtx.Logger.Log(message, "error")
		}
	}()

	alertStatusManager, err := db.NewAlertStatusManager(appCtx.DB.Instance, appCtx.Logger, appCtx.Config.Db)
	if err != nil {
		appCtx.Logger.Fatal("Failed to create AlertStatusManager", "error")
	}

	// 启动告警/恢复消费者（定时发送告警和恢复通知）
	interval := time.Duration(appCtx.Config.Alert.AggregateWindow) * time.Second
	consumer := notifier.NewConsumer(alertStatusManager, appCtx.Logger, interval, appCtx.AggregatorHandle)
	go consumer.Start()

	// 执行初始 ping 检查
	appCtx.Logger.Log("Performing initial ping check", "info")
	chk.PingHosts()

	// 动态 interval 定时器
	var intervalToUse time.Duration
	if appCtx.Config.Ping.Interval > 0 {
		intervalToUse = time.Duration(appCtx.Config.Ping.Interval) * time.Second
	} else {
		intervalToUse = time.Duration(appCtx.Config.Interval) * time.Second
	}
	ticker := time.NewTicker(intervalToUse)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			chk.PingHosts()
		case newInterval := <-tickerControlChan:
			appCtx.Logger.Log(fmt.Sprintf("Ping interval updated to %v", newInterval), "info")
			ticker.Stop()
			ticker = time.NewTicker(newInterval)
		}
	}
}
