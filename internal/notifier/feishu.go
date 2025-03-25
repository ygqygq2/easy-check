package notifier

import (
	"bytes"
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"encoding/json"
	"fmt"
	"net/http"
)

// FeishuNotifier 飞书通知器
type FeishuNotifier struct {
	WebhookURL string
	MsgType    string
	Logger     *logger.Logger
}

// 不同消息类型的接口
type FeishuMessageSender interface {
	PrepareMessage(content string) ([]byte, error)
}

// 文本消息发送器
type TextMessageSender struct{}

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
		Logger:     logger,
	}, nil
}

// SendNotification 发送告警通知
func (n *FeishuNotifier) SendNotification(title, content string) error {
	n.Logger.Log(fmt.Sprintf("Sending notification to webhook: %s", n.WebhookURL), "debug")

	// 准备消息内容
	messageSender := &TextMessageSender{}
	data, err := messageSender.PrepareMessage(title, content)
	if err != nil {
		n.Logger.Log("Failed to prepare message content", "error")
		return fmt.Errorf("failed to prepare message content: %w", err)
	}

	// 打印准备好的数据
	n.Logger.Log(fmt.Sprintf("Prepared message data: %s", string(data)), "debug")

	resp, err := http.Post(n.WebhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		n.Logger.Log("Failed to send notification", "error")
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	// 打印响应状态码
	n.Logger.Log(fmt.Sprintf("Response status code: %d", resp.StatusCode), "debug")

	if resp.StatusCode != http.StatusOK {
		n.Logger.Log(fmt.Sprintf("Failed to send notification, status code: %d", resp.StatusCode), "error")
		return fmt.Errorf("failed to send notification, status code: %d", resp.StatusCode)
	}

	n.Logger.Log("Successfully sent notification", "info")
	return nil
}

// TextMessageSender 实现 PrepareMessage 方法
func (s *TextMessageSender) PrepareMessage(title string, content string) ([]byte, error) {
	msg := FeishuTextMessage{
		MsgType: "text",
	}

	// 合并 title 和 content
	msg.Content.Text = fmt.Sprintf("%s\n%s", title, content)

	return json.Marshal(msg)
}
