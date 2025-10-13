# Scripts 目录说明

此目录包含 Easy-Check 的开机自启安装和卸载脚本。

## 文件列表

### Windows

- `install.bat` - 开机自启安装脚本
- `uninstall.bat` - 开机自启卸载脚本

### Linux

- `install.sh` - 开机自启安装脚本
- `uninstall.sh` - 开机自启卸载脚本

## 使用方法

### Windows

**安装**

双击 `install.bat`，然后选择要安装的版本：

- 选项 1: UI 版本
- 选项 2: CMD 版本

**卸载**

双击 `uninstall.bat`，会自动删除所有已安装的开机自启项。

### Linux

**安装**

```bash
sudo ./install.sh
```

然后选择要安装的版本：

- 选项 1: UI 版本
- 选项 2: CMD 版本

**卸载**

```bash
sudo ./uninstall.sh
```

会自动删除所有已安装的开机自启服务。

## 注意事项

- Windows 脚本无需管理员权限
- Linux 脚本需要使用 sudo 运行
- 安装脚本会检测可执行文件是否存在
- 可以重复运行安装脚本来切换版本
