package notifier

import (
	"easy-check/internal/config"
	"easy-check/internal/db"
	"easy-check/internal/types"
)

// NoopNotifier 是一个不执行任何操作的通知器
type NoopNotifier struct{}

func (n *NoopNotifier) SendNotification(host config.Host) error {
	return nil
}

func (n *NoopNotifier) SendAggregatedNotification(alerts []*db.AlertItem) error {
	return nil
}

func (n *NoopNotifier) SendRecoveryNotification(host config.Host, recoveryInfo *types.RecoveryInfo) error {
	return nil
}

func (n *NoopNotifier) Close() error {
	return nil
}
