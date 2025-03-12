# easy-check

easy-check 是一个用于定期检测网络连接的工具。它可以在 Linux 和 Windows 系统上运行，开机时自动启动。该项目的主要功能是通过 ping 指定的主机列表，记录检测结果并将日志写入文件。

## 功能

- 定期 ping 检测指定主机
- 将检测结果和时间戳记录到日志文件
- 支持 Linux 和 Windows 系统的开机自启动

## 项目结构

```
easy-check
├── cmd
│   └── main.go          # 应用程序入口点
├── internal
│   ├── checker
│   │   └── checker.go   # 网络检测逻辑
│   │   └── config.go    # 配置加载逻辑
│   │   └── pinger.go    # Ping 接口定义
│   │   └── pinger_linux.go # Linux 平台的 Ping 实现
│   │   └── pinger_windows.go # Windows 平台的 Ping 实现
│   └── logger
│       └── logger.go    # 日志记录功能
├── configs
│   └── config.yaml      # 配置文件
├── scripts
│   ├── install.sh       # 安装脚本
│   └── uninstall.sh     # 卸载脚本
├── go.mod               # Go模块配置
├── go.sum               # Go模块依赖
├── Makefile             # Makefile 文件
├── .gitignore           # Git 忽略文件
└── README.md            # 项目文档
```

## 使用方法

1. **配置**: 编辑 `configs/config.yaml` 文件，添加要检测的主机列表和检测间隔时间。
2. **安装**: 使用管理员用户运行 `scripts/install.sh` 脚本以设置开机自启动。
3. **启动**: 运行 `bin/easy-check` 启动应用程序。

## 配置说明

`config.yaml` 文件包含以下字段：

- `hosts`: 要 ping 的主机列表
- `ping`:
  - `count`: ping 的次数
  - `timeout`: ping 的超时时间，单位为秒
- `interval`: 检测间隔时间（单位：秒）
- `log`:
  - `file`: 日志文件路径

## 日志

检测结果将记录在指定的日志文件中，包含每次检测的时间戳和结果。

## 贡献

欢迎任何形式的贡献！请提交问题或拉取请求。
