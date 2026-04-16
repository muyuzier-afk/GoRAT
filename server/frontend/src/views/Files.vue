<template>
  <div class="files-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>文件管理</span>
          <div>
            <el-input
              v-model="searchQuery"
              placeholder="搜索文件"
              style="width: 200px; margin-right: 10px"
              prefix-icon="el-icon-search"
              clearable
            />
            <el-button size="small" @click="loadFiles">刷新</el-button>
          </div>
        </div>
      </template>
      <el-table :data="filteredFiles" style="width: 100%" v-loading="loading">
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
        <el-table-column label="所属设备" width="150">
          <template #default="scope">
            {{ scope.row.client ? scope.row.client.name || scope.row.client.device_id : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="scope">
            <el-button type="danger" size="small" @click="deleteFile(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Files',
  data() {
    return {
      files: [],
      searchQuery: '',
      loading: false
    }
  },
  computed: {
    filteredFiles() {
      if (!this.searchQuery) return this.files
      const q = this.searchQuery.toLowerCase()
      return this.files.filter(f =>
        (f.filename && f.filename.toLowerCase().includes(q)) ||
        (f.type && f.type.toLowerCase().includes(q))
      )
    }
  },
  mounted() {
    this.loadFiles()
  },
  methods: {
    async loadFiles() {
      this.loading = true
      try {
        const resp = await axios.get('/api/admin/files')
        this.files = resp.data.files || []
      } catch (err) {
        this.$message.error('加载文件列表失败')
        console.error(err)
      } finally {
        this.loading = false
      }
    },
    async deleteFile(file) {
      try {
        await this.$confirm(`确认删除文件 "${file.filename}"？`, '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        await axios.delete(`/api/admin/file/${file.id}`)
        this.$message.success('删除成功')
        this.loadFiles()
      } catch (err) {
        if (err !== 'cancel') {
          this.$message.error('删除失败')
        }
      }
    },
    formatTime(time) {
      if (!time) return '-'
      return new Date(time).toLocaleString()
    },
    formatSize(size) {
      if (!size) return '0 B'
      if (size < 1024) return size + ' B'
      if (size < 1024 * 1024) return (size / 1024).toFixed(2) + ' KB'
      return (size / (1024 * 1024)).toFixed(2) + ' MB'
    },
    getTypeTag(type) {
      switch (type) {
        case 'video': return 'warning'
        case 'telemetry': return 'info'
        case 'info': return 'success'
        default: return 'default'
      }
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
