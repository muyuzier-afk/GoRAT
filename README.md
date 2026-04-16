# GoRAT

## 技术栈

- **客户端**：Go 1.20+、FFmpeg
- **服务端后端**：Go 1.20+、Gin框架、PostgreSQL
- **服务端前端**：Vue 3、Element Plus、ECharts
- **存储**：S3兼容存储（MinIO、AWS S3、阿里云OSS等）

## 快速开始

### 前置要求

- Go 1.20或更高版本
- PostgreSQL数据库
- S3兼容存储服务（可选）
- FFmpeg（客户端需要）

### 1. 部署服务端

#### 1.1 配置环境变量

编辑 `server/backend/.env` 文件：

```env
# 服务端口
PORT=8000

# 数据库配置
DATABASE_URL=postgres://user:password@localhost:5432/gorat

# S3存储配置
S3_ENDPOINT=https://your-s3-endpoint:9000
S3_ACCESS_KEY=your-access-key
S3_SECRET_KEY=your-secret-key
S3_REGION=us-east-1
S3_BUCKET=gorat-data

# JWT密钥
JWT_SECRET=your-jwt-secret-key
```

1.2 启动后端服务

```bash
cd server/backend
go mod download
go run main.go
```

1.3 启动前端服务

```bash
cd server/frontend
npm install
npm run dev
```

服务端访问地址：

· 后端API：http://localhost:8000
· 前端界面：http://localhost:3000

2. 编译客户端

2.1 跨平台编译

使用提供的编译脚本：

```bash
cd client
chmod +x build.sh
./build.sh
```

编译选项：

1. 编译Windows版本
2. 编译Linux版本
3. 编译所有版本

2.2 手动编译

Windows版本：

```bash
GOOS=windows GOARCH=amd64 go build -o gorat-client.exe main.go
```

Linux版本：

```bash
GOOS=linux GOARCH=amd64 go build -o gorat-client main.go
```

3. 安装客户端

Windows系统

1. 以管理员身份运行命令提示符
2. 执行安装脚本：

```cmd
install.bat
```

Linux系统

1. 赋予执行权限并运行安装脚本：

```bash
chmod +x install.sh
sudo ./install.sh
```

使用说明

客户端配置

首次运行客户端时，需要配置服务端地址：

```bash
# Windows
gorat-client.exe --server http://your-server:8000

# Linux
./gorat-client --server http://your-server:8000
```

客户端会自动：

1. 向服务端注册
2. 获取Client ID和Client Key
3. 下载S3配置
4. 开始采集和上传数据

服务端管理

1. 访问Web管理界面：http://your-server:3000
2. 使用默认管理员账户登录（首次登录请修改密码）
3. 在客户端列表中查看所有已注册的客户端
4. 查看客户端上传的视频和系统信息
5. 执行远程控制操作

开发指南

本地开发

1. 克隆仓库：

```bash
git clone https://github.com/muyuzier-afk/GoRAT.git
cd GoRAT
```

1. 启动开发环境：

```bash
# 启动数据库（使用Docker）
docker run -d --name gorat-postgres -p 5432:5432 -e POSTGRES_PASSWORD=password postgres

# 启动MinIO（可选，用于S3存储）
docker run -d --name gorat-minio -p 9000:9000 -p 9001:9001 minio/minio server /data --console-address ":9001"
```

运行测试

```bash
cd test
go run test_cross_platform.go
```

测试内容包括：

· 服务端健康检查
· 客户端注册流程
· 跨平台功能验证
· 心跳检测

故障排除


许可证

本项目使用MIT Lincens.

贡献

欢迎提交Issue和Pull Request！

联系方式

如有问题或建议，请通过GitHub Issues联系。

---

注意：请合法使用本软件，仅用于授权的系统管理和安全研究目的。

