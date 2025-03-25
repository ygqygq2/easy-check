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

// SendNotification 发送单条告警通知
func (n *FeishuNotifier) SendNotification(title, content string) error {
	// 处理普通告警的模板
	message := processTemplate(n.Content, map[string]string{
		"Time":        time.Now().Format("2006-01-02 15:04:05"),
		"Host":        title, // 为了兼容之前的逻辑，使用title作为Host
		"Description": content,
	})

	// 通过通用方法发送消息
	return n.sendMessage(n.Title, message, fmt.Sprintf("host %s", title))
}

// SendAggregatedNotification 发送聚合告警通知
func (n *FeishuNotifier) SendAggregatedNotification(title string, alertCount int, alertList string, alerts []*AlertItem) error {
	// 处理聚合告警的模板
	message := processTemplate(n.Content, map[string]string{
		"Time":       time.Now().Format("2006-01-02 15:04:05"),
		"AlertCount": fmt.Sprintf("%d", alertCount),
		"AlertList":  alertList,
	})

	// 通过通用方法发送消息
	return n.sendMessage(title, message, fmt.Sprintf("%d hosts", alertCount))
}

// sendMessage 内部方法，处理实际的消息发送逻辑
func (n *FeishuNotifier) sendMessage(title, content, logContext string) error {
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
	data, err := sender.PrepareMessage(title, content)
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

	n.Logger.Log(fmt.Sprintf("Successfully sent notification for %s", logContext), "info")
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

// processTemplate 统一的模板处理函数
func processTemplate(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		result = strings.ReplaceAll(result, "{{."+key+"}}", value)
	}
	return result
}
