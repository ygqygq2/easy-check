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
}

func NewChecker(config *config.Config, pinger Pinger, logger *logger.Logger, db *db.AlertStatusManager) *Checker {
	return &Checker{
		Config: config,
		Pinger: pinger,
		Logger: logger,
		DB:     db,
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

func (c *Checker) handlePingFailure(host config.Host, reason string) {
	// 检查是否启用失败告警
	if !c.isFailAlertEnabled(host) {
		c.Logger.Log(fmt.Sprintf("Fail alert disabled for host: %s", host.Host), "debug")
		return
	}

	// 记录失败日志
	c.Logger.Log(fmt.Sprintf("Ping to [%s] %s failed: %s", host.Description, host.Host, reason), "error")

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

func (c *Checker) pingHost(host config.Host) {
	output, err := c.Pinger.Ping(host.Host, c.Config.Ping.Count, c.Config.Ping.Timeout)
	if err != nil {
		// 调用提取的函数处理 Ping 失败
		c.handlePingFailure(host, err.Error())
		return
	}
	c.handlePingSuccess(host)

	lines := strings.Split(output, "\n")

	// 使用平台特定的解析方法
	successCount, sampleLatency := c.Pinger.ParsePingOutput(lines, c.Config.Ping.Count)

	successRate := float64(successCount) / float64(c.Config.Ping.Count)
	if successRate < 0.8 {
		// 调用提取的函数处理 Ping 成功率过低
		c.handlePingFailure(host, fmt.Sprintf("success rate %.2f%%", successRate*100))
	} else {
		c.Logger.Log(fmt.Sprintf("Ping to [%s] %s succeeded: success rate %.2f%%, latency %s", host.Description, host.Host, successRate*100, sampleLatency), "info")
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
