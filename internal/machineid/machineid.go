package machineid

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GetMachineID 获取机器唯一 ID
func GetMachineID() (string, error) {
	id, err := getMachineID()
	if err != nil {
		return "", fmt.Errorf("failed to get machine ID: %w", err)
	}

	if err != nil {
		return "", err
	}

	// 对硬件信息进行哈希处理，生成固定长度的唯一 ID
	hash := sha256.Sum256([]byte(id))
	return hex.EncodeToString(hash[:]), nil
}
