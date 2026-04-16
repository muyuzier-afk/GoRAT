#!/bin/bash

echo "======================================="
echo "  GoRAT - Linux Install Script"
echo "======================================="

if ! command -v ffmpeg &> /dev/null; then
    echo "Error: ffmpeg not found, please install ffmpeg first"
    echo "Ubuntu/Debian: sudo apt install ffmpeg"
    echo "CentOS/RHEL: sudo yum install ffmpeg"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo "Error: go not found, please install go 1.20 or later"
    exit 1
fi

INSTALL_DIR="/opt/gorat-client"
SERVICE_FILE="/etc/systemd/system/gorat-client.service"

echo "Creating install directory..."
sudo mkdir -p $INSTALL_DIR

echo "Copying executable..."
sudo cp gorat-client-linux $INSTALL_DIR/
sudo chmod +x $INSTALL_DIR/gorat-client-linux

echo "Creating systemd service file..."
sudo cat > $SERVICE_FILE << EOF
[Unit]
Description=GoRAT Client
After=network.target

[Service]
Type=simple
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/gorat-client-linux
Restart=always
RestartSec=10
User=root

[Install]
WantedBy=multi-user.target
EOF

echo "Reloading systemd..."
sudo systemctl daemon-reload

echo "Enabling service..."
sudo systemctl enable gorat-client.service

echo "Starting service..."
sudo systemctl start gorat-client.service

echo "======================================="
echo "  Installation complete!"
echo "  Service started and enabled on boot"
echo "  Install path: $INSTALL_DIR"
echo "  Service name: gorat-client.service"
echo "  Check status: systemctl status gorat-client.service"
echo "  Stop service: systemctl stop gorat-client.service"
echo "  Start service: systemctl start gorat-client.service"
echo "======================================="
