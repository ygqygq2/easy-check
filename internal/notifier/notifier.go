package notifier

import (
	"easy-check/internal/config"
	"easy-check/internal/db"
	"easy-check/internal/logger"
	"easy-check/internal/queue"
	"easy-check/internal/types"
	"fmt"
	"time"
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

// SendNotification 实现 Notifier 接口，向所有启用的通知器发送消息
func (m *MultiNotifierWrapper) SendNotification(host config.Host) error {
	var errs []error
	for _, notifier := range m.Notifiers {
		if err := notifier.SendNotification(host); err != nil {
			m.Logger.Log(fmt.Sprintf("Error sending notification: %v", err), "error")
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to send notification to some notifiers: %v", errs)
	}
	return nil
}

// SendAggregatedNotification 实现 Notifier 接口，向所有启用的通知器发送聚合消息
// SendAggregatedNotification 发送聚合通知
// 该方法接收一个警报项的切片，并尝试通过所有注册的通知器发送聚合通知
// 如果所有通知器都成功发送通知，则返回nil；否则返回一个错误列表
func (m *MultiNotifierWrapper) SendAggregatedNotification(alerts []*db.AlertItem) error {
	// 初始化一个错误切片，用于存储发送通知时发生的错误
	var errs []error
	// 遍历所有注册的通知器
	for _, notifier := range m.Notifiers {
		if err := notifier.SendAggregatedNotification(alerts); err != nil {
			m.Logger.Log(fmt.Sprintf("Error sending aggregated notification: %v", err), "error")
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to send aggregated notification to some notifiers: %v", errs)
	}
	return nil
}

// SendRecoveryNotification 实现 Notifier 接口，向所有启用的通知器发送恢复通知
func (m *MultiNotifierWrapper) SendRecoveryNotification(host config.Host, recoveryInfo *types.RecoveryInfo) error {
	var errs []error
	for _, notifier := range m.Notifiers {
		if err := notifier.SendRecoveryNotification(host, recoveryInfo); err != nil {
			m.Logger.Log(fmt.Sprintf("Error sending recovery notification: %v", err), "error")
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to send recovery notification to some notifiers: %v", errs)
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

// Start 启动通知器管理器，处理队列中的事件
func (n *NotifierManagerWrapper) Start() {
	for {
		event, ok := n.Queue.Pop()
		if !ok {
			// 如果队列为空，等待一段时间再尝试
			time.Sleep(100 * time.Millisecond)
			continue
		}

		switch event.Type {
		case "ALERT":
			n.Logger.Log("Processing alert event", "info")
			host := config.Host{
				Host:        event.Host,
				Description: event.Description,
			}
			if err := n.MultiNotifier.SendNotification(host); err != nil {
				n.Logger.Log(fmt.Sprintf("Failed to send alert notification: %v", err), "error")
			}
		case "RECOVERY":
			n.Logger.Log("Processing recovery event", "info")
			host := config.Host{
				Host:        event.Host,
				Description: event.Description,
			}
			if err := n.MultiNotifier.SendNotification(host); err != nil {
				n.Logger.Log(fmt.Sprintf("Failed to send recovery notification: %v", err), "error")
			}
		default:
			n.Logger.Log(fmt.Sprintf("Unknown event type: %s", event.Type), "warn")
		}
	}
}
