package checker

import (
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"easy-check/internal/notifier"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Checker struct {
	Hosts    []config.Host
	Interval time.Duration
	Count    int
	Timeout  int
	Pinger   Pinger
	Logger   *logger.Logger
	Notifier notifier.Notifier
}

func NewChecker(config *config.Config, pinger Pinger, logger *logger.Logger, notifier notifier.Notifier) *Checker {
	return &Checker{
		Hosts:    config.Hosts,
		Interval: time.Duration(config.Interval) * time.Second,
		Count:    config.Ping.Count,
		Timeout:  config.Ping.Timeout,
		Pinger:   pinger,
		Logger:   logger,
		Notifier: notifier,
	}
}

func (c *Checker) PingHosts() {
	var wg sync.WaitGroup

	for _, host := range c.Hosts {
		wg.Add(1)
		go func(host config.Host) {
			defer wg.Done()
			c.pingHost(host)
		}(host)
	}

	wg.Wait()
}

func (c *Checker) handlePingFailure(host config.Host, reason string) {
	// 记录失败日志
	c.Logger.Log(fmt.Sprintf("Ping to [%s] %s failed: %s", host.Description, host.Host, reason), "error")
	c.Logger.Log(fmt.Sprintf("Attempting to send notification for [%s] %s", host.Description, host.Host), "debug")

	// 发送通知
	if c.Notifier != nil {
		err := c.Notifier.SendNotification(host)
		if err != nil {
			c.Logger.Log(fmt.Sprintf("Failed to send notification: %v", err), "error")
		}
	}
}

func (c *Checker) pingHost(host config.Host) {
	output, err := c.Pinger.Ping(host.Host, c.Count, c.Timeout)
	if err != nil {
		// 调用提取的函数处理 Ping 失败
		c.handlePingFailure(host, err.Error())
		return
	}

	lines := strings.Split(output, "\n")

	// 使用平台特定的解析方法
	successCount, sampleLatency := c.Pinger.ParsePingOutput(lines, c.Count)

	successRate := float64(successCount) / float64(c.Count)
	if successRate < 0.8 {
		// 调用提取的函数处理 Ping 成功率过低
		c.handlePingFailure(host, fmt.Sprintf("success rate %.2f%%", successRate*100))
	} else {
		c.Logger.Log(fmt.Sprintf("Ping to [%s] %s succeeded: success rate %.2f%%, latency %s", host.Description, host.Host, successRate*100, sampleLatency), "info")
	}
}
