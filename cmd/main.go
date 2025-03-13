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
	logger := initLogger(config.Log.File)
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
	log.Printf("Current working directory: %s", cwd)
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

func initLogger(logFilePath string) *logger.Logger {
	logDir := filepath.Dir(logFilePath)
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatalf("Error creating log directory: %v", err)
	}

	logger, err := logger.NewLogger(logFilePath)
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}

	return logger
}

func startPeriodicPingChecks(chk *checker.Checker, interval int, logger *logger.Logger) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	logger.Log("Starting periodic ping checks")

	if isWindows() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			for {
				select {
				case <-ticker.C:
					chk.PingHosts()
				case sig := <-sigChan:
					fmt.Printf("\nReceived signal: %v. Are you sure you want to exit? (y/n): ", sig)
					var response string
					fmt.Scanln(&response)
					if response == "y" || response == "Y" {
						logger.Log("Shutting down easy-check...")
						os.Exit(0)
					} else {
						fmt.Println("Continuing...")
					}
				}
			}
		}()
	} else {
		for range ticker.C {
			chk.PingHosts()
		}
	}

	select {}
}

func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}
