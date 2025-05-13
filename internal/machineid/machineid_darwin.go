//go:build darwin

package machineid

import (
	"fmt"
	"os/exec"
	"strings"
)

func getMachineID() (string, error) {
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "IOPlatformUUID") {
			parts := strings.Split(line, "\"")
			if len(parts) > 3 {
				return parts[3], nil
			}
		}
	}
	return "", fmt.Errorf("failed to get machine ID")
}
