package assets

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
)

//go:embed *
var embeddedAssets embed.FS

// GetAsset retrieves a file from the embedded assets.
func GetAsset(filePath string) ([]byte, error) {
	// 验证路径是否合法
	if !fs.ValidPath(filePath) {
		return nil, errors.New("invalid file path")
	}

	// 确保路径是相对路径，避免路径注入
	cleanPath := path.Clean(filePath)

	// 读取文件
	data, err := embeddedAssets.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read asset %s: %w", cleanPath, err)
	}
	return data, nil
}

// GetAssetsFS returns the embedded file system for advanced use cases.
func GetAssetsFS() fs.FS {
	return embeddedAssets
}
