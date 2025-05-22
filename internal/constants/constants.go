package constants

import (
	"easy-check/internal/initializer"
)

// PlatformInfo 定义平台信息
type PlatformInfo struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

// SharedConstants 定义需要共享的常量
type SharedConstants struct {
	AppName      string       `json:"appName"`
	AppVersion   string       `json:"appVersion"`
	PlatformInfo PlatformInfo `json:"platformInfo"`
	UpdateServer string       `json:"UpdateServer"`
	NeedsRestart bool         `json:"needsRestart"`
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
		UpdateServer: "https://easy-check-server.ygqygq2.com",
		NeedsRestart: appCtx.NeedsRestart,
	}
}
