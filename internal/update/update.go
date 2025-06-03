package update

import (
	"archive/zip"
	"bytes"
	"easy-check/internal/initializer"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/minio/selfupdate"
)

// 定义响应结构体
type VersionInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	DownloadUrl string `json:"downloadUrl"`
}
type UpdateResult struct {
	Updated      bool   // 是否进行了更新
	NeedsRestart bool   // 是否需要重启
	Message      string // 状态消息
}

// 解压 ZIP 文件到指定目录
func unzip(src []byte, dest string) error {
	reader, err := zip.NewReader(bytes.NewReader(src), int64(len(src)))
	if err != nil {
		return fmt.Errorf("解压失败: %v", err)
	}

	for _, file := range reader.File {
		filePath := filepath.Join(dest, file.Name)
		if file.FileInfo().IsDir() {
			// 创建目录
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				return fmt.Errorf("创建目录失败: %v", err)
			}
			continue
		}

		// 创建文件
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return fmt.Errorf("创建文件目录失败: %v", err)
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("创建文件失败: %v", err)
		}

		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("打开压缩文件失败: %v", err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
	}

	return nil
}

func findExecutable(tempDir, goos, arch string) (string, error) {
	var executableName string
	if runtime.GOOS == "windows" {
		executableName = fmt.Sprintf("easy-check-ui-%s-%s.exe", goos, arch)
	} else {
		executableName = fmt.Sprintf("easy-check-ui-%s-%s", goos, arch)
	}

	var executablePath string
	err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), executableName) {
			executablePath = path
			return io.EOF // 提前终止 Walk
		}
		return nil
	})

	if err != nil && err != io.EOF {
		return "", fmt.Errorf("查找可执行文件失败: %v", err)
	}
	if executablePath == "" {
		return "", fmt.Errorf("未找到可执行文件: %s", executableName)
	}
	return executablePath, nil
}

// CheckAndUpdate 检查更新并执行更新
func CheckAndUpdate(appCtx *initializer.AppContext, UpdateServer string) (*UpdateResult, error) {
	goos := appCtx.PlatformInfo.OS
	arch := appCtx.PlatformInfo.Arch
	currentVersion := appCtx.AppVersion

	// 1. 从 latestUrl 获取最新版本信息
	latestUrl := fmt.Sprintf("%s/api/versions/latest?os=%s&arch=%s", UpdateServer, goos, arch)
	fmt.Printf("检查最新版本地址: %s\n", latestUrl)

	client := &http.Client{}
	resp, err := client.Get(latestUrl)
	if err != nil {
		return nil, fmt.Errorf("无法访问最新版本地址: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取最新版本失败，HTTP 状态码: %d", resp.StatusCode)
	}

	// 解析最新版本信息
	var versionInfo VersionInfo
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取最新版本响应失败: %v", err)
	}
	err = json.Unmarshal(body, &versionInfo)
	if err != nil {
		return nil, fmt.Errorf("解析最新版本信息失败: %v", err)
	}
	fmt.Printf("最新版本: %s\n", versionInfo.Name)

	// 去掉版本号前缀 "v"
	currentVersion = strings.TrimPrefix(currentVersion, "v")
	latestVersion := strings.TrimPrefix(versionInfo.Name, "v")

	// 2. 使用 semver 对比当前版本号和最新版本号
	currentVer, err := semver.NewVersion(currentVersion)
	if err != nil {
		fmt.Printf("当前版本号格式错误: %v，使用默认版本 0.0.1\n", err)
		currentVersion = "0.0.1"
		currentVer, err = semver.NewVersion(currentVersion)
		if err != nil {
			return nil, fmt.Errorf("默认版本号解析失败: %v", err)
		}
	}
	latestVer, err := semver.NewVersion(latestVersion)
	if err != nil {
		return nil, fmt.Errorf("最新版本号格式错误: %v", err)
	}

	if !currentVer.LessThan(latestVer) {
		fmt.Println("当前已是最新版本，无需更新。")
		return &UpdateResult{
			Updated:      false,
			NeedsRestart: false,
			Message:      "当前已是最新版本，无需更新。",
		}, nil
	}

	// 3. 如果当前版本号小于最新版本号，则执行更新
	updateURL := versionInfo.DownloadUrl
	fmt.Printf("发现新版本，下载地址: %s\n", updateURL)

	resp, err = client.Get(updateURL)
	if err != nil {
		return nil, fmt.Errorf("无法访问更新地址: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载更新包失败，HTTP 状态码: %d", resp.StatusCode)
	}

	// 读取 ZIP 文件内容
	zipData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取更新包失败: %v", err)
	}

	// 解压 ZIP 文件到临时目录
	tempDir, err := os.MkdirTemp("", "update")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = unzip(zipData, tempDir)
	if err != nil {
		return nil, fmt.Errorf("解压更新包失败: %v", err)
	}

	// 找到解压后的可执行文件路径
	executablePath, err := findExecutable(tempDir, goos, arch)
	if err != nil {
		return nil, err
	}

	// 执行更新
	fmt.Println("开始更新…")
	executableFile, err := os.Open(executablePath)
	if err != nil {
		return nil, fmt.Errorf("打开解压后的可执行文件失败: %v", err)
	}
	defer executableFile.Close()

	err = selfupdate.Apply(executableFile, selfupdate.Options{})
	if err != nil {
		return nil, fmt.Errorf("更新失败: %v", err)
	}

	// 设置需要重启标记
	appCtx.NeedsRestart = true

	// 4. 提示更新成功
	fmt.Println("更新成功！请重新启动应用程序。")
	return &UpdateResult{
		Updated:      true,
		NeedsRestart: true,
		Message:      "更新成功！请重新启动应用程序。",
	}, nil
}
