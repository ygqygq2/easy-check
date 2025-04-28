package main

import (
	"context"
	"easy-check/internal/config"
	"easy-check/internal/initializer"
	"fmt"
)

// App struct
type App struct {
	ctx    context.Context
	appCtx *initializer.AppContext
}

// NewApp creates a new App application struct
func NewApp(appCtx *initializer.AppContext) *App {
	return &App{
		appCtx: appCtx,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	// 启动逻辑
	a.ctx = ctx
}

// GetConfig 获取配置文件内容
func (a *App) GetConfig() (string, error) {
	content, err := config.GetConfigContent(a.appCtx.ConfigPath)
	if err != nil {
		// a.appCtx.Logger.Log(fmt.Sprintf("获取配置失败: %v", err), "error")
		return "", fmt.Errorf("获取配置失败: %v", err)
	}
	return content, nil
}

// SaveConfig 保存配置文件内容
func (a *App) SaveConfig(content string) error {
	err := config.SaveConfigContent(a.appCtx.ConfigPath, content, a.appCtx.Logger)
	if err != nil {
		a.appCtx.Logger.Log(fmt.Sprintf("保存配置失败: %v", err), "error")
		return fmt.Errorf("保存失败: %v", err)
	}
	return nil
}
