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

func (c *Checker) PingHosts() {
	var wg sync.WaitGroup

	for _, host := range c.Config.Hosts {
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
	if host.FailAlert != nil {
		return *host.FailAlert // 优先使用主机的配置
	}
	return c.Config.Alert.FailAlert // 否则使用全局配置
}

type PingResult struct {
	Timestamp  string  `json:"timestamp"`
	RTT        float64 `json:"rtt"`
	PacketLoss float64 `json:"packet_loss"`
	Color      string  `json:"color"`
}

func (c *Checker) pingHost(host config.Host) {
	output, err := c.Pinger.Ping(host.Host, c.Config.Ping.Count, c.Config.Ping.Timeout)
	if err != nil {
		// 调用提取的函数处理 Ping 失败，同时传递原始输出
		c.handlePingFailure(host, err.Error(), output)
		return
	}
	c.handlePingSuccess(host)

	lines := strings.Split(output, "\n")

	// 使用平台特定的解析方法
	successCount, minLatency, avgLatency, maxLatency := c.Pinger.ParsePingOutput(lines, c.Config.Ping.Count)

	// 计算丢包率
	totalCount := c.Config.Ping.Count
	packetLossRate := c.calculatePacketLoss(successCount, totalCount)

	// 判断是否超过失败率阈值
	if packetLossRate > c.getFailRateThreshold() {
		c.handlePingFailure(host, fmt.Sprintf("packet loss rate %.2f%%", packetLossRate), output)
		return
	}

	c.Logger.Log(fmt.Sprintf("Ping to [%s] %s succeeded: packet loss rate %.2f%%, latency %s", host.Description, host.Host, packetLossRate, avgLatency), "info")

	// 将统计数据写入 TSDB
	c.writeMetricsToTSDB(host.Host, map[string]interface{}{
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

	// 将失败信息保存到数据库
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
	if c.Config.Ping.LossRate > 0 {
		return float64(c.Config.Ping.LossRate)
	}
	return 20.0 // 默认值（失败率超过 20% 触发告警）
}
