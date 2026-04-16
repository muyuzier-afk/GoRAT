<template>
  <div class="home-container">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <span>系统概览</span>
        </div>
      </template>
      <div class="overview-stats">
        <el-row :gutter="20">
          <el-col :span="6">
            <el-card class="stat-card">
              <div class="stat-content">
                <div class="stat-number">{{ totalClients }}</div>
                <div class="stat-label">总客户端</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card">
              <div class="stat-content">
                <div class="stat-number">{{ onlineClients }}</div>
                <div class="stat-label">在线客户端</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card">
              <div class="stat-content">
                <div class="stat-number">{{ totalFiles }}</div>
                <div class="stat-label">总文件数</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card">
              <div class="stat-content">
                <div class="stat-number">{{ totalStorage }} GB</div>
                <div class="stat-label">总存储</div>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>
      <div class="chart-container">
        <el-card class="chart-card">
          <template #header>
            <span>系统负载</span>
          </template>
          <div id="loadChart" style="width: 100%; height: 300px;"></div>
        </el-card>
      </div>
    </el-card>
  </div>
</template>

<script>
export default {
  name: 'Home',
  data() {
    return {
      totalClients: 0,
      onlineClients: 0,
      totalFiles: 0,
      totalStorage: 0,
      loadChart: null
    }
  },
  mounted() {
    this.loadData()
    this.initChart()
  },
  methods: {
    loadData() {
      // 模拟数据，实际项目中应该从API获取
      this.totalClients = 10
      this.onlineClients = 8
      this.totalFiles = 120
      this.totalStorage = 15.6
    },
    initChart() {
      // 模拟图表数据
      const echarts = require('echarts')
      this.loadChart = echarts.init(document.getElementById('loadChart'))
      
      const option = {
        tooltip: {
          trigger: 'axis'
        },
        legend: {
          data: ['CPU', '内存']
        },
        grid: {
          left: '3%',
          right: '4%',
          bottom: '3%',
          containLabel: true
        },
        xAxis: {
          type: 'category',
          boundaryGap: false,
          data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00']
        },
        yAxis: {
          type: 'value',
          max: 100
        },
        series: [
          {
            name: 'CPU',
            type: 'line',
            data: [30, 40, 20, 50, 60, 40]
          },
          {
            name: '内存',
            type: 'line',
            data: [60, 70, 50, 80, 70, 60]
          }
        ]
      }
      
      this.loadChart.setOption(option)
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

.chart-container {
  margin-top: 20px;
}

.chart-card {
  height: 350px;
}
</style>
