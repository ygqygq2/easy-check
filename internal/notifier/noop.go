package notifier

import (
	"easy-check/internal/db"
)

// NoopNotifier 是一个不执行任何操作的通知器
type NoopNotifier struct{}

func (n *NoopNotifier) SendNotification(alert *db.AlertStatus) error {
	return nil
}

func (n *NoopNotifier) SendRecoveryNotification(alerts *db.AlertStatus) error {
	return nil
}

func (n *NoopNotifier) SendAggregatedNotification(alerts []*db.AlertStatus) error {
	return nil
}

func (n *NoopNotifier) Close() error {
	return nil
}
