package services

import (
	"context"
	"easy-check/internal/config"
	"easy-check/internal/constants"
	"easy-check/internal/data"
	"easy-check/internal/initializer"
	"easy-check/internal/types"
	"easy-check/internal/update"
	"easy-check/internal/utils"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// AppService struct
type AppService struct {
	ctx    context.Context
	appCtx *initializer.AppContext
}

// NewAppService creates a new AppService
func NewAppService(appCtx *initializer.AppContext) *AppService {
	return &AppService{
		appCtx: appCtx,
	}
}

func (a *AppService) GetCurrentInstanceInfo() map[string]interface{} {
	return map[string]interface{}{
		"args":       os.Args,
		"workingDir": utils.GetCurrentWorkingDir(),
	}
}

// Startup is called when the service starts
func (a *AppService) ServiceStartup(ctx context.Context, options application.ServiceOptions) {
	a.ctx = ctx
}

// Shutdown is called when the service shuts down
func (a *AppService) ServiceShutdown(ctx context.Context, options application.ServiceOptions) {
	if a.appCtx != nil && a.appCtx.DB != nil && a.appCtx.DB.Instance != nil {
		a.appCtx.DB.Instance.Close()
	}
	if a.appCtx != nil && a.appCtx.TSDB != nil {
		a.appCtx.TSDB.Close()
	}
}

// GetConfig retrieves the configuration file content
func (a *AppService) GetConfig() (string, error) {
	content, err := config.GetConfigFromFile(a.appCtx.ConfigPath)
	if err != nil {
		return "", fmt.Errorf("获取配置失败: %v", err)
	}
	return content, nil
}

// SaveConfig saves the configuration file content
func (a *AppService) SaveConfig(content string) error {
	err := config.SaveConfigToFile(a.appCtx.ConfigPath, content, a.appCtx.Logger)
	if err != nil {
		a.appCtx.Logger.Log(fmt.Sprintf("保存配置失败: %v", err), "error")
		return fmt.Errorf("保存失败: %v", err)
	}
	return nil
}

// GetSharedConstant retrieves shared constants
func (a *AppService) GetSharedConstant() *constants.SharedConstants {
	constInfo := constants.GetSharedConstants(a.appCtx)
	return &constInfo
}

// CheckForUpdates checks for updates
func (a *AppService) CheckForUpdates() string {
	constInfo := constants.GetSharedConstants(a.appCtx)
	result, err := update.CheckAndUpdate(a.appCtx, constInfo.UpdateServer)
	if err != nil {
		return fmt.Sprintf("检查更新失败: %v", err)
	}
	if result.Updated {
		if result.NeedsRestart {
			return fmt.Sprintf("%s 请重新启动应用程序。", result.Message)
		}
		return result.Message
	}

	return result.Message
}

// RestartApp restarts the application
func (a *AppService) RestartApp() error {
	return update.RestartApp()
}

// GetHosts retrieves the list of hosts
func (a *AppService) GetHosts(page int, pageSize int, searchTerm string) *types.HostsResponse {
	hosts, total, err := data.GetHostsFromBadgerWithPagination(a.appCtx.DB, page, pageSize, searchTerm)
	if err != nil {
		return &types.HostsResponse{
			Hosts: nil,
			Total: 0,
			Error: fmt.Sprintf("获取主机列表失败: %v", err),
		}
	}

	var result []types.Host
	for _, h := range hosts {
		result = append(result, types.Host{
			Host:        h.Host,
			Description: h.Description,
		})
	}

	return &types.HostsResponse{
		Hosts: result,
		Total: total,
		Error: "",
	}
}

// GetStatusWithHosts retrieves latency data and status for hosts
func (a *AppService) GetStatusWithHosts(hosts []string) *types.HostsStatusResponse {
	metrics := []string{"min_latency", "avg_latency", "max_latency", "packet_loss"}
	latencyData := make([]types.HostStatusData, 0)

	// 获取主机状态
	statusMap, err := data.GetHostStatus(a.appCtx.DB, hosts)
	if err != nil {
		return &types.HostsStatusResponse{
			Hosts: nil,
			Total: 0,
			Error: fmt.Sprintf("查询主机状态失败: %v", err),
		}
	}

	for _, metric := range metrics {
		data, err := data.GetHostMetrics(a.appCtx.TSDB, hosts, metric)
		if err != nil {
			return &types.HostsStatusResponse{
				Hosts: nil,
				Total: 0,
				Error: fmt.Sprintf("查询指标 %s 失败: %v", metric, err),
			}
		}

		for host, value := range data {
			var existing *types.HostStatusData
			for i := range latencyData {
				if latencyData[i].Host == host {
					existing = &latencyData[i]
					break
				}
			}

			if existing == nil {
				latencyData = append(latencyData, types.HostStatusData{Host: host})
				existing = &latencyData[len(latencyData)-1]
			}

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

	// 合并主机状态到返回数据
	for i := range latencyData {
		if status, ok := statusMap[latencyData[i].Host]; ok {
			latencyData[i].Status = string(status)
		}
	}

	return &types.HostsStatusResponse{
		Hosts: latencyData,
		Total: len(latencyData),
		Error: "",
	}
}

// GetLogFiles retrieves the list of log files
func (a *AppService) GetLogFiles() ([]string, error) {
	logFilePath := a.appCtx.Config.Log.File
	logDir := filepath.Dir(logFilePath)
	if logDir == "" {
		return nil, fmt.Errorf("日志目录未配置，请检查配置文件")
	}

	files, err := os.ReadDir(logDir)
	if err != nil {
		return nil, fmt.Errorf("无法读取日志目录: %v", err)
	}

	var logFiles []string
	for _, file := range files {
		if !file.IsDir() && file.Name() != "app.log" && file.Name() != "error.log" {
			logFiles = append(logFiles, file.Name())
		}
	}

	return logFiles, nil
}

// GetLogFileContent retrieves the content of a specific log file
func (a *AppService) GetLogFileContent(fileName string, isLatest ...bool) (string, error) {
	logFilePath := a.appCtx.Config.Log.File
	logDir := filepath.Dir(logFilePath)

	// 检查 isLatest 参数是否传递，默认为 false
	useLatest := len(isLatest) > 0 && isLatest[0]

	// 如果 useLatest 为 true，直接读取 logFilePath
	if useLatest {
		content, err := os.ReadFile(logFilePath)
		if err != nil {
			return "", fmt.Errorf("无法读取最新日志文件: %v", err)
		}
		return string(content), nil
	}

	// 否则读取指定的日志文件
	filePath := filepath.Join(logDir, fileName)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("无法读取日志文件 %s: %v", fileName, err)
	}

	return string(content), nil
}
