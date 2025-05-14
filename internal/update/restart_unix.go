//go:build !windows

package update

import (
	"syscall"
	"time"
)

func restartOS(self string, args []string, env []string) error {
	time.Sleep(1 * time.Second)
	return syscall.Exec(self, args, env)
}
