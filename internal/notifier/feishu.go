package notifier

import (
	"bytes"
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"
)

// FeishuNotifier 飞书通知器
type FeishuNotifier struct {
	WebhookURL string
	MsgType    string
	Logger     *logger.Logger
	Config     *config.Config
	Options    map[string]interface{} // 从 NotifierConfig.Options 中读取
}

// FeishuResponse 飞书 API 响应结构
type FeishuResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// FeishuMessageSender 消息发送器接口
type FeishuMessageSender interface {
	PrepareMessage(title, content string) ([]byte, error)
}

// TextMessageSender 文本消息发送器
type TextMessageSender struct{}

// FeishuTextMessage 飞书文本消息结构
type FeishuTextMessage struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

// TODO Post 卡片消息发送器
type PostMessageSender struct{}

// TODO Interactive 卡片消息发送器
type InteractiveMessageSender struct{}

// 实现 Notifier 接口的 SendNotification 方法
func (f *FeishuNotifier) SendNotification(host config.Host) error {
	f.Logger.Log(fmt.Sprintf("Sending notification to webhook: %s", f.WebhookURL), "debug")

	// 准备消息内容
	content, err := f.prepareContent(host)
	if err != nil {
		return fmt.Errorf("failed to prepare content: %v", err)
	}

	// 发送消息
	err = f.sendMessage(content)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	f.Logger.Log("Successfully sent notification", "info")
	return nil
}

func NewFeishuNotifier(cfg *config.NotifierConfig, logger *logger.Logger) (*FeishuNotifier, error) {
	if !cfg.Enable {
		return nil, fmt.Errorf("feishu notifier is not enabled")
	}

	// 验证消息类型
	msgType, ok := cfg.Options["msg_type"].(string)
	if !ok || (msgType != "text" && msgType != "post" && msgType != "interactive") {
		return nil, fmt.Errorf("unsupported or missing message type in Feishu notifier options")
	}

	webhookURL, ok := cfg.Options["webhook"].(string)
	if !ok || webhookURL == "" {
		return nil, fmt.Errorf("missing webhook URL in Feishu notifier options")
	}

	return &FeishuNotifier{
		WebhookURL: webhookURL,
		MsgType:    msgType,
		Logger:     logger,
		Options:    cfg.Options,
	}, nil
}

// prepareContent 准备消息内容
func (f *FeishuNotifier) prepareContent(host config.Host) (string, error) {
	// 从 Options 中获取模板内容
	templateContent, ok := f.Options["content_template"].(string)
	if !ok || templateContent == "" {
		return "", fmt.Errorf("missing or invalid content_template in Feishu notifier options")
	}

	// 使用模板生成消息内容
	return processTemplate(templateContent, map[string]string{
		"Date":        time.Now().Format("2006-01-02"),
		"Time":        time.Now().Format("15:04:05"),
		"Host":        host.Host,
		"Description": host.Description,
	})
}

// sendMessage 发送消息
func (f *FeishuNotifier) sendMessage(content string) error {
	// 构造飞书消息
	data, err := json.Marshal(FeishuTextMessage{
		MsgType: "text",
		Content: struct {
			Text string `json:"text"`
		}{
			Text: content,
		},
	})
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

	return nil
}

// PrepareMessage 准备文本消息内容
func (s *TextMessageSender) PrepareMessage(title, content string) ([]byte, error) {
	msg := FeishuTextMessage{
		MsgType: "text",
	}
	msg.Content.Text = fmt.Sprintf("%s\n%s", title, content)
	return json.Marshal(msg)
}

func (n *FeishuNotifier) Close() error {
	n.Logger.Log("Closing FeishuNotifier", "info")
	// 如果有需要清理的资源，可以在这里处理
	return nil
}

// PrepareAggregatedContent 准备聚合告警的内容
func (f *FeishuNotifier) PrepareAggregatedContent(alerts []*AlertItem) (string, error) {
	// 从配置中获取行模板
	lineTemplate := f.Config.Alert.AggregateLineTemplate
	if lineTemplate == "" {
		lineTemplate = "- 时间：{{.FailTime}} | 主机：{{.Host}} | 描述：{{.Description}}" // 默认行模板
	}

	// 根据行模板生成 AlertList
	alertList := make([]string, len(alerts))
	for i, alert := range alerts {
		data := struct {
			Host        string
			Description string
			FailTime    string
		}{
			Host:        alert.Host,
			Description: alert.Description,
			FailTime:    alert.Timestamp.Format("15:04:05"),
		}

		var buffer bytes.Buffer
		tmpl, err := template.New("line").Parse(lineTemplate)
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

	// 准备聚合模板
	templateStr := f.Config.Alert.AggregateLineTemplate
	if templateStr == "" {
		templateStr = "检测到 {{.AlertCount}} 个主机异常:\n\n{{.AlertList}}" // 默认聚合模板
	}

	// 替换聚合模板中的 AlertList
	data := TemplateData{
		Date:       time.Now().Format("2006-01-02"),
		Time:       time.Now().Format("15:04:05"),
		AlertCount: len(alerts),
		AlertList:  alertListStr,
		Alerts:     alerts,
	}

	var buffer bytes.Buffer
	tmpl, err := template.New("aggregate").Parse(templateStr)
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

func (f *FeishuNotifier) SendAggregatedNotification(alerts []*AlertItem) error {
	content, err := f.PrepareAggregatedContent(alerts)
	if err != nil {
		f.Logger.Log(fmt.Sprintf("Error preparing aggregated content: %v", err), "error")
		return err
	}

	err = f.sendMessage(content)
	if err != nil {
		f.Logger.Log(fmt.Sprintf("Error sending aggregated notification: %v", err), "error")
		return err
	}

	f.Logger.Log("Successfully sent aggregated notification via Feishu", "info")
	return nil
}
