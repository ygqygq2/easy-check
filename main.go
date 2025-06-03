package main

import (
	"easy-check/internal/assets"
	"easy-check/internal/checker"
	"easy-check/internal/constants"
	"easy-check/internal/db"
	"easy-check/internal/initializer"
	"easy-check/internal/machineid"
	"easy-check/internal/services"
	"embed"
	"fmt"
	"os"
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

	// 创建主窗口
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

	// 启动后台任务
	go func() {
		runBackgroundTask(appCtx)
	}()

	err = app.Run()
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
