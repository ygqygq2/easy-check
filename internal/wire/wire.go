//go:build wireinject
// +build wireinject

package wire

import (
	"easy-check/internal/aggregator"
	"easy-check/internal/checker"
	"easy-check/internal/config"
	"easy-check/internal/db"
	"easy-check/internal/logger"
	"easy-check/internal/notifier"
	"easy-check/internal/types"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/wire"
)

// 应用程序上下文结构
type AppContext struct {
	Config           *config.Config
	Logger           *logger.Logger
	DB               *badger.DB
	AlertStatusMgr   *db.AlertStatusManager
	Pinger           checker.Pinger
	Notifier         types.Notifier
	AggregatorHandle types.AggregatorHandle
}

// 定义提供者集
var loggerSet = wire.NewSet(
	provideLogger,
)

var dbSet = wire.NewSet(
	provideBadgerDB,
	provideAlertStatusManager,
)

var checkerSet = wire.NewSet(
	checker.NewChecker,
	checker.NewPinger,
)

var notifierSet = wire.NewSet(
	provideNotifier,
)

var aggregatorSet = wire.NewSet(
	provideAggregator,
)

func provideDefaultLogger() *logger.Logger {
	return logger.NewDefaultLogger()
}

func provideConfig(configPath string) (*config.Config, error) {
	defaultLogger := logger.NewDefaultLogger() // 使用默认日志器
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

// provideBadgerDB 创建数据库连接
func provideBadgerDB(cfg *config.Config, log *logger.Logger) (*badger.DB, error) {
	opts := badger.DefaultOptions(cfg.Db.Path)
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to open database: %v", err))
		return nil, err
	}
	return db, nil
}

// provideAlertStatusManager 创建 AlertStatusManager
func provideAlertStatusManager(badgerDB *badger.DB, log *logger.Logger, cfg *config.Config) (*db.AlertStatusManager, error) {
	return db.NewAlertStatusManager(badgerDB, log, cfg.Db)
}

// provideNotifier 创建通知器
func provideNotifier(cfg *config.Config, log *logger.Logger) (types.Notifier, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	// 使用 notifier.CreateNotifiers 从配置中创建所有通知器
	notifiers := notifier.CreateNotifiers(cfg, log)
	if len(notifiers) == 0 {
		log.Log("No enabled notifiers found in configuration", "warn")
		return &notifier.NoopNotifier{}, nil // 返回一个空操作通知器
	}

	// 使用 MultiNotifierWrapper 包装 NewMultiNotifier
	return &notifier.MultiNotifierWrapper{
		MultiNotifier: notifier.NewMultiNotifier(notifiers, log),
	}, nil
}

// provideAggregator 创建聚合器
func provideAggregator(cfg *config.Config, log *logger.Logger, notifier types.Notifier) (types.AggregatorHandle, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	if cfg.Alert.AggregateAlerts {
		window := time.Duration(cfg.Alert.AggregateWindow) * time.Second
		return aggregator.NewAggregator(
			cfg.Alert.AggregateLineTemplate,
			notifier,
			log,
			window,
		), nil
	} else {
		return aggregator.NewNoAggregator(notifier, log), nil
	}
}

// InitializeApp 是 Wire 生成的初始化函数
func InitializeApp(configPath string) (*AppContext, error) {
	wire.Build(
		configSet,                         // 配置提供者集
		loggerSet,                         // 日志器提供者集
		dbSet,                             // 数据库提供者集
		checkerSet,                        // 检查器提供者集
		notifierSet,                       // 通知器提供者集
		aggregatorSet,                     // 聚合器提供者集
		wire.Struct(new(AppContext), "*"), // 构造 AppContext
	)
	return nil, nil
}
