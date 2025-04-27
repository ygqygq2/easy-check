package checker

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type ICMPPinger struct{}

func (p *ICMPPinger) Ping(host string, count int, timeout int) (string, error) {
	ipAddr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return "", fmt.Errorf("failed to resolve host: %v", err)
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "")
	if err != nil {
		return "", fmt.Errorf("failed to listen for ICMP packets: %v", err)
	}
	defer conn.Close()

	var buffer bytes.Buffer
	for i := 0; i < count; i++ {
		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   i + 1,
				Seq:  i + 1,
				Data: []byte("PING"),
			},
		}
		msgBytes, err := msg.Marshal(nil)
		if err != nil {
			return "", fmt.Errorf("failed to marshal ICMP message: %v", err)
		}

		start := time.Now()
		_, err = conn.WriteTo(msgBytes, ipAddr)
		if err != nil {
			return "", fmt.Errorf("failed to send ICMP message: %v", err)
		}

		conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
		reply := make([]byte, 1500)
		_, _, err = conn.ReadFrom(reply)
		if err != nil {
			buffer.WriteString(fmt.Sprintf("Request timeout for icmp_seq %d\n", i+1))
		} else {
			duration := time.Since(start)
			buffer.WriteString(fmt.Sprintf("Reply from %s: time=%v\n", ipAddr.String(), duration))
		}

		// 等待 1 秒再发送下一个包
		if i < count-1 {
			time.Sleep(1 * time.Second)
		}
	}

	return buffer.String(), nil
}

func (p *ICMPPinger) ParsePingOutput(lines []string, count int) (int, float64, float64, float64) {
	successCount := 0
	var latencies []float64

	// 遍历每一行输出，提取延迟值
	for _, line := range lines {
		if strings.Contains(line, "Reply from") {
			successCount++

			// 提取延迟值，例如 "time=12ms"
			start := strings.Index(line, "time=")
			if start != -1 {
				end := strings.Index(line[start:], " ")
				if end == -1 {
					end = len(line)
				} else {
					end += start
				}

				latencyStr := strings.TrimSuffix(line[start+5:end], "ms")
				latency, err := strconv.ParseFloat(latencyStr, 64)
				if err == nil {
					latencies = append(latencies, latency)
				}
			}
		}
	}

	// 如果没有延迟值，返回默认值
	if len(latencies) == 0 {
		return successCount, 0, 0, 0
	}

	// 计算最小、最大和平均延迟
	minLatency := latencies[0]
	maxLatency := latencies[0]
	var totalLatency float64

	for _, latency := range latencies {
		if latency < minLatency {
			minLatency = latency
		}
		if latency > maxLatency {
			maxLatency = latency
		}
		totalLatency += latency
	}

	avgLatency := totalLatency / float64(len(latencies))
	return successCount, minLatency, avgLatency, maxLatency
}
