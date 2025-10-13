//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"easy-check/internal/machineid"
)

func main() {
	fmt.Println("Machine ID Test Tool")
	fmt.Println("====================")
	fmt.Println()

	// 获取机器 ID
	id, err := machineid.GetMachineID()
	if err != nil {
		log.Fatalf("Error getting machine ID: %v", err)
	}

	fmt.Printf("Machine ID: %s\n", id)
	fmt.Printf("ID Length: %d characters\n", len(id))
	fmt.Println()

	// 显示配置目录信息
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}
	configDir := filepath.Join(homeDir, ".easy-check")
	machineIDFile := filepath.Join(configDir, "machine-id")

	fmt.Println("Configuration:")
	fmt.Printf("  Config Directory: %s\n", configDir)
	fmt.Printf("  Machine ID File: %s\n", machineIDFile)

	// 检查文件是否存在
	if _, err := os.Stat(machineIDFile); err == nil {
		fmt.Printf("  Status: File exists (using fallback ID)\n")
		
		// 读取文件内容
		data, err := os.ReadFile(machineIDFile)
		if err == nil {
			fmt.Printf("  Raw ID (before hash): %s\n", string(data))
		}
	} else {
		fmt.Printf("  Status: Using system-provided ID\n")
	}
}
