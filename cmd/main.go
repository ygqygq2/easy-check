package main

import (
	"easy-check/internal/checker"
	"easy-check/internal/config"
	"easy-check/internal/db"
	"easy-check/internal/initializer"
	"easy-check/internal/logger"
	"easy-check/internal/notifier"
	"easy-check/internal/scheduler"
	"easy-check/internal/signal"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var version string // 通过 ldflags 注入

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

	// 创建一个通道用于控制定时器
	tickerControlChan := make(chan time.Duration)

	// 首先启动配置文件监听
	go config.WatchConfigFile(filepath.Join("configs", "config.yaml"), appCtx.Logger, func(newConfig *config.Config) {
		// 更新配置
		oldInterval := appCtx.Config.Interval
		oldPingInterval := appCtx.Config.Ping.Interval

		// 保存旧的日志配置
		oldLogConfig := appCtx.Config.Log

		// 更新整个配置
		appCtx.Config = newConfig
		appCtx.Logger.Log("Configuration reloaded successfully", "info")

		// 如果日志配置发生变化，更新日志器
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

		// 修改这部分，检查 ping.interval 和全局 interval 是否有变化
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

		// 如果实际使用的间隔时间发生变化，通知定时器更新
		if oldIntervalInUse != newIntervalToUse {
			appCtx.Logger.Log(fmt.Sprintf("Interval changed from %d to %d seconds", oldIntervalInUse, newIntervalToUse), "info")
			tickerControlChan <- time.Duration(newIntervalToUse) * time.Second
		}
	})

	// 然后初始化 pinger 和 checker
	pinger := checker.NewPinger()
	// 创建 AlertStatusManager
	alertStatusManager, err := db.NewAlertStatusManager(appCtx.DB.Instance, appCtx.Logger, appCtx.Config.Db)
	if err != nil {
		appCtx.Logger.Fatal("Failed to create AlertStatusManager", "error")
	}
	appCtx.Logger.Log("AlertStatusManager initialized successfully", "debug")
	appCtx.Logger.Log("Application started successfully", "info")

	interval := time.Duration(appCtx.Config.Alert.AggregateWindow) * time.Second
	consumer := notifier.NewConsumer(alertStatusManager, appCtx.Logger, interval, appCtx.AggregatorHandle)
	go consumer.Start()

	chk := checker.NewChecker(appCtx.Config, pinger, appCtx.Logger, alertStatusManager)
	// 执行初始 ping 检查
	appCtx.Logger.Log("Performing initial ping check", "info")
	chk.PingHosts()

	// 最后启动定期 ping 检查，并传入控制通道
	stopChan := scheduler.StartPeriodicPingChecks(chk, appCtx.Config, appCtx.Logger, tickerControlChan)

	// 等待退出信号
	exitChan := signal.RegisterExitListener()
	<-exitChan

	// 清理资源
	close(stopChan)
	if appCtx.Notifier != nil {
		appCtx.Notifier.Close()
	}
	appCtx.Logger.Log("Application shutting down", "info")
	os.Exit(0)
}
