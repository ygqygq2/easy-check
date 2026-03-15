package heartbeat

import "time"

// HeartbeatData 定义上报的心跳数据结构
type HeartbeatData struct {
	MachineID        string                 `json:"machineId"`
	Hostname         string                 `json:"hostname"`
	ExternalIP       string                 `json:"externalIp"`
	Version          string                 `json:"version"`
	OSVersion        string                 `json:"osVersion"`
	LastCheckResults map[string]CheckResult `json:"lastCheckResults,omitempty"`
	Timestamp        int64                  `json:"timestamp"`
}

// CheckResult 定义单个主机的检查结果
type CheckResult struct {
	Host        string  `json:"host"`
	Description string  `json:"description"`
	IsReachable bool    `json:"isReachable"`
	LossRate    float64 `json:"lossRate"`
	LastCheck   string  `json:"lastCheck"` // ISO 8601 格式
}

// Config 心跳配置
type Config struct {
	Enabled   bool
	ServerURL string
	Interval  time.Duration
	Timeout   time.Duration
}
