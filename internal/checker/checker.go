package checker

import (
	"easy-check/internal/logger"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Checker struct {
	Hosts    []string
	Interval time.Duration
	Count    int
	Timeout  int
	Pinger   Pinger
	Logger   *logger.Logger
}

func NewChecker(hosts []string, interval int, count int, timeout int, pinger Pinger, logger *logger.Logger) *Checker {
	return &Checker{
		Hosts:    hosts,
		Interval: time.Duration(interval) * time.Second,
		Count:    count,
		Timeout:  timeout,
		Pinger:   pinger,
		Logger:   logger,
	}
}

func (c *Checker) PingHosts() {
	var wg sync.WaitGroup

	for _, host := range c.Hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			c.pingHost(host)
		}(host)
	}

	wg.Wait()
}

func (c *Checker) pingHost(host string) {
	output, err := c.Pinger.Ping(host, c.Count, c.Timeout)
	if err != nil {
		c.Logger.Log(fmt.Sprintf("Ping to %s failed: %v", host, err))
		c.Logger.Console(fmt.Sprintf("Ping to %s failed: %v", host, err))
		return
	}

	lines := strings.Split(output, "\n")
	successCount, sampleLatency := c.parsePingOutput(lines)

	successRate := float64(successCount) / float64(c.Count)
	if successRate < 0.8 {
		c.Logger.Console(fmt.Sprintf("Ping to %s failed: success rate %.2f%%", host, successRate*100))
	} else {
		c.Logger.Console(fmt.Sprintf("Ping to %s succeeded: success rate %.2f%%, latency %s", host, successRate*100, sampleLatency))
	}
}

func (c *Checker) parsePingOutput(lines []string) (int, string) {
	successCount := 0
	var sampleLatency string

	re := regexp.MustCompile(`time[=<]\d+ms|时间[=<]\d+ms`)

	for _, line := range lines {
		if strings.Contains(line, "TTL=") || strings.Contains(line, "ttl=") {
			c.Logger.Log(line)
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
