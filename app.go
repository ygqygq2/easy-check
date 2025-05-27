package main

import (
	"context"
	"easy-check/internal/config"
	"easy-check/internal/constants"
	"easy-check/internal/data"
	"easy-check/internal/initializer"
	"easy-check/internal/types"
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
	wailsContext = &ctx
}

func (a *App) shutdown(ctx context.Context) {
	// 关闭数据库连接
	if a.appCtx != nil && a.appCtx.DB != nil && a.appCtx.DB.Instance != nil {
		a.appCtx.DB.Instance.Close()
	}
	if a.appCtx != nil && a.appCtx.TSDB != nil {
		a.appCtx.TSDB.Close()
	}
}

// GetConfig 获取配置文件内容
func (a *App) GetConfig() (string, error) {
	content, err := config.GetConfigFromFile(a.appCtx.ConfigPath)
	if err != nil {
		// a.appCtx.Logger.Log(fmt.Sprintf("获取配置失败: %v", err), "error")
		return "", fmt.Errorf("获取配置失败: %v", err)
	}
	return content, nil
}

// SaveConfig 保存配置文件内容
func (a *App) SaveConfig(content string) error {
	err := config.SaveConfigToFile(a.appCtx.ConfigPath, content, a.appCtx.Logger)
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
	err := update.CheckAndUpdate(a.appCtx, constInfo.UpdateServer)
	if err != nil {
		return fmt.Sprintf("检查更新失败: %v", err)
	}
	return "更新成功！请重新启动应用程序。"
}

// RestartApp 重启应用程序
func (a *App) RestartApp() error {
	return update.RestartApp()
}

// GetHosts 获取主机列表
func (a *App) GetHosts(page int, pageSize int) *types.HostsResponse {
	hosts, total, err := data.GetHostsFromBadgerWithPagination(a.appCtx.DB, page, pageSize)
	if err != nil {
		// 如果发生错误，将错误信息封装到响应中
		return &types.HostsResponse{
			Hosts: nil,
			Total: 0,
			Error: fmt.Sprintf("获取主机列表失败: %v", err),
		}
	}

	// 转换为前端需要的类型
	var result []types.Host
	for _, h := range hosts {
		result = append(result, types.Host{
			Host:        h.Host,
			Description: h.Description,
		})
	}

	// 返回封装的结构体
	return &types.HostsResponse{
		Hosts: result,
		Total: total,
		Error: "",
	}
}

// GetLatencyWithHosts 获取主机延迟数据
func (a *App) GetLatencyWithHosts(hosts []string) *types.HostsLatencyResponse {
	// 定义需要查询的指标
	metrics := []string{"min_latency", "avg_latency", "max_latency", "packet_loss"}

	// 存储结果
	latencyData := make([]types.HostLatencyData, 0)

	// 遍历每个指标查询数据
	for _, metric := range metrics {
		data, err := data.GetHostMetrics(a.appCtx.TSDB, hosts, metric)
		if err != nil {
			return &types.HostsLatencyResponse{
				Hosts: nil,
				Total: 0,
				Error: fmt.Sprintf("查询指标 %s 失败: %v", metric, err),
			}
		}

		// 合并数据到 latencyData
		for host, value := range data {
			// 查找是否已有该主机的数据
			var existing *types.HostLatencyData
			for i := range latencyData {
				if latencyData[i].Host == host {
					existing = &latencyData[i]
					break
				}
			}

			// 如果没有，创建新的
			if existing == nil {
				latencyData = append(latencyData, types.HostLatencyData{Host: host})
				existing = &latencyData[len(latencyData)-1]
			}

			// 根据指标名称填充数据
			switch metric {
			case "min_latency":
				existing.MinLatency = value
			case "avg_latency":
				existing.AvgLatency = value
			case "max_latency":
				existing.MaxLatency = value
			case "packet_loss":
				existing.PacketLoss = value
			}
		}
	}

	// 返回结果
	return &types.HostsLatencyResponse{
		Hosts: latencyData,
		Total: len(latencyData),
		Error: "",
	}
}
