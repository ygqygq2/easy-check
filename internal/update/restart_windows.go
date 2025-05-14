//go:build windows

package update

import (
	"os"
	"os/exec"

	"golang.org/x/sys/windows"
)

func restartOS(self string, args []string, env []string) error {
	// 为Windows添加延迟启动，确保当前实例有足够时间退出并释放锁
	cmd := exec.Command("cmd", "/c", "timeout", "/t", "1", "&&", self)
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
	cmd.Args = append(cmd.Args, args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = env
	err := cmd.Start()
	if err == nil {
		// 成功启动新实例后退出当前实例
		os.Exit(0)
	}
	return err
}
