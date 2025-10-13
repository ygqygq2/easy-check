//go:build darwin

package machineid

import (
	"fmt"
	"os/exec"
	"strings"
)

func getMachineID() (string, error) {
	// 方法1: 使用 ioreg 获取 IOPlatformUUID
	if id, err := getIDFromIOReg(); err == nil && id != "" {
		return id, nil
	}

	// 方法2: 使用 system_profiler 获取硬件 UUID
	if id, err := getIDFromSystemProfiler(); err == nil && id != "" {
		return id, nil
	}

	// 所有方法都失败，返回错误让主函数使用备用方案
	return "", fmt.Errorf("all methods to get machine ID failed")
}

// getIDFromIOReg 使用 ioreg 命令获取 UUID
func getIDFromIOReg() (string, error) {
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "IOPlatformUUID") {
			parts := strings.Split(line, "\"")
			if len(parts) > 3 {
				uuid := parts[3]
				if uuid != "" {
					return uuid, nil
				}
			}
		}
	}
	return "", fmt.Errorf("IOPlatformUUID not found")
}

// getIDFromSystemProfiler 使用 system_profiler 获取硬件 UUID
func getIDFromSystemProfiler() (string, error) {
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Hardware UUID") || strings.Contains(line, "UUID") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				uuid := strings.TrimSpace(parts[1])
				if uuid != "" {
					return uuid, nil
				}
			}
		}
	}
	return "", fmt.Errorf("hardware UUID not found")
}
