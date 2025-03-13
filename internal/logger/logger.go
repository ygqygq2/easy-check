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

func (l *Logger) formatMessage(message string) string {
	timestamp := time.Now().Format(time.RFC3339)
	return timestamp + " - " + message + "\n"
}

func (l *Logger) Log(message string) error {
	logMessage := l.formatMessage(message)

	// 写入日志文件
	_, err := l.file.WriteString(logMessage)
	if err != nil {
		return err
	}

	return nil
}

func (l *Logger) Console(message string) {
	logMessage := l.formatMessage(message)

	// 打印到控制台
	fmt.Print(logMessage)
}

func (l *Logger) Close() error {
	return l.file.Close()
}
