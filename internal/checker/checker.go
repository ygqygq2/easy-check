package checker

import (
	"easy-check/internal/logger"
	"fmt"
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
			output, err := c.Pinger.Ping(host, c.Count, c.Timeout)
			if err != nil {
				c.Logger.Log(fmt.Sprintf("Ping to %s failed: %v", host, err))
			} else {
				c.Logger.Log(output)
			}

			successCount := c.Count
			if err != nil {
				successCount = 0
			}
			successRate := float64(successCount) / float64(c.Count)
			if successRate < 0.8 {
				c.Logger.Console(fmt.Sprintf("Ping to %s failed: success rate %.2f%%", host, successRate*100))
			} else {
				c.Logger.Console(fmt.Sprintf("Ping to %s succeeded: success rate %.2f%%", host, successRate*100))
			}
		}(host)
	}

	wg.Wait()
}
