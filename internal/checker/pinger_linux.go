//go:build linux
// +build linux

package checker

import (
	"fmt"
	"os/exec"
)

type LinuxPinger struct{}

func (p *LinuxPinger) Ping(host string, count int, timeout int) error {
    cmd := exec.Command("ping", "-c", fmt.Sprintf("%d", count), "-W", fmt.Sprintf("%d", timeout), host)
    return cmd.Run()
}

func NewPinger() Pinger {
    return &LinuxPinger{}
}
