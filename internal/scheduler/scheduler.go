package scheduler

import (
	"easy-check/internal/checker"
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"fmt"
	"time"
)

func StartPeriodicPingChecks(chk *checker.Checker, cfg *config.Config, logger *logger.Logger, tickerControlChan chan time.Duration) chan struct{} {
    ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)

    logger.Log("Starting periodic ping checks", "info") // 将日志级别从 debug 改为 info

    // 创建一个停止通道，用于通知程序退出
    stopChan := make(chan struct{})

    // 处理 ping 检查的 goroutine
    go func() {
        defer ticker.Stop() // 移到 goroutine 内部，这样只有在 goroutine 结束时才会停止 ticker
        for {
            select {
            case <-ticker.C:
                logger.Log("Executing scheduled ping check", "debug")
                chk.PingHosts()
            case newInterval := <-tickerControlChan:
                // 收到新的间隔时间，更新定时器
                ticker.Stop()
                ticker = time.NewTicker(newInterval)
                logger.Log(fmt.Sprintf("Updated ping check interval to %v", newInterval), "info")
            case <-stopChan:
                return
            }
        }
    }()

    return stopChan
}
