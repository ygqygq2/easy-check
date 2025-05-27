package update

import (
	"fmt"
	"os"
)

// RestartApp 重启应用程序
func RestartApp() error {
	fmt.Println("应用程序重启...")

	// 获取当前可执行文件路径
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %v", err)
	}

	// 获取当前命令行参数和环境变量
	args := os.Args
	env := os.Environ()

	// 添加一个环境变量标记，指示是重启操作
	env = append(env, "EASY_CHECK_RESTART=true")

	return restartOS(self, args, env)
}
