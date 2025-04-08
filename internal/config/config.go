package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Host 主机配置
type Host struct {
	Host        string `yaml:"host"`
	Description string `yaml:"description"`
	FailAlert   *bool  `yaml:"fail_alert"`
}

// PingConfig Ping相关配置
type PingConfig struct {
	Count   int `yaml:"count"`
	Timeout int `yaml:"timeout"`
}

// LogConfig 日志配置
type LogConfig struct {
	File         string `yaml:"file"`
	MaxSize      int    `yaml:"max_size"`
	MaxAge       int    `yaml:"max_age"`
	MaxBackups   int    `yaml:"max_backups"`
	Compress     bool   `yaml:"compress"`
	ConsoleLevel string `yaml:"console_level"`
	FileLevel    string `yaml:"file_level"`
}

// NotifierConfig 通知器配置
type NotifierConfig struct {
	Name    string                 `yaml:"name"`
	Type    string                 `yaml:"type"`
	Enable  bool                   `yaml:"enable"`
	Options map[string]interface{} `yaml:",inline"` // 存储特定通知器的配置
}

// AlertConfig 告警配置
type AlertConfig struct {
	FailAlert             bool             `yaml:"fail_alert"`
	AggregateAlerts       bool             `yaml:"aggregate_alerts"`
	AggregateWindow       int              `yaml:"aggregate_window"`
	AggregateLineTemplate string           `yaml:"aggregate_line_template"`
	AggregateTemplate     string           `yaml:"aggregate_template"` // 聚合告警的整体模板
	Notifiers             []NotifierConfig `yaml:"notifiers"`
}

// Config 应用总配置
type Config struct {
	Hosts    []Host      `yaml:"hosts"`
	Interval int         `yaml:"interval"`
	Ping     PingConfig  `yaml:"ping"`
	Log      LogConfig   `yaml:"log"`
	Alert    AlertConfig `yaml:"alert"`
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// 确保每个通知器都有唯一的名称
	for i, notifier := range config.Alert.Notifiers {
		if notifier.Name == "" {
			config.Alert.Notifiers[i].Name = fmt.Sprintf("%s-%d", notifier.Type, i)
		}
	}

	return &config, nil
}

// GetNotifierByType 根据类型获取指定通知器配置
func (c *Config) GetNotifierByType(notifierType string) (*NotifierConfig, bool) {
	for _, n := range c.Alert.Notifiers {
		if n.Type == notifierType && n.Enable {
			return &n, true
		}
	}
	return nil, false
}

// HasEnabledNotifiers 检查是否有启用的通知器
func (c *Config) HasEnabledNotifiers() bool {
	for _, n := range c.Alert.Notifiers {
		if n.Enable {
			return true
		}
	}
	return false
}
