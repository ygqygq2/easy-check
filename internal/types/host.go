package types

// Host 定义前端需要的主机类型
type Host struct {
	Host        string `json:"host"`
	Description string `json:"description"`
}

// HostsResponse 定义返回给前端的结构体
type HostsResponse struct {
	Hosts []Host `json:"hosts"`
	Total int    `json:"total"`
	Error string `json:"error,omitempty"`
}

type HostStatusData struct {
	Host       string  `json:"host"`
	MinLatency float64 `json:"min_latency"`
	AvgLatency float64 `json:"avg_latency"`
	MaxLatency float64 `json:"max_latency"`
	PacketLoss float64 `json:"packet_loss"`
	Status     string  `json:"status"`
	// Sent       bool    `json:"sent"`
}

// HostsStatusResponse
type HostsStatusResponse struct {
	Hosts []HostStatusData `json:"hosts"`
	Total int              `json:"total"`
	Error string           `json:"error,omitempty"`
}

// TimeSeriesPoint 时间序列数据点
type TimeSeriesPoint struct {
	Timestamp int64   `json:"timestamp"` // 毫秒时间戳
	Value     float64 `json:"value"`     // 数据值
}

// HostRangeData 主机历史数据
type HostRangeData struct {
	Host   string            `json:"host"`
	Series map[string][]TimeSeriesPoint `json:"series"` // metric -> points
}

// HostsRangeResponse 主机历史数据响应
type HostsRangeResponse struct {
	Hosts []HostRangeData `json:"hosts"`
	Total int             `json:"total"`
	Error string          `json:"error,omitempty"`
}
