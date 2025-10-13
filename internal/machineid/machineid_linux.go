//go:build linux

package machineid

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func getMachineID() (string, error) {
	// 方法1: 尝试读取 /etc/machine-id
	if id, err := readMachineIDFile("/etc/machine-id"); err == nil && id != "" {
		return id, nil
	}

	// 方法2: 尝试读取 /var/lib/dbus/machine-id
	if id, err := readMachineIDFile("/var/lib/dbus/machine-id"); err == nil && id != "" {
		return id, nil
	}

	// 方法3: 尝试使用 cat 命令
	if id, err := readMachineIDWithCat(); err == nil && id != "" {
		return id, nil
	}

	// 所有方法都失败，返回错误让主函数使用备用方案
	return "", fmt.Errorf("all methods to get machine ID failed")
}

// readMachineIDFile 直接读取机器 ID 文件
func readMachineIDFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	id := strings.TrimSpace(string(data))
	if id == "" {
		return "", fmt.Errorf("empty machine ID")
	}
	return id, nil
}

// readMachineIDWithCat 使用 cat 命令读取机器 ID
func readMachineIDWithCat() (string, error) {
	cmd := exec.Command("cat", "/etc/machine-id")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	id := strings.TrimSpace(string(output))
	if id == "" {
		return "", fmt.Errorf("empty machine ID")
	}
	return id, nil
}
