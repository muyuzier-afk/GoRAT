#!/bin/bash

# 启动后端服务
echo "Starting backend server..."
cd backend && go run main.go &

# 等待后端服务启动
sleep 5

# 启动前端服务
echo "Starting frontend server..."
cd ../frontend && npm install && npm run dev
