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
	fmt.Println("🚀 ~ file: consumer.go:40 ~ alerts:", alerts)
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
			return err
		}
	}

	return nil
}

func (c *Consumer) processRecoveryNotifications() {
	// 从数据库中获取未发送的恢复通知
	recoveries, err := c.db.GetAllUnsentRecoveries()
	if err != nil {
		c.logger.Log(fmt.Sprintf("Failed to get unsent recoveries: %v", err), "error")
		return
	}

	for _, recovery := range recoveries {
		c.logger.Log(fmt.Sprintf("Sending recovery notification for host: %s", recovery.Host), "debug")

		// 构造主机信息
		host := config.Host{
			Host:        recovery.Host,
			Description: recovery.Description,
		}

		// 将数据库中的时间字符串解析为time.Time
		failTime, err := time.Parse(time.RFC3339, recovery.Timestamp)
		if err != nil {
			c.logger.Log(fmt.Sprintf("Failed to parse timestamp for host %s: %v", recovery.Host, err), "error")
			failTime = time.Now() // 解析失败则使用当前时间
		}

		// 构造恢复信息
		recoveryInfo := &RecoveryInfo{
			FailTime:     failTime,
			RecoveryTime: time.Now(),
		}

		// 发送恢复通知，传入所有需要的信息
		err = c.Notifier.SendRecoveryNotification(host, recoveryInfo)
		if err != nil {
			c.logger.Log(fmt.Sprintf("Failed to send recovery notification: %v", err), "error")
			continue
		}

		// 标记为已发送
		recovery.Sent = true
		err = c.db.SetOrUpdateAlertStatus(recovery)
		if err != nil {
			c.logger.Log("Failed to update recovery status: "+err.Error(), "error")
		}
	}
}
