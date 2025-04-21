package notifier

import (
	"easy-check/internal/db"
	"easy-check/internal/logger"
	"easy-check/internal/queue"
	"easy-check/internal/types"
	"fmt"
)

// MultiNotifierWrapper 是对 types.MultiNotifier 的本地包装器
type MultiNotifierWrapper struct {
	*types.MultiNotifier
}

// NotifierManagerWrapper 包装了 NotifierManager
type NotifierManagerWrapper struct {
	Queue         *queue.Queue
	MultiNotifier *MultiNotifierWrapper
	Logger        *logger.Logger
}

// NewNotifierManager 创建一个新的 NotifierManager
func NewNotifierManager(queue *queue.Queue, multiNotifier *types.MultiNotifier, logger *logger.Logger) *types.NotifierManager {
	return &types.NotifierManager{
		Queue:         queue,
		MultiNotifier: multiNotifier,
		Logger:        logger,
	}
}

// NewMultiNotifier 创建一个新的 MultiNotifier
func NewMultiNotifier(allNotifiers []types.Notifier, logger *logger.Logger) *types.MultiNotifier {
	enabledNotifiers := []types.Notifier{}
	for _, notifier := range allNotifiers {
		if notifier != nil { // 确保通知器有效
			enabledNotifiers = append(enabledNotifiers, notifier)
		}
	}
	if len(enabledNotifiers) == 0 {
		logger.Log("No enabled notifiers found", "warn")
	}
	return &types.MultiNotifier{Notifiers: enabledNotifiers, Logger: logger}
}

// SendNotification 实现 Notifier 接口，向所有启用的通知器发送消息d
func (m *MultiNotifierWrapper) SendNotification(alert *db.AlertStatus, isRecovery bool) error {
	var errs []error
	for _, notifier := range m.Notifiers {
		if err := notifier.SendNotification(alert, isRecovery); err != nil {
			m.Logger.Log(fmt.Sprintf("Error sending notification: %v", err), "error")
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to send notification to some notifiers: %v", errs)
	}
	return nil
}

func (m *MultiNotifierWrapper) SendAggregatedNotification(alerts []*db.AlertStatus, isRecovery bool) error {
	// 初始化一个错误切片，用于存储发送通知时发生的错误
	var errs []error
	// 遍历所有注册的通知器
	for _, notifier := range m.Notifiers {
		if err := notifier.SendAggregatedNotification(alerts, isRecovery); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to send aggregated notification to some notifiers: %v", errs)
	}
	return nil
}

// Close 实现 Notifier 接口，关闭所有启用的通知器
func (m *MultiNotifierWrapper) Close() error {
	var errs []error
	for _, notifier := range m.Notifiers {
		if err := notifier.Close(); err != nil {
			m.Logger.Log(fmt.Sprintf("Error closing notifier: %v", err), "error")
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to close some notifiers: %v", errs)
	}
	return nil
}
