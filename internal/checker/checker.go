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

func (c *Checker) pingHost(host config.Host) {
	output, err := c.Pinger.Ping(host.Host, c.Count, c.Timeout)
	if err != nil {
		c.Logger.Log(fmt.Sprintf("Ping to [%s] %s failed: %v", host.Description, host.Host, err), "error")
		// 发送通知
		c.Logger.Log(fmt.Sprintf("Ping command to [%s] %s failed, sending notification", host.Description, host.Host), "debug")
		if c.Notifier != nil {
			err := c.Notifier.SendNotification(host.Host, host.Description)
			if err != nil {
				c.Logger.Log(fmt.Sprintf("Failed to send notification: %v", err), "error")
			}
		}
		return
	}

	lines := strings.Split(output, "\n")

	// 使用平台特定的解析方法
	successCount, sampleLatency := c.Pinger.ParsePingOutput(lines, c.Count)

	successRate := float64(successCount) / float64(c.Count)
	if successRate < 0.8 {
		c.Logger.Log(fmt.Sprintf("Ping to [%s] %s failed: success rate %.2f%%", host.Description, host.Host, successRate*100), "error")
		// 发送通知
		c.Logger.Log(fmt.Sprintf("Ping to %s failed, attempting to send notification", host.Host), "debug")
		if c.Notifier != nil {
			c.Notifier.SendNotification(host.Host, host.Description)
		}
	} else {
		c.Logger.Log(fmt.Sprintf("Ping to [%s] %s succeeded: success rate %.2f%%, latency %s", host.Description, host.Host, successRate*100, sampleLatency), "info")
	}
}
