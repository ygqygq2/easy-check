//go:build linux
// +build linux

package checker

import (
	"fmt"
	"os/exec"
)

type LinuxPinger struct{}

func (p *LinuxPinger) Ping(host string, count int, timeout int) (string, error) {
	cmd := exec.Command("ping", "-4", "-c", fmt.Sprintf("%d", count), "-W", fmt.Sprintf("%d", timeout), host)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func NewPinger() Pinger {
	return &LinuxPinger{}
}
