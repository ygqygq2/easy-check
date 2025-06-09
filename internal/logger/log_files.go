package logger

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetLogFiles 获取日志目录中的所有日志文件
func GetLogFiles(logDir string) ([]string, error) {
	files, err := os.ReadDir(logDir)
	if err != nil {
		return nil, fmt.Errorf("无法读取日志目录: %v", err)
	}

	var logFiles []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".txt" {
			logFiles = append(logFiles, file.Name())
		}
	}

	return logFiles, nil
}
