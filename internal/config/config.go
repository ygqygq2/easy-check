package config

import (
	"easy-check/internal/logger"
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
	Count    int `yaml:"count"`
	Timeout  int `yaml:"timeout"`
	Interval int `yaml:"interval"`
	LossRate int `yaml:"loss_rate"`
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

// DbConfig 数据库配置
type DbConfig struct {
	Path      string `yaml:"path"`
	Expire    int    `yaml:"expire"`
	Retention string `yaml:"retention"`
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
	FailAlert                     bool             `yaml:"fail_alert"`
	AggregateAlerts               bool             `yaml:"aggregate_alerts"`
	AggregateWindow               int              `yaml:"aggregate_window"`
	AggregateAlertLineTemplate    string           `yaml:"aggregate_alert_line_template"`
	AggregateRecoveryLineTemplate string           `yaml:"aggregate_recovery_line_template"`
	Notifiers                     []NotifierConfig `yaml:"notifiers"`
}

// Config 应用总配置
type Config struct {
	Hosts    []Host      `yaml:"hosts"`
	Interval int         `yaml:"interval"`
	Ping     PingConfig  `yaml:"ping"`
	Log      LogConfig   `yaml:"log"`
	Db       DbConfig    `yaml:"db"`
	Alert    AlertConfig `yaml:"alert"`
}

// LoadConfig 从文件加载配置
func LoadConfig(configPath string, logger *logger.Logger) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, logger.LogAndError("Failed to read config file: %v", "error", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, logger.LogAndError("Failed to parse config file: %v", "error", err)
	}

	// 为没有指定名称的通知器生成默认名称
	for i := range config.Alert.Notifiers {
		if config.Alert.Notifiers[i].Name == "" {
			config.Alert.Notifiers[i].Name = fmt.Sprintf("%s-%d", config.Alert.Notifiers[i].Type, i)
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

// GetConfigFromFile 获取配置文件内容
func GetConfigFromFile(configPath string) (string, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("读取配置文件失败: %v", err)
	}
	return string(data), nil
}

// SaveConfigToFile 保存配置文件内容
func SaveConfigToFile(configPath string, content string, logger *logger.Logger) error {
	// 先验证YAML格式是否正确
	var config Config
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return fmt.Errorf("无效的YAML格式: %v", err)
	}

	// 备份原配置
	backupPath := configPath + ".bak"
	if err := os.Rename(configPath, backupPath); err != nil {
		logger.Log(fmt.Sprintf("创建配置备份失败: %v", err), "warn")
		// 继续执行，即使备份失败
	}

	// 写入新配置
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		// 尝试恢复备份
		os.Rename(backupPath, configPath)
		return fmt.Errorf("保存配置失败: %v", err)
	}

	logger.Log("配置已成功保存", "info")
	return nil
}
