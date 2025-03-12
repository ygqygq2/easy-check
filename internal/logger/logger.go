package logger

import (
	"fmt"
	"os"
	"time"
)

type Logger struct {
	file *os.File
}

func NewLogger(filePath string) (*Logger, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	return &Logger{file: file}, nil
}

func (l *Logger) Log(message string) error {
	timestamp := time.Now().Format(time.RFC3339)
	logMessage := timestamp + " - " + message + "\n"

	// 写入日志文件
	_, err := l.file.WriteString(logMessage)
	if err != nil {
		return err
	}

	// 打印到控制台
	fmt.Print(logMessage)

	return nil
}

func (l *Logger) Close() error {
	return l.file.Close()
}
