package signal

import (
	"easy-check/internal/logger"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	exitChan     chan os.Signal
	listeners    []chan<- os.Signal
	listenerLock sync.Mutex
	once         sync.Once
)

// RegisterExitListener 注册一个退出信号监听者
func RegisterExitListener() <-chan os.Signal {
	listenerChan := make(chan os.Signal, 1)

	listenerLock.Lock()
	defer listenerLock.Unlock()

	once.Do(func() {
		exitChan = make(chan os.Signal, 1)
		signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM)

		// 启动信号分发器
		go signalDispatcher()
	})

	listeners = append(listeners, listenerChan)
	return listenerChan
}

// 信号分发器，将收到的信号分发给所有监听者
func signalDispatcher() {
	for {
		sig := <-exitChan

		listenerLock.Lock()
		for _, listener := range listeners {
			select {
			case listener <- sig:
				// 信号已发送
			default:
				// 如果监听者没有准备好接收，不要阻塞
			}
		}
		listenerLock.Unlock()

		// 对于强制退出的信号，可以在分发后直接退出程序
		if sig == syscall.SIGINT || sig == syscall.SIGTERM {
			// 给一点时间让其他 goroutine 处理信号
			time.Sleep(time.Second)
			os.Exit(0)
		}
	}
}

// WaitForExitSignal 等待退出信号（向后兼容的方法）
func WaitForExitSignal(logger *logger.Logger) {
	sigChan := RegisterExitListener()
	sig := <-sigChan
	logger.Log(fmt.Sprintf("Received exit signal (%s), shutting down...", sig), "info")
}
