//go:build darwin
// +build darwin

package checker

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type DarwinPinger struct{}

func (p *DarwinPinger) Ping(host string, count int, timeout int) (string, error) {
	// macOS 使用的 ping 命令参数
	cmd := exec.Command("ping", "-c", fmt.Sprintf("%d", count), "-W", fmt.Sprintf("%d", timeout), host)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (p *DarwinPinger) ParsePingOutput(lines []string, count int) (int, float64, float64, float64) {
	var successCount int
	var minLatency, avgLatency, maxLatency float64

	// 使用正则表达式解析 ping 输出
	latencyRegex := regexp.MustCompile(`min/avg/max/.+=\s*([\d.]+)/([\d.]+)/([\d.]+)`)
	for _, line := range lines {
		if strings.Contains(line, "bytes from") {
			successCount++
		}
		if matches := latencyRegex.FindStringSubmatch(line); matches != nil {
			minLatency, _ = strconv.ParseFloat(matches[1], 64)
			avgLatency, _ = strconv.ParseFloat(matches[2], 64)
			maxLatency, _ = strconv.ParseFloat(matches[3], 64)
		}
	}

	return successCount, minLatency, avgLatency, maxLatency
}

func NewPinger() Pinger {
	return &DarwinPinger{}
}
