package machineid

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// GetMachineID 获取机器唯一 ID
func GetMachineID() (string, error) {
	// 首先尝试从系统获取 machine ID
	id, err := getMachineID()
	if err != nil {
		// 如果系统方法失败，使用备用方案：从文件读取或生成新的 ID
		id, err = getFallbackMachineID()
		if err != nil {
			return "", fmt.Errorf("failed to get machine ID: %w", err)
		}
	}

	// 对硬件信息进行哈希处理，生成固定长度的唯一 ID
	hash := sha256.Sum256([]byte(id))
	return hex.EncodeToString(hash[:]), nil
}

// getFallbackMachineID 获取备用机器 ID（从文件读取或生成新的）
func getFallbackMachineID() (string, error) {
	// 获取配置目录路径
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}

	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	machineIDFile := filepath.Join(configDir, "machine-id")

	// 尝试读取现有的 machine-id 文件
	if data, err := os.ReadFile(machineIDFile); err == nil {
		id := string(data)
		if len(id) > 0 {
			return id, nil
		}
	}

	// 生成新的 machine ID
	id, err := generateMachineID()
	if err != nil {
		return "", fmt.Errorf("failed to generate machine ID: %w", err)
	}

	// 保存到文件
	if err := os.WriteFile(machineIDFile, []byte(id), 0644); err != nil {
		return "", fmt.Errorf("failed to save machine ID: %w", err)
	}

	return id, nil
}

// getConfigDir 获取配置目录路径
func getConfigDir() (string, error) {
	// 优先使用用户配置目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// 如果无法获取用户目录，使用临时目录
		return filepath.Join(os.TempDir(), ".easy-check"), nil
	}
	return filepath.Join(homeDir, ".easy-check"), nil
}

// generateMachineID 生成随机的机器 ID
func generateMachineID() (string, error) {
	// 生成 32 字节的随机数据
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	return hex.EncodeToString(randomBytes), nil
}
