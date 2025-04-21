package types

import (
	"easy-check/internal/config"
	"easy-check/internal/db"
)

// AggregatorHandle 定义了处理告警和恢复通知的接口
type AggregatorHandle interface {
	// 处理多个告警
	ProcessAlerts(alerts []*db.AlertStatus, dbManager *db.AlertStatusManager) error

	// 发送恢复通知
	SendRecoveryNotification(host config.Host, recoveryInfo *RecoveryInfo) error
}
