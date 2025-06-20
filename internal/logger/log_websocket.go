package logger

import (
	"bufio"
	"io"
	"net/http"
	"os"
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

	// 从文件末尾开始读取
	file.Seek(0, io.SeekEnd)
	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			time.Sleep(1 * time.Second) // 如果没有新内容，等待一段时间
			continue
		}

		// 将日志内容发送给前端
		err = conn.WriteMessage(websocket.TextMessage, []byte(line))
		if err != nil {
			break
		}
	}
}
