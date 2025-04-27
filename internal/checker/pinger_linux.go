//go:build linux
// +build linux

package checker

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type LinuxPinger struct{}

func (p *LinuxPinger) Ping(host string, count int, timeout int) (string, error) {
	cmd := exec.Command("ping", "-4", "-c", fmt.Sprintf("%d", count), "-W", fmt.Sprintf("%d", timeout), host)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (p *LinuxPinger) ParsePingOutput(lines []string, count int) (int, float64, float64, float64) {
	successCount := 0
	var latencies []float64

	// 正则表达式匹配延迟值
	re := regexp.MustCompile(`time=(\d+(\.\d+)?) ms`)

	for _, line := range lines {
		if strings.Contains(line, "ttl=") || strings.Contains(line, "time=") {
			successCount++
			match := re.FindStringSubmatch(line)
			if len(match) > 1 {
				latency, _ := strconv.ParseFloat(match[1], 64)
				latencies = append(latencies, latency)
			}
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
	return &LinuxPinger{}
}

func isAdmin() bool {
	// 检查是否有管理员权限
	return os.Geteuid() == 0
}
