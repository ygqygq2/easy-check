package heartbeat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"
)

// Logger 日志接口
type Logger interface {
	Log(message string, level ...string)
}

// AppInfo 应用信息接口
type AppInfo interface {
	GetMachineID() string
	GetVersion() string
	GetHostname() (string, error)
}

// Reporter 心跳上报器
type Reporter struct {
	config    *Config
	logger    Logger
	appInfo   AppInfo
	client    *http.Client
	cancel    context.CancelFunc
	machineID string
	version   string
}

// NewReporter 创建新的心跳上报器
func NewReporter(config *Config, logger Logger, machineID, version string) *Reporter {
	return &Reporter{
		config:    config,
		logger:    logger,
		machineID: machineID,
		version:   version,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Start 启动心跳上报
func (r *Reporter) Start(ctx context.Context) {
	if !r.config.Enabled {
		r.logger.Log("心跳上报已禁用", "info")
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel

	// 立即发送一次心跳
	r.sendHeartbeat()

	// 启动定时器
	ticker := time.NewTicker(r.config.Interval)
	defer ticker.Stop()

	r.logger.Log(fmt.Sprintf("心跳上报已启动，间隔：%v", r.config.Interval), "info")

	for {
		select {
		case <-ctx.Done():
			r.logger.Log("心跳上报已停止", "info")
			return
		case <-ticker.C:
			r.sendHeartbeat()
		}
	}
}

// Stop 停止心跳上报
func (r *Reporter) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
}

// sendHeartbeat 发送心跳数据
func (r *Reporter) sendHeartbeat() {
	data := r.collectHeartbeatData()

	jsonData, err := json.Marshal(data)
	if err != nil {
		r.logger.Log(fmt.Sprintf("心跳数据序列化失败: %v", err), "error")
		return
	}

	req, err := http.NewRequest("POST", r.config.ServerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		r.logger.Log(fmt.Sprintf("创建心跳请求失败: %v", err), "error")
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		r.logger.Log(fmt.Sprintf("发送心跳失败: %v", err), "error")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		r.logger.Log(fmt.Sprintf("心跳上报失败，状态码: %d", resp.StatusCode), "error")
		return
	}

	r.logger.Log("心跳上报成功", "debug")
}

// collectHeartbeatData 收集心跳数据
func (r *Reporter) collectHeartbeatData() *HeartbeatData {
	// 获取外网 IP
	externalIP := getExternalIP(r.client)

	// 获取主机名
	hostname, _ := getHostname()

	// 获取 OS 版本信息
	osVersion := fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)

	return &HeartbeatData{
		MachineID:        r.machineID,
		Hostname:         hostname,
		ExternalIP:       externalIP,
		Version:          r.version,
		OSVersion:        osVersion,
		LastCheckResults: r.getLastCheckResults(),
		Timestamp:        time.Now().Unix(),
	}
}

// getExternalIP 获取外网 IP
func getExternalIP(client *http.Client) string {
	// 使用多个服务尝试获取外网 IP
	services := []string{
		"https://api.ipify.org?format=text",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
	}

	for _, service := range services {
		resp, err := client.Get(service)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			ip := buf.String()
			if len(ip) > 0 {
				return ip
			}
		}
	}

	return "unknown"
}

// getHostname 获取主机名
func getHostname() (string, error) {
	return os.Hostname()
}

// getLastCheckResults 获取最近的检查结果
func (r *Reporter) getLastCheckResults() map[string]CheckResult {
	// TODO: 从数据库或内存中获取最近的检查结果
	// 暂时返回空结果
	return make(map[string]CheckResult)
}
