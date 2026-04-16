<template>
  <div v-if="!isAuthenticated" class="login-container">
    <el-card class="login-card">
      <template #header>
        <div class="login-header">GoRAT Login</div>
      </template>
      <el-form :model="loginForm" @submit.prevent="handleLogin">
        <el-form-item label="Username">
          <el-input v-model="loginForm.username" placeholder="Admin username" />
        </el-form-item>
        <el-form-item label="Password">
          <el-input v-model="loginForm.password" type="password" placeholder="Password" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" style="width: 100%" @click="handleLogin" :loading="loginLoading">Login</el-button>
        </el-form-item>
        <el-alert v-if="loginError" :title="loginError" type="error" show-icon :closable="false" />
      </el-form>
    </el-card>
  </div>
  <div v-else class="app-container">
    <el-container style="height: 100vh; border: 1px solid #eee;">
      <el-aside width="200px" style="background-color: #303133;">
        <div class="logo">GoRAT</div>
        <el-menu
          :default-active="activeMenu"
          class="el-menu-vertical-demo"
          background-color="#303133"
          text-color="#fff"
          active-text-color="#409EFF"
          router
        >
          <el-menu-item index="/" @click="$router.push('/')">
            <el-icon><House /></el-icon>
            <span>Home</span>
          </el-menu-item>
          <el-menu-item index="/clients" @click="$router.push('/clients')">
            <el-icon><User /></el-icon>
            <span>Clients</span>
          </el-menu-item>
          <el-menu-item index="/files" @click="$router.push('/files')">
            <el-icon><Document /></el-icon>
            <span>Files</span>
          </el-menu-item>
          <el-menu-item index="/telemetry" @click="$router.push('/telemetry')">
            <el-icon><DataAnalysis /></el-icon>
            <span>Monitoring</span>
          </el-menu-item>
          <el-menu-item index="/settings" @click="$router.push('/settings')">
            <el-icon><Setting /></el-icon>
            <span>Settings</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      <el-container>
        <el-header style="text-align: right; font-size: 12px;">
          <el-dropdown @command="handleDropdownCommand">
            <span class="el-dropdown-link">
              {{ username }} <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">Logout</el-dropdown-item>
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
import axios from 'axios'

export default {
  name: 'App',
  components: { House, User, Document, DataAnalysis, Setting, ArrowDown },
  data() {
    return {
      loginForm: { username: '', password: '' },
      loginLoading: false,
      loginError: '',
      activeMenu: '/'
    }
  },
  computed: {
    isAuthenticated() {
      return !!localStorage.getItem('token')
    },
    username() {
      return localStorage.getItem('username') || 'admin'
    }
  },
  methods: {
    async handleLogin() {
      this.loginLoading = true
      this.loginError = ''
      try {
        const resp = await axios.post('/api/admin/login', {
          username: this.loginForm.username,
          password: this.loginForm.password
        })
        localStorage.setItem('token', resp.data.token)
        localStorage.setItem('username', this.loginForm.username)
      } catch (err) {
        if (err.response && err.response.data && err.response.data.error) {
          this.loginError = err.response.data.error
        } else {
          this.loginError = 'Login failed'
        }
      } finally {
        this.loginLoading = false
      }
    },
    handleDropdownCommand(cmd) {
      if (cmd === 'logout') {
        localStorage.removeItem('token')
        localStorage.removeItem('username')
      }
    }
  }
}
</script>

<style>
.app-container {
  width: 100%;
  height: 100vh;
}

.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #f0f2f5;
}

.login-card {
  width: 400px;
}

.login-header {
  font-size: 18px;
  font-weight: bold;
  text-align: center;
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
