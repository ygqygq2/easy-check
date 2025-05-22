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

type HostLatencyData struct {
	Host       string  `json:"host"`
	MinLatency float64 `json:"min_latency"`
	AvgLatency float64 `json:"avg_latency"`
	MaxLatency float64 `json:"max_latency"`
	PacketLoss float64 `json:"packet_loss"`
}

// HostsLatencyResponse
type HostsLatencyResponse struct {
	Hosts []HostLatencyData `json:"hosts"`
	Total int               `json:"total"`
	Error string            `json:"error,omitempty"`
}
