//go:build windows
// +build windows

package checker

import (
	"fmt"
	"os/exec"
)

type WindowsPinger struct{}

func (p *WindowsPinger) Ping(host string, count int, timeout int) error {
	cmd := exec.Command("ping", "-n", fmt.Sprintf("%d", count), "-w", fmt.Sprintf("%d", timeout*1000), host)
	return cmd.Run()
}

func NewPinger() Pinger {
	return &WindowsPinger{}
}
