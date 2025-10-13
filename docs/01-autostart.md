# 开机自启

本项目支持 Windows 和 Linux 的开机自启功能，可通过脚本或应用内设置管理。

## 实现方式

### Windows

- 使用快捷方式放置到启动文件夹
- 无需管理员权限
- 兼容 Windows 7/8/10/11
- 不依赖 wmic 命令
- 启动文件夹位置: `%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup`

### Linux

- 使用 systemd 服务
- 需要 root 权限
- 支持服务管理
- 服务名称: easy-check 或 easy-check-ui

## 脚本安装

### Windows

双击 `scripts/install.bat`，选择要安装的版本（UI 或 CMD）

卸载: 双击 `scripts/uninstall.bat`

### Linux

```bash
sudo ./scripts/install.sh  # 选择要安装的版本
sudo ./scripts/uninstall.sh  # 卸载
```

## 应用内管理

在 Wails 应用的帮助菜单中，勾选"开机自启"可直接启用或禁用。

### Go 服务方法

- IsAutoStartEnabled() - 检查状态
- EnableAutoStart() - 启用
- DisableAutoStart() - 禁用
- GetAutoStartInfo() - 获取详情

## 手动管理

### Windows

按 Win + R，输入 `shell:startup`，删除快捷方式即可取消开机自启

### Linux

```bash
systemctl status easy-check-ui  # 查看状态
systemctl stop easy-check-ui    # 停止服务
systemctl disable easy-check-ui # 禁用开机自启
```

## 注意事项

- 确保程序路径中没有中文或特殊字符
- 某些杀毒软件可能会拦截自启动项，需添加信任
- Windows 快捷方式会设置正确的工作目录
- Linux 需要 root 权限才能操作 systemd 服务
