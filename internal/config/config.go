package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Host struct {
	Host        string `yaml:"host"`
	Description string `yaml:"description"`
	FailAlert   *bool   `yaml:"fail_alert"`
}

type Config struct {
	Hosts    []Host `yaml:"hosts"`
	Interval int    `yaml:"interval"`
	Ping     struct {
		Count   int `yaml:"count"`
		Timeout int `yaml:"timeout"`
	} `yaml:"ping"`
	Log struct {
		File         string `yaml:"file"`
		MaxSize      int    `yaml:"max_size"`
		MaxAge       int    `yaml:"max_age"`
		MaxBackups   int    `yaml:"max_backups"`
		Compress     bool   `yaml:"compress"`
		ConsoleLevel string `yaml:"console_level"`
		FileLevel    string `yaml:"file_level"`
	} `yaml:"log"`
	Alert struct {
		FailAlert             bool         `yaml:"fail_alert"`
		AggregateAlerts       bool         `yaml:"aggregate_alerts"`
		AggregateWindow       int          `yaml:"aggregate_window"`
		AggregateLineTemplate string       `yaml:"aggregate_line_template"`
		Feishu                FeishuConfig `yaml:"feishu"`
	} `yaml:"alert"`
}

type FeishuConfig struct {
	Enable  bool   `yaml:"enable"`
	Webhook string `yaml:"webhook"`
	MsgType string `yaml:"msg_type"`
	Title   string `yaml:"title"`
	Content string `yaml:"content"`
}

func LoadConfig(configPath string) (*Config, error) {
  data, err := os.ReadFile(configPath)
  if err != nil {
      return nil, fmt.Errorf("failed to read config file: %v", err)
  }

  var config Config
  if err := yaml.Unmarshal(data, &config); err != nil {
      return nil, fmt.Errorf("failed to parse config file: %v", err)
  }

  return &config, nil
}
