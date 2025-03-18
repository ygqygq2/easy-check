package main

import (
	"easy-check/internal/checker"
	"easy-check/internal/logger"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
    // 切换到项目根目录
    changeToProjectRoot()

    // 加载配置
    config := loadConfig()

    // 初始化日志记录器
    logConfig := logger.Config{
        File:         config.Log.File,
        MaxSize:      config.Log.MaxSize,
        MaxAge:       config.Log.MaxAge,
        MaxBackups:   config.Log.MaxBackups,
        Compress:     config.Log.Compress,
        ConsoleLevel: config.Log.ConsoleLevel,
        FileLevel:    config.Log.FileLevel,
    }
    logger := logger.NewLogger(logConfig)
    defer logger.Close()

    logger.Log("Starting easy-check...")
    logger.Log("Configuration loaded successfully")

    // 初始化 pinger 和 checker
    pinger := checker.NewPinger()
    chk := checker.NewChecker(config.Hosts, config.Interval, config.Ping.Count, config.Ping.Timeout, pinger, logger)

    // 执行初始 ping 检查
    logger.Log("Performing initial ping check")
    chk.PingHosts()

    // 启动定期 ping 检查
    startPeriodicPingChecks(chk, config.Interval, logger)
}

func changeToProjectRoot() {
    projectRoot := filepath.Dir(filepath.Dir(os.Args[0]))
    err := os.Chdir(projectRoot)
    if err != nil {
        log.Fatalf("Error changing directory to project root: %v", err)
    }

    cwd, err := os.Getwd()
    if err != nil {
        log.Fatalf("Error getting current working directory: %v", err)
    }
    log.Printf("Current working directory: %s\n", cwd)
}

func loadConfig() *checker.Config {
    configPath := filepath.Join("configs", "config.yaml")
    config, err := checker.LoadConfig(configPath)
    if err != nil {
        log.Fatalf("Error loading configuration: %v", err)
    }

    if config.Log.File == "" {
        log.Fatalf("Log file path is empty in the configuration")
    }

    return config
}

func startPeriodicPingChecks(chk *checker.Checker, interval int, logger *logger.Logger) {
    ticker := time.NewTicker(time.Duration(interval) * time.Second)
    defer ticker.Stop()

    logger.Log("Starting periodic ping checks")

    if isWindows() {
        // 创建一个停止通道，用于通知程序退出
        stopChan := make(chan struct{})
        // 创建一个信号通道
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

        // 处理 ping 检查的 goroutine
        go func() {
            for {
                select {
                case <-ticker.C:
                    chk.PingHosts()
                case <-stopChan:
                    return
                }
            }
        }()

        // 处理信号的 goroutine
        go func() {
            for {
                sig := <-sigChan
                exePath := getExecutablePath()
                fmt.Printf("\nReceived signal: %v. Executable path: %s\n", sig, exePath)
                fmt.Println("Waiting for some seconds before exit...")

                time.Sleep(5 * time.Second)
            }
        }()

        // 主线程保持运行
        select {}
    } else {
        for range ticker.C {
            chk.PingHosts()
        }
    }
}

func getExecutablePath() string {
    cwd, err := os.Getwd()
    if err != nil {
        log.Fatalf("Error getting current working directory: %v", err)
    }
    return filepath.Join(cwd, "bin", "easy-check.exe")
}

func isWindows() bool {
    return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}
