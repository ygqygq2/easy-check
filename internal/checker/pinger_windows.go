//go:build windows
// +build windows

package checker

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"

	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type WindowsPinger struct{}

func (p *WindowsPinger) Ping(host string, count int, timeout int) (string, error) {
	// Windows ping命令参数不同
	cmd := exec.Command("ping", "-4", "-n", fmt.Sprintf("%d", count), "-w", fmt.Sprintf("%d", timeout*1000), host)
	// 隐藏黑色控制台窗口
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
	output, err := cmd.CombinedOutput()

	// 尝试将GBK编码转换为UTF-8
	reader := transform.NewReader(bytes.NewReader(output), simplifiedchinese.GBK.NewDecoder())
	utf8Output, _ := io.ReadAll(reader)
	return string(utf8Output), err
}

func (p *WindowsPinger) ParsePingOutput(lines []string, count int) (int, float64, float64, float64) {
	successCount := 0
	var latencies []float64

	// 使用正则表达式匹配延迟值
	reLatencyAndTTL := regexp.MustCompile(`[=<](\d+)ms TTL=(\d+)`)

	for _, line := range lines {
		// 同时匹配延迟值和 TTL
		matches := reLatencyAndTTL.FindStringSubmatch(line)
		if len(matches) > 2 {
			successCount++
			latency, _ := strconv.ParseFloat(matches[1], 64) // 提取延迟值
			latencies = append(latencies, latency)
			// TTL 值可以根据需要提取，但这里暂时不使用
			// ttl, _ := strconv.Atoi(matches[2])
		}
	}

	// 如果没有延迟值，返回默认值
	if len(latencies) == 0 {
		return successCount, 0, 0, 0
	}

	// 计算最小、最大和平均延迟
	minLatency := latencies[0]
	maxLatency := latencies[0]
	var totalLatency float64

	for _, latency := range latencies {
		if latency < minLatency {
			minLatency = latency
		}
		if latency > maxLatency {
			maxLatency = latency
		}
		totalLatency += latency
	}

	avgLatency := totalLatency / float64(len(latencies))
	return successCount, minLatency, avgLatency, maxLatency
}

func NewPinger() Pinger {
	if isAdmin() {
		return &ICMPPinger{}
	}
	return &WindowsPinger{}
}

func isAdmin() bool {
	// Windows 上检查管理员权限
	cmd := exec.Command("net", "session")
	// 隐藏黑色控制台窗口
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
	err := cmd.Run()
	return err == nil
}
