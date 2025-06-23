package logger

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

// ServeLogWebSocket 启动 WebSocket 服务，实时推送日志内容
func ServeLogWebSocket(logFilePath string, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	file, err := os.Open(logFilePath)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("无法打开日志文件"))
		return
	}
	defer file.Close()

	file.Seek(0, io.SeekEnd)
	reader := bufio.NewReader(file)

	timePrefix := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} \[`)
	var buffer strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if buffer.Len() > 0 {
				conn.WriteMessage(websocket.TextMessage, []byte(buffer.String()))
				buffer.Reset()
			}
			time.Sleep(1 * time.Second)
			continue
		}

		if timePrefix.MatchString(line) && buffer.Len() > 0 {
			conn.WriteMessage(websocket.TextMessage, []byte(buffer.String()))
			buffer.Reset()
		}
		buffer.WriteString(line)
	}
}
