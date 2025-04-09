package notifier

import (
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"fmt"
)

// NotifierCreator 通知器创建函数类型
type NotifierCreator func(options map[string]interface{}, logger *logger.Logger) (Notifier, error)

// 全局注册表
var notifierRegistry = make(map[string]NotifierCreator)

// RegisterNotifier 注册通知器创建函数
func RegisterNotifier(typeName string, creator NotifierCreator) {
	notifierRegistry[typeName] = creator
}

// CreateNotifiers 从配置创建所有通知器
func CreateNotifiers(cfg *config.Config, logger *logger.Logger) []Notifier {
	var notifiers []Notifier

	for _, notifierCfg := range cfg.Alert.Notifiers {
		if !notifierCfg.Enable {
			logger.Log(fmt.Sprintf("Notifier %s is disabled, skipping", notifierCfg.Name), "debug")
			continue
		}

		creator, exists := notifierRegistry[notifierCfg.Type]
		if !exists {
			logger.Log(fmt.Sprintf("Unknown notifier type: %s", notifierCfg.Type), "error")
			continue
		}

		notifier, err := creator(notifierCfg.Options, logger)
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to initialize notifier %s: %v", notifierCfg.Name, err), "error")
			continue
		}

		notifiers = append(notifiers, notifier)
		logger.Log(fmt.Sprintf("Successfully initialized notifier %s", notifierCfg.Name), "debug")
	}

	return notifiers
}
