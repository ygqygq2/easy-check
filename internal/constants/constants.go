package constants

import (
	"easy-check/internal/initializer"
)

// PlatformInfo 定义平台信息
type PlatformInfo struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

// ServerConfig 定义服务器配置
type ServerConfig struct {
	BaseURL      string `json:"baseUrl"`
	UpdateAPI    string `json:"updateApi"`
	HeartbeatAPI string `json:"heartbeatApi"`
}

// HeartbeatConfig 定义心跳配置
type HeartbeatConfig struct {
	Enabled  bool `json:"enabled"`
	Interval int  `json:"interval"` // 秒
	Timeout  int  `json:"timeout"`  // 秒
}

// SharedConstants 定义需要共享的常量
type SharedConstants struct {
	AppName      string          `json:"appName"`
	AppVersion   string          `json:"appVersion"`
	PlatformInfo PlatformInfo    `json:"platformInfo"`
	Server       ServerConfig    `json:"server"`
	Heartbeat    HeartbeatConfig `json:"heartbeat"`
	NeedsRestart bool            `json:"needsRestart"`
}

// getServerConfig 根据版本返回服务器配置
func getServerConfig(version string) ServerConfig {
	isDev := version == "dev"
	baseURL := "https://easy-check-server.ygqygq2.com"
	
	if isDev {
		baseURL = "http://localhost:3000"
	}
	
	return ServerConfig{
		BaseURL:      baseURL,
		UpdateAPI:    "/api/update",
		HeartbeatAPI: "/api/heartbeat",
	}
}

// getHeartbeatConfig 根据版本返回心跳配置
func getHeartbeatConfig(version string) HeartbeatConfig {
	isDev := version == "dev"
	interval := 1800 // 生产环境：30分钟
	
	if isDev {
		interval = 60 // 开发环境：1分钟
	}
	
	return HeartbeatConfig{
		Enabled:  true,
		Interval: interval,
		Timeout:  10,
	}
}

// GetSharedConstants 返回共享常量
func GetSharedConstants(appCtx *initializer.AppContext) SharedConstants {
	return SharedConstants{
		AppName:    "easy-check",
		AppVersion: appCtx.AppVersion,
		PlatformInfo: PlatformInfo{
			OS:   appCtx.PlatformInfo.OS,
			Arch: appCtx.PlatformInfo.Arch,
		},
		Server:       getServerConfig(appCtx.AppVersion),
		Heartbeat:    getHeartbeatConfig(appCtx.AppVersion),
		NeedsRestart: appCtx.NeedsRestart,
	}
}
