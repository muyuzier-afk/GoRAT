<template>
  <div class="telemetry-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>系统监控</span>
          <div>
            <el-select v-model="selectedClientId" placeholder="选择客户端" @change="loadTelemetry" style="width: 200px; margin-right: 10px">
              <el-option
                v-for="client in clients"
                :key="client.id"
                :label="client.name || client.device_id"
                :value="client.id"
              />
            </el-select>
            <el-button size="small" @click="loadTelemetry">刷新</el-button>
          </div>
        </div>
      </template>
      <div v-if="!selectedClientId" class="empty-hint">
        <el-empty description="请选择一个客户端查看遥测数据" />
      </div>
      <div v-else>
        <div class="chart-row">
          <el-col :span="12">
            <el-card class="chart-card">
              <template #header><span>CPU使用率</span></template>
              <div id="cpuChart" style="width: 100%; height: 300px;"></div>
            </el-card>
          </el-col>
          <el-col :span="12">
            <el-card class="chart-card">
              <template #header><span>内存使用率</span></template>
              <div id="memoryChart" style="width: 100%; height: 300px;"></div>
            </el-card>
          </el-col>
        </div>
        <div class="process-section">
          <el-card class="process-card">
            <template #header><span>进程列表 (最新快照)</span></template>
            <el-table :data="processes" style="width: 100%" v-loading="loading">
              <el-table-column prop="pid" label="PID" width="80"></el-table-column>
              <el-table-column prop="name" label="进程名称"></el-table-column>
              <el-table-column prop="cpu" label="CPU%" width="100"></el-table-column>
              <el-table-column prop="mem" label="内存%" width="100"></el-table-column>
            </el-table>
            <el-empty v-if="!loading && processes.length === 0" description="暂无进程数据" />
          </el-card>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Telemetry',
  data() {
    return {
      clients: [],
      selectedClientId: null,
      telemetryData: [],
      processes: [],
      loading: false,
      cpuChart: null,
      memoryChart: null
    }
  },
  mounted() {
    this.loadClients()
  },
  beforeUnmount() {
    if (this.cpuChart) this.cpuChart.dispose()
    if (this.memoryChart) this.memoryChart.dispose()
  },
  methods: {
    async loadClients() {
      try {
        const resp = await axios.get('/api/admin/clients')
        this.clients = resp.data.clients || []
      } catch (err) {
        console.error('Failed to load clients:', err)
      }
    },
    async loadTelemetry() {
      if (!this.selectedClientId) return
      this.loading = true
      try {
        const resp = await axios.get(`/api/admin/client/${this.selectedClientId}`)
        this.telemetryData = resp.data.telemetry || []
        this.updateCharts()
        this.updateProcesses()
      } catch (err) {
        this.$message.error('加载遥测数据失败')
        console.error(err)
      } finally {
        this.loading = false
      }
    },
    updateCharts() {
      const echarts = require('echarts')
      const times = this.telemetryData.map(t => new Date(t.created_at).toLocaleTimeString()).reverse()
      const cpuData = this.telemetryData.map(t => t.cpu).reverse()
      const memData = this.telemetryData.map(t => t.memory).reverse()

      if (this.cpuChart) this.cpuChart.dispose()
      if (this.memoryChart) this.memoryChart.dispose()

      this.cpuChart = echarts.init(document.getElementById('cpuChart'))
      this.cpuChart.setOption({
        tooltip: { trigger: 'axis' },
        xAxis: { type: 'category', data: times },
        yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
        series: [{ data: cpuData, type: 'line', smooth: true, areaStyle: { opacity: 0.3 } }]
      })

      this.memoryChart = echarts.init(document.getElementById('memoryChart'))
      this.memoryChart.setOption({
        tooltip: { trigger: 'axis' },
        xAxis: { type: 'category', data: times },
        yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
        series: [{ data: memData, type: 'line', smooth: true, areaStyle: { opacity: 0.3 }, itemStyle: { color: '#67C23A' } }]
      })
    },
    updateProcesses() {
      if (this.telemetryData.length === 0) {
        this.processes = []
        return
      }
      const latest = this.telemetryData[0]
      try {
        const parsed = typeof latest.processes === 'string' ? JSON.parse(latest.processes) : latest.processes
        this.processes = Array.isArray(parsed) ? parsed : []
      } catch {
        this.processes = []
      }
    }
  }
}
</script>

<style scoped>
.telemetry-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.empty-hint {
  padding: 60px 0;
  text-align: center;
}

.chart-row {
  display: flex;
  gap: 20px;
  margin-bottom: 20px;
}

.chart-card {
  flex: 1;
  height: 350px;
}

.process-section {
  margin-top: 20px;
}

.process-card {
  min-height: 300px;
}
</style>
