<template>
  <div class="settings-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>系统设置</span>
        </div>
      </template>
      <el-form :model="settings" label-width="120px">
        <el-form-item label="S3 端点">
          <el-input v-model="settings.s3Endpoint" placeholder="S3兼容存储地址" />
        </el-form-item>
        <el-form-item label="S3 Access Key">
          <el-input v-model="settings.s3AccessKey" placeholder="Access Key" />
        </el-form-item>
        <el-form-item label="S3 Secret Key">
          <el-input v-model="settings.s3SecretKey" type="password" placeholder="Secret Key" />
        </el-form-item>
        <el-form-item label="S3 存储桶">
          <el-input v-model="settings.s3Bucket" placeholder="存储桶名称" />
        </el-form-item>
        <el-form-item label="数据库连接">
          <el-input v-model="settings.databaseUrl" placeholder="数据库连接字符串" />
        </el-form-item>
        <el-form-item label="服务器端口">
          <el-input v-model="settings.port" placeholder="服务器端口" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveSettings">保存设置</el-button>
          <el-button @click="resetSettings">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
export default {
  name: 'Settings',
  data() {
    return {
      settings: {
        s3Endpoint: 'https://your-minio:9000',
        s3AccessKey: 'your-key',
        s3SecretKey: 'your-secret',
        s3Bucket: 'factory-telemetry',
        databaseUrl: 'postgres://postgres:postgres@localhost:5432/campus_management',
        port: '8000'
      }
    }
  },
  methods: {
    saveSettings() {
      // 保存设置
      console.log('Save settings:', this.settings)
      this.$message.success('设置保存成功')
    },
    resetSettings() {
      // 重置设置
      this.settings = {
        s3Endpoint: 'https://your-minio:9000',
        s3AccessKey: 'your-key',
        s3SecretKey: 'your-secret',
        s3Bucket: 'factory-telemetry',
        databaseUrl: 'postgres://postgres:postgres@localhost:5432/campus_management',
        port: '8000'
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

.el-form {
  max-width: 600px;
}
</style>
