package update

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

// RestartApp 重启应用程序
func RestartApp() error {
	fmt.Println("应用程序需要重启...")

	// 获取当前可执行文件路径
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %v", err)
	}

	// 获取当前命令行参数和环境变量
	args := os.Args
	env := os.Environ()

	if runtime.GOOS == "windows" {
		cmd := exec.Command(self, args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = env
		err := cmd.Start()
		if err == nil {
			os.Exit(0)
		}
		return err
	}
	return syscall.Exec(self, args, env)
}
