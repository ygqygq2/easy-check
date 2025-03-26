package signal

import (
	"easy-check/internal/logger"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	sig := <-exitChan

	listenerLock.Lock()
	defer listenerLock.Unlock()

	for _, listener := range listeners {
		listener <- sig
	}
}

// WaitForExitSignal 等待退出信号（向后兼容的方法）
func WaitForExitSignal(logger *logger.Logger) {
	sigChan := RegisterExitListener()
	sig := <-sigChan
	logger.Log(fmt.Sprintf("Received exit signal (%s), shutting down...", sig), "info")
}
