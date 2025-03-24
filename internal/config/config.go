package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Host struct {
    Host        string `yaml:"host"`
    Description string `yaml:"description"`
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
        Feishu FeishuConfig `yaml:"feishu"`
    } `yaml:"alert"`
}

type FeishuConfig struct {
    Enable   bool   `yaml:"enable"`
    Webhook  string `yaml:"webhook"`
    MsgType  string `yaml:"msg_type"`
    Title    string `yaml:"title"`
    Content  string `yaml:"content"`
}

func LoadConfig(configPath string) (*Config, error) {
    data, err := os.ReadFile(configPath)
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
