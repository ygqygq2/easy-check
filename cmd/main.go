package main

import (
	"easy-check/internal/checker"
	"easy-check/internal/config"
	"easy-check/internal/initializer"
	"easy-check/internal/logger"
	"easy-check/internal/scheduler"
	"easy-check/internal/signal"
	"fmt"
	"path/filepath"
	"time"
)

func main() {
	// 初始化配置和通知器
	cfg, notifierInstance, err := initializer.Initialize()
	if err != nil {
		initializer.GlobalLogger.Fatal(fmt.Sprintf("Initialization error: %v", err), "fatal")
	}

	// 创建一个通道用于控制定时器
	tickerControlChan := make(chan time.Duration)

	// 首先启动配置文件监听
	go config.WatchConfigFile(filepath.Join("configs", "config.yaml"), initializer.GlobalLogger, func(newConfig *config.Config) {
		// 更新配置
		oldInterval := cfg.Interval

		// 保存旧的日志配置
		oldLogConfig := cfg.Log

		// 更新整个配置
		cfg = newConfig
		initializer.GlobalLogger.Log("Configuration reloaded successfully", "info")

		// 如果日志配置发生变化，更新日志器
		if oldLogConfig != cfg.Log {
			logConfig := logger.Config{
				File:         cfg.Log.File,
				MaxSize:      cfg.Log.MaxSize,
				MaxAge:       cfg.Log.MaxAge,
				MaxBackups:   cfg.Log.MaxBackups,
				Compress:     cfg.Log.Compress,
				ConsoleLevel: cfg.Log.ConsoleLevel,
				FileLevel:    cfg.Log.FileLevel,
			}
			initializer.GlobalLogger.UpdateConfig(logConfig)
			initializer.GlobalLogger.Log("Logger configuration updated", "info")
		}

		// 如果间隔时间发生变化，通知定时器更新
		if oldInterval != cfg.Interval {
			initializer.GlobalLogger.Log(fmt.Sprintf("Interval changed from %d to %d seconds", oldInterval, cfg.Interval), "info")
			tickerControlChan <- time.Duration(cfg.Interval) * time.Second
		}
	})

	// 然后初始化 pinger 和 checker
	pinger := checker.NewPinger()
	chk := checker.NewChecker(cfg, pinger, initializer.GlobalLogger, notifierInstance)

	// 执行初始 ping 检查
	initializer.GlobalLogger.Log("Performing initial ping check", "info")
	chk.PingHosts()

	// 最后启动定期 ping 检查，并传入控制通道
	stopChan := scheduler.StartPeriodicPingChecks(chk, cfg, initializer.GlobalLogger, tickerControlChan)

	// 等待退出信号
	signal.WaitForExitSignal(initializer.GlobalLogger)
	close(stopChan)
}
