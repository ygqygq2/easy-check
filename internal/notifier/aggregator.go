package notifier

import (
	"bytes"
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"fmt"
	"sync"
	"text/template"
	"time"
)

// AlertItem 表示一条告警信息
type AlertItem struct {
	Host        string
	Description string
	Timestamp   time.Time
}

// AlertAggregator 用于聚合一段时间内的告警
type AlertAggregator struct {
	alerts         []*AlertItem
	aggregateTimer *time.Timer
	window         time.Duration
	notifier       Notifier
	logger         *logger.Logger
	mu             sync.Mutex
	active         bool
	config         *config.Config // 修改为通用配置类型
}

// TemplateData 保存传递给模板的数据
type TemplateData struct {
	Time       string
	AlertCount int
	AlertList  string
	Alerts     []*AlertItem
}

// NewAlertAggregator 创建一个新的告警聚合器
func NewAlertAggregator(window time.Duration, notifier Notifier, logger *logger.Logger, config *config.Config) *AlertAggregator {
	agg := &AlertAggregator{
		alerts:   make([]*AlertItem, 0),
		window:   window,
		notifier: notifier,
		logger:   logger,
		active:   true,
		config:   config,
	}

	// 启动聚合定时器
	agg.resetTimer()

	return agg
}

// AddAlert 添加一条告警到聚合队列
func (a *AlertAggregator) AddAlert(host, description string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 如果不活跃，直接返回
	if !a.active {
		return
	}

	alert := &AlertItem{
		Host:        host,
		Description: description,
		Timestamp:   time.Now(),
	}

	a.alerts = append(a.alerts, alert)
	a.logger.Log(fmt.Sprintf("Alert for host %s queued for aggregation", host), "debug")
}

// sendAggregatedAlerts 发送聚合的告警
func (a *AlertAggregator) sendAggregatedAlerts() {
	a.mu.Lock()

	if len(a.alerts) == 0 {
		a.mu.Unlock()
		return
	}

	// 复制告警数组以便在解锁后使用
	alerts := make([]*AlertItem, len(a.alerts))
	copy(alerts, a.alerts)
	alertCount := len(alerts)

	// 清空告警队列
	a.alerts = make([]*AlertItem, 0)

	// 在处理告警之前解锁
	a.mu.Unlock()

	// 构建聚合告警列表
	alertList := a.formatAlertList(alerts)

	// 获取标题
	title := "聚合告警"
	if a.config != nil && a.config.Alert.Feishu.Title != "" {
		title = a.config.Alert.Feishu.Title
	}

	// 尝试使用适当的方法发送告警
	var err error
	if feishuNotifier, ok := a.notifier.(*FeishuNotifier); ok {
		// 如果是飞书通知器，使用专门的聚合方法
		err = feishuNotifier.SendAggregatedNotification(title, alertCount, alertList, alerts)
	} else {
		// 对于其他通知器，使用通用接口方法
		message := fmt.Sprintf("检测到 %d 个主机异常:\n\n%s", alertCount, alertList)
		err = a.notifier.SendNotification(title, message)
	}

	if err != nil {
		a.logger.Log(fmt.Sprintf("Error sending aggregated alerts: %v", err), "error")
	} else {
		a.logger.Log(fmt.Sprintf("Successfully sent aggregated alerts for %d hosts", alertCount), "info")
	}

	// 重置计时器
	a.mu.Lock()
	a.resetTimer()
	a.mu.Unlock()
}

// formatAlertList 根据配置的行模板格式化告警列表
func (a *AlertAggregator) formatAlertList(alerts []*AlertItem) string {
	var buffer bytes.Buffer

	if a.config != nil && a.config.Alert.AggregateLineTemplate != "" {
		// 使用配置的行模板
		lineTemplate := a.config.Alert.AggregateLineTemplate
		tmpl, err := template.New("line").Parse(lineTemplate)
		if err != nil {
			a.logger.Log(fmt.Sprintf("Error parsing line template: %v", err), "error")
			// 使用默认格式
			for _, alert := range alerts {
				timeStr := alert.Timestamp.Format("15:04:05")
				buffer.WriteString(fmt.Sprintf("- [%s] %s: %s\n", timeStr, alert.Host, alert.Description))
			}
			return buffer.String()
		}

		for _, alert := range alerts {
			data := struct {
				Host        string
				Description string
				Time        string
			}{
				Host:        alert.Host,
				Description: alert.Description,
				Time:        alert.Timestamp.Format("15:04:05"),
			}

			var lineBuffer bytes.Buffer
			if err := tmpl.Execute(&lineBuffer, data); err != nil {
				a.logger.Log(fmt.Sprintf("Error applying line template: %v", err), "error")
				buffer.WriteString(fmt.Sprintf("- [%s] %s: %s\n", data.Time, data.Host, data.Description))
			} else {
				buffer.WriteString("- ")
				buffer.WriteString(lineBuffer.String())
				buffer.WriteString("\n")
			}
		}
	} else {
		// 使用默认格式
		for _, alert := range alerts {
			timeStr := alert.Timestamp.Format("15:04:05")
			buffer.WriteString(fmt.Sprintf("- [%s] %s: %s\n", timeStr, alert.Host, alert.Description))
		}
	}

	return buffer.String()
}

// applyTemplate 应用模板到数据
func (a *AlertAggregator) applyTemplate(templateStr string, data interface{}) string {
	tmpl, err := template.New("content").Parse(templateStr)
	if err != nil {
		a.logger.Log(fmt.Sprintf("Error parsing content template: %v", err), "error")
		// 如果模板解析失败，返回简单格式
		if data, ok := data.(TemplateData); ok {
			return fmt.Sprintf("检测到 %d 个主机异常:\n\n%s", data.AlertCount, data.AlertList)
		}
		return templateStr
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		a.logger.Log(fmt.Sprintf("Error applying content template: %v", err), "error")
		return templateStr
	}

	return buffer.String()
}

// resetTimer 重置聚合定时器
func (a *AlertAggregator) resetTimer() {
	if a.aggregateTimer != nil {
		a.aggregateTimer.Stop()
	}

	a.aggregateTimer = time.AfterFunc(a.window, func() {
		a.sendAggregatedAlerts()
	})
}

// Close 关闭聚合器，发送剩余告警
func (a *AlertAggregator) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 标记为非活跃
	a.active = false

	// 停止定时器
	if a.aggregateTimer != nil {
		a.aggregateTimer.Stop()
	}

	// 如果还有未发送的告警，立即发送
	if len(a.alerts) > 0 {
		a.mu.Unlock()
		a.sendAggregatedAlerts()
		a.mu.Lock()
	}
}
