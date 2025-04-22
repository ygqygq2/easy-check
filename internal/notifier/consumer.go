package notifier

import (
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
	alerts, err := c.db.GetAllUnsentStatuses(statusType)
	if err != nil {
		c.logError(fmt.Sprintf("Failed to fetch unsent %s", eventType), err)
		return
	}

	if len(alerts) == 0 {
		c.logger.Log(fmt.Sprintf("No unsent %s to process", eventType), "debug")
		return
	}

	switch eventType {
	case "alerts":
		if err := c.handler.ProcessAlerts(alerts, c.db); err != nil {
			c.logError("Failed to process alerts", err)
		}
	case "recoveries":
		if err := c.handler.ProcessRecoveries(alerts, c.db); err != nil {
			c.logError("Failed to process recoveries", err)
		}
	default:
		c.logger.Log(fmt.Sprintf("Unknown event type: %s", eventType), "warn")
	}
}

// logError 记录错误日志
func (c *Consumer) logError(message string, err error) {
	c.logger.Log(fmt.Sprintf("%s: %v", message, err), "error")
}
