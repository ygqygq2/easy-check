//go:build windows

package machineid

import (
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows"
)

func getMachineID() (string, error) {
	cmd := exec.Command("wmic", "csproduct", "get", "UUID")
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("failed to get machine ID")
	}
	return strings.TrimSpace(lines[1]), nil
}
