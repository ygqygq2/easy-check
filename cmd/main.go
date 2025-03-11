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
    // Load configuration
    configPath := filepath.Join("configs", "config.yaml")
    config, err := checker.LoadConfig(configPath)
    if (err != nil) {
        log.Fatalf("Error loading configuration: %v", err)
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

    logger.Log("Configuration loaded successfully.")

    // Initialize pinger
    pinger := checker.NewPinger()

    // Initialize checker
    chk := checker.NewChecker(config.Hosts, config.Interval, config.Ping.Count, config.Ping.Timeout, pinger)

    // Start periodic ping checks
    ticker := time.NewTicker(time.Duration(config.Interval) * time.Second)
    defer ticker.Stop()

    logger.Log("Starting periodic ping checks.")

    for range ticker.C {
        results := chk.PingHosts()
        logger.Log(results)
        package main
    
    import (
        "easy-check/internal/checker"
        "easy-check/internal/logger"
        "fmt"
        "log"
        "os"
        "path/filepath"
        "time"
    )
    
    func main() {
        // 打印当前工作目录
        cwd, err := os.Getwd()
        if err != nil {
            log.Fatalf("Error getting current working directory: %v", err)
        }
        fmt.Println("🚀 ~ Current working directory:", cwd)
    
        // Load configuration
        configPath := filepath.Join("configs", "config.yaml")
        fmt.Println("🚀 ~ file: main.go:15 ~ configPath:", configPath)
        config, err := checker.LoadConfig(configPath)
        if err != nil {
            log.Fatalf("Error loading configuration: %v", err)
        }
    
        // Debug: Print the log file path
        log.Printf("Log file path from config: %s", config.Log.File)
    
        // Ensure the log file path is not empty
        fmt.Printf("%+v\n", config)
    
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
    
        logger.Log("Configuration loaded successfully.")
    
        // Initialize pinger
        pinger := checker.NewPinger()
    
        // Initialize checker
        chk := checker.NewChecker(config.Hosts, config.Interval, config.Ping.Count, config.Ping.Timeout, pinger, logger)
    
        // Start periodic ping checks
        ticker := time.NewTicker(time.Duration(config.Interval) * time.Second)
        defer ticker.Stop()
    
        logger.Log("Starting periodic ping checks.")
    
        for {
            select {
            case <-ticker.C:
                chk.PingHosts()
            }
        }
    }}
}
