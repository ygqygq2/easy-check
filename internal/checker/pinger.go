package checker

// Pinger 定义了ping主机的接口
type Pinger interface {
	// Ping 执行ping操作，返回输出结果，错误
	Ping(host string, count int, timeout int) (string, error)

	// ParsePingOutput 解析ping输出，返回成功次数和最小、平均、最大延迟
	ParsePingOutput(lines []string, count int) (int, float64, float64, float64)
}

// NewPinger 函数在相应的平台特定文件中实现
// 见 pinger_linux.go 和 pinger_windows.go
