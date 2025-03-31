//go:build windows
// +build windows

package checker

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type WindowsPinger struct{}

func (p *WindowsPinger) Ping(host string, count int, timeout int) (string, error) {
	// Windows ping命令参数不同
	cmd := exec.Command("ping", "-4", "-n", fmt.Sprintf("%d", count), "-w", fmt.Sprintf("%d", timeout*1000), host)
	output, err := cmd.CombinedOutput()

	// 尝试将GBK编码转换为UTF-8
	reader := transform.NewReader(bytes.NewReader(output), simplifiedchinese.GBK.NewDecoder())
	utf8Output, _ := io.ReadAll(reader)
	return string(utf8Output), err
}

func (p *WindowsPinger) ParsePingOutput(lines []string, count int) (int, string) {
	successCount := 0
	var sampleLatency string

	// 使用更通用的正则表达式匹配时间值，不依赖特定的前缀文本
	// 这个正则表达式会匹配任何"数字+ms"的模式
	reTime := regexp.MustCompile(`(\d+)ms`)

	for _, line := range lines {
		// 只检查TTL存在与否，这在不同语言版本中都是一致的
		if strings.Contains(line, "TTL=") || strings.Contains(line, "TTL=") {
			successCount++
			if sampleLatency == "" {
				matches := reTime.FindStringSubmatch(line)
				if len(matches) > 1 {
					// 构造延迟字符串
					sampleLatency = fmt.Sprintf("time=%sms", matches[1])
				}
			}
		}
	}

	return successCount, sampleLatency
}

func NewPinger() Pinger {
	return &WindowsPinger{}
}
