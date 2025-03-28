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

	// 根据消息类型选择消息发送器
	var messageSender FeishuMessageSender
	switch n.MsgType {
	case "text":
		messageSender = &TextMessageSender{}
	default:
		return fmt.Errorf("unsupported message type: %s", n.MsgType)
	}

	// 准备消息内容
	data, err := messageSender.PrepareMessage(title, content)
	if err != nil {
		n.Logger.Log(fmt.Sprintf("Failed to prepare message content: %v", err), "error")
		return fmt.Errorf("failed to prepare message content: %w", err)
	}

	// 打印准备好的数据
	n.Logger.Log(fmt.Sprintf("Prepared message data: %s", string(data)), "debug")

	// 发送 HTTP 请求
	resp, err := http.Post(n.WebhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		n.Logger.Log(fmt.Sprintf("Failed to send notification: %v", err), "error")
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应内容
	var feishuResp FeishuResponse
	if err := json.NewDecoder(resp.Body).Decode(&feishuResp); err != nil {
		n.Logger.Log(fmt.Sprintf("Failed to parse response: %v", err), "error")
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查响应码
	if feishuResp.Code != 0 {
		errMsg := fmt.Sprintf("API error: code=%d, message=%s", feishuResp.Code, feishuResp.Msg)
		n.Logger.Log(errMsg, "error")
		return fmt.Errorf("send notification failed: %s", errMsg)
	}

	n.Logger.Log("Successfully sent notification", "info")
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
