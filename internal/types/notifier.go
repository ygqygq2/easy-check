package types

import (
	"easy-check/internal/db"
	"easy-check/internal/logger"
	"easy-check/internal/queue"
)

// Notifier 接口定义了通知器的基本行为
type Notifier interface {
	// 发送单个主机的告警通知
	SendNotification(alert *db.AlertStatus) error
	// 发送单个恢复通知
	SendRecoveryNotification(alert *db.AlertStatus) error
	// 发送聚合告警
	SendAggregatedNotification(alerts []*db.AlertStatus) error
	// 关闭通知器
	Close() error
}

// MultiNotifier 将消息发送到多个启用的通知器
type MultiNotifier struct {
	Notifiers []Notifier
	Logger    *logger.Logger
}

// NotifierManager 管理通知器的队列和发送逻辑
type NotifierManager struct {
	Queue         *queue.Queue
	MultiNotifier *MultiNotifier
	Logger        *logger.Logger
}
