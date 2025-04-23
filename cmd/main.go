package main

import (
	"easy-check/internal/assets"
	"easy-check/internal/checker"
	"easy-check/internal/db"
	"easy-check/internal/initializer"
	"easy-check/internal/signal"
	_ "embed"
	"fmt"
	"os"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/getlantern/systray"
)

var version string // 通过 ldflags 注入
var (
	logBuffer []string
	logMutex  sync.Mutex
)

func main() {
	if version == "" {
		version = "dev" // 默认值
	}
	fmt.Printf("easy-check version: %s\n", version)

	// 初始化配置和通知器
	appCtx, err := initializer.Initialize()
	if err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}
	defer appCtx.Logger.Close()

	// 启动系统托盘
	go func() {
		systray.Run(onReady, onExit)
	}()

	// 启动 Gio UI
	go func() {
		window := new(app.Window)
    window.Option(app.Title("Easy Check 监控工具"))

		err := run(window, appCtx)
		if err != nil {
			appCtx.Logger.Fatal(fmt.Sprintf("Failed to run UI: %v", err))
		}
		os.Exit(0)
	}()

	// 启动后台任务
	go runBackgroundTask(appCtx)

	// 等待退出信号
	exitChan := signal.RegisterExitListener()
	<-exitChan

	// 清理资源
	if appCtx.Notifier != nil {
		appCtx.Notifier.Close()
	}
	appCtx.Logger.Log("Application shutting down", "info")
	os.Exit(0)
}

func runBackgroundTask(appCtx *initializer.AppContext) {
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Recovered from panic: %v", r)
			appCtx.Logger.Log(message, "error")
			AppendLog(message)
		}
	}()

	pinger := checker.NewPinger()
	alertStatusManager, err := db.NewAlertStatusManager(appCtx.DB.Instance, appCtx.Logger, appCtx.Config.Db)
	if err != nil {
		message := "Failed to create AlertStatusManager"
		appCtx.Logger.Fatal(message, "error")
		AppendLog(message)
	}
	chk := checker.NewChecker(appCtx.Config, pinger, appCtx.Logger, alertStatusManager)

	for {
		chk.PingHosts()
		AppendLog("Pinged hosts successfully")
		time.Sleep(time.Duration(appCtx.Config.Ping.Interval) * time.Second)
	}
}

func onReady() {
	systray.SetIcon(getIcon())
	systray.SetTitle("easy-check")
	systray.SetTooltip("easy-check is running")

	mQuit := systray.AddMenuItem("Quit", "Quit the application")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {
	// 清理资源
}

func getIcon() []byte {
	icon, err := assets.GetAsset("favicon.ico")
	if err != nil {
		fmt.Printf("Failed to load icon: %v\n", err)
		return nil
	}
	return icon
}

func run(window *app.Window, appCtx *initializer.AppContext) error {
	theme := material.NewTheme()
	var ops op.Ops

	// 定义按钮状态
	openConfigButton := new(widget.Clickable)
	viewLogsButton := new(widget.Clickable)
	showLatestLogsButton := new(widget.Clickable)

	// 定义日志列表
	var logList layout.List
	logList.Axis = layout.Vertical // 垂直滚动

	for {
		// 使用 window.Event() 获取事件
		e := window.Event()
		switch e := e.(type) {
		case app.DestroyEvent:
			// 窗口关闭事件
			fmt.Println("Window closed, application continues running in the background.")
			return e.Err
		case app.FrameEvent:
			// 创建布局上下文
			gtx := layout.Context{
				Ops: &ops,
				Now: e.Now,
			}

			// 布局窗口内容
			layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				// 顶部按钮区域
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Horizontal,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return material.Button(theme, openConfigButton, "打开配置文件").Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Spacer{Width: 10}.Layout(gtx) // 添加间距
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return material.Button(theme, viewLogsButton, "查看日志文件").Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Spacer{Width: 10}.Layout(gtx) // 添加间距
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return material.Button(theme, showLatestLogsButton, "显示最新日志").Layout(gtx)
						}),
					)
				}),
				// 日志显示区域
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					// 使用 layout.List 显示日志
					return logList.Layout(gtx, len(logBuffer), func(gtx layout.Context, index int) layout.Dimensions {
						logMutex.Lock()
						defer logMutex.Unlock()
						if index < len(logBuffer) {
							return material.Label(theme, unit.Sp(14), logBuffer[index]).Layout(gtx)
						}
						return layout.Dimensions{}
					})
				}),
			)

			// 提交绘图操作
			e.Frame(gtx.Ops)
		}
	}
}

// AppendLog 将日志追加到缓冲区
func AppendLog(message string) {
	logMutex.Lock()
	defer logMutex.Unlock()
	if len(logBuffer) > 100 { // 限制日志条数
		logBuffer = logBuffer[1:]
	}
	logBuffer = append(logBuffer, message)
}

// GetLogs 获取当前日志内容
func GetLogs() string {
	logMutex.Lock()
	defer logMutex.Unlock()
	return fmt.Sprintf("%s", logBuffer)
}
