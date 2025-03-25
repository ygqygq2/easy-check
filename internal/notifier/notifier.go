package notifier

import (
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"fmt"
	"time"
)

// Notifier 接口定义了通知器的基本行为
type Notifier interface {
	SendNotification(title, content string) error
}

// AggregatingNotifier 是一个装饰器，为任何通知器添加聚合功能
type AggregatingNotifier struct {
	baseNotifier Notifier
	aggregator   *AlertAggregator
}

// NewAggregatingNotifier 创建一个新的聚合通知器
func NewAggregatingNotifier(baseNotifier Notifier, config *config.Config, logger *logger.Logger) *AggregatingNotifier {
	window := time.Duration(config.Alert.AggregateWindow) * time.Second
	return &AggregatingNotifier{
		baseNotifier: baseNotifier,
		aggregator:   NewAlertAggregator(window, baseNotifier, logger, config),
	}
}

// SendNotification 实现 Notifier 接口，将告警添加到聚合队列
func (n *AggregatingNotifier) SendNotification(host, description string) error {
	// 添加到聚合队列
	n.aggregator.AddAlert(host, description)
	return nil
}

// SendDirectNotification 直接发送通知，不经过聚合
func (n *AggregatingNotifier) SendDirectNotification(title, content string) error {
	if n.baseNotifier != nil {
		return n.baseNotifier.SendNotification(title, content)
	}
	return fmt.Errorf("no base notifier available")
}

// Close 关闭聚合器
func (n *AggregatingNotifier) Close() {
	n.aggregator.Close()
}
