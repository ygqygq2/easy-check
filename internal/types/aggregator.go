package types

import (
	"easy-check/internal/db"
)

// AggregatorHandle 定义了处理告警和恢复通知的接口
type AggregatorHandle interface {
	// 处理多个告警
	ProcessAlerts(alerts []*db.AlertStatus, dbManager *db.AlertStatusManager) error

	// 处理多个恢复通知
	ProcessRecoveries(recoveries []*db.AlertStatus, dbManager *db.AlertStatusManager) error
}
