package update

import (
	"fmt"
	"net/http"

	"github.com/minio/selfupdate"
)

// CheckAndUpdate 检查更新并执行更新
func CheckAndUpdate(updateURL string, currentVersion string) error {
	fmt.Printf("当前版本: %s\n", currentVersion)
	fmt.Printf("检查更新地址: %s\n", updateURL)

	// 创建 HTTP 客户端
	client := &http.Client{}

	// 检查更新
	resp, err := client.Get(updateURL)
	if err != nil {
		return fmt.Errorf("无法访问更新地址: %v", err)
	}
	defer resp.Body.Close()

	// 执行更新
	if resp.StatusCode == http.StatusOK {
		fmt.Println("发现新版本，开始更新...")
		err = selfupdate.Apply(resp.Body, selfupdate.Options{})
		if err != nil {
			return fmt.Errorf("更新失败: %v", err)
		}
		fmt.Println("更新成功！请重新启动应用程序。")
		return nil
	}

	fmt.Println("未发现新版本。")
	return nil
}
