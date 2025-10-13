# Machine ID 模块

这个模块用于获取机器的唯一标识符,支持多个操作系统,并提供了可靠的备用方案。

## 功能特性

### 1. 多平台支持

- **Linux**: 从 `/etc/machine-id` 或 `/var/lib/dbus/machine-id` 读取
- **macOS**: 使用 `ioreg` 或 `system_profiler` 获取硬件 UUID
- **Windows**: 支持 Windows 10 和 Windows 11
  - Windows 10: 使用 `wmic` 命令
  - Windows 11: 使用 PowerShell 的 `Get-CimInstance` 命令(兼容没有 wmic 的系统)

### 2. 备用方案

当系统原生方法获取机器 ID 失败时,会自动使用备用方案:

1. 首次运行时,生成一个随机的 64 字符十六进制 ID
2. 将 ID 保存到用户配置目录: `~/.easy-check/machine-id`
3. 后续运行时,直接从文件读取,确保 ID 的一致性

### 3. 安全性

所有获取到的机器 ID 都会经过 SHA256 哈希处理,确保:

- 输出长度一致(64 字符)
- 不会暴露原始硬件信息
- 提供足够的唯一性

## 使用方法

```go
import "easy-check/internal/machineid"

// 获取机器 ID
id, err := machineid.GetMachineID()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Machine ID: %s\n", id)
```

## 工作流程

```
1. 尝试系统原生方法获取 ID
   └─ 成功 → 返回哈希后的 ID
   └─ 失败 ↓

2. 尝试从配置文件读取
   └─ 文件存在 → 返回哈希后的 ID
   └─ 文件不存在 ↓

3. 生成新的随机 ID
   └─ 保存到配置文件
   └─ 返回哈希后的 ID
```

## Windows 11 兼容性

Windows 11 已经移除或弃用了 `wmic` 命令,本模块通过以下方式兼容:

1. 首先尝试使用 `wmic` (兼容 Windows 10)
2. 如果失败,使用 PowerShell 的 `Get-CimInstance` (Windows 11 推荐方式)
3. 如果都失败,使用备用的文件存储方案

## 配置文件位置

备用 machine-id 文件存储位置:

- Linux/macOS: `~/.easy-check/machine-id`
- Windows: `%USERPROFILE%\.easy-check\machine-id`

如果无法获取用户目录,会使用系统临时目录:

- Linux/macOS: `/tmp/.easy-check/machine-id`
- Windows: `%TEMP%\.easy-check\machine-id`

## 测试

运行测试:

```bash
go test -v ./internal/machineid/
```

## 注意事项

1. 生成的 machine-id 文件应该被保留,删除后会重新生成一个新的 ID
2. 所有 ID 经过 SHA256 哈希,输出长度固定为 64 个十六进制字符
3. 备用方案生成的 ID 是随机的,但会持久化存储,确保同一台机器的 ID 不会改变
