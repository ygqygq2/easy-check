//go:build !windows
// +build !windows

package main

import "github.com/wailsapp/wails/v3/pkg/application"

// 在非 windows 平台下什么也不做，避免编译错误
func watchTaskbarCreated(
	tray *application.SystemTray,
	window *application.WebviewWindow,
	icon []byte,
	menu *application.Menu,
) {
	// no-op
}
