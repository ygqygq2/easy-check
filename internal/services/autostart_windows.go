//go:build windows

package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsAutoStartEnabled 检查开机自启是否已启用（快捷方式方式）
func (s *AppService) IsAutoStartEnabled() (bool, error) {
	startupFolder := os.Getenv("APPDATA") +
		"\\Microsoft\\Windows\\Start Menu\\Programs\\Startup"
	shortcutPath := filepath.Join(startupFolder, "Easy-Check.lnk")
	
	_, err := os.Stat(shortcutPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// EnableAutoStart 启用开机自启（快捷方式方式）
func (s *AppService) EnableAutoStart() error {
	// 获取当前程序的执行路径
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %w", err)
	}

	// 解析符号链接，获取真实路径
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("解析程序路径失败: %w", err)
	}

	// 获取启动文件夹路径
	startupFolder := os.Getenv("APPDATA") +
		"\\Microsoft\\Windows\\Start Menu\\Programs\\Startup"
	
	// 确保启动文件夹存在
	if err := os.MkdirAll(startupFolder, 0755); err != nil {
		return fmt.Errorf("创建启动文件夹失败: %w", err)
	}

	shortcutPath := filepath.Join(startupFolder, "Easy-Check.lnk")
	workingDir := filepath.Dir(exePath)

	// 转义路径中的特殊字符
	escapedShortcutPath := strings.ReplaceAll(shortcutPath, "'", "''")
	escapedExePath := strings.ReplaceAll(exePath, "'", "''")
	escapedWorkingDir := strings.ReplaceAll(workingDir, "'", "''")

	// 使用 PowerShell 创建快捷方式
	psScript := fmt.Sprintf(`
		$WS = New-Object -ComObject WScript.Shell
		$SC = $WS.CreateShortcut('%s')
		$SC.TargetPath = '%s'
		$SC.WorkingDirectory = '%s'
		$SC.Description = 'Easy-Check - 网络检测工具'
		$SC.Save()
	`, escapedShortcutPath, escapedExePath, escapedWorkingDir)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("创建快捷方式失败: %w, 输出: %s", err, string(output))
	}

	s.appCtx.Logger.Log("开机自启已启用", "info")
	return nil
}

// DisableAutoStart 禁用开机自启（快捷方式方式）
func (s *AppService) DisableAutoStart() error {
	startupFolder := os.Getenv("APPDATA") +
		"\\Microsoft\\Windows\\Start Menu\\Programs\\Startup"
	shortcutPath := filepath.Join(startupFolder, "Easy-Check.lnk")

	// 删除快捷方式文件
	err := os.Remove(shortcutPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除快捷方式失败: %w", err)
	}

	s.appCtx.Logger.Log("开机自启已禁用", "info")
	return nil
}

// GetAutoStartInfo 获取开机自启的详细信息
func (s *AppService) GetAutoStartInfo() (map[string]interface{}, error) {
	enabled, err := s.IsAutoStartEnabled()
	if err != nil {
		return nil, err
	}

	startupFolder := os.Getenv("APPDATA") +
		"\\Microsoft\\Windows\\Start Menu\\Programs\\Startup"
	shortcutPath := filepath.Join(startupFolder, "Easy-Check.lnk")

	exePath, _ := os.Executable()
	exePath, _ = filepath.EvalSymlinks(exePath)

	info := map[string]interface{}{
		"enabled":        enabled,
		"method":         "shortcut",
		"shortcutPath":   shortcutPath,
		"executablePath": exePath,
		"startupFolder":  startupFolder,
	}

	return info, nil
}
