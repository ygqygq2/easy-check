//go:build linux
// +build linux

package checker

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type LinuxPinger struct{}

func (p *LinuxPinger) Ping(host string, count int, timeout int) (string, error) {
	cmd := exec.Command("ping", "-4", "-c", fmt.Sprintf("%d", count), "-W", fmt.Sprintf("%d", timeout), host)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (p *LinuxPinger) ParsePingOutput(lines []string, count int) (int, string) {
	successCount := 0
	var sampleLatency string

	// Linux平台的ping输出解析逻辑
	re := regexp.MustCompile(`time=\d+(\.\d+)? ms`)

	for _, line := range lines {
		if strings.Contains(line, "ttl=") || strings.Contains(line, "time=") {
			successCount++
			if sampleLatency == "" {
				match := re.FindString(line)
				if match != "" {
					sampleLatency = match
				}
			}
		}
	}

	return successCount, sampleLatency
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
