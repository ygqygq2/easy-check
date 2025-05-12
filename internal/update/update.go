package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// CheckAndUpdate 检查更新并执行更新
func CheckAndUpdate(UpdateServer string, os, arch, currentVersion string) error {
	// 1. 从 latestUrl 获取最新版本信息
	latestUrl := fmt.Sprintf("%s/api/versions/latest?os=%s&arch=%s", UpdateServer, os, arch)
	fmt.Printf("检查最新版本地址: %s\n", latestUrl)

	client := &http.Client{}
	resp, err := client.Get(latestUrl)
	if err != nil {
		return fmt.Errorf("无法访问最新版本地址: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("获取最新版本失败，HTTP 状态码: %d", resp.StatusCode)
	}

	// 解析最新版本信息
	var versionInfo VersionInfo
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取最新版本响应失败: %v", err)
	}
	err = json.Unmarshal(body, &versionInfo)
	if err != nil {
		return fmt.Errorf("解析最新版本信息失败: %v", err)
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
			return fmt.Errorf("默认版本号解析失败: %v", err)
		}
	}
	latestVer, err := semver.NewVersion(latestVersion)
	if err != nil {
		return fmt.Errorf("最新版本号格式错误: %v", err)
	}

	if !currentVer.LessThan(latestVer) {
		fmt.Println("当前已是最新版本，无需更新。")
		return nil
	}

	// 3. 如果当前版本号小于最新版本号，则执行更新
	updateURL := versionInfo.DownloadUrl
	fmt.Printf("发现新版本，下载地址: %s\n", updateURL)

	resp, err = client.Get(updateURL)
	if err != nil {
		return fmt.Errorf("无法访问更新地址: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载更新包失败，HTTP 状态码: %d", resp.StatusCode)
	}

	// 执行更新
	fmt.Println("开始更新...")
	err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	if err != nil {
		return fmt.Errorf("更新失败: %v", err)
	}

	// 4. 提示更新成功
	fmt.Println("更新成功！请重新启动应用程序。")
	return nil
}
