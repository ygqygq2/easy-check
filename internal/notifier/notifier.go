package notifier

import (
	"bytes"
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"fmt"
	"text/template"
	"time"
)

// Notifier 接口定义了通知器的基本行为
type Notifier interface {
	SendNotification(title, content string) error
}

// AggregatingNotifier 是一个装饰器，为任何通知器添加聚合功能
type AggregatingNotifier struct {
	notifiers  []Notifier
	aggregator *AlertAggregator
	config     *config.Config
}

// NewAggregatingNotifier 创建一个新的聚合通知器
func NewAggregatingNotifier(notifiers []Notifier, config *config.Config, logger *logger.Logger) *AggregatingNotifier {
	window := time.Duration(config.Alert.AggregateWindow) * time.Second
	return &AggregatingNotifier{
		notifiers:  notifiers,
		aggregator: NewAlertAggregator(window, notifiers, logger, config),
		config:     config,
	}
}

// SendNotification 实现 Notifier 接口，根据配置判断是否开启聚合告警
func (n *AggregatingNotifier) SendNotification(host, description string) error {
  // 开启聚合告警
	if n.config.Alert.AggregateAlerts {
		// 添加到聚合队列
		n.aggregator.AddAlert(host, description)
	} else {
		// 直接发送通知
		title := n.config.Alert.Feishu.Title
		content, err := processTemplate(n.config.Alert.Feishu.Content, map[string]string{
      "Date":         time.Now().Format("2006-01-02"),
			"Time":        time.Now().Format("2006-01-02 15:04:05"),
			"Host":        host,
			"Description": description,
		})
		if err != nil {
			return fmt.Errorf("failed to process template: %v", err)
		}
		return n.SendDirectNotification(title, content)
	}
	return nil
}

// SendDirectNotification 直接发送通知，不经过聚合
func (n *AggregatingNotifier) SendDirectNotification(title, content string) error {
	var err error
	for _, notifier := range n.notifiers {
		err = notifier.SendNotification(title, content)
		if err != nil {
			return err
		}
	}
	return nil
}

// applyTemplate 应用模板到数据
func (n *AggregatingNotifier) applyTemplate(templateStr string, data interface{}) string {
	tmpl, err := template.New("content").Parse(templateStr)
	if err != nil {
		return fmt.Sprintf("Error parsing template: %v", err)
	}
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, data)
	if err != nil {
		return fmt.Sprintf("Error executing template: %v", err)
	}
	return buffer.String()
}

// Close 关闭聚合器
func (n *AggregatingNotifier) Close() {
	n.aggregator.Close()
}
