package signal

import (
	"easy-check/internal/logger"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func WaitForExitSignal(logger *logger.Logger) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    sig := <-sigChan
    logger.Log(fmt.Sprintf("Received exit signal (%s), shutting down...", sig), "info")
}
