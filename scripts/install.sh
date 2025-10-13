#!/bin/bash

echo "========================================"
echo "Easy-Check 开机自启安装"
echo "========================================"
echo

select_version() {
    echo "请选择要安装的版本:"
    echo "1. UI 版本 (easy-check-ui-linux-amd64)"
    echo "2. CMD 版本 (easy-check-linux-amd64)"
    echo "3. 退出"
    echo
    read -p "请输入选项 (1-3): " choice
    
    case $choice in
        1)
            install_ui
            ;;
        2)
            install_cmd
            ;;
        3)
            echo "已取消安装"
            exit 0
            ;;
        *)
            echo "[错误] 无效的选项，请重新选择"
            echo
            select_version
            ;;
    esac
}

install_ui() {
    SERVICE_NAME="easy-check-ui"
    EXECUTABLE_NAME="easy-check-ui-linux-amd64"
    install_service
}

install_cmd() {
    SERVICE_NAME="easy-check"
    EXECUTABLE_NAME="easy-check-linux-amd64"
    install_service
}

install_service() {
    EXECUTABLE_PATH="$(cd "$(dirname "$0")/.." && pwd)/bin/$EXECUTABLE_NAME"
    SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME.service"
    
    echo
    echo "服务名称: $SERVICE_NAME"
    echo "可执行文件: $EXECUTABLE_PATH"
    echo "服务文件: $SERVICE_FILE"
    echo
    
    # 检查可执行文件是否存在
    if [ ! -f "$EXECUTABLE_PATH" ]; then
        echo "[错误] 找不到可执行文件: $EXECUTABLE_PATH"
        echo "请确保程序已正确编译"
        exit 1
    fi
    
    # 检查是否有 root 权限
    if [ "$EUID" -ne 0 ]; then
        echo "[错误] 请使用 sudo 运行此脚本"
        exit 1
    fi
    
    # 创建 systemd 服务文件
    echo "正在创建 systemd 服务..."
    WORKING_DIR="$(cd "$(dirname "$EXECUTABLE_PATH")/.." && pwd)"
    
    # 确定运行用户：优先使用 SUDO_USER，如果为空则使用当前用户
    RUN_USER="${SUDO_USER:-$USER}"
    if [ "$RUN_USER" = "root" ]; then
        echo "[警告] 将以 root 用户运行服务，建议使用普通用户"
    fi
    
    cat <<EOF >$SERVICE_FILE
[Unit]
Description=Easy-Check Network Monitor ($SERVICE_NAME)
After=network.target

[Service]
Type=simple
ExecStart=$EXECUTABLE_PATH
Restart=always
User=$RUN_USER
WorkingDirectory=$WORKING_DIR

[Install]
WantedBy=multi-user.target
EOF
    
    # 重新加载 systemd
    systemctl daemon-reload
    
    # 启用服务
    systemctl enable $SERVICE_NAME
    
    echo
    echo "[成功] $SERVICE_NAME 开机自启已设置成功！"
    echo
    echo "服务管理命令:"
    echo "  启动服务: sudo systemctl start $SERVICE_NAME"
    echo "  查看状态: sudo systemctl status $SERVICE_NAME"
    echo "  停止服务: sudo systemctl stop $SERVICE_NAME"
    echo
}

# 开始执行
select_version
