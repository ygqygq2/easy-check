package notifier

import (
	"bytes"
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// FeishuNotifier 飞书通知器
type FeishuNotifier struct {
	WebhookURL string
	MsgType    string
	Title      string
	Content    string
	Logger     *logger.Logger
}

// 不同消息类型的接口
type FeishuMessageSender interface {
	PrepareMessage(title string, content string) ([]byte, error)
}

// 文本消息发送器
type TextMessageSender struct{}

type FeishuTextMessage struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

// TODO Post 卡片消息发送器 TODO
type PostMessageSender struct{}

// TODO Interactive 卡片消息发送器
type InteractiveMessageSender struct{}

func NewFeishuNotifier(config *config.FeishuConfig, logger *logger.Logger) (*FeishuNotifier, error) {
	if !config.Enable {
		return nil, fmt.Errorf("Feishu notifier is not enabled")
	}

	// 验证消息类型
	switch config.MsgType {
	case "text", "post", "interactive":
		// 支持的消息类型
	default:
		return nil, fmt.Errorf("Unsupported message type: %s", config.MsgType)
	}

	return &FeishuNotifier{
		WebhookURL: config.Webhook,
		MsgType:    config.MsgType,
		Title:      config.Title,
		Content:    config.Content,
		Logger:     logger,
	}, nil
}

func (n *FeishuNotifier) SendNotification(host, description string) error {
	title := n.Title
	message := n.Content
	message = replaceTemplateVariables(message, host, description)

	// 根据消息类型选择不同的发送器
	var sender FeishuMessageSender
	switch n.MsgType {
	case "text":
		sender = &TextMessageSender{}
	case "post":
		// sender = &PostMessageSender{} // 待实现
		n.Logger.Log("Post message type is not implemented yet", "warn")
		return fmt.Errorf("post message type not implemented yet")
	case "interactive":
		// sender = &InteractiveMessageSender{} // 待实现
		n.Logger.Log("Interactive message type is not implemented yet", "warn")
		return fmt.Errorf("interactive message type not implemented yet")
	default:
		n.Logger.Log(fmt.Sprintf("Unsupported message type: %s", n.MsgType), "error")
		return fmt.Errorf("unsupported message type: %s", n.MsgType)
	}

	// 准备消息
	data, err := sender.PrepareMessage(title, message)
	if err != nil {
		n.Logger.Log(fmt.Sprintf("Failed to prepare message: %v", err), "error")
		return err
	}

	// 发送消息
	n.Logger.Log(fmt.Sprintf("Sending notification to webhook: %s", n.WebhookURL), "debug")
	resp, err := http.Post(n.WebhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		n.Logger.Log(fmt.Sprintf("Failed to send notification: %v", err), "error")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		n.Logger.Log(fmt.Sprintf("Failed to send notification, status code: %d", resp.StatusCode), "error")
		return fmt.Errorf("failed to send notification, status code: %d", resp.StatusCode)
	}

	n.Logger.Log(fmt.Sprintf("Successfully sent notification for host %s", host), "info")
	return nil
}

// TextMessageSender 实现 PrepareMessage 方法
func (s *TextMessageSender) PrepareMessage(title string, content string) ([]byte, error) {
	msg := FeishuTextMessage{
		MsgType: "text",
	}

	// 标题添加到内容的第一行，并添加换行
	if title != "" {
		msg.Content.Text = title + "\n" + content
	} else {
		msg.Content.Text = content
	}

	return json.Marshal(msg)
}

func replaceTemplateVariables(template, host, description string) string {
	// 当前时间
	time := time.Now().Format("2006-01-02 15:04:05")
	template = strings.ReplaceAll(template, "{{.Time}}", time)
	template = strings.ReplaceAll(template, "{{.Host}}", host)
	template = strings.ReplaceAll(template, "{{.Description}}", description)
	return template
}

// 添加一个新的方法用于处理聚合消息
func (n *FeishuNotifier) SendAggregatedNotification(title string, alertCount int, alertList string, alerts []*AlertItem) error {
	message := n.Content

	// 替换通用变量
	time := time.Now().Format("2006-01-02 15:04:05")
	message = strings.ReplaceAll(message, "{{.Time}}", time)

	// 替换聚合特定变量
	message = strings.ReplaceAll(message, "{{.AlertCount}}", fmt.Sprintf("%d", alertCount))
	message = strings.ReplaceAll(message, "{{.AlertList}}", alertList)

	// 根据消息类型选择不同的发送器
	var sender FeishuMessageSender
	switch n.MsgType {
	case "text":
		sender = &TextMessageSender{}
	case "post":
		n.Logger.Log("Post message type is not implemented yet", "warn")
		return fmt.Errorf("post message type not implemented yet")
	case "interactive":
		n.Logger.Log("Interactive message type is not implemented yet", "warn")
		return fmt.Errorf("interactive message type not implemented yet")
	default:
		n.Logger.Log(fmt.Sprintf("Unsupported message type: %s", n.MsgType), "error")
		return fmt.Errorf("unsupported message type: %s", n.MsgType)
	}

	// 准备消息
	data, err := sender.PrepareMessage(title, message)
	if err != nil {
		n.Logger.Log(fmt.Sprintf("Failed to prepare message: %v", err), "error")
		return err
	}

	// 发送消息
	n.Logger.Log("Sending aggregated notification", "debug")
	resp, err := http.Post(n.WebhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		n.Logger.Log(fmt.Sprintf("Failed to send notification: %v", err), "error")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		n.Logger.Log(fmt.Sprintf("Failed to send notification, status code: %d", resp.StatusCode), "error")
		return fmt.Errorf("failed to send notification, status code: %d", resp.StatusCode)
	}

	n.Logger.Log(fmt.Sprintf("Successfully sent aggregated notification for %d hosts", alertCount), "info")
	return nil
}
