package aggregator

import (
	"easy-check/internal/db"
	"easy-check/internal/logger"
	"easy-check/internal/types"
	"fmt"
)

// NoAggregator 实现了非聚合告警的逻辑
type NoAggregator struct {
	notifier types.Notifier // 使用接口类型
	logger   *logger.Logger
}

// 确保 NoAggregator 实现了 AggregatorHandle 接口
var _ types.AggregatorHandle = (*NoAggregator)(nil)

func NewNoAggregator(notifier types.Notifier, logger *logger.Logger) *NoAggregator {
	return &NoAggregator{
		notifier: notifier,
		logger:   logger,
	}
}

func (n *NoAggregator) ProcessNotifications(notifications []*db.AlertStatus, dbManager *db.AlertStatusManager, isRecovery bool) error {
	for _, notification := range notifications {
		action := "alert"
		if isRecovery {
			action = "recovery"
		}

		n.logger.Log(fmt.Sprintf("Sending %s for host: %s", action, notification.Host), "debug")
		err := n.notifier.SendNotification(notification, isRecovery)
		if err != nil {
			n.logger.Log(fmt.Sprintf("Failed to send %s for host %s: %v", action, notification.Host, err), "error")
			continue
		}

		err = dbManager.UpdateSentStatus(notification.Host, true)
		if err != nil {
			n.logger.Log(fmt.Sprintf("Failed to update status for host %s: %v", notification.Host, err), "error")
		}
	}
	return nil
}

// ProcessAlerts 调用通用方法处理告警
func (n *NoAggregator) ProcessAlerts(alerts []*db.AlertStatus, dbManager *db.AlertStatusManager) error {
	return n.ProcessNotifications(alerts, dbManager, false)
}

// ProcessRecoveries 调用通用方法处理恢复
func (n *NoAggregator) ProcessRecoveries(recoveries []*db.AlertStatus, dbManager *db.AlertStatusManager) error {
	return n.ProcessNotifications(recoveries, dbManager, true)
}
