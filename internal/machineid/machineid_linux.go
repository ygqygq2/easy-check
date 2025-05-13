//go:build linux

package machineid

import (
	"os/exec"
	"strings"
)

func getMachineID() (string, error) {
	cmd := exec.Command("cat", "/etc/machine-id")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
