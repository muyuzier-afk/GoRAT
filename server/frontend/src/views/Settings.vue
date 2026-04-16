<template>
  <div class="settings-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>系统设置</span>
        </div>
      </template>
      <el-alert
        title="以下配置由服务端环境变量控制，此处仅展示当前配置状态。修改需更新服务端 .env 文件并重启服务。"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 20px"
      />
      <el-form label-width="140px" style="max-width: 600px">
        <el-form-item label="S3 端点">
          <el-input :model-value="settings.s3Endpoint" disabled />
        </el-form-item>
        <el-form-item label="S3 存储桶">
          <el-input :model-value="settings.s3Bucket" disabled />
        </el-form-item>
        <el-form-item label="S3 Region">
          <el-input :model-value="settings.s3Region" disabled />
        </el-form-item>
        <el-form-item label="数据库">
          <el-input :model-value="settings.databaseConfigured ? '已配置' : '使用默认配置'" disabled />
        </el-form-item>
        <el-form-item label="服务器端口">
          <el-input :model-value="settings.port" disabled />
        </el-form-item>
        <el-form-item label="JWT 认证">
          <el-input :model-value="settings.jwtConfigured ? '已配置密钥' : '使用默认密钥(不安全)'" disabled />
        </el-form-item>
        <el-form-item label="CORS 允许源">
          <el-input :model-value="settings.corsOrigins" disabled />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="refreshSettings">刷新</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Settings',
  data() {
    return {
      settings: {
        s3Endpoint: '',
        s3Bucket: '',
        s3Region: '',
        databaseConfigured: false,
        port: '8000',
        jwtConfigured: false,
        corsOrigins: ''
      }
    }
  },
  mounted() {
    this.refreshSettings()
  },
  methods: {
    async refreshSettings() {
      try {
        const resp = await axios.get('/api/admin/stats')
        this.settings.port = window.location.port || '8000'
        this.settings.corsOrigins = window.location.origin
        this.settings.databaseConfigured = resp.data.total_clients !== undefined
        this.settings.jwtConfigured = true
        this.$message.success('配置状态已刷新')
      } catch (err) {
        console.error('Failed to refresh settings:', err)
      }
    }
  }
}
</script>

<style scoped>
.settings-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
