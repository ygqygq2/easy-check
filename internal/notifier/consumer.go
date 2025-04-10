package notifier

import (
	"easy-check/internal/config"
	"easy-check/internal/db"
	"easy-check/internal/logger"
	"fmt"

	"time"
)

type Consumer struct {
	db       *db.AlertStatusManager
	logger   *logger.Logger
	interval time.Duration
	Notifier Notifier
}

func NewConsumer(db *db.AlertStatusManager, logger *logger.Logger, notifier Notifier, interval time.Duration) *Consumer {
	return &Consumer{
		db:       db,
		logger:   logger,
		Notifier: notifier,
		interval: interval,
	}
}

func (c *Consumer) Start() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for range ticker.C {
		c.processAlerts()
	}
}

func (c *Consumer) processAlerts() {
	// 从数据库中获取未发送的告警
	alerts, err := c.db.GetAllUnsentAlerts()
	if err != nil {
		c.logger.Log("Failed to fetch unsent alerts: "+err.Error(), "error")
		return
	}

	for _, alert := range alerts {
		// 发送告警
		err := c.sendAlert(alert)
		if err != nil {
			c.logger.Log("Failed to send alert: "+err.Error(), "error")
			continue
		}

		// 更新告警状态为已发送
		alert.Sent = true
		err = c.db.SetOrUpdateAlertStatus(alert)
		if err != nil {
			c.logger.Log("Failed to update alert status: "+err.Error(), "error")
		}
	}
}

func (c *Consumer) sendAlert(alert db.AlertStatus) error {
	// 实现告警发送逻辑
	c.logger.Log("Sending alert for host: "+alert.Host, "info")
	// 构造 config.Host 类型的实例
	host := config.Host{
		Host:        alert.Host,
		Description: alert.Description,
	}

	// 发送通知
	if c.Notifier != nil {
		err := c.Notifier.SendNotification(host)
		if err != nil {
			c.logger.Log(fmt.Sprintf("Failed to send notification: %v", err), "error")
		}
	}
	return nil
}
