<template>
  <div class="telemetry-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>系统监控</span>
          <el-select v-model="selectedClient" placeholder="选择客户端">
            <el-option
              v-for="client in clients"
              :key="client.id"
              :label="client.name"
              :value="client.id"
            />
          </el-select>
        </div>
      </template>
      <div class="chart-row">
        <el-col :span="12">
          <el-card class="chart-card">
            <template #header>
              <span>CPU使用率</span>
            </template>
            <div id="cpuChart" style="width: 100%; height: 300px;"></div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card class="chart-card">
            <template #header>
              <span>内存使用率</span>
            </template>
            <div id="memoryChart" style="width: 100%; height: 300px;"></div>
          </el-card>
        </el-col>
      </div>
      <div class="process-section">
        <el-card class="process-card">
          <template #header>
            <span>进程列表</span>
          </template>
          <el-table :data="processes" style="width: 100%">
            <el-table-column prop="pid" label="PID" width="80"></el-table-column>
            <el-table-column prop="name" label="进程名称"></el-table-column>
            <el-table-column prop="cpu" label="CPU%" width="100"></el-table-column>
            <el-table-column prop="mem" label="内存%" width="100"></el-table-column>
          </el-table>
        </el-card>
      </div>
    </el-card>
  </div>
</template>

<script>
export default {
  name: 'Telemetry',
  data() {
    return {
      clients: [
        { id: 1, name: '实验室电脑1' },
        { id: 2, name: '实验室电脑2' },
        { id: 3, name: '实验室电脑3' }
      ],
      selectedClient: 1,
      processes: [],
      cpuChart: null,
      memoryChart: null
    }
  },
  mounted() {
    this.loadProcesses()
    this.initCharts()
  },
  methods: {
    loadProcesses() {
      // 模拟数据，实际项目中应该从API获取
      this.processes = [
        { pid: 1234, name: 'chrome.exe', cpu: 10.5, mem: 5.2 },
        { pid: 5678, name: 'explorer.exe', cpu: 2.3, mem: 3.1 },
        { pid: 9012, name: 'factoryeye.exe', cpu: 1.5, mem: 2.0 },
        { pid: 3456, name: 'svchost.exe', cpu: 0.8, mem: 1.2 }
      ]
    },
    initCharts() {
      // 初始化图表
      const echarts = require('echarts')
      
      // CPU图表
      this.cpuChart = echarts.init(document.getElementById('cpuChart'))
      this.cpuChart.setOption({
        tooltip: { trigger: 'axis' },
        xAxis: {
          type: 'category',
          data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00']
        },
        yAxis: { type: 'value', max: 100 },
        series: [{
          data: [30, 40, 20, 50, 60, 40],
          type: 'line',
          smooth: true
        }]
      })
      
      // 内存图表
      this.memoryChart = echarts.init(document.getElementById('memoryChart'))
      this.memoryChart.setOption({
        tooltip: { trigger: 'axis' },
        xAxis: {
          type: 'category',
          data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00']
        },
        yAxis: { type: 'value', max: 100 },
        series: [{
          data: [60, 70, 50, 80, 70, 60],
          type: 'line',
          smooth: true
        }]
      })
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
