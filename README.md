# GoRAT

## 项目结构

```
GoRAT/
├── client/            # 客户端程序
│   ├── main.go        # 客户端主程序（支持Windows和Linux）
│   ├── go.mod         # Go依赖配置
│   ├── build.sh       # 编译脚本（支持跨平台编译）
│   ├── install.bat    # Windows安装脚本
│   └── install.sh     # Linux安装脚本
├── server/            # 服务端
│   ├── backend/       # Go后端服务
│   │   ├── controllers/  # API控制器
│   │   ├── models/       # 数据库模型
│   │   ├── utils/         # 工具函数
│   │   ├── .env           # 环境配置
│   │   ├── go.mod         # Go依赖配置
│   │   └── main.go         # 后端主程序
│   ├── frontend/      # Vue3前端
│   │   ├── public/         # 静态资源
│   │   ├── src/             # 源代码
│   │   ├── package.json     # npm依赖配置
│   │   └── vite.config.js   # Vite配置
│   └── start.sh       # 服务端启动脚本
├── test/              # 测试程序
│   ├── README.md          # 测试说明
│   ├── test_client_server.go     # 客户端-服务端连接测试
│   └── test_cross_platform.go     # 跨平台功能测试
└── README.md          # 项目说明（本文档）
```

## 系统架构

### 1. 客户端
- **支持平台**：Windows 10+、Linux
- **核心功能**：
  - 摄像头和麦克风数据采集（每5秒一个视频片段）
  - 系统信息采集（CPU、内存、进程列表）
  - 自动注册和配置获取
  - 数据上传到服务端
  - 开机自启和后台运行

### 2. 服务端
- **后端**：Go 1.20 + Gin + PostgreSQL
- **前端**：Vue 3 + Element Plus + ECharts
- **核心功能**：
  - 客户端管理（注册、心跳、状态监控）
  - 数据存储和转发（上传到S3）
  - 远程控制（进程管理、开关机）
  - 实时监控和数据分析

### 3. 数据流向
1. 客户端采集数据
2. 上传到服务端
3. 服务端存储并转发到S3
4. 管理员通过Web界面查看

## 快速开始

### 1. 启动服务端

#### 1.1 配置服务端
编辑 `server/backend/.env` 文件：

```env
PORT=8000
S3_ENDPOINT=https://your-minio:9000
S3_ACCESS_KEY=your-key
S3_SECRET_KEY=your-secret
S3_REGION=us-east-1
S3_BUCKET=factory-telemetry
DATABASE_URL=postgres://postgres:postgres@localhost:5432/campus_management
```

#### 1.2 启动服务

```bash
cd server
./start.sh
```

服务端将在以下地址运行：
- 后端API：http://localhost:8000
- 前端界面：http://localhost:3000

### 2. 编译和安装客户端

#### 2.1 编译客户端

```bash
cd client
chmod +x build.sh
./build.sh
```

编译选项：
1. 编译Windows版本
2. 编译Linux版本
3. 编译所有版本

#### 2.2 安装客户端

**Windows**：
```bash
# 以管理员身份运行
install.bat
```

**Linux**：
```bash
chmod +x install.sh
sudo ./install.sh
```

### 3. 测试系统

```bash
cd test
go run test_cross_platform.go
```

测试程序会验证：
- 服务端健康状态
- Windows客户端注册和配置获取
- Linux客户端注册和配置获取
- 客户端心跳功能

## 核心功能

### 1. 客户端功能
- **视频采集**：1080P30视频，每5秒一个片段
- **系统监控**：CPU、内存、进程列表（每30秒）
- **自动注册**：首次运行自动注册并获取凭证
- **配置管理**：从服务端获取S3配置
- **跨平台支持**：Windows和Linux

### 2. 服务端功能
- **客户端管理**：注册、心跳、状态监控
- **数据管理**：视频、遥测数据存储和转发
- **远程控制**：进程管理、开关机
- **Web界面**：实时监控和数据分析
- **S3集成**：数据备份到S3兼容存储

### 3. 管理界面
- **客户端列表**：显示所有客户端状态和操作系统类型
- **文件管理**：查看和下载客户端上传的文件
- **系统监控**：实时监控客户端CPU、内存和进程
- **设置管理**：配置S3和数据库连接

## 技术栈

- **后端**：Go 1.20, Gin, PostgreSQL, AWS SDK
- **前端**：Vue 3, Element Plus, ECharts, Vue Router
- **客户端**：Go 1.20, FFmpeg
- **存储**：S3兼容存储（MinIO/AWS/OSS）

## 故障排除

### 1. 客户端注册失败
- 检查网络连接
- 确认服务端是否运行
- 查看客户端日志

### 2. 视频采集失败
- 检查摄像头是否可用
- 确认FFmpeg已安装
- 检查设备权限

### 3. 服务端启动失败
- 检查数据库连接
- 确认端口未被占用
- 查看服务端日志

## 安全注意事项
- 客户端凭证存储在本地配置文件中
- 服务端API需要认证
- 建议在生产环境中使用HTTPS
- 定期更新客户端和服务端

## 后续扩展
- 支持更多操作系统（macOS）
- 添加更多传感器数据采集
- 实现AI分析功能
- 支持集群部署

## 联系方式

如有问题或建议，请联系项目维护人员。
