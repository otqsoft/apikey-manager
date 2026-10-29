<template>
  <el-container class="layout-container">
    <el-aside width="220px" class="layout-aside">
      <div class="logo">
        <div class="logo-mark">
          <img src="/logo.png" alt="logo" />
        </div>
        <span class="logo-text">API Key 控制台</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        background-color="transparent"
        text-color="#8fa5c2"
        active-text-color="#ffffff"
        class="side-menu"
      >
        <el-menu-item v-for="item in visibleMenus" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>
      </el-menu>

      <!-- <div class="aside-footer">
        <span class="env-dot" />
        后端 :8090 · MySQL
      </div> -->
    </el-aside>

    <el-container>
      <el-header class="layout-header" height="56px">
        <div class="header-title">{{ currentTitle }}</div>
        <div class="header-user">
          <el-dropdown @command="onCommand">
            <span class="user-trigger">
              <span class="avatar-circle">{{ avatarChar }}</span>
              <span class="user-name">{{ store.displayName }}</span>
              <el-tag size="small" :type="store.isAdmin ? 'danger' : 'info'" effect="light" class="role-tag" round>
                {{ store.isAdmin ? '管理员' : '普通用户' }}
              </el-tag>
              <el-icon class="arrow"><ArrowDown /></el-icon>
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

/** 头像字符：展示名首字符，兜底「用」 */
const avatarChar = computed(() => (store.displayName || '用').trim().charAt(0).toUpperCase())

interface MenuItem {
  path: string
  title: string
  icon: string
  roles: string[]
}

/** 菜单：admin 全量 + 全局数据；user 除用户管理外均可（数据范围限自己名下） */
const allMenus: MenuItem[] = [
  { path: '/dashboard', title: '仪表盘', icon: 'Odometer', roles: ['admin', 'user'] },
  { path: '/providers', title: '供应商管理', icon: 'Connection', roles: ['admin', 'user'] },
  { path: '/models', title: '模型管理', icon: 'Files', roles: ['admin', 'user'] },
  { path: '/keys', title: 'API-KEY 管理', icon: 'Key', roles: ['admin', 'user'] },
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
/* 暗科技蓝侧边栏 */
.layout-aside {
  background: linear-gradient(180deg, #0a1628 0%, #0d1d36 100%);
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}
.logo {
  height: 56px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 16px;
  flex-shrink: 0;
}
.logo-mark {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  background: none;
}
.logo-mark img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.logo-text {
  color: #f1f5f9;
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0.5px;
  white-space: nowrap;
}
.side-menu {
  border-right: none;
  padding: 10px 10px;
  flex: 1;
}
.side-menu :deep(.el-menu-item) {
  border-radius: 8px;
  margin-bottom: 3px;
  position: relative;
  height: 42px;
  line-height: 42px;
  transition: background 0.15s ease, color 0.15s ease;
}
.side-menu :deep(.el-menu-item:hover) {
  background: rgba(255, 255, 255, 0.05);
}
.side-menu :deep(.el-menu-item.is-active) {
  background: linear-gradient(90deg, rgba(59, 130, 246, 0.22), rgba(0, 212, 255, 0.1));
  color: #fff;
}
.side-menu :deep(.el-menu-item.is-active)::before {
  content: '';
  position: absolute;
  left: 0;
  top: 9px;
  bottom: 9px;
  width: 3px;
  border-radius: 2px;
  background: #00d4ff;
}
.aside-footer {
  padding: 12px 18px;
  font-size: 11px;
  color: #4b6079;
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}
.env-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #10b981;
}

/* 轻量顶栏：白底、无阴影 */
.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-bottom: 1px solid #e8eef6;
}
.header-title {
  font-size: 15px;
  font-weight: 600;
  color: #1f2937;
}
.user-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  outline: none;
}
.avatar-circle {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(135deg, #3b82f6, #3b82f6);
}
.user-name {
  font-size: 14px;
  color: #303133;
}
.arrow {
  font-size: 12px;
  color: #94a3b8;
}
.layout-main {
  background-color: #f4f7fc;
  padding: 14px 16px;
  overflow-y: auto;
}
</style>
