package notifier

import (
	"bytes"
	"easy-check/internal/config"
	"easy-check/internal/db"
	"easy-check/internal/logger"
	"easy-check/internal/types"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"
)

// 定义配置项的 key 枚举
type FeishuOptionKey string

const (
	OptionKeyWebhook         FeishuOptionKey = "webhook"
	OptionKeyMsgType         FeishuOptionKey = "msg_type"
	OptionKeyAlertTitle      FeishuOptionKey = "alert_title"
	OptionKeyAlertContent    FeishuOptionKey = "alert_content"
	OptionKeyRecoveryTitle   FeishuOptionKey = "recovery_title"
	OptionKeyRecoveryContent FeishuOptionKey = "recovery_content"
)

// FeishuNotifier 飞书通知器
type FeishuNotifier struct {
	WebhookURL string
	MsgType    string
	Logger     *logger.Logger
	Config     *config.Config
	Options    map[string]interface{} // 从 NotifierConfig.Options 中读取
}

// FeishuTextMessage 飞书文本消息结构
type FeishuTextMessage struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

// FeishuResponse 飞书 API 响应结构
type FeishuResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// TemplateData 定义用于模板渲染的数据结构
type TemplateData struct {
	Date       string
	Time       string
	AlertCount int
	AlertList  string
	Alerts     []*db.AlertStatus
}

// FeishuMessageSender 消息发送器接口
type FeishuMessageSender interface {
	PrepareMessage(title, content string) ([]byte, error)
}

// TextMessageSender 文本消息发送器
type TextMessageSender struct{}

// TODO Post 卡片消息发送器
type PostMessageSender struct{}

// TODO Interactive 卡片消息发送器
type InteractiveMessageSender struct{}

// 实现 Notifier 接口的 SendNotification 方法
func (f *FeishuNotifier) SendNotification(alert *db.AlertStatus, isRecovery bool) error {
	// 根据类型设置标题和模板
	var titleKey, contentKey FeishuOptionKey
	var defaultTitle, defaultTemplate string

	if isRecovery {
		titleKey = OptionKeyRecoveryTitle
		contentKey = OptionKeyRecoveryContent
		defaultTitle = "💚【easy-check】：恢复通知"
		defaultTemplate = "🧭【恢复时间】：{{.RecoveryTime}}\n📝【恢复详情】：以下主机已恢复：\n- 开始时间：{{.FailTime}} | 主机：{{.Host}} | 描述：{{.Description}} | 恢复时间：{{.RecoveryTime}}"
	} else {
		titleKey = OptionKeyAlertTitle
		contentKey = OptionKeyAlertContent
		defaultTitle = "💔【easy-check】：告警通知"
		defaultTemplate = "🧭【告警时间】：{{.Date}} {{.Time}}\n📝【告警详情】：以下主机不可达：\n- 开始时间：{{.FailTime}} | 主机：{{.Host}} | 描述：{{.Description}}"
	}

	// 获取标题
	title, ok := f.Options[string(titleKey)].(string)
	if !ok || title == "" {
		title = defaultTitle
	}

	// 获取模板
	templateContent, ok := f.Options[string(contentKey)].(string)
	if !ok || templateContent == "" {
		templateContent = defaultTemplate
	}

	// 准备模板数据
	data := map[string]string{
		"Date":         time.Now().Format("2006-01-02"),
		"Time":         time.Now().Format("15:04:05"),
		"Host":         alert.Host,
		"Description":  alert.Description,
		"FailTime":     formatTime(alert.FailTime),
		"RecoveryTime": "", // 默认值为空字符串
	}

	// 检查 RecoveryTime 是否存在
	if alert.RecoveryTime != "" {
		data["RecoveryTime"] = formatTime(alert.RecoveryTime)
	}

	// 使用模板生成消息内容
	var buffer bytes.Buffer
	tmpl, err := template.New("notification").Parse(templateContent)
	if err != nil {
		f.Logger.Log(fmt.Sprintf("Error parsing template: %v", err), "error")
		return fmt.Errorf("failed to parse template: %v", err)
	}

	if err := tmpl.Execute(&buffer, data); err != nil {
		f.Logger.Log(fmt.Sprintf("Error applying template: %v", err), "error")
		return fmt.Errorf("failed to apply template: %v", err)
	}

	content := buffer.String()
	f.Logger.Log(fmt.Sprintf("Generated notification content: %s", content), "debug")

	// 准备完整消息（包含标题和内容）
	sender := &TextMessageSender{}
	message, err := sender.PrepareMessage(title, content)
	if err != nil {
		return fmt.Errorf("failed to concatenate title and content: %v", err)
	}

	// 发送消息
	err = f.sendMessage(string(message))
	if err != nil {
		return fmt.Errorf("failed to send notification: %v", err)
	}

	f.Logger.Log("Successfully sent notification", "debug")
	return nil
}

func NewFeishuNotifier(options map[string]interface{}, logger *logger.Logger) (types.Notifier, error) {
	webhookURL, ok := options[string(OptionKeyWebhook)].(string)
	if !ok || webhookURL == "" {
		return nil, fmt.Errorf("missing webhook URL in Feishu notifier options")
	}

	msgType, ok := options[string(OptionKeyMsgType)].(string)
	if !ok || (msgType != "text" && msgType != "post" && msgType != "interactive") {
		return nil, fmt.Errorf("unsupported or missing message type in Feishu notifier options")
	}

	return &FeishuNotifier{
		WebhookURL: webhookURL,
		MsgType:    msgType,
		Logger:     logger,
		Options:    options,
	}, nil
}

// prepareContent 准备消息内容
func (f *FeishuNotifier) prepareContent(host config.Host, failTime time.Time) (string, error) {
	// 从配置中获取模板内容
	templateContent, ok := f.Options[string(OptionKeyAlertContent)].(string)
	if !ok || templateContent == "" {
		return "", fmt.Errorf("missing or invalid content template in configuration")
	}

	// 使用模板生成消息内容
	data := map[string]string{
		"Date":        time.Now().Format("2006-01-02"),
		"Time":        time.Now().Format("15:04:05"),
		"FailTime":    failTime.Format("15:04:05"),
		"Host":        host.Host,
		"Description": host.Description,
	}

	var buffer bytes.Buffer
	tmpl, err := template.New(string(OptionKeyAlertContent)).Parse(templateContent)
	if err != nil {
		f.Logger.Log(fmt.Sprintf("Error parsing content template: %v", err), "error")
		return "", fmt.Errorf("failed to parse content template: %v", err)
	}

	if err := tmpl.Execute(&buffer, data); err != nil {
		f.Logger.Log(fmt.Sprintf("Error applying content template: %v", err), "error")
		return "", fmt.Errorf("failed to apply content template: %v", err)
	}

	return buffer.String(), nil
}

// sendMessage 发送消息
func (f *FeishuNotifier) sendMessage(content string) error {
	// 打印发送的消息内容
	f.Logger.Log(fmt.Sprintf("Sending message: %s", content), "debug")

	// 构造飞书消息
	message := FeishuTextMessage{
		MsgType: "text",
	}
	message.Content.Text = content

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// 发送 HTTP 请求
	resp, err := http.Post(f.WebhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应内容
	var feishuResp FeishuResponse
	if err := json.NewDecoder(resp.Body).Decode(&feishuResp); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	// 检查响应码
	if feishuResp.Code != 0 {
		return fmt.Errorf("API error: code=%d, message=%s", feishuResp.Code, feishuResp.Msg)
	}

	f.Logger.Log("Message sent successfully", "debug")
	return nil
}

// PrepareMessage 简单拼接标题和内容
func (s *TextMessageSender) PrepareMessage(title, content string) ([]byte, error) {
	// 简单拼接标题和内容，用换行分隔
	message := fmt.Sprintf("%s\n%s", title, content)
	return []byte(message), nil
}

func (n *FeishuNotifier) Close() error {
	n.Logger.Log("Closing FeishuNotifier", "debug")
	// 如果有需要清理的资源，可以在这里处理
	return nil
}

// PrepareAggregatedContent 通用方法，准备聚合内容（告警或恢复）
func (f *FeishuNotifier) PrepareAggregatedContent(alerts []*db.AlertStatus, isRecovery bool) (string, error) {
	// 检查输入是否为空
	if len(alerts) == 0 {
		return "", fmt.Errorf("no alerts to process")
	}

	// 根据类型选择模板
	var lineTemplateContent, aggregateTemplateContent string
	if isRecovery {
		lineTemplateContent, _ = f.Options["recovery_line_template"].(string)
		aggregateTemplateContent, _ = f.Options[string(OptionKeyRecoveryContent)].(string)
		if lineTemplateContent == "" {
			lineTemplateContent = "- 开始时间：{{.FailTime}} | 恢复时间：{{.RecoveryTime}} | 主机：{{.Host}} | 描述：{{.Description}}"
		}
		if aggregateTemplateContent == "" {
			aggregateTemplateContent = "🧭【发送时间】：{{.Date}} {{.Time}}\n📝【恢复详情】：以下 {{.AlertCount}} 主机已恢复：\n{{.AlertList}}"
		}
	} else {
		lineTemplateContent, _ = f.Options["alert_line_template"].(string)
		aggregateTemplateContent, _ = f.Options[string(OptionKeyAlertContent)].(string)
		if lineTemplateContent == "" {
			lineTemplateContent = "- 开始时间：{{.FailTime}} | 主机：{{.Host}} | 描述：{{.Description}}"
		}
		if aggregateTemplateContent == "" {
			aggregateTemplateContent = "🧭【告警时间】：{{.Date}} {{.Time}}\n📝【告警详情】：以下 {{.AlertCount}} 个主机不可达：\n{{.AlertList}}"
		}
	}

	// 根据行模板生成 AlertList
	alertList := make([]string, len(alerts))
	for i, alert := range alerts {
		data := struct {
			Host         string
			Description  string
			FailTime     string
			RecoveryTime string
			Date         string
			Time         string
		}{
			Host:         alert.Host,
			Description:  alert.Description,
			FailTime:     formatTime(alert.FailTime),
			RecoveryTime: formatTime(alert.RecoveryTime),
			Date:         time.Now().Format("2006-01-02"),
			Time:         time.Now().Format("15:04:05"),
		}

		var buffer bytes.Buffer
		tmpl, err := template.New("line").Parse(lineTemplateContent)
		if err != nil {
			f.Logger.Log(fmt.Sprintf("Error parsing line template: %v", err), "error")
			return "", fmt.Errorf("failed to parse line template: %v", err)
		}

		if err := tmpl.Execute(&buffer, data); err != nil {
			f.Logger.Log(fmt.Sprintf("Error applying line template: %v", err), "error")
			return "", fmt.Errorf("failed to apply line template: %v", err)
		}

		alertList[i] = buffer.String()
	}

	// 将 AlertList 拼接为字符串
	alertListStr := strings.Join(alertList, "\n")

	// 准备聚合模板数据
	data := TemplateData{
		Date:       time.Now().Format("2006-01-02"),
		Time:       time.Now().Format("15:04:05"),
		AlertCount: len(alerts),
		AlertList:  alertListStr,
		Alerts:     alerts,
	}

	var buffer bytes.Buffer
	tmpl, err := template.New("aggregate").Parse(aggregateTemplateContent)
	if err != nil {
		f.Logger.Log(fmt.Sprintf("Error parsing aggregate template: %v", err), "error")
		return "", fmt.Errorf("failed to parse aggregate template: %v", err)
	}

	if err := tmpl.Execute(&buffer, data); err != nil {
		f.Logger.Log(fmt.Sprintf("Error applying aggregate template: %v", err), "error")
		return "", fmt.Errorf("failed to apply aggregate template: %v", err)
	}

	return buffer.String(), nil
}

// SendAggregatedNotification 通用方法，发送聚合通知（告警或恢复）
func (f *FeishuNotifier) SendAggregatedNotification(alerts []*db.AlertStatus, isRecovery bool) error {
	// 准备聚合内容
	content, err := f.PrepareAggregatedContent(alerts, isRecovery)
	if err != nil {
		f.Logger.Log(fmt.Sprintf("Error preparing aggregated content: %v", err), "error")
		return err
	}

	// 获取标题
	var titleKey FeishuOptionKey
	var defaultTitle string
	if isRecovery {
		titleKey = OptionKeyRecoveryTitle
		defaultTitle = "💚【easy-check】：恢复通知"
	} else {
		titleKey = OptionKeyAlertTitle
		defaultTitle = "💔【easy-check】：告警通知"
	}

	title, ok := f.Options[string(titleKey)].(string)
	if !ok || title == "" {
		title = defaultTitle
	}

	// 准备完整消息
	sender := &TextMessageSender{}
	message, err := sender.PrepareMessage(title, content)
	if err != nil {
		f.Logger.Log(fmt.Sprintf("Error preparing message: %v", err), "error")
		return fmt.Errorf("failed to prepare message: %v", err)
	}

	// 发送消息
	err = f.sendMessage(string(message))
	if err != nil {
		f.Logger.Log(fmt.Sprintf("Error sending aggregated notification: %v", err), "error")
		return err
	}

	if isRecovery {
		f.Logger.Log("Successfully sent aggregated recovery notification via Feishu", "debug")
	} else {
		f.Logger.Log("Successfully sent aggregated alert notification via Feishu", "debug")
	}
	return nil
}
