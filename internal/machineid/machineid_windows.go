//go:build windows

package machineid

import (
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows"
)

func getMachineID() (string, error) {
	// 方法1: 尝试使用 wmic (Windows 10 及更早版本)
	id, err := getIDFromWMIC()
	if err == nil && id != "" {
		return id, nil
	}

	// 方法2: 尝试使用 PowerShell (Windows 11 备用方案)
	id, err = getIDFromPowerShell()
	if err == nil && id != "" {
		return id, nil
	}

	// 所有方法都失败，返回错误让主函数使用备用方案
	return "", fmt.Errorf("all methods to get machine ID failed")
}

// getIDFromWMIC 使用 wmic 命令获取机器 UUID
func getIDFromWMIC() (string, error) {
	cmd := exec.Command("wmic", "csproduct", "get", "UUID")
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("unexpected wmic output format")
	}
	uuid := strings.TrimSpace(lines[1])
	if uuid == "" {
		return "", fmt.Errorf("empty UUID from wmic")
	}
	return uuid, nil
}

// getIDFromPowerShell 使用 PowerShell 获取机器 UUID (Windows 11 兼容)
func getIDFromPowerShell() (string, error) {
	// 使用 PowerShell 的 Get-CimInstance 命令
	cmd := exec.Command("powershell", "-Command", 
		"(Get-CimInstance -ClassName Win32_ComputerSystemProduct).UUID")
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	uuid := strings.TrimSpace(string(output))
	if uuid == "" {
		return "", fmt.Errorf("empty UUID from PowerShell")
	}
	return uuid, nil
}
