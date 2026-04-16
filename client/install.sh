#!/bin/bash

echo "======================================="
echo "  GoRAT - Linux安装脚本"
echo "======================================="

# 检查ffmpeg是否安装
if ! command -v ffmpeg &> /dev/null; then
    echo "错误: 未找到ffmpeg，请先安装ffmpeg"
    echo "Ubuntu/Debian: sudo apt install ffmpeg"
    echo "CentOS/RHEL: sudo yum install ffmpeg"
    exit 1
fi

# 检查go是否安装
if ! command -v go &> /dev/null; then
    echo "错误: 未找到go，请先安装go 1.20或更高版本"
    exit 1
fi

# 安装目录
INSTALL_DIR="/opt/factoryeye"
SERVICE_FILE="/etc/systemd/system/factoryeye.service"

# 创建安装目录
echo "创建安装目录..."
sudo mkdir -p $INSTALL_DIR

# 复制可执行文件
echo "复制可执行文件..."
sudo cp factoryeye-linux $INSTALL_DIR/
sudo chmod +x $INSTALL_DIR/factoryeye-linux

# 创建systemd服务文件
echo "创建systemd服务文件..."
sudo cat > $SERVICE_FILE << EOF
[Unit]
Description=FactoryEye Client
After=network.target

[Service]
Type=simple
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/factoryeye-linux
Restart=always
RestartSec=10
User=root

[Install]
WantedBy=multi-user.target
EOF

# 重新加载systemd
echo "重新加载systemd..."
sudo systemctl daemon-reload

# 启用服务
echo "启用服务..."
sudo systemctl enable factoryeye.service

# 启动服务
echo "启动服务..."
sudo systemctl start factoryeye.service

echo "======================================="
echo "  安装完成!"
echo "  服务已启动并设置为开机自启"
echo "  安装路径: $INSTALL_DIR"
echo "  服务名称: factoryeye.service"
echo "  查看状态: systemctl status factoryeye.service"
echo "  停止服务: systemctl stop factoryeye.service"
echo "  启动服务: systemctl start factoryeye.service"
echo "======================================="
