#!/bin/bash

echo "========================================"
echo "Easy-Check 开机自启卸载"
echo "========================================"
echo

UI_SERVICE="easy-check-ui"
CMD_SERVICE="easy-check"

# 检查是否有 root 权限
if [ "$EUID" -ne 0 ]; then
    echo "[错误] 请使用 sudo 运行此脚本"
    exit 1
fi

FOUND=0

# 检查并停止 UI 服务
if systemctl is-enabled $UI_SERVICE &>/dev/null || systemctl is-active $UI_SERVICE &>/dev/null; then
    echo "正在停止并禁用 UI 版本服务..."
    systemctl stop $UI_SERVICE 2>/dev/null
    systemctl disable $UI_SERVICE 2>/dev/null
    rm -f /etc/systemd/system/$UI_SERVICE.service
    echo "[成功] UI 版本服务已删除"
    FOUND=1
fi

# 检查并停止 CMD 服务
if systemctl is-enabled $CMD_SERVICE &>/dev/null || systemctl is-active $CMD_SERVICE &>/dev/null; then
    echo "正在停止并禁用 CMD 版本服务..."
    systemctl stop $CMD_SERVICE 2>/dev/null
    systemctl disable $CMD_SERVICE 2>/dev/null
    rm -f /etc/systemd/system/$CMD_SERVICE.service
    echo "[成功] CMD 版本服务已删除"
    FOUND=1
fi

if [ $FOUND -eq 1 ]; then
    systemctl daemon-reload
    echo
    echo "[完成] 开机自启已取消"
else
    echo "[提示] 未找到任何开机自启服务"
fi

echo
