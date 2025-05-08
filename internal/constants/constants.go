package constants

import (
	"easy-check/internal/initializer"
)

// SharedConstants 定义需要共享的常量
type SharedConstants struct {
	AppName      string `json:"appName"`
	AppVersion   string `json:"appVersion"`
	UpdateServer string `json:"UpdateServer"`
}

// GetSharedConstants 返回共享常量
func GetSharedConstants(appCtx *initializer.AppContext) SharedConstants {
	return SharedConstants{
		AppName:      "easy-check",
		AppVersion:   appCtx.AppVersion,
		UpdateServer: "https://easy-check.ygqygq2.com",
	}
}
