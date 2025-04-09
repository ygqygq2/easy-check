// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package wire

import (
	"easy-check/internal/checker"
	"easy-check/internal/config"
	"easy-check/internal/db"
	"easy-check/internal/logger"
	"easy-check/internal/notifier"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/wire"
)

// Injectors from wire.go:

// InitializeApp 是 Wire 生成的初始化函数
func InitializeApp(configPath string) (*AppContext, error) {
	config, err := provideConfig(configPath)
	if err != nil {
		return nil, err
	}
	logger, err := provideLogger(config)
	if err != nil {
		return nil, err
	}
	badgerDB, err := provideBadgerDB(config, logger)
	if err != nil {
		return nil, err
	}
	alertStatusManager, err := db.NewAlertStatusManager(badgerDB, logger)
	if err != nil {
		return nil, err
	}
	pinger := checker.NewPinger()
	notifier, err := provideNotifier(config, logger)
	if err != nil {
		return nil, err
	}
	checkerChecker := checker.NewChecker(config, pinger, logger, notifier, alertStatusManager)
	appContext := &AppContext{
		Config:         config,
		Logger:         logger,
		DB:             badgerDB,
		AlertStatusMgr: alertStatusManager,
		Checker:        checkerChecker,
		Notifier:       notifier,
	}
	return appContext, nil
}

// wire.go:

// 应用程序上下文结构
type AppContext struct {
	Config         *config.Config
	Logger         *logger.Logger
	DB             *badger.DB
	AlertStatusMgr *db.AlertStatusManager
	Checker        *checker.Checker
	Notifier       notifier.Notifier
}

// 定义提供者集
var loggerSet = wire.NewSet(
	provideLogger,
)

var dbSet = wire.NewSet(
	provideBadgerDB, db.NewAlertStatusManager,
)

var checkerSet = wire.NewSet(checker.NewChecker, checker.NewPinger)

var notifierSet = wire.NewSet(
	provideNotifier,
)

func provideDefaultLogger() *logger.Logger {
	return logger.NewDefaultLogger()
}

func provideConfig(configPath string) (*config.Config, error) {
	defaultLogger := logger.NewDefaultLogger()
	return config.LoadConfig(configPath, defaultLogger)
}

var configSet = wire.NewSet(
	provideConfig,
)

// provideLogger 创建日志记录器
func provideLogger(cfg *config.Config) (*logger.Logger, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	logConfig := logger.Config{
		File:         cfg.Log.File,
		MaxSize:      cfg.Log.MaxSize,
		MaxAge:       cfg.Log.MaxAge,
		MaxBackups:   cfg.Log.MaxBackups,
		Compress:     cfg.Log.Compress,
		ConsoleLevel: cfg.Log.ConsoleLevel,
		FileLevel:    cfg.Log.FileLevel,
	}
	return logger.NewLogger(logConfig), nil
}

// provideBadgerDB 创建数据库连接func provideBadgerDB(cfg *config.Config, log *logger.Logger) (*badger.DB, error) {
func provideBadgerDB(cfg *config.Config, log *logger.Logger) (*badger.DB, error) {
	opts := badger.DefaultOptions(cfg.Db.Path)
	db2, err := badger.Open(opts)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to open database: %v", err))
		return nil, err
	}
	return db2, nil
}

// provideNotifier 创建通知器
func provideNotifier(cfg *config.Config, log *logger.Logger) (notifier.Notifier, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	notifiers := notifier.CreateNotifiers(cfg, log)
	if len(notifiers) == 0 {
		log.Log("No enabled notifiers found in configuration", "warn")
		return &notifier.NoopNotifier{}, nil
	}

	return notifier.NewMultiNotifier(notifiers, log), nil
}
