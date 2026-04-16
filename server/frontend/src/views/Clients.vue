<template>
  <div class="clients-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>客户端管理</span>
          <el-button type="primary" size="small">添加客户端</el-button>
        </div>
      </template>
      <el-table :data="clients" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80"></el-table-column>
        <el-table-column prop="device_id" label="设备ID" width="180"></el-table-column>
        <el-table-column prop="name" label="设备名称"></el-table-column>
        <el-table-column prop="ip" label="IP地址" width="150"></el-table-column>
        <el-table-column prop="os" label="操作系统" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.os === 'Windows' ? 'primary' : 'info'">
              {{ scope.row.os }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'online' ? 'success' : 'danger'">
              {{ scope.row.status === 'online' ? '在线' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_heartbeat" label="最后心跳" width="180">
          <template #default="scope">
            {{ formatTime(scope.row.last_heartbeat) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <el-button type="primary" size="small" @click="viewClient(scope.row)">查看</el-button>
            <el-button type="warning" size="small" @click="sendCommand(scope.row)">命令</el-button>
            <el-button type="danger" size="small" @click="powerControl(scope.row)">电源</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script>
export default {
  name: 'Clients',
  data() {
    return {
      clients: []
    }
  },
  mounted() {
    this.loadClients()
  },
  methods: {
    loadClients() {
      // 模拟数据，实际项目中应该从API获取
      this.clients = [
        {
          id: 1,
          device_id: 'factory-001',
          name: '实验室电脑1',
          ip: '192.168.1.101',
          os: 'Windows',
          status: 'online',
          last_heartbeat: new Date().toISOString()
        },
        {
          id: 2,
          device_id: 'factory-002',
          name: '实验室电脑2',
          ip: '192.168.1.102',
          os: 'Windows',
          status: 'offline',
          last_heartbeat: new Date(Date.now() - 3600000).toISOString()
        },
        {
          id: 3,
          device_id: 'linux-001',
          name: 'Linux服务器1',
          ip: '192.168.1.201',
          os: 'Linux',
          status: 'online',
          last_heartbeat: new Date().toISOString()
        }
      ]
    },
    formatTime(time) {
      return new Date(time).toLocaleString()
    },
    viewClient(client) {
      // 查看客户端详情
      console.log('View client:', client)
    },
    sendCommand(client) {
      // 发送命令
      console.log('Send command to:', client)
    },
    powerControl(client) {
      // 电源控制
      console.log('Power control for:', client)
    }
  }
}
</script>

<style scoped>
.clients-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
