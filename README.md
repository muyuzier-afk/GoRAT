# GoRAT

## Tech Stack

- **Client**: Go 1.20+, FFmpeg, gopsutil
- **Server Backend**: Go 1.20+, Gin framework, PostgreSQL, JWT authentication
- **Server Frontend**: Vue 3, Element Plus, ECharts
- **Storage**: S3-compatible storage (MinIO, AWS S3, Aliyun OSS, etc.)

## Quick Start

### Prerequisites

- Go 1.20+
- PostgreSQL database
- S3-compatible storage service (optional)
- FFmpeg (for client video recording)
- Node.js 16+ (for frontend development)

### 1. Deploy Server

#### 1.1 Configure Environment Variables

Create a `.env` file in `server/backend/` (see `.env.example` for reference):

```env
# Server port
PORT=8000

# Database connection string
DATABASE_URL=postgres://user:password@localhost:5432/gorat

# S3 storage configuration
S3_ENDPOINT=https://your-s3-endpoint:9000
S3_ACCESS_KEY=your-access-key
S3_SECRET_KEY=your-secret-key
S3_REGION=us-east-1
S3_BUCKET=gorat-data

# JWT secret (change in production!)
JWT_SECRET=your-jwt-secret-key

# Admin credentials (change in production!)
ADMIN_USERNAME=admin
ADMIN_PASSWORD=changeme

# CORS allowed origins (comma-separated)
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

#### 1.2 Start Backend Server

```bash
cd server/backend
go mod download
go run main.go
```

#### 1.3 Start Frontend Server

```bash
cd server/frontend
npm install
npm run dev
```

Server access addresses:

- Backend API: http://localhost:8000
- Frontend UI: http://localhost:3000

### 2. Build Client

#### 2.1 Cross-Platform Build

Use the provided build script:

```bash
cd client
chmod +x build.sh
./build.sh
```

Build options:

1. Build Windows version
2. Build Linux version
3. Build all versions

#### 2.2 Manual Build

Windows version:

```bash
GOOS=windows GOARCH=amd64 go build -o gorat-client.exe main.go
```

Linux version:

```bash
GOOS=linux GOARCH=amd64 go build -o gorat-client main.go
```

### 3. Install Client

#### Windows

1. Run Command Prompt as Administrator
2. Execute the install script:

```cmd
install.bat
```

#### Linux

1. Grant execute permission and run the install script:

```bash
chmod +x install.sh
sudo ./install.sh
```

## Usage

### Client Configuration

On first run, specify the server address:

```bash
# Windows
gorat-client.exe --server http://your-server:8000

# Linux
./gorat-client --server http://your-server:8000
```

Optional flags:

- `--server`: Server endpoint URL (required on first run)
- `--device`: Custom device ID (auto-generated if not specified)

The client will automatically:

1. Register with the server
2. Receive a Client ID and Client Key
3. Fetch upload configuration (presigned S3 URLs)
4. Begin collecting and uploading system telemetry and video data

### Server Management

1. Access the web management UI: http://your-server:3000
2. Login with admin credentials (default: admin / changeme)
3. View all registered clients in the Clients page
4. View uploaded videos and system information in the Files page
5. Monitor client CPU/memory in the Telemetry page
6. Send remote commands (shell, power control) to clients

## API Endpoints

### Authentication

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/admin/login` | Admin login, returns JWT token |

### Client API (no auth required)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/client/register` | Register a new client |
| POST | `/api/client/config` | Get client config (presigned upload URLs) |
| POST | `/api/client/heartbeat` | Send heartbeat |
| POST | `/api/client/upload/video` | Upload video file |
| POST | `/api/client/upload/telemetry` | Upload telemetry data |
| POST | `/api/client/upload/info` | Upload device info |

### Admin API (JWT auth required)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/admin/stats` | Dashboard statistics |
| GET | `/api/admin/clients` | List all clients |
| GET | `/api/admin/client/:id` | Get client detail |
| POST | `/api/admin/client/:id/command` | Send command to client |
| POST | `/api/admin/client/:id/power` | Send power command |
| POST | `/api/admin/client/:id/process` | Send process command |
| GET | `/api/admin/files` | List all files |
| GET | `/api/admin/file/:id` | Get file detail |
| DELETE | `/api/admin/file/:id` | Delete file |

### Other

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/api/ws/:clientId` | WebSocket connection |

## Development

### Local Development

1. Clone the repository:

```bash
git clone https://github.com/muyuzier-afk/GoRAT.git
cd GoRAT
```

2. Start dependencies:

```bash
# Start PostgreSQL (using Docker)
docker run -d --name gorat-postgres -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_DB=gorat postgres

# Start MinIO (optional, for S3 storage)
docker run -d --name gorat-minio -p 9000:9000 -p 9001:9001 minio/minio server /data --console-address ":9001"
```

3. Start backend and frontend servers (see Quick Start)

### Run Tests

```bash
cd test
go run test_client_server.go
```

Test coverage includes:

- Health check
- Admin login (valid and invalid credentials)
- Authentication enforcement (unauthorized access rejected)
- Client registration flow
- Client config retrieval
- Heartbeat
- Telemetry upload
- Video upload
- Device info upload
- Admin API endpoints (clients, files, stats)

## Security Notes

- Change default admin credentials (`ADMIN_USERNAME`, `ADMIN_PASSWORD`) in production
- Set a strong `JWT_SECRET` in production
- Configure `CORS_ALLOWED_ORIGINS` to restrict allowed frontend origins
- S3 credentials are never exposed to clients; presigned URLs are used for uploads
- WebSocket connections validate the Origin header against CORS allowed origins
- All admin API endpoints require JWT Bearer token authentication

## License

This project uses MIT License.

## Contributing

Issues and Pull Requests are welcome!

## Contact

For questions or suggestions, please reach out via GitHub Issues.

---

Note: Please use this software legally and only for authorized system administration and security research purposes.
