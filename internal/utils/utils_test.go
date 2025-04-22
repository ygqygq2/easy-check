package utils

import (
	"testing"
)

func TestFormatTime(t *testing.T) {
	timestamp := "2025-04-22T15:04:05Z"
	expected := "2025-04-22 15:04:05"
	result := FormatTime(timestamp)
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestProcessTemplate(t *testing.T) {
	templateStr := "Hello, {{.Name}}!"
	data := map[string]string{"Name": "World"}
	expected := "Hello, World!"
	result, err := ProcessTemplate(templateStr, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}
