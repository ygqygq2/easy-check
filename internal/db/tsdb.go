package db

import (
	"context"
	"easy-check/internal/config"
	"easy-check/internal/utils"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/tsdb"
)

type TSDB struct {
	db *tsdb.DB
	// 添加写入锁保护并发写入
	writeMu sync.Mutex
}

// NewTSDB 初始化 Prometheus TSDB
func NewTSDB(isDev bool, dbConfig *config.DbConfig) (*TSDB, error) {
	var tsdbPath string
	var path string
	
	if isDev {
		// 开发模式：使用时间戳命名，保证新目录总是最新的
		basePath := utils.AddDirectorySuffix(dbConfig.Path)
		timestamp := time.Now().Format("20060102-150405")
		tsdbPath = fmt.Sprintf("tsdb-dev-%s", timestamp)
		path = basePath + tsdbPath
		
		if err := copyFromLatestDevDirectory(basePath, path); err != nil {
			// 复制失败不应该阻止启动，只记录警告
			fmt.Printf("Warning: Failed to copy data from latest dev directory: %v\n", err)
		}
	} else {
		// 生产模式下使用固定路径
		tsdbPath = "tsdb"
		path = utils.AddDirectorySuffix(dbConfig.Path) + tsdbPath
	}

	// 确保路径存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, fmt.Errorf("failed to create TSDB directory: %v", err)
		}
	}

	// 解析 retention 配置
	retentionDuration, err := parseRetention(dbConfig.Retention)
	if err != nil {
		return nil, fmt.Errorf("invalid retention duration: %v", err)
	}

	// 配置 TSDB
	opts := tsdb.DefaultOptions()
	opts.RetentionDuration = retentionDuration // 设置保留时间

	db, err := tsdb.Open(path, nil, nil, opts, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open TSDB: %v", err)
	}

	return &TSDB{db: db}, nil
}

// parseRetention 将 "30d" 等格式转换为秒数
func parseRetention(retention string) (int64, error) {
	unit := retention[len(retention)-1:]
	value := retention[:len(retention)-1]

	days, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid retention format: %v", err)
	}

	switch unit {
	case "d":
		return int64(days * 24 * 60 * 60 * 1000), nil
	case "h":
		return int64(days * 60 * 60 * 1000), nil
	case "m":
		return int64(days * 60 * 1000), nil
	default:
		return 0, fmt.Errorf("unsupported retention unit: %s", unit)
	}
}

// Close 关闭 TSDB
func (t *TSDB) Close() error {
	return t.db.Close()
}

// AppendMetrics 批量添加监控数据
func (t *TSDB) AppendMetrics(metrics map[string]float64, timestamp int64, labelsMap map[string]string) error {
	// 加锁保护并发写入
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	
	app := t.db.Appender(context.Background())
	defer func() {
		// 确保在出错时也能正确回滚
		if err := recover(); err != nil {
			app.Rollback()
			panic(err)
		}
	}()

	for metric, value := range metrics {
		// 构建标签，包含指标名称和额外标签
		metricLabels := map[string]string{"__name__": metric}
		for k, v := range labelsMap {
			metricLabels[k] = v
		}

		// 写入数据点
		if _, err := app.Append(0, labels.FromMap(metricLabels), timestamp, value); err != nil {
			app.Rollback()
			return fmt.Errorf("failed to append metric %s with value %f: %v", metric, value, err)
		}
	}

	// 提交数据
	if err := app.Commit(); err != nil {
		return fmt.Errorf("failed to commit metrics: %v", err)
	}

	return nil
}

// QueryLatestMetricsForHosts 使用 PromQL 查询多个主机的最新监控数据
func (t *TSDB) QueryLatestMetricsForHosts(hosts []string, metric string) (map[string]float64, error) {
	result := make(map[string]float64)
	engine := promql.NewEngine(promql.EngineOpts{
		MaxSamples: 50000,
		Timeout:    10 * time.Second,
	})
	queryable := t.db

	// 构造 PromQL 表达式
	// 例如 metric{host=~"host1|host2|host3"}
	var hostFilter string
	for i, h := range hosts {
		if i > 0 {
			hostFilter += "|"
		}
		hostFilter += h // 直接拼接主机名
	}
	expr := fmt.Sprintf(`%s{host=~"%s"}`, metric, hostFilter) // 将整个正则表达式用双引号包裹

	// 指定查询时间点
	queryTime := time.Now()

	// 创建 PromQL 查询
	ctx := context.Background()
	q, err := engine.NewInstantQuery(ctx, queryable, nil, expr, queryTime)
	if err != nil {
		return nil, fmt.Errorf("failed to create PromQL query: %v", err)
	}
	defer q.Close()

	// 执行查询
	res := q.Exec(ctx)
	if res.Err != nil {
		return nil, fmt.Errorf("failed to execute PromQL query: %v", res.Err)
	}

	// 解析查询结果
	if vector, ok := res.Value.(promql.Vector); ok {
		for _, sample := range vector {
			host := sample.Metric.Get("host") // 使用 Get 方法获取标签值
			if host != "" {
				result[host] = float64(sample.F)
			}
		}
	} else {
		return nil, fmt.Errorf("unexpected query result type: %T", res.Value)
	}

	return result, nil
}

// QueryRangeMetricsForHosts 使用 PromQL 查询多个主机的历史监控数据
func (t *TSDB) QueryRangeMetricsForHosts(hosts []string, metric string, startTime, endTime time.Time, step time.Duration) (map[string][]TimeSeriesPoint, error) {
	result := make(map[string][]TimeSeriesPoint)
	engine := promql.NewEngine(promql.EngineOpts{
		MaxSamples: 50000,
		Timeout:    30 * time.Second,
	})
	queryable := t.db

	// 构造 PromQL 表达式
	var hostFilter string
	for i, h := range hosts {
		if i > 0 {
			hostFilter += "|"
		}
		hostFilter += h
	}
	expr := fmt.Sprintf(`%s{host=~"%s"}`, metric, hostFilter)

	// 创建范围查询
	ctx := context.Background()
	q, err := engine.NewRangeQuery(ctx, queryable, nil, expr, startTime, endTime, step)
	if err != nil {
		return nil, fmt.Errorf("failed to create PromQL range query: %v", err)
	}
	defer q.Close()

	// 执行查询
	res := q.Exec(ctx)
	if res.Err != nil {
		return nil, fmt.Errorf("failed to execute PromQL range query: %v", res.Err)
	}

	// 解析查询结果
	if matrix, ok := res.Value.(promql.Matrix); ok {
		for _, series := range matrix {
			host := series.Metric.Get("host")
			if host == "" {
				continue
			}
			
			var points []TimeSeriesPoint
			for _, point := range series.Floats {
				points = append(points, TimeSeriesPoint{
					Timestamp: point.T,
					Value:     point.F,
				})
			}
			result[host] = points
		}
	} else {
		return nil, fmt.Errorf("unexpected query result type: %T", res.Value)
	}

	return result, nil
}

// TimeSeriesPoint 时间序列数据点
type TimeSeriesPoint struct {
	Timestamp int64   `json:"timestamp"` // 毫秒时间戳
	Value     float64 `json:"value"`     // 数据值
}

// copyFromLatestDevDirectory 从最新的开发目录复制数据到新目录
func copyFromLatestDevDirectory(basePath, newPath string) error {
	// 查找最新的开发目录
	latestDir, err := findLatestDevDirectory(basePath)
	if err != nil || latestDir == "" {
		return fmt.Errorf("no previous dev directory found")
	}
	
	latestPath := filepath.Join(basePath, latestDir)
	
	// 检查源目录是否存在且可读
	if _, err := os.Stat(latestPath); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", latestPath)
	}
	
	// 创建目标目录
	if err := os.MkdirAll(newPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}
	
	// 复制目录内容
	return copyDirectory(latestPath, newPath)
}

// findLatestDevDirectory 找到最新的开发目录
func findLatestDevDirectory(basePath string) (string, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return "", err
	}
	
	var devDirs []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "tsdb-dev-") {
			devDirs = append(devDirs, entry.Name())
		}
	}
	
	if len(devDirs) == 0 {
		return "", fmt.Errorf("no dev directories found")
	}
	
	// 按名称排序，最新的在最后
	sort.Strings(devDirs)
	return devDirs[len(devDirs)-1], nil
}

// copyDirectory 递归复制目录
func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// 计算目标路径
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)
		
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		
		return copyFile(path, dstPath)
	})
}

// copyFile 复制单个文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	
	// 复制文件权限
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	return os.Chmod(dst, sourceInfo.Mode())
}
