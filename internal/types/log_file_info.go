// types.go
package types

// LogFileInfo 包含日志文件的详细信息
type LogFileInfo struct {
	Name    string `json:"name"`
	ModTime int64  `json:"modTime"` // Unix a a second
	Size    int64  `json:"size"`
}
