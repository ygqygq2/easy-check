package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
    *logrus.Logger
    consoleLogger *logrus.Logger
    fileLevel     logrus.Level
    consoleLevel  logrus.Level
}

type Config struct {
    File         string
    MaxSize      int
    MaxAge       int
    MaxBackups   int
    Compress     bool
    ConsoleLevel string
    FileLevel    string
}

type CustomFormatter struct {
    TimestampFormat string
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
    timestamp := entry.Time.Format(f.TimestampFormat)
    msg := fmt.Sprintf("%s [%s] %s\n", timestamp, entry.Level.String(), entry.Message)
    return []byte(msg), nil
}

func NewLogger(config Config) *Logger {
    log := logrus.New()
    consoleLogger := logrus.New()

    // 设置日志输出到文件并进行轮转
    log.SetOutput(&lumberjack.Logger{
        Filename:   config.File,
        MaxSize:    config.MaxSize,    // MB
        MaxAge:     config.MaxAge,     // days
        MaxBackups: config.MaxBackups, // number of backups
        Compress:   config.Compress,   // compress old logs
    })

    // 设置自定义日志格式
    formatter := &CustomFormatter{
        TimestampFormat: time.RFC3339,
    }
    log.SetFormatter(formatter)
    consoleLogger.SetFormatter(formatter)

    // 设置文件日志级别
    fileLevel, err := logrus.ParseLevel(config.FileLevel)
    if err != nil {
        fileLevel = logrus.InfoLevel
    }
    log.SetLevel(fileLevel)

    // 设置控制台日志级别
    consoleLevel, err := logrus.ParseLevel(config.ConsoleLevel)
    if err != nil {
        consoleLevel = logrus.ErrorLevel
    }
    consoleLogger.SetLevel(consoleLevel)

    // 设置控制台输出
    consoleLogger.SetOutput(os.Stdout)

    return &Logger{log, consoleLogger, fileLevel, consoleLevel}
}

func (l *Logger) Log(message string, level ...string) {
    logLevel := l.fileLevel
    if len(level) > 0 {
        parsedLevel, err := logrus.ParseLevel(level[0])
        if err == nil {
            logLevel = parsedLevel
        }
    }
    l.Logger.Log(logLevel, message)
}

func (l *Logger) Console(message string, level ...string) {
    logLevel := l.consoleLevel
    if len(level) > 0 {
        parsedLevel, err := logrus.ParseLevel(level[0])
        if err == nil {
            logLevel = parsedLevel
        }
    }
    l.consoleLogger.Log(logLevel, message)
}

func (l *Logger) Close() error {
    // logrus 没有显式的关闭方法，这里可以留空
    return nil
}
