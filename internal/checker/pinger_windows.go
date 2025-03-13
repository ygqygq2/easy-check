//go:build windows
// +build windows

package checker

import (
	"fmt"
	"os/exec"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type WindowsPinger struct{}

func (p *WindowsPinger) Ping(host string, count int, timeout int) (string, error) {
	cmd := exec.Command("ping", "-n", fmt.Sprintf("%d", count), "-w", fmt.Sprintf("%d", timeout*1000), host)
	output, err := cmd.CombinedOutput()

	// 将 GBK 编码转换为 UTF-8
	utf8Output, _, err := transform.String(simplifiedchinese.GBK.NewDecoder(), string(output))
	if err != nil {
		return "", fmt.Errorf("failed to decode GBK output: %v", err)
	}

	// 调试信息：打印 ping 命令的输出和错误
	// fmt.Printf("Command output: %s\n", utf8Output)
	// if err != nil {
	// 	fmt.Printf("Command error: %v\n", err)
	// }

	return utf8Output, err
}

func NewPinger() Pinger {
	return &WindowsPinger{}
}
