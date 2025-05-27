package utils

import (
  "testing"
)

func TestFormatTime(t *testing.T) {
  tests := []struct {
    name      string
    input     string
    want      string
  }{
    {
      name:  "empty string",
      input: "",
      want:  "",
    },
    {
      name:  "valid ISO8601",
      input: "2024-06-01T12:34:56+08:00",
      want:  "2024-06-01 12:34:56",
    },
    {
      name:  "invalid format",
      input: "not-a-time",
      want:  "not-a-time",
    },
    {
      name:  "valid UTC",
      input: "2024-06-01T04:34:56Z",
      want:  "2024-06-01 04:34:56",
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      got := FormatTime(tt.input)
      if got != tt.want {
        t.Errorf("FormatTime(%q) = %q, want %q", tt.input, got, tt.want)
      }
    })
  }
}

func TestIsDirectorySuffix(t *testing.T) {
  tests := []struct {
    name  string
    input string
    want  bool
  }{
    {"empty string", "", false},
    {"ends with slash", "foo/bar/", true},
    {"does not end with slash", "foo/bar", false},
    {"only slash", "/", true},
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      got := IsDirectorySuffix(tt.input)
      if got != tt.want {
        t.Errorf("IsDirectorySuffix(%q) = %v, want %v", tt.input, got, tt.want)
      }
    })
  }
}

func TestAddDirectorySuffix(t *testing.T) {
  tests := []struct {
    name  string
    input string
    want  string
  }{
    {"already has slash", "foo/bar/", "foo/bar/"},
    {"no slash", "foo/bar", "foo/bar/"},
    {"empty string", "", "/"},
    {"only slash", "/", "/"},
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      got := AddDirectorySuffix(tt.input)
      if got != tt.want {
        t.Errorf("AddDirectorySuffix(%q) = %q, want %q", tt.input, got, tt.want)
      }
    })
  }
}
