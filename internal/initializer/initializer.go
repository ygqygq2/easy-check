package initializer

import (
	"easy-check/internal/checker"
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"easy-check/internal/notifier"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AppContext 包含应用程序的所有依赖
type AppContext struct {
	Config   *config.Config
	Logger   *logger.Logger
	Pinger   checker.Pinger
	Notifier notifier.Notifier
}

// Initialize 初始化应用程序上下文
func Initialize() (*AppContext, error) {
	// 切换到项目根目录
	if err := changeToProjectRoot(); err != nil {
		return nil, fmt.Errorf("failed to change to project root: %w", err)
	}

	// 初始化日志
	defaultLogger := logger.NewDefaultLogger()
	// 不再defer关闭defaultLogger，移交给initLogger做统一关闭

	// 加载配置
	cfg, err := loadConfig(defaultLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// 初始化日志器，内部关闭defaultLogger
	appLogger := initLogger(cfg, defaultLogger)

	// 注册通知器
	RegisterNotifiers(appLogger)

	// 创建通知器实例
	baseNotifier, err := createNotifier(cfg, appLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create notifier: %w", err)
	}

	// 初始化聚合告警逻辑
	finalNotifier := initializeAlertAggregator(cfg, baseNotifier, appLogger)

	// 创建 AppContext
	appContext := &AppContext{
		Config:   cfg,
		Logger:   appLogger,
		Pinger:   checker.NewPinger(),
		Notifier: finalNotifier,
	}

	appLogger.Log("Application initialized successfully", "info")
	return appContext, nil
}

// changeToProjectRoot 切换到项目根目录
func changeToProjectRoot() error {
	// 首先尝试使用当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %w", err)
	}

	// 检查当前目录下是否存在configs目录
	configPath := filepath.Join(cwd, "configs")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 尝试上一级目录
		parentDir := filepath.Dir(cwd)
		configPath = filepath.Join(parentDir, "configs")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// 尝试使用二进制文件所在目录
			execPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("error getting executable path: %w", err)
			}
			execDir := filepath.Dir(execPath)
			projectRoot := filepath.Dir(execDir)
			if err := os.Chdir(projectRoot); err != nil {
				return fmt.Errorf("error changing to project root: %w", err)
			}
		} else {
			if err := os.Chdir(parentDir); err != nil {
				return fmt.Errorf("error changing to parent directory: %w", err)
			}
		}
	}

	// 建议使用日志输出当前工作目录（无法获取日志器时可暂用fmt.Printf）
	fmt.Printf("Current working directory: %s\n", os.Getenv("PWD"))
	return nil
}

// loadConfig 加载配置文件
func loadConfig(logger *logger.Logger) (*config.Config, error) {
	configPath := filepath.Join("configs", "config.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to load configuration: %v", err), "error")
		return nil, err
	}
	logger.Log("Configuration loaded successfully", "info")
	return cfg, nil
}

// initLogger 初始化日志器
func initLogger(cfg *config.Config, defaultLogger *logger.Logger) *logger.Logger {
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
	// 关闭默认日志器，避免重复释放
	defaultLogger.Close()
	appLogger.Log("Logger initialized successfully", "info")
	return appLogger
}

// RegisterNotifiers 注册所有支持的通知器
func RegisterNotifiers(logger *logger.Logger) {
	notifier.RegisterNotifier("feishu", notifier.NewFeishuNotifier)
	// 可以在这里添加其他通知器的注册
	logger.Log("All notifiers registered successfully", "info")
}

// createNotifier 创建通知器实例
func createNotifier(cfg *config.Config, logger *logger.Logger) (notifier.Notifier, error) {
	// 从配置中创建所有启用的通知器
	notifiers := notifier.CreateNotifiers(cfg, logger)
	if len(notifiers) == 0 {
		logger.Log("No enabled notifiers found in configuration", "warn")
		return &notifier.NoopNotifier{}, nil // 返回空操作通知器
	}
	return notifier.NewMultiNotifier(notifiers, logger), nil
}

func initializeAlertAggregator(cfg *config.Config, baseNotifier notifier.Notifier, logger *logger.Logger) notifier.Notifier {
	if cfg.Alert.AggregateAlerts {
		logger.Log(fmt.Sprintf("Alert aggregation enabled with %d second window", cfg.Alert.AggregateWindow), "info")
		window := time.Duration(cfg.Alert.AggregateWindow) * time.Second

		// 创建 AlertAggregator，但不作为返回值
		aggregator := notifier.NewAlertAggregator(window, []notifier.Notifier{baseNotifier}, logger, cfg)

		// 启动一个 goroutine，监听告警队列
		go func() {
			for {
				// 模拟告警添加逻辑
				time.Sleep(10 * time.Second)
				aggregator.AddAlert("example-host", "example-description")
			}
		}()

		// 返回基础通知器
		return baseNotifier
	}

	return baseNotifier
}
