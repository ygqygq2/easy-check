package heartbeat

import (
	"context"
	"time"
)

var globalReporter *Reporter

// Initialize 初始化心跳模块
func Initialize(logger Logger, machineID, version, serverBaseURL, heartbeatAPI string, enabled bool, interval, timeout int) error {
	if !enabled {
		logger.Log("心跳上报已禁用", "info")
		return nil
	}

	config := &Config{
		Enabled:   enabled,
		ServerURL: serverBaseURL + heartbeatAPI,
		Interval:  time.Duration(interval) * time.Second,
		Timeout:   time.Duration(timeout) * time.Second,
	}

	globalReporter = NewReporter(config, logger, machineID, version)

	// 在后台启动心跳上报
	go globalReporter.Start(context.Background())

	return nil
}

// Stop 停止心跳上报
func Stop() {
	if globalReporter != nil {
		globalReporter.Stop()
	}
}
