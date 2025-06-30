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
	listeners    map[chan os.Signal]bool // 使用map来管理监听器
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
		listeners = make(map[chan os.Signal]bool)
		signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM)

		// 启动信号分发器
		go signalDispatcher()
	})

	listeners[listenerChan] = true
	return listenerChan
}

// CleanupClosedListeners 清理已关闭的监听器，防止内存泄漏
func CleanupClosedListeners() {
	listenerLock.Lock()
	defer listenerLock.Unlock()
	
	for listener := range listeners {
		select {
		case <-listener:
			// channel已关闭，删除
			delete(listeners, listener)
		default:
			// channel还活着，保留
		}
	}
}

// 信号分发器，将收到的信号分发给所有监听者
func signalDispatcher() {
	for {
		sig := <-exitChan

		listenerLock.Lock()
		for listener := range listeners {
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
