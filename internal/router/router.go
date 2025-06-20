package router

import (
	"easy-check/internal/initializer"
	"easy-check/internal/logger"
	"net/http"
)

func RegisterRoutes(appCtx *initializer.AppContext) {
	// 其它路由...
	http.HandleFunc("/ws/logs", func(w http.ResponseWriter, r *http.Request) {
		// 从配置里拿日志文件路径
		logFilePath := appCtx.Config.Log.File
		logger.ServeLogWebSocket(logFilePath, w, r)
	})
}
