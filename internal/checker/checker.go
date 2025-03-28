package checker

import (
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"easy-check/internal/notifier"
	"fmt"
	"regexp"
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
	successCount, sampleLatency := c.parsePingOutput(lines)

	successRate := float64(successCount) / float64(c.Count)
	if successRate < 0.8 {
		c.Logger.Log(fmt.Sprintf("Ping to [%s] %s failed: success rate %.2f%%", host.Description, host.Host, successRate*100), "error")
		// 发送通知
		c.Logger.Log(fmt.Sprintf("Ping to %s failed, attempting to send notification", host.Host), "debug")
		if c.Notifier != nil {
			c.Notifier.SendNotification(host.Host, host.Description)
		}
	} else {
		c.Logger.Log(fmt.Sprintf("Ping to [%s] %s succeeded: success rate %.2f%%, latency %s", host.Description, host.Host, successRate*100, sampleLatency), "info") // 显式指定info级别
	}
}

func (c *Checker) parsePingOutput(lines []string) (int, string) {
	successCount := 0
	var sampleLatency string

	// 更新正则表达式以匹配 Linux 和 Windows 的 ping 输出
	re := regexp.MustCompile(`time[=<]\d+(\.\d+)? ms|时间[=<]\d+(\.\d+)? ms`)

	for _, line := range lines {
		if strings.Contains(line, "TTL=") || strings.Contains(line, "ttl=") || strings.Contains(line, "time=") {
			c.Logger.Log(line, "info")
			successCount++
			if sampleLatency == "" {
				match := re.FindString(line)
				if match != "" {
					sampleLatency = match
				}
			}
		}
	}

	return successCount, sampleLatency
}
