//go:build windows
// +build windows

package main

import (
	"syscall"
	"unsafe"

	"github.com/lxn/win"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// watchTaskbarCreated 监控 Explorer 重启后 TaskbarCreated 消息，重建托盘
func watchTaskbarCreated(
	tray *application.SystemTray,
	_ *application.WebviewWindow,
	icon []byte,
	menu *application.Menu,
) {
	// 注册系统消息
	name, _ := syscall.UTF16PtrFromString("TaskbarCreated")
	msgID := win.RegisterWindowMessage((*uint16)(unsafe.Pointer(name)))

	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) > 0 {
		if msg.Message == msgID {
			tray.SetIcon(icon)
			tray.SetMenu(menu)
		}
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}
}
