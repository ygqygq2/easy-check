package main

import (
	"context"
	"easy-check/internal/config"
	"easy-check/internal/constants"
	"easy-check/internal/initializer"
	"easy-check/internal/update"
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

func (a *App) shutdown(ctx context.Context) {
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

// GetSharedConstant 获取共享常量
func (a *App) GetSharedConstant() *constants.SharedConstants {
    constInfo := constants.GetSharedConstants(a.appCtx)
    return &constInfo
}

// CheckForUpdates 检查更新
func (a *App) CheckForUpdates() string {
  constInfo := constants.GetSharedConstants(a.appCtx)
  err := update.CheckAndUpdate(constInfo.UpdateServer, a.appCtx.AppVersion)
  if err != nil {
      return fmt.Sprintf("检查更新失败: %v", err)
  }
  return "更新成功！请重新启动应用程序。"
}
