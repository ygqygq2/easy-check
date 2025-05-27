package data

import (
	"easy-check/internal/config"
	"easy-check/internal/db"
	"encoding/json"
	"fmt"
	"strings"
)

// HostData 包含主机的状态和性能数据
type HostData struct {
	Host     string             `json:"host"`
	TSDBData map[string]float64 `json:"tsdb_data"`
}

// GetHostsFromBadgerWithPagination 从 BadgerDB 分页获取主机列表
func GetHostsFromBadgerWithPagination(db *db.DB, page, pageSize int, searchTerm string) ([]config.Host, int, error) {
	val, err := db.Get("hosts")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get hosts from BadgerDB: %v", err)
	}
	var hosts []config.Host
	if err := json.Unmarshal([]byte(val), &hosts); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal hosts: %v", err)
	}

	// 如果 searchTerm 不为空，进行过滤
	if searchTerm != "" {
		filteredHosts := []config.Host{}
		for _, host := range hosts {
			if containsIgnoreCase(host.Host, searchTerm) || containsIgnoreCase(host.Description, searchTerm) {
				filteredHosts = append(filteredHosts, host)
			}
		}
		hosts = filteredHosts
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

// containsIgnoreCase 检查字符串是否包含子字符串（忽略大小写）
func containsIgnoreCase(str, substr string) bool {
	return str != "" && substr != "" && strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

func GetHostMetrics(tsdb *db.TSDB, hosts []string, metric string) (map[string]float64, error) {
	return tsdb.QueryLatestMetricsForHosts(hosts, metric)
}
