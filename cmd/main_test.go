package main

import (
	"os"
	"path/filepath"
	"testing"
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
	config := loadConfig()

	if config.Log.File == "" {
		t.Error("Expected log file path to be set in the configuration")
	}
}

func TestInitLogger(t *testing.T) {
	logFilePath := "logs/test.log"
	logger := initLogger(logFilePath)
	defer logger.Close()

	if logger == nil {
		t.Error("Expected logger to be initialized")
	}

	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Errorf("Expected log file to be created at %s", logFilePath)
	}

	os.RemoveAll(filepath.Dir(logFilePath))
}
