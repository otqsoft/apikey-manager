/**
 * 路由表 + 全局守卫（对齐契约第 4 节路由与角色可见性）
 * - admin：全部菜单；user：仅仪表盘 + 对话测试
 * - 未登录访问受保护页面 → /login；登录态但角色不匹配 → 提示并回 /dashboard
 */
import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/store/user'
import MainLayout from '@/layout/MainLayout.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { public: true, title: '登录' }
  },
  {
    path: '/',
    component: MainLayout,
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { title: '仪表盘', roles: ['admin', 'user'] }
      },
      {
        path: 'providers',
        name: 'Providers',
        component: () => import('@/views/providers/index.vue'),
        meta: { title: '供应商管理', roles: ['admin'] }
      },
      {
        path: 'models',
        name: 'Models',
        component: () => import('@/views/models/index.vue'),
        meta: { title: '模型管理', roles: ['admin'] }
      },
      {
        path: 'keys',
        name: 'Keys',
        component: () => import('@/views/keys/index.vue'),
        meta: { title: 'API-KEY 管理', roles: ['admin'] }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/users/index.vue'),
        meta: { title: '用户管理', roles: ['admin'] }
      },
      {
        path: 'chat',
        name: 'Chat',
        component: () => import('@/views/chat/index.vue'),
        meta: { title: '对话测试', roles: ['admin', 'user'] }
      }
    ]
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

router.beforeEach(async (to) => {
  const store = useUserStore()

  // 公开页面：已登录则不再停留登录页
  if (to.meta.public) {
    if (store.token) return { path: '/dashboard' }
    return true
  }

  // 未登录 → 跳登录，记录回跳地址
  if (!store.token) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }

  // 有 token 但无用户档案（刷新页面 / 首次挂载）→ 拉取 profile
  if (!store.user) {
    try {
      await store.fetchProfile()
    } catch {
      store.clearAuth()
      return { path: '/login', query: { redirect: to.fullPath } }
    }
  }

  // 角色权限校验（契约第 4 节）
  const roles = (to.meta.roles as string[] | undefined) || []
  if (roles.length > 0 && store.user && !roles.includes(store.user.role)) {
    ElMessage.error('当前账号无权访问该页面')
    return { path: '/dashboard' }
  }

  document.title = `${(to.meta.title as string) || ''} · LLM 密钥管理台`
  return true
})

export default router
