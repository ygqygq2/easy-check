package notifier

import (
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"easy-check/internal/queue"
	"fmt"
	"time"
)

// Notifier 接口定义了通知器的基本行为
type Notifier interface {
	// 发送单个主机的告警通知
	SendNotification(host config.Host) error
	// 发送聚合告警
	SendAggregatedNotification(alerts []*AlertItem) error
	// 关闭通知器
	Close() error
}

// MultiNotifier 将消息发送到多个启用的通知器
type MultiNotifier struct {
	notifiers []Notifier
	logger    *logger.Logger
}

// NewMultiNotifier 创建一个新的 MultiNotifier
func NewMultiNotifier(allNotifiers []Notifier, logger *logger.Logger) *MultiNotifier {
	enabledNotifiers := []Notifier{}
	for _, notifier := range allNotifiers {
		if notifier != nil { // 确保通知器有效
			enabledNotifiers = append(enabledNotifiers, notifier)
		}
	}
	if len(enabledNotifiers) == 0 {
		logger.Log("No enabled notifiers found", "warn")
	}
	return &MultiNotifier{notifiers: enabledNotifiers, logger: logger}
}

// SendNotification 实现 Notifier 接口，向所有启用的通知器发送消息
func (m *MultiNotifier) SendNotification(host config.Host) error {
	var errs []error
	for _, notifier := range m.notifiers {
		if err := notifier.SendNotification(host); err != nil {
			m.logger.Log(fmt.Sprintf("Error sending notification: %v", err), "error")
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to send notification to some notifiers: %v", errs)
	}
	return nil
}

// SendAggregatedNotification 实现 Notifier 接口，向所有启用的通知器发送聚合消息
func (m *MultiNotifier) SendAggregatedNotification(alerts []*AlertItem) error {
	var errs []error
	for _, notifier := range m.notifiers {
		if err := notifier.SendAggregatedNotification(alerts); err != nil {
			m.logger.Log(fmt.Sprintf("Error sending aggregated notification: %v", err), "error")
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to send aggregated notification to some notifiers: %v", errs)
	}
	return nil
}

// Close 实现 Notifier 接口，关闭所有启用的通知器
func (m *MultiNotifier) Close() error {
	var errs []error
	for _, notifier := range m.notifiers {
		if err := notifier.Close(); err != nil {
			m.logger.Log(fmt.Sprintf("Error closing notifier: %v", err), "error")
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to close some notifiers: %v", errs)
	}
	return nil
}

// NotifierManager 管理通知器的队列和发送逻辑
type NotifierManager struct {
	queue         *queue.Queue
	multiNotifier *MultiNotifier
	logger        *logger.Logger
}

// NewNotifierManager 创建一个新的 NotifierManager
func NewNotifierManager(queue *queue.Queue, multiNotifier *MultiNotifier, logger *logger.Logger) *NotifierManager {
	return &NotifierManager{
		queue:         queue,
		multiNotifier: multiNotifier,
		logger:        logger,
	}
}

// Start 启动通知器管理器，处理队列中的事件
func (n *NotifierManager) Start() {
	for {
		event, ok := n.queue.Pop()
		if !ok {
			// 如果队列为空，等待一段时间再尝试
			time.Sleep(100 * time.Millisecond)
			continue
		}

		switch event.Type {
		case "ALERT":
			n.logger.Log("Processing alert event", "info")
			host := config.Host{
				Host:        event.Host,
				Description: event.Description,
			}
			if err := n.multiNotifier.SendNotification(host); err != nil {
				n.logger.Log(fmt.Sprintf("Failed to send alert notification: %v", err), "error")
			}
		case "RECOVERY":
			n.logger.Log("Processing recovery event", "info")
			host := config.Host{
				Host:        event.Host,
				Description: event.Description,
			}
			if err := n.multiNotifier.SendNotification(host); err != nil {
				n.logger.Log(fmt.Sprintf("Failed to send recovery notification: %v", err), "error")
			}
		default:
			n.logger.Log(fmt.Sprintf("Unknown event type: %s", event.Type), "warn")
		}
	}
}
