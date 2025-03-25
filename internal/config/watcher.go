package config

import (
	"easy-check/internal/logger"
	"fmt"

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

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
				logger.Log(fmt.Sprintf("Config file changed: %s", event.Name))
				newConfig, err := LoadConfig(configPath)
				if err != nil {
					logger.Log(fmt.Sprintf("Error reloading configuration: %v", err))
				} else {
					onChange(newConfig)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.Log(fmt.Sprintf("File watcher error: %v", err))
		}
	}
}
