#!/bin/bash

echo "======================================="
echo "  GoRAT - 编译脚本"
echo "======================================="

echo "1. 编译Windows版本"
echo "2. 编译Linux版本"
echo "3. 编译所有版本"
echo "4. 退出"

echo -n "请选择: "
read choice

case $choice in
    1)
        echo "编译Windows版本..."
        GOOS=windows GOARCH=amd64 go build -ldflags "-X main.osType=Windows" -o factoryeye-windows.exe main.go
        echo "Windows版本编译完成: factoryeye-windows.exe"
        ;;
    2)
        echo "编译Linux版本..."
        GOOS=linux GOARCH=amd64 go build -ldflags "-X main.osType=Linux" -o factoryeye-linux main.go
        echo "Linux版本编译完成: factoryeye-linux"
        ;;
    3)
        echo "编译Windows版本..."
        GOOS=windows GOARCH=amd64 go build -ldflags "-X main.osType=Windows" -o factoryeye-windows.exe main.go
        echo "Windows版本编译完成: factoryeye-windows.exe"
        
        echo "编译Linux版本..."
        GOOS=linux GOARCH=amd64 go build -ldflags "-X main.osType=Linux" -o factoryeye-linux main.go
        echo "Linux版本编译完成: factoryeye-linux"
        ;;
    4)
        echo "退出"
        exit 0
        ;;
    *)
        echo "无效选择"
        ;;
esac

echo "======================================="
echo "  编译完成!"
echo "======================================="
