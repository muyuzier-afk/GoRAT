<template>
  <div class="clients-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>客户端管理</span>
          <el-button size="small" @click="loadClients">刷新</el-button>
        </div>
      </template>
      <el-table :data="clients" style="width: 100%" v-loading="loading">
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
            <el-button type="warning" size="small" @click="openCommandDialog(scope.row)">命令</el-button>
            <el-button type="danger" size="small" @click="openPowerDialog(scope.row)">电源</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="detailVisible" title="客户端详情" width="600px">
      <template v-if="clientDetail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ clientDetail.client.id }}</el-descriptions-item>
          <el-descriptions-item label="设备ID">{{ clientDetail.client.device_id }}</el-descriptions-item>
          <el-descriptions-item label="名称">{{ clientDetail.client.name }}</el-descriptions-item>
          <el-descriptions-item label="IP">{{ clientDetail.client.ip }}</el-descriptions-item>
          <el-descriptions-item label="操作系统">{{ clientDetail.client.os }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="clientDetail.client.status === 'online' ? 'success' : 'danger'">
              {{ clientDetail.client.status === 'online' ? '在线' : '离线' }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>
        <div v-if="clientDetail.files && clientDetail.files.length" style="margin-top: 16px">
          <h4>关联文件 ({{ clientDetail.files.length }})</h4>
          <el-table :data="clientDetail.files" size="small">
            <el-table-column prop="filename" label="文件名"></el-table-column>
            <el-table-column prop="type" label="类型" width="80"></el-table-column>
          </el-table>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="commandVisible" title="发送命令" width="400px">
      <el-form :model="commandForm" label-width="80px">
        <el-form-item label="命令类型">
          <el-select v-model="commandForm.type" style="width: 100%">
            <el-option label="Shell" value="shell" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="命令内容">
          <el-input v-model="commandForm.content" type="textarea" :rows="3" placeholder="输入命令内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="commandVisible = false">取消</el-button>
        <el-button type="primary" @click="sendCommand" :loading="sending">发送</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="powerVisible" title="电源控制" width="400px">
      <el-form label-width="80px">
        <el-form-item label="操作">
          <el-radio-group v-model="powerAction">
            <el-radio-button label="shutdown">关机</el-radio-button>
            <el-radio-button label="restart">重启</el-radio-button>
            <el-radio-button label="sleep">休眠</el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="powerVisible = false">取消</el-button>
        <el-button type="danger" @click="sendPowerControl" :loading="sending">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Clients',
  data() {
    return {
      clients: [],
      loading: false,
      detailVisible: false,
      clientDetail: null,
      commandVisible: false,
      commandForm: { type: 'shell', content: '' },
      selectedClient: null,
      powerVisible: false,
      powerAction: 'shutdown',
      sending: false
    }
  },
  mounted() {
    this.loadClients()
  },
  methods: {
    async loadClients() {
      this.loading = true
      try {
        const resp = await axios.get('/api/admin/clients')
        this.clients = resp.data.clients || []
      } catch (err) {
        this.$message.error('加载客户端列表失败')
        console.error(err)
      } finally {
        this.loading = false
      }
    },
    async viewClient(client) {
      try {
        const resp = await axios.get(`/api/admin/client/${client.id}`)
        this.clientDetail = resp.data
        this.detailVisible = true
      } catch (err) {
        this.$message.error('加载客户端详情失败')
      }
    },
    openCommandDialog(client) {
      this.selectedClient = client
      this.commandForm = { type: 'shell', content: '' }
      this.commandVisible = true
    },
    async sendCommand() {
      this.sending = true
      try {
        await axios.post(`/api/admin/client/${this.selectedClient.id}/command`, this.commandForm)
        this.$message.success('命令已发送')
        this.commandVisible = false
      } catch (err) {
        this.$message.error('发送命令失败')
      } finally {
        this.sending = false
      }
    },
    openPowerDialog(client) {
      this.selectedClient = client
      this.powerAction = 'shutdown'
      this.powerVisible = true
    },
    async sendPowerControl() {
      this.sending = true
      try {
        await axios.post(`/api/admin/client/${this.selectedClient.id}/power`, { action: this.powerAction })
        this.$message.success('电源命令已发送')
        this.powerVisible = false
      } catch (err) {
        this.$message.error('发送电源命令失败')
      } finally {
        this.sending = false
      }
    },
    formatTime(time) {
      if (!time) return '-'
      return new Date(time).toLocaleString()
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
