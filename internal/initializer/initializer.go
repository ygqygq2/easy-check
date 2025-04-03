package initializer

import (
	"easy-check/internal/checker"
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"easy-check/internal/notifier"
	"fmt"
	"os"
	"path/filepath"
)

// AppContext 包含应用程序的所有依赖
type AppContext struct {
	Config   *config.Config
	Logger   *logger.Logger
	Pinger   checker.Pinger
	Notifier notifier.Notifier
}

func Initialize() (*AppContext, error) {
	// 切换到项目根目录
	err := changeToProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("error changing directory to project root: %v", err)
	}

	// 尽早初始化一个默认日志记录器（控制台输出）
	defaultLogger := logger.NewDefaultLogger()
	defer defaultLogger.Close()

	// 加载配置
	configPath := filepath.Join("configs", "config.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("error loading configuration: %v", err)
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

	appLogger := logger.NewLogger(logConfig)

	appLogger.Log("Starting easy-check...", "info")
	appLogger.Log("Configuration loaded successfully", "info")

	// 从配置创建所有通知器
	notifiers := notifier.CreateNotifiers(cfg, appLogger)

	// 创建多通知器
	var notifierInstance notifier.Notifier
	if cfg.Alert.AggregateAlerts {
		appLogger.Log(fmt.Sprintf("Alert aggregation enabled with %d second window", cfg.Alert.AggregateWindow), "info")
		notifierInstance = notifier.NewAlertAggregator(notifiers, cfg, appLogger)
	} else {
		notifierInstance = notifier.NewMultiNotifier(notifiers, appLogger)
	}

	// 创建 AppContext
	appContext := &AppContext{
		Config:   cfg,
		Logger:   appLogger,
		Pinger:   checker.NewPinger(),
		Notifier: notifierInstance,
	}

	return appContext, nil
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
