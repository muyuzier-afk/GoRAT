#!/bin/bash

echo "======================================="
echo "  GoRAT - Build Script"
echo "======================================="

echo "1. Build Windows version"
echo "2. Build Linux version"
echo "3. Build all versions"
echo "4. Exit"

echo -n "Select: "
read choice

case $choice in
    1)
        echo "Building Windows version..."
        GOOS=windows GOARCH=amd64 go build -ldflags "-X main.osType=Windows" -o gorat-client-windows.exe main.go
        echo "Windows build complete: gorat-client-windows.exe"
        ;;
    2)
        echo "Building Linux version..."
        GOOS=linux GOARCH=amd64 go build -ldflags "-X main.osType=Linux" -o gorat-client-linux main.go
        echo "Linux build complete: gorat-client-linux"
        ;;
    3)
        echo "Building Windows version..."
        GOOS=windows GOARCH=amd64 go build -ldflags "-X main.osType=Windows" -o gorat-client-windows.exe main.go
        echo "Windows build complete: gorat-client-windows.exe"

        echo "Building Linux version..."
        GOOS=linux GOARCH=amd64 go build -ldflags "-X main.osType=Linux" -o gorat-client-linux main.go
        echo "Linux build complete: gorat-client-linux"
        ;;
    4)
        echo "Exit"
        exit 0
        ;;
    *)
        echo "Invalid selection"
        ;;
esac

echo "======================================="
echo "  Build complete!"
echo "======================================="
