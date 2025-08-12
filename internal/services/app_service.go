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
	"time"

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

// FrontendConfig 前端需要的配置信息结构体
type FrontendConfig struct {
	PingInterval   int `json:"pingInterval"`   // ping间隔时间
	GlobalInterval int `json:"globalInterval"` // 全局间隔时间
	HostsCount     int `json:"hostsCount"`     // 主机数量
}

// GetFrontendConfig 获取前端需要的配置信息
// 只返回前端实际使用的配置项，而不是完整的配置文件
func (a *AppService) GetFrontendConfig() (*FrontendConfig, error) {
	if a.appCtx == nil || a.appCtx.Config == nil {
		return nil, fmt.Errorf("配置未初始化")
	}

	config := a.appCtx.Config
	
	// 获取ping间隔，如果ping.interval未设置则使用全局interval
	pingInterval := config.Interval
	if config.Ping.Interval > 0 {
		pingInterval = config.Ping.Interval
	}

	return &FrontendConfig{
		PingInterval:   pingInterval,
		GlobalInterval: config.Interval,
		HostsCount:     len(config.Hosts),
	}, nil
}

// GetConfigValue 根据YAML路径获取配置值
// 支持路径如: "ping.interval", "interval", "ping.count" 等
func (a *AppService) GetConfigValue(path string) (interface{}, error) {
	if a.appCtx == nil || a.appCtx.Config == nil {
		return nil, fmt.Errorf("配置未初始化")
	}

	config := a.appCtx.Config
	
	// 解析配置路径
	switch path {
	case "ping.interval":
		if config.Ping.Interval > 0 {
			return config.Ping.Interval, nil
		}
		return config.Interval, nil // fallback到全局interval
	case "ping.count":
		return config.Ping.Count, nil
	case "ping.timeout":
		return config.Ping.Timeout, nil
	case "ping.loss_rate":
		return config.Ping.LossRate, nil
	case "interval":
		return config.Interval, nil
	case "hosts.count":
		return len(config.Hosts), nil
	case "db.retention":
		return config.Db.Retention, nil
	case "log.console_level":
		return config.Log.ConsoleLevel, nil
	case "log.file_level":
		return config.Log.FileLevel, nil
	default:
		return nil, fmt.Errorf("不支持的配置路径: %s", path)
	}
}

// AppInfo 应用程序信息结构体
type AppInfo struct {
	AppName         string       `json:"appName"`         // 应用名称
	AppVersion      string       `json:"appVersion"`      // 应用版本
	Author          string       `json:"author"`          // 作者
	Copyright       string       `json:"copyright"`       // 版权信息
	License         string       `json:"license"`         // 许可证
	Repository      string       `json:"repository"`      // 代码仓库
	Description     string       `json:"description"`     // 应用描述
	BuildTime       string       `json:"buildTime"`       // 构建时间
	GoVersion       string       `json:"goVersion"`       // Go版本
	PlatformInfo    PlatformInfo `json:"platformInfo"`    // 平台信息
	UpdateServer    string       `json:"updateServer"`    // 更新服务器
	NeedsRestart    bool         `json:"needsRestart"`    // 是否需要重启
}

// PlatformInfo 平台信息结构体
type PlatformInfo struct {
	OS   string `json:"os"`   // 操作系统
	Arch string `json:"arch"` // 架构
}

// GetAppInfo 获取应用程序信息（类似VSCode的关于页面）
func (a *AppService) GetAppInfo() (*AppInfo, error) {
	if a.appCtx == nil {
		return nil, fmt.Errorf("应用上下文未初始化")
	}

	return &AppInfo{
		AppName:     "Easy Check",
		AppVersion:  a.appCtx.AppVersion,
		Author:      "Chinge Yang (ygqygq2)",
		Copyright:   "Copyright (c) 2025 Chinge Yang",
		License:     "MIT License",
		Repository:  "https://github.com/ygqygq2/easy-check",
		Description: "简单网络检测工具 - 定期检测网络连接状态，支持多种告警方式",
		BuildTime:   "2025-08-11", // TODO: 可以通过ldflags注入实际构建时间
		GoVersion:   "go1.21+",    // TODO: 可以通过runtime.Version()获取
		PlatformInfo: PlatformInfo{
			OS:   a.appCtx.PlatformInfo.OS,
			Arch: a.appCtx.PlatformInfo.Arch,
		},
		UpdateServer: "https://easy-check-server.ygqygq2.com",
		NeedsRestart: a.appCtx.NeedsRestart,
	}, nil
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

// GetHistoryWithHosts 获取主机历史数据
func (a *AppService) GetHistoryWithHosts(hosts []string, startTime, endTime int64, step int64) *types.HostsRangeResponse {
	metrics := []string{"min_latency", "avg_latency", "max_latency", "packet_loss"}

	if endTime <= startTime {
		return &types.HostsRangeResponse{Hosts: nil, Total: 0, Error: "endTime must be > startTime"}
	}

	// 动态步长：如果前端传来的 step<=0 或过小，则按范围与预期点数估算
	// 自适应 maxPoints：按主机数限制总点数上限，将每条序列的 maxPoints = clamp(totalCap/len(hosts), 300, 1200)
	// 这里把传入的 step 当作“最小允许步长”兼容旧调用（旧前端会传硬编码值）
	rangeMs := endTime - startTime
	const totalPointsCap = 3000 // 所有主机总点数目标上限（近似）
	perSeriesMin := int64(300)
	perSeriesMax := int64(1200)
	hostsLen := int64(len(hosts))
	if hostsLen == 0 {
		hostsLen = 1
	}
	effMaxPoints := totalPointsCap / hostsLen
	if effMaxPoints < perSeriesMin {
		effMaxPoints = perSeriesMin
	}
	if effMaxPoints > perSeriesMax {
		effMaxPoints = perSeriesMax
	}

	computedStep := pickStep(rangeMs, effMaxPoints, a.getPingIntervalSeconds())
	if step > 0 && step > computedStep { // 旧前端给的较大步长，沿用较大值以避免更多点
		computedStep = step
	}

	// 边界对齐到步长网格（避免“半桶”），end 向上对齐但不超过 now
	nowMs := time.Now().UnixMilli()
	// 转为秒做取整
	startSec := startTime / 1000
	endSec := endTime / 1000
	stepSec := computedStep
	if stepSec <= 0 {
		stepSec = a.getPingIntervalSeconds()
		if stepSec <= 0 {
			stepSec = 1
		}
	}
	alignedStartSec := (startSec / stepSec) * stepSec // 向下取整
	alignedEndSec := ((endSec + stepSec - 1) / stepSec) * stepSec // 向上取整
	if alignedEndSec*1000 > nowMs { // 不超过当前时间
		alignedEndSec = (nowMs / 1000 / stepSec) * stepSec
	}
	// 若对齐后 end<=start，强制至少一个步长
	if alignedEndSec <= alignedStartSec {
		alignedEndSec = alignedStartSec + stepSec
	}

	startT := time.Unix(alignedStartSec, 0)
	endT := time.Unix(alignedEndSec, 0)
	stepDuration := time.Duration(stepSec) * time.Second

	var hostRangeData []types.HostRangeData
	for _, host := range hosts {
		hostData := types.HostRangeData{Host: host, Series: make(map[string][]types.TimeSeriesPoint)}
		for _, metric := range metrics {
			rangeData, err := data.GetHostRangeMetrics(a.appCtx.TSDB, []string{host}, metric, startT, endT, stepDuration)
			if err != nil {
				return &types.HostsRangeResponse{Hosts: nil, Total: 0, Error: fmt.Sprintf("查询主机 %s 指标 %s 历史数据失败: %v", host, metric, err), StepSeconds: computedStep}
			}
			if points, ok := rangeData[host]; ok {
				seriesPoints := make([]types.TimeSeriesPoint, 0, len(points))
				for _, p := range points {
					seriesPoints = append(seriesPoints, types.TimeSeriesPoint{Timestamp: p.Timestamp, Value: p.Value})
				}
				hostData.Series[metric] = seriesPoints
			}
		}
		hostRangeData = append(hostRangeData, hostData)
	}

	return &types.HostsRangeResponse{Hosts: hostRangeData, Total: len(hostRangeData), Error: "", StepSeconds: stepSec}
}

// 获取当前有效的 ping 间隔（秒）
func (a *AppService) getPingIntervalSeconds() int64 {
	cfg := a.appCtx.Config
	if cfg != nil && cfg.Ping.Interval > 0 {
		return int64(cfg.Ping.Interval)
	}
	if cfg != nil && cfg.Interval > 0 { // 兜底全局 interval
		return int64(cfg.Interval)
	}
	return 30 // 默认 30s
}

// 计算步长，借鉴 Grafana：range / maxPoints 并对齐阶梯
func pickStep(rangeMs int64, maxPoints int64, scrapeIntervalSec int64) int64 {
	if rangeMs <= 0 {
		return scrapeIntervalSec
	}
	targetSec := (rangeMs / 1000) / maxPoints
	if (rangeMs/1000)%maxPoints != 0 { // 向上取整
		targetSec += 1
	}
	if targetSec < 1 {
		targetSec = 1
	}
	// 与抓取间隔和最小间隔比较
	if targetSec < scrapeIntervalSec {
		targetSec = scrapeIntervalSec
	}
	ladder := []int64{1, 2, 5, 10, 15, 30, 60, 120, 300, 600, 900, 1800, 3600, 7200, 21600, 43200, 86400}
	chosen := ladder[len(ladder)-1]
	for _, v := range ladder {
		if v >= targetSec {
			chosen = v
			break
		}
	}
	// 安全点数限制，再校正（避免返回 > 1.5*maxPoints 的点数）
	if (rangeMs/1000)/chosen > int64(float64(maxPoints)*1.5) {
		// 尝试升一个阶梯
		for i, v := range ladder {
			if v == chosen && i+1 < len(ladder) {
				chosen = ladder[i+1]
				break
			}
		}
	}
	return chosen
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
