package aggregator

import (
	"easy-check/internal/db"
	"easy-check/internal/logger"
	"easy-check/internal/types"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Aggregator 实现了聚合告警的逻辑
type Aggregator struct {
	alertLineTemplate    string
	recoveryLineTemplate string
	notifier             types.Notifier
	logger               *logger.Logger
	window               time.Duration
	alerts               []*db.AlertStatus
	mu                   sync.Mutex
}

// 确保 Aggregator 实现了 AggregatorHandle 接口
var _ types.AggregatorHandle = (*Aggregator)(nil)

func NewAggregator(alertLineTemplate string, recoveryLineTemplate string, notifier types.Notifier, logger *logger.Logger, window time.Duration) *Aggregator {
	return &Aggregator{
		alertLineTemplate:    alertLineTemplate,
		recoveryLineTemplate: recoveryLineTemplate,
		notifier:             notifier,
		logger:               logger,
		window:               window,
		alerts:               make([]*db.AlertStatus, 0),
	}
}

func (a *Aggregator) ProcessAlerts(alerts []*db.AlertStatus, dbManager *db.AlertStatusManager) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(alerts) == 0 {
		return nil
	}

	// 格式化告警内容
	content, err := a.formatAlerts(alerts, false) // false 表示这是告警
	if err != nil {
		a.logger.Log(fmt.Sprintf("Failed to format alerts: %v", err), "error")
		return err
	}

	// 发送聚合通知
	a.logger.Log(fmt.Sprintf("Sending aggregated alerts:\n%s", content), "debug")
	if err := a.notifier.SendAggregatedNotification(alerts, false); err != nil {
		return err
	}

	// 更新告警状态
	if err := a.updateAlertStatuses(alerts, dbManager); err != nil {
		a.logger.Log(fmt.Sprintf("Failed to update alert statuses: %v", err), "error")
		return err
	}

	return nil
}

// 格式化告警内容
func (a *Aggregator) formatAlerts(alerts []*db.AlertStatus, isRecovery bool) (string, error) {
	var template string
	if isRecovery {
		template = a.recoveryLineTemplate
	} else {
		template = a.alertLineTemplate
	}

	alertList := make([]string, len(alerts))
	for i, alert := range alerts {
		line := strings.ReplaceAll(template, "{{.FailTime}}", alert.FailTime)
		line = strings.ReplaceAll(line, "{{.RecoveryTime}}", alert.RecoveryTime)
		line = strings.ReplaceAll(line, "{{.Host}}", alert.Host)
		line = strings.ReplaceAll(line, "{{.Description}}", alert.Description)
		alertList[i] = line
	}
	return strings.Join(alertList, "\n"), nil
}

// 更新告警状态
func (a *Aggregator) updateAlertStatuses(alerts []*db.AlertStatus, dbManager *db.AlertStatusManager) error {
	for _, alert := range alerts {
		if err := dbManager.UpdateSentStatus(alert.Host, true); err != nil {
			a.logger.Log(fmt.Sprintf("Failed to update alert status for host %s: %v", alert.Host, err), "error")
			return err
		}
	}
	return nil
}

func (a *Aggregator) ProcessRecoveries(recoveries []*db.AlertStatus, dbManager *db.AlertStatusManager) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(recoveries) == 0 {
		return nil
	}

	// 格式化恢复通知内容
	content, err := a.formatAlerts(recoveries, true) // true 表示这是恢复
	if err != nil {
		a.logger.Log(fmt.Sprintf("Failed to format recoveries: %v", err), "error")
		return err
	}

	// 发送恢复通知
	a.logger.Log(fmt.Sprintf("Sending aggregated recoveries:\n%s", content), "info")
	if err := a.notifier.SendAggregatedNotification(recoveries, true); err != nil {
		a.logger.Log(fmt.Sprintf("Failed to send aggregated recoveries: %v", err), "error")
		return err
	}

	// 更新恢复状态
	if err := a.updateAlertStatuses(recoveries, dbManager); err != nil {
		a.logger.Log(fmt.Sprintf("Failed to update recovery statuses: %v", err), "error")
		return err
	}

	return nil
}
