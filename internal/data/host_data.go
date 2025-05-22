package data

import (
	"easy-check/internal/config"
	"easy-check/internal/db"
	"encoding/json"
	"fmt"
)

// HostData 包含主机的状态和性能数据
type HostData struct {
	Host     string             `json:"host"`
	TSDBData map[string]float64 `json:"tsdb_data"`
}

// GetHostsFromBadgerWithPagination 从 BadgerDB 分页获取主机列表
func GetHostsFromBadgerWithPagination(db *db.DB, page, pageSize int) ([]config.Host, int, error) {
	val, err := db.Get("hosts")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get hosts from BadgerDB: %v", err)
	}
	var hosts []config.Host
	if err := json.Unmarshal([]byte(val), &hosts); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal hosts: %v", err)
	}
	total := len(hosts)
	start := (page - 1) * pageSize
	if start > total {
		return []config.Host{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return hosts[start:end], total, nil
}

func GetHostMetrics(tsdb *db.TSDB, hosts []string, metric string) (map[string]float64, error) {
	return tsdb.QueryLatestMetricsForHosts(hosts, metric)
}
