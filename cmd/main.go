package main

import (
	"easy-check/internal/checker"
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"easy-check/internal/notifier"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	// 切换到项目根目录
	changeToProjectRoot()

	// 加载配置
	configPath := filepath.Join("configs", "config.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// 初始化日志记录器
	logConfig := logger.Config{
		File:         cfg.Log.File,
		MaxSize:      cfg.Log.MaxSize,
		MaxAge:       cfg.Log.MaxAge,
		MaxBackups:   cfg.Log.MaxBackups,
		Compress:     cfg.Log.Compress,
		ConsoleLevel: cfg.Log.ConsoleLevel,
		FileLevel:    cfg.Log.FileLevel,
	}
	logger := logger.NewLogger(logConfig)
	defer logger.Close()

	logger.Log("Starting easy-check...")
	logger.Log("Configuration loaded successfully")

	// 初始化 Notifier
	var notifierInstance notifier.Notifier
	if cfg.Alert.Feishu.Enable && cfg.Alert.Feishu.MsgType == "text" {
		feishuNotifier, err := notifier.NewFeishuNotifier(&cfg.Alert.Feishu)
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to initialize FeishuNotifier: %v", err))
			return
		}
		notifierInstance = feishuNotifier
	} else {
		logger.Log("No valid notifier configuration found")
	}

	// 首先启动配置文件监听
	go config.WatchConfigFile(configPath, logger, func(newConfig *config.Config) {
		// 更新配置
		cfg = newConfig
		logger.Log("Configuration reloaded successfully")
	})

	// 然后初始化 pinger 和 checker
	pinger := checker.NewPinger()
	chk := checker.NewChecker(cfg, pinger, logger, notifierInstance)

	// 执行初始 ping 检查
	logger.Log("Performing initial ping check")
	chk.PingHosts()

	// 最后启动定期 ping 检查
	startPeriodicPingChecks(chk, cfg, logger) // 注意：传递整个cfg对象

	// 等待退出信号
	waitForExitSignal()
}

func waitForExitSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("Received exit signal, shutting down...")
}

func changeToProjectRoot() {
	projectRoot := filepath.Dir(filepath.Dir(os.Args[0]))
	err := os.Chdir(projectRoot)
	if err != nil {
		log.Fatalf("Error changing directory to project root: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}
	log.Printf("Current working directory: %s\n", cwd)
}

func startPeriodicPingChecks(chk *checker.Checker, cfg *config.Config, logger *logger.Logger) {
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()

	logger.Log("Starting periodic ping checks")

	// 创建一个停止通道，用于通知程序退出
	stopChan := make(chan struct{})

	// 处理 ping 检查的 goroutine
	go func() {
		for {
			select {
			case <-ticker.C:
				chk.PingHosts()
			case <-stopChan:
				return
			}
		}
	}()

	// 不要在这里阻塞，让main函数继续执行
}

func getExecutablePath() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}
	return filepath.Join(cwd, "bin", "easy-check.exe")
}

func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}
