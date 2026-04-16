# GoRAT - 测试说明

## 系统架构

本系统包含以下组件：

1. **服务端**：[server/backend](file:///workspace/server/backend) - Go后端服务
2. **前端**：[server/frontend](file:///workspace/server/frontend) - Vue3管理界面
3. **客户端**：[client/main.go](file:///workspace/client/main.go) - 客户端程序（支持Windows和Linux）
4. **测试程序**：[test_client_server.go](file:///workspace/test/test_client_server.go) - 连接测试程序

## 快速开始

### 1. 启动服务端

#### 1.1 配置服务端

编辑 [server/backend/.env](file:///workspace/server/backend/.env) 文件：

```env
PORT=8000
S3_ENDPOINT=https://your-minio:9000
S3_ACCESS_KEY=your-key
S3_SECRET_KEY=your-secret
S3_REGION=us-east-1
S3_BUCKET=factory-telemetry
DATABASE_URL=postgres://postgres:postgres@localhost:5432/campus_management
```

#### 1.2 启动数据库（可选）

使用Docker启动PostgreSQL：

```bash
docker run --name postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=campus_management -p 5432:5432 -d postgres
```

#### 1.3 启动后端服务

```bash
cd server/backend
go mod tidy
go run main.go
```

后端服务将在 `http://localhost:8000` 启动。

### 2. 运行测试程序

#### 2.1 编译并运行测试

```bash
cd test
go run test_client_server.go
```

测试程序会依次执行以下测试：

- **健康检查**：验证服务端是否正常运行
- **客户端注册**：注册新客户端并获取ClientID和ClientKey
- **获取配置**：使用ClientID和ClientKey获取S3配置
- **心跳检测**：测试客户端心跳功能

#### 2.2 测试输出示例

```
=======================================
  GoRAT - 测试
=======================================

🔧 测试0: 健康检查
=======================
响应: {"status":"ok"}

✅ 健康检查通过!

🔧 测试1: 客户端注册
=======================
响应: {"message":"Client registered successfully","client_id":"...","client_key":"...","client":{...}}

✅ 注册成功!
   ClientID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
   ClientKey: yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy

🔧 测试2: 获取配置
=======================
响应: {"message":"Config retrieved successfully","config":{...}}

✅ 获取配置成功!

🔧 测试3: 心跳检测
=======================
响应: {"message":"Heartbeat received"}

✅ 心跳成功!

=======================================
  ✅ 所有测试通过!
=======================================
```

## 客户端工作流程

1. **首次运行**：
   - 客户端向服务端注册，获取ClientID和ClientKey
   - 将凭证保存到本地配置文件 `factoryeye_config.json`
   - 使用凭证从服务端获取S3配置

2. **后续运行**：
   - 从本地加载ClientID和ClientKey
   - 使用凭证从服务端获取S3配置
   - 开始采集数据并上传

## 主要API接口

### 客户端API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/client/register | 注册新客户端 |
| POST | /api/client/config | 获取S3配置 |
| POST | /api/client/heartbeat | 发送心跳 |
| POST | /api/client/upload/video | 上传视频 |
| POST | /api/client/upload/telemetry | 上传遥测数据 |
| POST | /api/client/upload/info | 上传设备信息 |

### 管理员API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/admin/clients | 获取所有客户端 |
| GET | /api/admin/client/:id | 获取客户端详情 |
| POST | /api/admin/client/:id/command | 发送命令 |
| GET | /api/admin/files | 获取所有文件 |
| GET | /api/admin/file/:id | 获取文件详情 |
| DELETE | /api/admin/file/:id | 删除文件 |
| POST | /api/admin/client/:id/power | 电源控制 |
| POST | /api/admin/client/:id/process | 进程控制 |

## 故障排除

### 测试失败：健康检查失败

**问题**：无法连接到服务端

**解决方案**：
1. 确认服务端已启动
2. 检查服务端端口是否被占用
3. 确认防火墙设置

### 测试失败：获取配置失败

**问题**：无效的客户端凭证

**解决方案**：
1. 确认客户端已成功注册
2. 检查配置文件是否正确保存
3. 尝试删除配置文件重新注册

### 数据库连接失败

**问题**：无法连接到PostgreSQL

**解决方案**：
1. 确认PostgreSQL已启动
2. 检查数据库连接字符串
3. 确认数据库和用户存在

## 后续步骤

1. 测试通过后，可以：
   - 启动前端界面：`cd server/frontend && npm install && npm run dev`
   - 编译客户端：`cd client && ./build.sh`
   - 在Windows或Linux设备上安装客户端

2. 访问管理界面：
   - 前端：http://localhost:3000
   - 后端API：http://localhost:8000
