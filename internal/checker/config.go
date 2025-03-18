package checker

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Hosts    []Host     `yaml:"hosts"`
	Ping     PingConfig `yaml:"ping"`
	Interval int        `yaml:"interval"`
	Log      LogConfig  `yaml:"log"`
}

type PingConfig struct {
	Count   int `yaml:"count"`
	Timeout int `yaml:"timeout"`
}

type LogConfig struct {
	File         string `yaml:"file"`
	MaxSize      int    `yaml:"max_size"`
	MaxAge       int    `yaml:"max_age"`
	MaxBackups   int    `yaml:"max_backups"`
	Compress     bool   `yaml:"compress"`
	ConsoleLevel string `yaml:"console_level"`
	FileLevel    string `yaml:"file_level"`
}

func LoadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
