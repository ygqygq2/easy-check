package checker

import (
	"easy-check/internal/logger"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Host struct {
    Host        string
    Description string
}

type Checker struct {
    Hosts    []Host
    Interval time.Duration
    Count    int
    Timeout  int
    Pinger   Pinger
    Logger   *logger.Logger
}

func NewChecker(hosts []Host, interval int, count int, timeout int, pinger Pinger, logger *logger.Logger) *Checker {
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
        go func(host Host) {
            defer wg.Done()
            c.pingHost(host)
        }(host)
    }

    wg.Wait()
}

func (c *Checker) pingHost(host Host) {
    output, err := c.Pinger.Ping(host.Host, c.Count, c.Timeout)
    if err != nil {
        c.Logger.Log(fmt.Sprintf("Ping to [%s] %s failed: %v", host.Description, host.Host, err), "error")
        c.Logger.Console(fmt.Sprintf("Ping to [%s] %s failed: %v", host.Description, host.Host, err))
        return
    }

    lines := strings.Split(output, "\n")
    successCount, sampleLatency := c.parsePingOutput(lines)

    successRate := float64(successCount) / float64(c.Count)
    if successRate < 0.8 {
        c.Logger.Log(fmt.Sprintf("Ping to [%s] %s failed: success rate %.2f%%", host.Description, host.Host, successRate*100), "error")
        c.Logger.Console(fmt.Sprintf("Ping to [%s] %s failed: success rate %.2f%%", host.Description, host.Host, successRate*100))
    } else {
        c.Logger.Log(fmt.Sprintf("Ping to [%s] %s succeeded: success rate %.2f%%, latency %s", host.Description, host.Host, successRate*100, sampleLatency))
    }
}

func (c *Checker) parsePingOutput(lines []string) (int, string) {
    successCount := 0
    var sampleLatency string

    // 更新正则表达式以匹配 Linux 和 Windows 的 ping 输出
    re := regexp.MustCompile(`time[=<]\d+(\.\d+)? ms|时间[=<]\d+(\.\d+)? ms`)

    for _, line := range lines {
        if strings.Contains(line, "TTL=") || strings.Contains(line, "ttl=") || strings.Contains(line, "time=") {
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
