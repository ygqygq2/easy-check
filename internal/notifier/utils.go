package notifier

import (
	"bytes"
	"fmt"
	"text/template"
)

// 通用的模板处理函数
func processTemplate(templateStr string, data interface{}) (string, error) {
	tmpl, err := template.New("template").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %v", err)
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, data)
	if err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}

	return buffer.String(), nil
}

// 通用的告警列表格式化函数
func formatAlertList(alerts []*AlertItem, templateStr string) string {
	var buffer bytes.Buffer

	// 如果提供了模板字符串，尝试解析模板
	if templateStr != "" {
		tmpl, err := template.New("alertList").Parse(templateStr)
		if err != nil {
			// 如果模板解析失败，记录错误并使用默认格式
			fmt.Printf("Error parsing alert list template: %v\n", err)
		} else {
			// 使用模板格式化每条告警
			for _, alert := range alerts {
				data := struct {
					Host        string
					Description string
					FailTime    string
				}{
					Host:        alert.Host,
					Description: alert.Description,
					FailTime:    alert.Timestamp.Format("15:04:05"),
				}

				var lineBuffer bytes.Buffer
				if err := tmpl.Execute(&lineBuffer, data); err != nil {
					// 如果模板执行失败，记录错误并使用默认格式
					fmt.Printf("Error applying alert list template: %v\n", err)
					buffer.WriteString(fmt.Sprintf("- [%s] %s: %s\n", data.FailTime, data.Host, data.Description))
				} else {
					buffer.WriteString(lineBuffer.String())
          buffer.WriteString("\n")
				}
			}
			return buffer.String()
		}
	}

	// 如果没有模板或模板解析失败，使用默认格式
	for _, alert := range alerts {
		timeStr := alert.Timestamp.Format("15:04:05")
		buffer.WriteString(fmt.Sprintf("- [%s] %s: %s\n", timeStr, alert.Host, alert.Description))
	}

	return buffer.String()
}
