package machineid

import (
	"testing"
)

func TestGetMachineID(t *testing.T) {
	// 测试获取机器 ID
	id1, err := GetMachineID()
	if err != nil {
		t.Fatalf("Failed to get machine ID: %v", err)
	}

	if id1 == "" {
		t.Fatal("Machine ID is empty")
	}

	// 再次获取，应该返回相同的 ID
	id2, err := GetMachineID()
	if err != nil {
		t.Fatalf("Failed to get machine ID on second call: %v", err)
	}

	if id1 != id2 {
		t.Errorf("Machine ID changed between calls: %s != %s", id1, id2)
	}

	t.Logf("Machine ID: %s (length: %d)", id1, len(id1))
}

func TestGenerateMachineID(t *testing.T) {
	// 生成两个 ID，应该不同
	id1, err := generateMachineID()
	if err != nil {
		t.Fatalf("Failed to generate machine ID: %v", err)
	}

	id2, err := generateMachineID()
	if err != nil {
		t.Fatalf("Failed to generate second machine ID: %v", err)
	}

	if id1 == id2 {
		t.Error("Generated machine IDs should be different")
	}

	// 验证 ID 长度（32 字节 = 64 个十六进制字符）
	if len(id1) != 64 {
		t.Errorf("Expected machine ID length 64, got %d", len(id1))
	}

	t.Logf("Generated ID 1: %s", id1)
	t.Logf("Generated ID 2: %s", id2)
}
