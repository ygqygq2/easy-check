//go:build wireinject
// +build wireinject

package initializer

import (
	"easy-check/internal/checker"
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"easy-check/internal/notifier"

	"github.com/google/wire"
)

// InitializeApp 使用 wire 构建应用程序的依赖关系
func InitializeApp() (*AppContext, error) {
	wire.Build(
		config.LoadConfig,
		logger.NewLogger,
		checker.NewPinger,
		notifier.NewAggregatingNotifier,
		NewAppContext,
	)
	return &AppContext{}, nil
}
