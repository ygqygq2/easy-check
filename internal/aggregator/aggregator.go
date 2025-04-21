package aggregator

import (
	"easy-check/internal/config"
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
	lineTemplate string
	notifier     types.Notifier
	logger       *logger.Logger
	window       time.Duration
	alerts       []*db.AlertStatus
	mu           sync.Mutex
}

// 确保 Aggregator 实现了 AggregatorHandle 接口
var _ types.AggregatorHandle = (*Aggregator)(nil)

func NewAggregator(lineTemplate string, notifier types.Notifier, logger *logger.Logger, window time.Duration) *Aggregator {
	return &Aggregator{
		lineTemplate: lineTemplate,
		notifier:     notifier,
		logger:       logger,
		window:       window,
		alerts:       make([]*db.AlertStatus, 0),
	}
}

// parseTime 辅助函数，将时间字符串解析为 time.Time
func parseTime(timestamp string) time.Time {
	// 假设时间戳格式为 "2006-01-02 15:04:05"
	t, err := time.Parse("2006-01-02 15:04:05", timestamp)
	if err != nil {
		// 如果解析失败，返回当前时间
		return time.Now()
	}
	return t
}

func (a *Aggregator) ProcessAlerts(alerts []*db.AlertStatus, dbManager *db.AlertStatusManager) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(alerts) == 0 {
		return nil
	}

	// 格式化告警内容
	content, err := a.formatAlerts(alerts)
	if err != nil {
		a.logger.Log(fmt.Sprintf("Failed to format alerts: %v", err), "error")
		return err
	}

	// 将 AlertStatus 转换为 AlertItem
	alertItems := make([]*db.AlertItem, len(alerts))
	for i, alert := range alerts {
		alertItems[i] = &db.AlertItem{
			Host:        alert.Host,
			Description: alert.Description,
			Timestamp:   parseTime(alert.Timestamp), // 需要将时间字符串转换为 time.Time
		}
	}

	// 发送聚合通知
	a.logger.Log(fmt.Sprintf("Sending aggregated alerts:\n%s", content), "info")
	if err := a.notifier.SendAggregatedNotification(alertItems); err != nil {
		a.logger.Log(fmt.Sprintf("Failed to send aggregated alerts: %v", err), "error")
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
func (a *Aggregator) formatAlerts(alerts []*db.AlertStatus) (string, error) {
	alertList := make([]string, len(alerts))
	for i, alert := range alerts {
		line := strings.ReplaceAll(a.lineTemplate, "{{.FailTime}}", alert.Timestamp)
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

func (a *Aggregator) SendRecoveryNotification(host config.Host, recoveryInfo *types.RecoveryInfo) error {
	err := a.notifier.SendRecoveryNotification(host, recoveryInfo)
	if err != nil {
		a.logger.Log(fmt.Sprintf("Failed to send recovery notification for host %s: %v", host.Host, err), "error")
		return err
	}
	return nil
}
