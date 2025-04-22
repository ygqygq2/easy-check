package checker

import (
	"bytes"
	"fmt"
	"net"
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

func (p *ICMPPinger) ParsePingOutput(lines []string, count int) (int, string) {
	successCount := 0
	var sampleLatency string

	for _, line := range lines {
		if strings.Contains(line, "Reply from") {
			successCount++
			if sampleLatency == "" {
				sampleLatency = line
			}
		}
	}

	return successCount, sampleLatency
}
