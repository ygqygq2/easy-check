package utils

import (
	"fmt"
	"time"
)

// FormatTime 格式化时间字符串
func FormatTime(timestamp string) string {
	if timestamp == "" {
		return "" // 如果时间为空，返回空字符串
	}

	// 假设 JSON 中的时间格式为 "2006-01-02T15:04:05Z07:00"（ISO 8601 格式）
	parsedTime, err := time.Parse("2006-01-02T15:04:05Z07:00", timestamp)
	if err != nil {
		// 如果解析失败，记录错误并返回原始时间字符串
		fmt.Printf("Failed to parse time: %v\n", err)
		return timestamp
	}

	// 格式化为目标时间格式
	return parsedTime.Format("2006-01-02 15:04:05")
}

// IsDirectorySuffix 检查路径是否以 "/" 结尾
func IsDirectorySuffix(path string) bool {
	if len(path) == 0 {
		return false
	}
	return path[len(path)-1] == '/'
}

// 目录添加 "/"
func AddDirectorySuffix(path string) string {
	if IsDirectorySuffix(path) {
		return path
	}
	return path + "/"
}
