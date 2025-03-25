package main

import (
	"easy-check/internal/checker"
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"easy-check/internal/notifier"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// 声明一个全局变量
var globalLogger *logger.Logger

func main() {
	// 切换到项目根目录
	err := changeToProjectRoot()
	if err != nil {
		globalLogger.Fatal(fmt.Sprintf("Error changing directory to project root: %v", err), "fatal")
	}

	// 尽早初始化一个默认日志记录器（控制台输出）
	globalLogger = logger.NewDefaultLogger()
	defer globalLogger.Close()

	// 加载配置
	configPath := filepath.Join("configs", "config.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		globalLogger.Fatal(fmt.Sprintf("Error loading configuration: %v", err), "error")
	}

	// 使用配置更新日志记录器
	logConfig := logger.Config{
		File:         cfg.Log.File,
		MaxSize:      cfg.Log.MaxSize,
		MaxAge:       cfg.Log.MaxAge,
		MaxBackups:   cfg.Log.MaxBackups,
		Compress:     cfg.Log.Compress,
		ConsoleLevel: cfg.Log.ConsoleLevel,
		FileLevel:    cfg.Log.FileLevel,
	}
	globalLogger.UpdateConfig(logConfig)

	globalLogger.Log("Starting easy-check...", "info")
	globalLogger.Log("Configuration loaded successfully", "info")

	// 初始化 Notifier
	var notifierInstance notifier.Notifier
	if cfg.Alert.Feishu.Enable && cfg.Alert.Feishu.MsgType == "text" {
		feishuNotifier, err := notifier.NewFeishuNotifier(&cfg.Alert.Feishu, globalLogger)
		if err != nil {
			globalLogger.Log(fmt.Sprintf("Failed to initialize FeishuNotifier: %v", err), "error")
			return
		}
		notifierInstance = feishuNotifier
	} else {
		globalLogger.Log("No valid notifier configuration found", "warn")
	}

	// 创建一个通道用于控制定时器
	tickerControlChan := make(chan time.Duration)

	// 首先启动配置文件监听
	go config.WatchConfigFile(configPath, globalLogger, func(newConfig *config.Config) {
		// 更新配置
		oldInterval := cfg.Interval

		// 保存旧的日志配置
		oldLogConfig := cfg.Log

		// 更新整个配置
		cfg = newConfig
		globalLogger.Log("Configuration reloaded successfully", "info")

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
			globalLogger.UpdateConfig(logConfig)
			globalLogger.Log("Logger configuration updated", "info")
		}

		// 如果间隔时间发生变化，通知定时器更新
		if oldInterval != cfg.Interval {
			globalLogger.Log(fmt.Sprintf("Interval changed from %d to %d seconds", oldInterval, cfg.Interval), "info")
			tickerControlChan <- time.Duration(cfg.Interval) * time.Second
		}
	})

	// 然后初始化 pinger 和 checker
	pinger := checker.NewPinger()
	chk := checker.NewChecker(cfg, pinger, globalLogger, notifierInstance)

	// 执行初始 ping 检查
	globalLogger.Log("Performing initial ping check", "info")
	chk.PingHosts()

	// 最后启动定期 ping 检查，并传入控制通道
	stopChan := startPeriodicPingChecks(chk, cfg, globalLogger, tickerControlChan)

	// 等待退出信号
	waitForExitSignal(globalLogger)
	close(stopChan)
}

func waitForExitSignal(logger *logger.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	logger.Log(fmt.Sprintf("Received exit signal (%s), shutting down...", sig), "info")
}

func changeToProjectRoot() error {
	projectRoot := filepath.Dir(filepath.Dir(os.Args[0]))
	err := os.Chdir(projectRoot)
	if err != nil {
		return fmt.Errorf("error changing directory to project root: %v", err)
	}

	cwd, err := os.Getwd()
	if (err != nil) {
		return fmt.Errorf("error getting current working directory: %v", err)
	}
	fmt.Printf("Current working directory: %s\n", cwd)
	return nil
}

func startPeriodicPingChecks(chk *checker.Checker, cfg *config.Config, logger *logger.Logger, tickerControlChan chan time.Duration) chan struct{} {
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)

	logger.Log("Starting periodic ping checks", "info") // 将日志级别从 debug 改为 info

	// 创建一个停止通道，用于通知程序退出
	stopChan := make(chan struct{})

	// 处理 ping 检查的 goroutine
	go func() {
			defer ticker.Stop() // 移到 goroutine 内部，这样只有在 goroutine 结束时才会停止 ticker
			for {
					select {
					case <-ticker.C:
							logger.Log("Executing scheduled ping check", "debug")
							chk.PingHosts()
					case newInterval := <-tickerControlChan:
							// 收到新的间隔时间，更新定时器
							ticker.Stop()
							ticker = time.NewTicker(newInterval)
							logger.Log(fmt.Sprintf("Updated ping check interval to %v", newInterval), "info")
					case <-stopChan:
							return
					}
			}
	}()

	return stopChan
}

func getExecutablePath() string {
	cwd, err := os.Getwd()
	if err != nil {
		globalLogger.Fatal(fmt.Sprintf("Error getting current working directory: %v", err), "fatal")
	}
	return filepath.Join(cwd, "bin", "easy-check.exe")
}

func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}
