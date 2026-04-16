<template>
  <div class="app-container">
    <el-container style="height: 100vh; border: 1px solid #eee;">
      <el-aside width="200px" style="background-color: #303133;">
        <div class="logo">GoRAT</div>
        <el-menu
          default-active="1"
          class="el-menu-vertical-demo"
          background-color="#303133"
          text-color="#fff"
          active-text-color="#409EFF"
        >
          <el-menu-item index="1">
            <el-icon><House /></el-icon>
            <span>首页</span>
          </el-menu-item>
          <el-menu-item index="2" @click="$router.push('/clients')">
            <el-icon><User /></el-icon>
            <span>客户端管理</span>
          </el-menu-item>
          <el-menu-item index="3" @click="$router.push('/files')">
            <el-icon><Document /></el-icon>
            <span>文件管理</span>
          </el-menu-item>
          <el-menu-item index="4" @click="$router.push('/telemetry')">
            <el-icon><DataAnalysis /></el-icon>
            <span>系统监控</span>
          </el-menu-item>
          <el-menu-item index="5" @click="$router.push('/settings')">
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      <el-container>
        <el-header style="text-align: right; font-size: 12px;">
          <el-dropdown>
            <span class="el-dropdown-link">
              管理员 <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item>个人中心</el-dropdown-item>
                <el-dropdown-item>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </el-header>
        <el-main>
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script>
import { House, User, Document, DataAnalysis, Setting, ArrowDown } from '@element-plus/icons-vue'

export default {
  name: 'App',
  components: {
    House,
    User,
    Document,
    DataAnalysis,
    Setting,
    ArrowDown
  }
}
</script>

<style>
.app-container {
  width: 100%;
  height: 100vh;
}

.el-header {
  background-color: #B3C0D1;
  color: #333;
  line-height: 60px;
}

.el-aside {
  color: #333;
}

.logo {
  color: white;
  font-size: 16px;
  font-weight: bold;
  text-align: center;
  padding: 20px 0;
  border-bottom: 1px solid #409EFF;
}
</style>
