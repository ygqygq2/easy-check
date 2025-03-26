package initializer

import (
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"easy-check/internal/notifier"
	"easy-check/internal/signal"
	"fmt"
	"os"
	"path/filepath"
)

var GlobalLogger *logger.Logger

func Initialize() (*config.Config, notifier.Notifier, error) {
	// 切换到项目根目录
	err := changeToProjectRoot()
	if err != nil {
		return nil, nil, fmt.Errorf("error changing directory to project root: %v", err)
	}

	// 尽早初始化一个默认日志记录器（控制台输出）
	GlobalLogger = logger.NewDefaultLogger()
	defer GlobalLogger.Close()

	// 加载配置
	configPath := filepath.Join("configs", "config.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error loading configuration: %v", err)
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
	GlobalLogger.UpdateConfig(logConfig)

	GlobalLogger.Log("Starting easy-check...", "info")
	GlobalLogger.Log("Configuration loaded successfully", "info")

	// 初始化 Notifier
	var notifiers []notifier.Notifier
	if cfg.Alert.Feishu.Enable {
		feishuNotifier, err := notifier.NewFeishuNotifier(&cfg.Alert.Feishu, GlobalLogger)
		if err != nil {
			GlobalLogger.Log(fmt.Sprintf("Failed to initialize FeishuNotifier: %v", err), "error")
			return nil, nil, err
		}
		notifiers = append(notifiers, feishuNotifier)
	}

	// 如果开启了告警聚合
	var notifierInstance notifier.Notifier
	if cfg.Alert.AggregateAlerts {
		GlobalLogger.Log(fmt.Sprintf("Alert aggregation enabled with %d second window", cfg.Alert.AggregateWindow), "info")
		aggregatingNotifier := notifier.NewAggregatingNotifier(notifiers, cfg, GlobalLogger)
		notifierInstance = aggregatingNotifier

		// 将关闭逻辑移到程序退出时
		go func() {
			sigChan := signal.RegisterExitListener()
			<-sigChan
			aggregatingNotifier.Close()
		}()
	} else {
		if len(notifiers) > 0 {
			notifierInstance = notifiers[0]
		} else {
			GlobalLogger.Log("No valid notifier configuration found", "warn")
		}
	}

	for i := range cfg.Hosts {
		if cfg.Hosts[i].FailAlert == false {
			continue
		}
		cfg.Hosts[i].FailAlert = true // 默认值为 true
	}

	return cfg, notifierInstance, nil
}

func changeToProjectRoot() error {
	projectRoot := filepath.Dir(filepath.Dir(os.Args[0]))
	err := os.Chdir(projectRoot)
	if err != nil {
		return fmt.Errorf("error changing directory to project root: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %v", err)
	}
	fmt.Printf("Current working directory: %s\n", cwd)
	return nil
}
