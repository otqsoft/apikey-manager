<template>
  <el-container class="layout-container">
    <el-aside width="220px" class="layout-aside">
      <div class="logo">
        <el-icon><Key /></el-icon>
        <span>LLM 密钥管理台</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        background-color="transparent"
        text-color="#c3cede"
        active-text-color="#ffffff"
        class="side-menu"
      >
        <el-menu-item v-for="item in visibleMenus" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="layout-header" height="56px">
        <div class="header-title">{{ currentTitle }}</div>
        <div class="header-user">
          <el-dropdown @command="onCommand">
            <span class="user-trigger">
              <el-icon class="avatar-icon"><UserFilled /></el-icon>
              <span class="user-name">{{ store.displayName }}</span>
              <el-tag size="small" :type="store.isAdmin ? 'danger' : 'info'" class="role-tag">
                {{ store.isAdmin ? '管理员' : '普通用户' }}
              </el-tag>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="password">修改密码</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="layout-main">
        <router-view />
      </el-main>
    </el-container>

    <ChangePasswordDialog v-model:visible="pwdDialogVisible" />
  </el-container>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/store/user'
import ChangePasswordDialog from './components/ChangePasswordDialog.vue'

const store = useUserStore()
const route = useRoute()
const router = useRouter()

const pwdDialogVisible = ref(false)

interface MenuItem {
  path: string
  title: string
  icon: string
  roles: string[]
}

/** 契约第 4 节菜单：admin 全量；user 仅仪表盘 + 对话测试 */
const allMenus: MenuItem[] = [
  { path: '/dashboard', title: '仪表盘', icon: 'Odometer', roles: ['admin', 'user'] },
  { path: '/providers', title: '供应商管理', icon: 'Connection', roles: ['admin'] },
  { path: '/models', title: '模型管理', icon: 'Files', roles: ['admin'] },
  { path: '/keys', title: 'API-KEY 管理', icon: 'Key', roles: ['admin'] },
  { path: '/users', title: '用户管理', icon: 'User', roles: ['admin'] },
  { path: '/chat', title: '对话测试', icon: 'ChatDotRound', roles: ['admin', 'user'] }
]

const visibleMenus = computed(() =>
  allMenus.filter((m) => store.user && m.roles.includes(store.user.role))
)

const activeMenu = computed(() => route.path)
const currentTitle = computed(() => (route.meta.title as string) || '')

async function onCommand(cmd: string) {
  if (cmd === 'password') {
    pwdDialogVisible.value = true
  } else if (cmd === 'logout') {
    await ElMessageBox.confirm('确定退出登录吗？', '提示', { type: 'warning' })
    await store.logoutApi()
    store.clearAuth()
    ElMessage.success('已退出登录')
    router.replace('/login')
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}
.layout-aside {
  background: linear-gradient(180deg, #1e293b 0%, #172033 100%);
  overflow-y: auto;
}
.logo {
  height: 58px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 18px;
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  background: linear-gradient(90deg, #2563eb, #1d4fc0);
}
.logo .el-icon {
  font-size: 20px;
  color: #dbeafe;
}
.side-menu {
  border-right: none;
  padding: 8px 10px;
}
.side-menu :deep(.el-menu-item) {
  border-radius: 8px;
  margin-bottom: 4px;
  position: relative;
}
.side-menu :deep(.el-menu-item.is-active) {
  background: rgba(37, 99, 235, 0.18) !important;
}
.side-menu :deep(.el-menu-item.is-active)::before {
  content: '';
  position: absolute;
  left: 0;
  top: 8px;
  bottom: 8px;
  width: 3px;
  border-radius: 2px;
  background: #60a5fa;
}
.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-bottom: 1px solid #e8edf3;
  box-shadow: 0 1px 4px rgba(15, 23, 42, 0.05);
}
.header-title {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}
.user-trigger {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  outline: none;
}
.avatar-icon {
  font-size: 20px;
  color: #2563eb;
}
.user-name {
  font-size: 14px;
  color: #303133;
}
.role-tag {
  margin: 0 4px;
}
.layout-main {
  background-color: #f3f5f9;
  padding: 16px;
}
</style>
