<template>
  <div class="files-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>文件管理</span>
          <el-input
            v-model="searchQuery"
            placeholder="搜索文件"
            style="width: 300px"
            prefix-icon="el-icon-search"
          />
        </div>
      </template>
      <el-table :data="files" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80"></el-table-column>
        <el-table-column prop="filename" label="文件名"></el-table-column>
        <el-table-column prop="type" label="类型" width="100">
          <template #default="scope">
            <el-tag :type="getTypeTag(scope.row.type)">
              {{ scope.row.type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="size" label="大小" width="120">
          <template #default="scope">
            {{ formatSize(scope.row.size) }}
          </template>
        </el-table-column>
        <el-table-column prop="client.name" label="所属设备" width="150"></el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template #default="scope">
            <el-button type="primary" size="small" @click="downloadFile(scope.row)">下载</el-button>
            <el-button type="danger" size="small" @click="deleteFile(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script>
export default {
  name: 'Files',
  data() {
    return {
      files: [],
      searchQuery: ''
    }
  },
  mounted() {
    this.loadFiles()
  },
  methods: {
    loadFiles() {
      // 模拟数据，实际项目中应该从API获取
      this.files = [
        {
          id: 1,
          filename: '2026-04-16_14_30_00_001.mp4',
          type: 'video',
          size: 1024 * 1024 * 10, // 10MB
          client: { name: '实验室电脑1' },
          created_at: new Date().toISOString()
        },
        {
          id: 2,
          filename: '2026-04-16_14_30_00.json',
          type: 'telemetry',
          size: 1024 * 5, // 5KB
          client: { name: '实验室电脑1' },
          created_at: new Date().toISOString()
        },
        {
          id: 3,
          filename: 'factory-001.json',
          type: 'info',
          size: 1024 * 2, // 2KB
          client: { name: '实验室电脑1' },
          created_at: new Date().toISOString()
        }
      ]
    },
    formatTime(time) {
      return new Date(time).toLocaleString()
    },
    formatSize(size) {
      if (size < 1024) {
        return size + ' B'
      } else if (size < 1024 * 1024) {
        return (size / 1024).toFixed(2) + ' KB'
      } else {
        return (size / (1024 * 1024)).toFixed(2) + ' MB'
      }
    },
    getTypeTag(type) {
      switch (type) {
        case 'video': return 'warning'
        case 'telemetry': return 'info'
        case 'info': return 'success'
        default: return 'default'
      }
    },
    downloadFile(file) {
      // 下载文件
      console.log('Download file:', file)
    },
    deleteFile(file) {
      // 删除文件
      console.log('Delete file:', file)
    }
  }
}
</script>

<style scoped>
.files-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
