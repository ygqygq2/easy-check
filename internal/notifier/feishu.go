package notifier

import (
	"bytes"
	"easy-check/internal/config"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type FeishuNotifier struct {
    WebhookURL string
    MsgType    string
    Title      string
    Content    string
}

type FeishuMessage struct {
    MsgType string `json:"msg_type"`
    Content struct {
        Text string `json:"text"`
    } `json:"content"`
}

func NewFeishuNotifier(config *config.FeishuConfig) (*FeishuNotifier, error) {
    if !config.Enable || config.MsgType != "text" {
        return nil, fmt.Errorf("Feishu notifier is not enabled or msg_type is not 'text'")
    }

    return &FeishuNotifier{
        WebhookURL: config.Webhook,
        MsgType:    config.MsgType,
        Title:      config.Title,
        Content:    config.Content,
    }, nil
}

func (n *FeishuNotifier) SendNotification(host, description string) error {
    message := n.Content
    message = replaceTemplateVariables(message, host, description)

    msg := FeishuMessage{
        MsgType: n.MsgType,
    }
    msg.Content.Text = message

    data, err := json.Marshal(msg)
    if err != nil {
        return err
    }

    resp, err := http.Post(n.WebhookURL, "application/json", bytes.NewBuffer(data))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to send notification, status code: %d", resp.StatusCode)
    }

    return nil
}

func replaceTemplateVariables(template, host, description string) string {
    template = strings.ReplaceAll(template, "{{.Host}}", host)
    template = strings.ReplaceAll(template, "{{.Description}}", description)
    return template
}
