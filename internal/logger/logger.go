package logger

import (
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
    _, err := l.file.WriteString(logMessage)
    return err
}

func (l *Logger) Close() error {
    return l.file.Close()
}