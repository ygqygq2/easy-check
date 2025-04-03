package aggregator

import (
	"easy-check/internal/db"
	"easy-check/internal/logger"
	"easy-check/internal/queue"
	"time"
)

// AggregatorManager 负责管理告警的聚合逻辑
type AggregatorManager struct {
	window time.Duration
	logger *logger.Logger
	alerts chan *AlertItem
	done   chan struct{}
	db     *db.DB
	queue  *queue.Queue
}

// NewAggregatorManager 创建一个新的 AggregatorManager
func NewAggregatorManager(window int, logger *logger.Logger, db *db.DB, queue *queue.Queue) *AggregatorManager {
	return &AggregatorManager{
		window: time.Duration(window) * time.Second,
		logger: logger,
		alerts: make(chan *AlertItem, 100),
		done:   make(chan struct{}),
		db:     db,
		queue:  queue,
	}
}

// AddAlert 添加告警到聚合队列
func (am *AggregatorManager) AddAlert(host, description string) {
	// 检查当前状态
	currentStatus, _ := am.db.GetAlertStatus(host)
	if currentStatus == "ALERT" {
		// 如果已经是告警状态，忽略
		return
	}

	// 更新状态为 ALERT
	am.db.SetAlertStatus(host, "ALERT")

	// 推送告警事件到队列
	am.queue.Push(queue.AlertEvent{
		Host:        host,
		Description: description,
		Type:        "ALERT",
	})
}

// AddRecovery 添加恢复事件到队列
func (am *AggregatorManager) AddRecovery(host string) {
	// 检查当前状态
	currentStatus, _ := am.db.GetAlertStatus(host)
	if currentStatus != "ALERT" {
		// 如果不是告警状态，忽略
		return
	}

	// 更新状态为 RECOVERY
	am.db.SetAlertStatus(host, "RECOVERY")

	// 推送恢复事件到队列
	am.queue.Push(queue.AlertEvent{
		Host: host,
		Type: "RECOVERY",
	})
}

// Start 启动聚合逻辑
func (am *AggregatorManager) Start(sendFunc func(alerts []*AlertItem)) {
	ticker := time.NewTicker(am.window)
	defer ticker.Stop()

	var batch []*AlertItem
	for {
		select {
		case alert := <-am.alerts:
			batch = append(batch, alert)
		case <-ticker.C:
			if len(batch) > 0 {
				sendFunc(batch) // 调用外部传入的发送函数
				batch = nil
			}
		case <-am.done:
			return
		}
	}
}

// Stop 停止聚合逻辑
func (am *AggregatorManager) Stop() {
	close(am.done)
}
