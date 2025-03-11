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
            err := c.Pinger.Ping(host, c.Count, c.Timeout)
            if (err != nil) {
                c.Logger.Log(fmt.Sprintf("Ping to %s failed: %v", host, err))
            } else {
                c.Logger.Log(fmt.Sprintf("Ping to %s succeeded", host))
            }
        }(host)
    }

    wg.Wait()
}
