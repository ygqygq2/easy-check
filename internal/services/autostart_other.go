//go:build !windows

package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// IsAutoStartEnabled 检查开机自启是否已启用
func (s *AppService) IsAutoStartEnabled() (bool, error) {
	if runtime.GOOS == "linux" {
		// 检查 systemd 服务是否启用
		cmd := exec.Command("systemctl", "is-enabled", "easy-check-ui")
		err := cmd.Run()
		return err == nil, nil
	}
	
	// macOS 和其他系统暂不支持
	return false, fmt.Errorf("开机自启功能暂不支持 %s 平台", runtime.GOOS)
}

// EnableAutoStart 启用开机自启
func (s *AppService) EnableAutoStart() error {
	if runtime.GOOS == "linux" {
		return s.enableAutoStartLinux()
	}
	
	return fmt.Errorf("开机自启功能暂不支持 %s 平台", runtime.GOOS)
}

// DisableAutoStart 禁用开机自启
func (s *AppService) DisableAutoStart() error {
	if runtime.GOOS == "linux" {
		return s.disableAutoStartLinux()
	}
	
	return fmt.Errorf("开机自启功能暂不支持 %s 平台", runtime.GOOS)
}

// enableAutoStartLinux Linux 平台启用开机自启
func (s *AppService) enableAutoStartLinux() error {
	// 获取当前程序路径
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %w", err)
	}
	
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("解析程序路径失败: %w", err)
	}
	
	workingDir := filepath.Dir(exePath)
	serviceName := "easy-check-ui"
	serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	
	// 创建 systemd 服务文件内容
	serviceContent := fmt.Sprintf(`[Unit]
Description=Easy-Check UI Network Monitor
After=network.target

[Service]
Type=simple
ExecStart=%s
Restart=always
User=%s
WorkingDirectory=%s

[Install]
WantedBy=multi-user.target
`, exePath, os.Getenv("USER"), workingDir)
	
	// 写入服务文件（需要 root 权限）
	err = os.WriteFile(serviceFile, []byte(serviceContent), 0644)
	if err != nil {
		return fmt.Errorf("创建服务文件失败（可能需要 root 权限）: %w", err)
	}
	
	// 重新加载 systemd
	cmd := exec.Command("systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("重新加载 systemd 失败: %w", err)
	}
	
	// 启用服务
	cmd = exec.Command("systemctl", "enable", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("启用服务失败: %w", err)
	}
	
	s.appCtx.Logger.Log("Linux 开机自启已启用", "info")
	return nil
}

// disableAutoStartLinux Linux 平台禁用开机自启
func (s *AppService) disableAutoStartLinux() error {
	serviceName := "easy-check-ui"
	
	// 停止服务
	cmd := exec.Command("systemctl", "stop", serviceName)
	cmd.Run() // 忽略错误，服务可能未运行
	
	// 禁用服务
	cmd = exec.Command("systemctl", "disable", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("禁用服务失败: %w", err)
	}
	
	// 删除服务文件
	serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	os.Remove(serviceFile) // 忽略错误
	
	// 重新加载 systemd
	cmd = exec.Command("systemctl", "daemon-reload")
	cmd.Run() // 忽略错误
	
	s.appCtx.Logger.Log("Linux 开机自启已禁用", "info")
	return nil
}

// GetAutoStartInfo 获取开机自启的详细信息
func (s *AppService) GetAutoStartInfo() (map[string]interface{}, error) {
	enabled, _ := s.IsAutoStartEnabled()
	
	info := map[string]interface{}{
		"enabled":  enabled,
		"platform": runtime.GOOS,
		"support":  runtime.GOOS == "linux",
	}
	
	if runtime.GOOS == "linux" {
		exePath, _ := os.Executable()
		info["executablePath"] = exePath
		info["method"] = "systemd"
	}
	
	return info, nil
}
