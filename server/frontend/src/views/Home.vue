<template>
  <div class="home-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>系统概览</span>
          <el-button size="small" @click="loadData">刷新</el-button>
        </div>
      </template>
      <div class="overview-stats">
        <el-row :gutter="20">
          <el-col :span="6">
            <el-card class="stat-card">
              <div class="stat-content">
                <div class="stat-number">{{ stats.totalClients }}</div>
                <div class="stat-label">总客户端</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card">
              <div class="stat-content">
                <div class="stat-number">{{ stats.onlineClients }}</div>
                <div class="stat-label">在线客户端</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card">
              <div class="stat-content">
                <div class="stat-number">{{ stats.totalFiles }}</div>
                <div class="stat-label">总文件数</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card">
              <div class="stat-content">
                <div class="stat-number">{{ stats.onlineClients }} / {{ stats.totalClients }}</div>
                <div class="stat-label">在线率</div>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>
      <div class="recent-section">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-card>
              <template #header><span>最近上线客户端</span></template>
              <el-table :data="recentClients" style="width: 100%" size="small">
                <el-table-column prop="name" label="名称"></el-table-column>
                <el-table-column prop="device_id" label="设备ID" width="160"></el-table-column>
                <el-table-column prop="status" label="状态" width="80">
                  <template #default="scope">
                    <el-tag :type="scope.row.status === 'online' ? 'success' : 'danger'" size="small">
                      {{ scope.row.status === 'online' ? '在线' : '离线' }}
                    </el-tag>
                  </template>
                </el-table-column>
              </el-table>
            </el-card>
          </el-col>
          <el-col :span="12">
            <el-card>
              <template #header><span>最近上传文件</span></template>
              <el-table :data="recentFiles" style="width: 100%" size="small">
                <el-table-column prop="filename" label="文件名"></el-table-column>
                <el-table-column prop="type" label="类型" width="80">
                  <template #default="scope">
                    <el-tag size="small">{{ scope.row.type }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="大小" width="100">
                  <template #default="scope">{{ formatSize(scope.row.size) }}</template>
                </el-table-column>
              </el-table>
            </el-card>
          </el-col>
        </el-row>
      </div>
    </el-card>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Home',
  data() {
    return {
      stats: { totalClients: 0, onlineClients: 0, totalFiles: 0 },
      recentClients: [],
      recentFiles: []
    }
  },
  mounted() {
    this.loadData()
  },
  methods: {
    async loadData() {
      try {
        const [statsResp, clientsResp, filesResp] = await Promise.all([
          axios.get('/api/admin/stats'),
          axios.get('/api/admin/clients'),
          axios.get('/api/admin/files')
        ])
        this.stats = {
          totalClients: statsResp.data.total_clients || 0,
          onlineClients: statsResp.data.online_clients || 0,
          totalFiles: statsResp.data.total_files || 0
        }
        const clients = clientsResp.data.clients || []
        this.recentClients = clients.slice(-5).reverse()
        const files = filesResp.data.files || []
        this.recentFiles = files.slice(-5).reverse()
      } catch (err) {
        console.error('Failed to load dashboard data:', err)
      }
    },
    formatSize(size) {
      if (!size) return '0 B'
      if (size < 1024) return size + ' B'
      if (size < 1024 * 1024) return (size / 1024).toFixed(2) + ' KB'
      return (size / (1024 * 1024)).toFixed(2) + ' MB'
    }
  }
}
</script>

<style scoped>
.home-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.overview-stats {
  margin-bottom: 20px;
}

.stat-card {
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-content {
  text-align: center;
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  color: #606266;
}

.recent-section {
  margin-top: 20px;
}
</style>
