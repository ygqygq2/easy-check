package notifier

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"
)

// 通用的模板处理函数
func processTemplate(templateStr string, variables map[string]string) string {
	result := templateStr
	for key, value := range variables {
		result = strings.ReplaceAll(result, "{{."+key+"}}", value)
	}
	return result
}

// 通用的告警列表格式化函数
func formatAlertList(alerts []*AlertItem, templateStr string) string {
	var buffer bytes.Buffer

	tmpl, err := template.New("alertList").Parse(templateStr)
	if err != nil {
		// 如果模板解析失败，使用默认格式
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
			buffer.WriteString(fmt.Sprintf("- [%s] %s: %s\n", data.Time, data.Host, data.Description))
		} else {
			buffer.WriteString(lineBuffer.String())
			buffer.WriteString("\n")
		}
	}

	return buffer.String()
}

// 通用的告警内容生成函数
func generateContent(templateStr string, alerts []*AlertItem) (string, error) {
	alertCount := len(alerts)
	alertList := formatAlertList(alerts, templateStr)

	message := processTemplate(templateStr, map[string]string{
		"Time":       time.Now().Format("2006-01-02 15:04:05"),
		"AlertCount": fmt.Sprintf("%d", alertCount),
		"AlertList":  alertList,
	})

	return message, nil
}
