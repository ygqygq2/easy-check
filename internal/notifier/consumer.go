package notifier

import (
	"easy-check/internal/config"
	"easy-check/internal/db"
	"easy-check/internal/logger"
	"easy-check/internal/types"
	"fmt"
	"time"
)

type Consumer struct {
	db       *db.AlertStatusManager
	logger   *logger.Logger
	interval time.Duration
	handler  types.AggregatorHandle
}

func NewConsumer(
	db *db.AlertStatusManager,
	logger *logger.Logger,
	interval time.Duration,
	handler types.AggregatorHandle,
) *Consumer {
	return &Consumer{
		db:       db,
		logger:   logger,
		interval: interval,
		handler:  handler,
	}
}

func (c *Consumer) Start() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for range ticker.C {
		c.processEvents(db.StatusAlert, "alerts")
		c.processEvents(db.StatusRecovery, "recoveries")
	}
}

// 通用事件处理方法
func (c *Consumer) processEvents(statusType db.StatusType, eventType string) {
	events, err := c.db.GetAllUnsentStatuses(statusType)
	if err != nil {
		c.logError(fmt.Sprintf("Failed to fetch unsent %s", eventType), err)
		return
	}

	if len(events) == 0 {
		c.logger.Log(fmt.Sprintf("No unsent %s to process", eventType), "debug")
		return
	}

	if eventType == "alerts" {
		if err := c.handler.ProcessAlerts(events, c.db); err != nil {
			c.logError("Failed to process alerts", err)
		}
	} else if eventType == "recoveries" {
		for _, recovery := range events {
			c.processSingleRecovery(recovery)
		}
	}
}

func (c *Consumer) processSingleRecovery(recovery *db.AlertStatus) {
	c.logger.Log(fmt.Sprintf("Processing recovery for host: %s", recovery.Host), "debug")

	host := config.Host{
		Host:        recovery.Host,
		Description: recovery.Description,
	}

	failTime, err := time.Parse(time.RFC3339, recovery.Timestamp)
	if err != nil {
		c.logError(fmt.Sprintf("Failed to parse timestamp for host %s", recovery.Host), err)
		failTime = time.Now()
	}

	recoveryInfo := &types.RecoveryInfo{
		FailTime:     failTime,
		RecoveryTime: time.Now(),
	}

	if err := c.handler.SendRecoveryNotification(host, recoveryInfo); err != nil {
		c.logError(fmt.Sprintf("Failed to send recovery notification for host %s", recovery.Host), err)
		return
	}

	if err := c.db.UpdateSentStatus(recovery.Host, true); err != nil {
		c.logError(fmt.Sprintf("Failed to update recovery status for host %s", recovery.Host), err)
	}
}

// logError 记录错误日志
func (c *Consumer) logError(message string, err error) {
	c.logger.Log(fmt.Sprintf("%s: %v", message, err), "error")
}
