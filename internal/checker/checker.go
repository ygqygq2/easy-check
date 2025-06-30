package checker

import (
	"easy-check/internal/config"
	"easy-check/internal/db"
	"easy-check/internal/logger"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Checker struct {
	Config *config.Config
	Pinger Pinger
	Logger *logger.Logger
	DB     *db.AlertStatusManager
	TSDB   *db.TSDB
	// 添加配置读写锁
	configMu sync.RWMutex
}

func NewChecker(config *config.Config, pinger Pinger, logger *logger.Logger, db *db.AlertStatusManager, tsdb *db.TSDB) *Checker {
	return &Checker{
		Config: config,
		Pinger: pinger,
		Logger: logger,
		DB:     db,
		TSDB:   tsdb,
	}
}

// UpdateConfig 提供线程安全的配置更新方法
func (c *Checker) UpdateConfig(newConfig *config.Config) {
	c.configMu.Lock()
	defer c.configMu.Unlock()
	c.Config = newConfig
}

// getConfig 提供线程安全的配置读取
func (c *Checker) getConfig() *config.Config {
	c.configMu.RLock()
	defer c.configMu.RUnlock()
	return c.Config
}

func (c *Checker) PingHosts() {
	var wg sync.WaitGroup

	// 使用线程安全的配置读取
	cfg := c.getConfig()

	for _, host := range cfg.Hosts {
		wg.Add(1)
		go func(host config.Host) {
			defer wg.Done()
			c.pingHost(host)
		}(host)
	}

	wg.Wait()
}

// 判断是否启用失败告警
func (c *Checker) isFailAlertEnabled(host config.Host) bool {
	cfg := c.getConfig()
	if host.FailAlert != nil {
		return *host.FailAlert // 优先使用主机的配置
	}
	return cfg.Alert.FailAlert // 否则使用全局配置
}

type PingResult struct {
	Timestamp  string  `json:"timestamp"`
	RTT        float64 `json:"rtt"`
	PacketLoss float64 `json:"packet_loss"`
	Color      string  `json:"color"`
}

func (c *Checker) pingHost(host config.Host) {
	cfg := c.getConfig()
	output, err := c.Pinger.Ping(host.Host, cfg.Ping.Count, cfg.Ping.Timeout)

	// 解析 Ping 输出
	lines := strings.Split(output, "\n")
	successCount, minLatency, avgLatency, maxLatency := c.Pinger.ParsePingOutput(lines, cfg.Ping.Count)
	packetLossRate := c.calculatePacketLoss(successCount, cfg.Ping.Count)

	// 根据 Ping 结果处理逻辑
	var reason string
	if err != nil {
		reason = err.Error()
	} else if packetLossRate > c.getFailRateThreshold() {
		// 如果 Ping 失败或丢包率超过阈值，处理失败逻辑
		reason = fmt.Sprintf("packet loss rate %.2f%%", packetLossRate)
	}

	if reason != "" {
		// 记录失败日志
		c.Logger.Log(fmt.Sprintf("Ping to [%s] %s failed: packet loss %.2f%%, avg latency time=%.2fms",
			host.Description, host.Host, packetLossRate, avgLatency), "error")
		c.handlePingFailure(host, reason, output)
	} else {
		successRate := 100.0 - packetLossRate
		c.Logger.Log(fmt.Sprintf("Ping to [%s] %s succeeded: success rate %.2f%%, latency time=%.2fms",
			host.Description, host.Host, successRate, avgLatency), "info")
		// 如果 Ping 成功且丢包率在阈值内，处理成功逻辑
		c.handlePingSuccess(host)
	}

	// 无论成功或失败，都写入统计数据到 TSDB
	c.writeMetricsToTSDB(host.Host, map[string]any{
		"packet_loss": packetLossRate,
		"min_latency": minLatency,
		"avg_latency": avgLatency,
		"max_latency": maxLatency,
	})
}

func (c *Checker) handlePingFailure(host config.Host, reason string, output string) {
	// 检查是否启用失败告警
	if !c.isFailAlertEnabled(host) {
		c.Logger.Log(fmt.Sprintf("Fail alert disabled for host: %s", host.Host), "debug")
		return
	}

	// 记录失败日志以及完整的ping输出
	c.Logger.Log(fmt.Sprintf("Ping to [%s] %s failed: %s", host.Description, host.Host, reason), "error")
	c.Logger.Log(fmt.Sprintf("Ping output for [%s] %s: \n%s", host.Description, host.Host, output), "info")

	// 构造 AlertStatus 结构体
	status := db.AlertStatus{
		Host:         host.Host,
		Description:  host.Description,
		Status:       db.StatusAlert,
		FailTime:     time.Now().Format(time.RFC3339),
		RecoveryTime: "",
		FailAlert:    true,
		Sent:         false,
	}

	// 将失败信息保存到数据库
	err := c.DB.MarkAsAlert(status)
	if err != nil {
		c.Logger.Log(fmt.Sprintf("Failed to record ping failure in DB: %v", err), "error")
	}
}

func (c *Checker) handlePingSuccess(host config.Host) {
	// 构造 AlertStatus 结构体
	status := db.AlertStatus{
		Host:         host.Host,
		Description:  host.Description,
		Status:       db.StatusRecovery,
		RecoveryTime: time.Now().Format(time.RFC3339),
	}

	err := c.DB.MarkAsRecovered(status)
	if err != nil {
		c.Logger.Log(fmt.Sprintf("Failed to update host recovery status: %v", err), "error")
	}
}

func (c *Checker) calculatePacketLoss(successCount, totalCount int) float64 {
	return float64(totalCount-successCount) / float64(totalCount) * 100
}

// 将统计数据写入 TSDB
func (c *Checker) writeMetricsToTSDB(host string, metrics map[string]interface{}) {
	labels := map[string]string{
		"host": host,
	}

	// 转换 metrics 数据类型为 map[string]float64
	floatMetrics := make(map[string]float64)
	for key, value := range metrics {
		if floatValue, ok := value.(float64); ok {
			floatMetrics[key] = floatValue
		} else {
			c.Logger.Log(fmt.Sprintf("Invalid metric value for %s: %v", key, value), "error")
		}
	}

	// 获取当前时间戳
	timestamp := time.Now().UnixMilli()

	// 写入 TSDB
	err := c.TSDB.AppendMetrics(floatMetrics, timestamp, labels)
	if err != nil {
		c.Logger.Log(fmt.Sprintf("Failed to write metrics to TSDB for host %s: %v", host, err), "error")
	}
}

func (c *Checker) getFailRateThreshold() float64 {
	cfg := c.getConfig()
	if cfg.Ping.LossRate > 0 {
		return float64(cfg.Ping.LossRate) * 100
	}
	return 20 // 默认值（失败率超过 20% 触发告警）
}
