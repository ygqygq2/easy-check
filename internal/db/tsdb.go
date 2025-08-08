package db

import (
	"context"
	"easy-check/internal/config"
	"easy-check/internal/utils"
	"fmt"
	"math/rand"
	"os"
	"strconv"
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
	if isDev {
		// 开发模式下使用随机路径
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomSuffix := r.Intn(10000)
		tsdbPath = fmt.Sprintf("tsdb-dev-%d", randomSuffix)
	} else {
		// 生产模式下使用固定路径
		tsdbPath = "tsdb"
	}

	// 确保路径存在
	path := utils.AddDirectorySuffix(dbConfig.Path) + tsdbPath
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
