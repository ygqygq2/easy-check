package config

import (
	"easy-check/internal/logger"
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
)

func WatchConfigFile(configPath string, logger *logger.Logger, onChange func(newConfig *Config)) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Log(fmt.Sprintf("Error creating file watcher: %v", err))
		return
	}
	defer watcher.Close()

	err = watcher.Add(configPath)
	if err != nil {
		logger.Log(fmt.Sprintf("Error adding file to watcher: %v", err))
		return
	}

	var debounceTimer *time.Timer
	const debounceDelay = 500 * time.Millisecond

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
				// 如果已有定时器，先停止它
				if debounceTimer != nil {
					debounceTimer.Stop()
				}

				// 创建新的定时器
				debounceTimer = time.AfterFunc(debounceDelay, func() {
					logger.Log(fmt.Sprintf("Config file changed: %s", event.Name), "info")
					newConfig, err := LoadConfig(configPath, logger)
					if err != nil {
						logger.Log(fmt.Sprintf("Error reloading configuration: %v", err), "error")
					} else {
						onChange(newConfig)
					}
				})
			} else if event.Op&fsnotify.Remove != 0 {
				logger.Log(fmt.Sprintf("Config file removed: %s. Using previous configuration.", event.Name), "warn")
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.Log(fmt.Sprintf("File watcher error: %v", err), "error")
		}
	}
}
