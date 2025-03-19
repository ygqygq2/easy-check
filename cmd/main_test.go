package main

import (
	"easy-check/internal/logger"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestChangeToProjectRoot(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting current working directory: %v", err)
	}
	defer os.Chdir(originalDir)

	changeToProjectRoot()

	projectRoot := filepath.Dir(filepath.Dir(os.Args[0]))
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting current working directory: %v", err)
	}

	if cwd != projectRoot {
		t.Errorf("Expected current working directory to be %s, but got %s", projectRoot, cwd)
	}
}

func TestLoadConfig(t *testing.T) {
	configPath := filepath.Join("..", "configs", "config.yaml")
	config := loadConfig(configPath)

	if config.Log.File == "" {
		t.Error("Expected log file path to be set in the configuration")
	}
}

func TestInitLogger(t *testing.T) {
	configPath := filepath.Join("..", "configs", "config.yaml")
	config := loadConfig(configPath)

	// 使用临时目录存放日志文件
	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "test.log")

	logConfig := logger.Config{
		File:         logFilePath,
		MaxSize:      config.Log.MaxSize,
		MaxAge:       config.Log.MaxAge,
		MaxBackups:   config.Log.MaxBackups,
		Compress:     config.Log.Compress,
		ConsoleLevel: config.Log.ConsoleLevel,
		FileLevel:    config.Log.FileLevel,
	}
	logger := logger.NewLogger(logConfig)
	defer logger.Close()

	if logger == nil {
		t.Error("Expected logger to be initialized")
	}

	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Errorf("Expected log file to be created at %s", logFilePath)
	}
}

func TestLogRotation(t *testing.T) {
	// 使用临时目录存放日志文件
	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "test-rotation.log")

	config := logger.Config{
		File:         logFilePath,
		MaxSize:      1, // MB
		MaxAge:       7, // days
		MaxBackups:   100,
		Compress:     false,
		ConsoleLevel: "error",
		FileLevel:    "info",
	}

	log := logger.NewLogger(config)
	defer log.Close()

	// 生成大量日志条目
	for i := 0; i < 10000; i++ {
		log.Log(fmt.Sprintf("This is a test log entry number %d", i), "info")
		time.Sleep(1 * time.Millisecond) // 可选：添加延迟以模拟实际日志生成
	}

	// 检查日志文件是否被截断
	files, err := filepath.Glob(logFilePath + "*")
	if err != nil {
		t.Fatalf("Error checking log files: %v", err)
	}

	if len(files) <= 1 {
		t.Errorf("Expected log rotation to occur, but found only %d log file(s)", len(files))
	}
}
