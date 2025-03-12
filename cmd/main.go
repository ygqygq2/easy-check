package main

import (
	"easy-check/internal/checker"
	"easy-check/internal/logger"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// 切换到项目根目录
	projectRoot := filepath.Dir(filepath.Dir(os.Args[0]))
	err := os.Chdir(projectRoot)
	if err != nil {
		log.Fatalf("Error changing directory to project root: %v", err)
	}

	// 打印当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}
	log.Printf("Current working directory: %s", cwd)

	// Load configuration
	configPath := filepath.Join("configs", "config.yaml")
	config, err := checker.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	if config.Log.File == "" {
		log.Fatalf("Log file path is empty in the configuration")
	}

	// Ensure the log file directory exists
	logDir := filepath.Dir(config.Log.File)
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatalf("Error creating log directory: %v", err)
	}

	// Initialize logger
	logFile, err := os.OpenFile(config.Log.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()

	logger, err := logger.NewLogger(config.Log.File)
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}

	logger.Log("Starting easy-check...")

	logger.Log("Configuration loaded successfully")

	// Initialize pinger
	pinger := checker.NewPinger()

	// Initialize checker
	chk := checker.NewChecker(config.Hosts, config.Interval, config.Ping.Count, config.Ping.Timeout, pinger, logger)

	// Perform an initial ping check
	logger.Log("Performing initial ping check")
	chk.PingHosts()

	// Start periodic ping checks
	ticker := time.NewTicker(time.Duration(config.Interval) * time.Second)
	defer ticker.Stop()

	logger.Log("Starting periodic ping checks")

	for range ticker.C {
		chk.PingHosts()
	}
}
