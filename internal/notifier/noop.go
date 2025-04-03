package notifier

import (
	"easy-check/internal/config"
)

// NoopNotifier 是一个不执行任何操作的通知器
type NoopNotifier struct{}

func (n *NoopNotifier) SendNotification(host config.Host) error {
	return nil
}

func (n *NoopNotifier) SendAggregatedNotification(alerts []*AlertItem) error {
	return nil
}

func (n *NoopNotifier) Close() error {
	return nil
}
