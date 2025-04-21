package aggregator

import (
	"easy-check/internal/config"
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

func (n *NoAggregator) ProcessAlerts(alerts []*db.AlertStatus, dbManager *db.AlertStatusManager) error {
	for _, alert := range alerts {
		host := config.Host{
			Host:        alert.Host,
			Description: alert.Description,
		}

		n.logger.Log(fmt.Sprintf("Sending alert for host: %s", alert.Host), "debug")
		err := n.notifier.SendNotification(host)
		if err != nil {
			n.logger.Log(fmt.Sprintf("Failed to send alert for host %s: %v", alert.Host, err), "error")
			continue
		}

		err = dbManager.UpdateSentStatus(alert.Host, true)
		if err != nil {
			n.logger.Log(fmt.Sprintf("Failed to update status for host %s: %v", alert.Host, err), "error")
		}
	}
	return nil
}

func (n *NoAggregator) SendRecoveryNotification(host config.Host, recoveryInfo *types.RecoveryInfo) error {
	return n.notifier.SendRecoveryNotification(host, recoveryInfo)
}
