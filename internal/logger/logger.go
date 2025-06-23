package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*logrus.Logger
	consoleLogger *logrus.Logger
	fileLevel     logrus.Level
	consoleLevel  logrus.Level
	fileLogger    *lumberjack.Logger
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
	// 使用自定义时间戳格式
	timestamp := entry.Time.Local().Format(f.TimestampFormat)
	msg := fmt.Sprintf("%s [%s] %s\n", timestamp, entry.Level.String(), entry.Message)
	return []byte(msg), nil
}

func NewLogger(config Config) *Logger {
	log := logrus.New()
	consoleLogger := logrus.New()

	// 设置日志输出到文件并进行轮转
	fileLogger := &lumberjack.Logger{
		Filename:   config.File,
		MaxSize:    config.MaxSize,    // MB
		MaxAge:     config.MaxAge,     // days
		MaxBackups: config.MaxBackups, // number of backups
		Compress:   config.Compress,   // compress old logs
	}
	log.SetOutput(fileLogger)

	// 设置自定义日志格式
	formatter := &CustomFormatter{
		TimestampFormat: "2006-01-02 15:04:05", // 自定义时间戳格式
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

	return &Logger{log, consoleLogger, fileLevel, consoleLevel, fileLogger}
}

// NewDefaultLogger 创建一个默认的日志记录器，仅输出到控制台
func NewDefaultLogger() *Logger {
	config := Config{
		ConsoleLevel: "debug",
		FileLevel:    "info",
	}
	return NewLogger(config)
}

// UpdateConfig 更新日志记录器的配置
func (l *Logger) UpdateConfig(config Config) {
	// 如果要更新文件日志，先关闭现有的文件
	if l.fileLogger != nil {
		l.fileLogger.Close()
	}

	// 如果配置了文件日志，初始化它
	if config.File != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(config.File)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			err = os.MkdirAll(logDir, 0755)
			if err != nil {
				fmt.Printf("Error creating log directory: %v\n", err)
			}
		}

		l.fileLogger = &lumberjack.Logger{
			Filename:   config.File,
			MaxSize:    config.MaxSize,
			MaxAge:     config.MaxAge,
			MaxBackups: config.MaxBackups,
			Compress:   config.Compress,
		}
		l.Logger.SetOutput(l.fileLogger)
	}

	// 更新日志级别
	l.consoleLevel = getLevelFromString(config.ConsoleLevel)
	l.fileLevel = getLevelFromString(config.FileLevel)
	l.Logger.SetLevel(l.fileLevel)
	l.consoleLogger.SetLevel(l.consoleLevel)
}

func getLevelFromString(levelStr string) logrus.Level {
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		level = logrus.InfoLevel
	}
	return level
}

func (l *Logger) Log(message string, level ...string) {
	logLevel := l.fileLevel
	if len(level) > 0 {
		parsedLevel, err := logrus.ParseLevel(level[0])
		if err == nil {
			logLevel = parsedLevel
		} else {
			fmt.Printf("Error parsing log level: %v\n", err)
		}
	}

	// 根据日志级别分别输出到控制台和文件
	if logLevel <= l.consoleLevel {
		l.consoleLogger.Log(logLevel, message)
	}
	if logLevel <= l.fileLevel {
		l.Logger.Log(logLevel, message)
	}
}

func (l *Logger) Fatal(message string, level ...string) {
	l.Log(message, level...)
	os.Exit(1)
}

// LogAndError 记录日志并返回错误
func (l *Logger) LogAndError(message string, level string, args ...interface{}) error {
	formattedMessage := fmt.Sprintf(message, args...) // 格式化消息
	l.Log(formattedMessage, level)                    // 使用 Logger 的 Log 方法记录日志
	return fmt.Errorf("%s", formattedMessage)         // 返回格式化后的错误
}

func (l *Logger) Close() error {
	// logrus 没有显式的关闭方法，这里可以留空
	return nil
}

// WailsLogger 实现 Wails 的 logger.Logger 接口
type WailsLogger struct {
	*Logger
}

// Print 实现 Wails Logger 的 Print 方法
func (wl *WailsLogger) Print(message string) {
	wl.Log(message, "info")
}

// Trace 实现 Wails Logger 的 Trace 方法
func (wl *WailsLogger) Trace(message string) {
	wl.Log(message, "trace")
}

// Debug 实现 Wails Logger 的 Debug 方法
func (wl *WailsLogger) Debug(message string) {
	wl.Log(message, "debug")
}

// Info 实现 Wails Logger 的 Info 方法
func (wl *WailsLogger) Info(message string) {
	wl.Log(message, "info")
}

// Warning 实现 Wails Logger 的 Warning 方法
func (wl *WailsLogger) Warning(message string) {
	wl.Log(message, "warning")
}

// Error 实现 Wails Logger 的 Error 方法
func (wl *WailsLogger) Error(message string) {
	wl.Log(message, "error")
}

// Fatal 实现 Wails Logger 的 Fatal 方法
func (wl *WailsLogger) Fatal(message string) {
	wl.Logger.Fatal(message)
}
