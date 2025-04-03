package db

import (
	"os"
	"testing"
)

func TestDB(t *testing.T) {
	// 创建临时目录
	path := "./testdb"
	defer os.RemoveAll(path)

	// 初始化数据库
	database, err := NewDB(path)
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}
	defer database.Close()

	// 测试 Set 和 Get
	err = database.Set("key1", "value1")
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	val, err := database.Get("key1")
	if err != nil || val != "value1" {
		t.Fatalf("Expected value1, got %v (err: %v)", val, err)
	}

	// 测试 Delete
	err = database.Delete("key1")
	if err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	_, err = database.Get("key1")
	if err == nil {
		t.Fatalf("Expected error for deleted key, got nil")
	}
}
