/** 用户态 Pinia store：token + 用户信息 + 权限；localStorage 持久化 */
import { defineStore } from 'pinia'
import { login as loginApi, logout as logoutApiCall, changePassword as passwordApi, getProfile } from '@/api/auth'
import type { LoginParams, PasswordParams, Profile, UserInfo } from '@/types'

const TOKEN_KEY = 'ak_token'
const USER_KEY = 'ak_user'

function readUser(): UserInfo | null {
  try {
    const raw = localStorage.getItem(USER_KEY)
    return raw ? (JSON.parse(raw) as UserInfo) : null
  } catch {
    return null
  }
}

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) || '',
    user: readUser(),
    permissions: [] as string[]
  }),
  getters: {
    isAdmin: (s) => s.user?.role === 'admin',
    displayName: (s) => s.user?.display_name || s.user?.username || '未登录',
    role: (s) => s.user?.role || ''
  },
  actions: {
    persist() {
      if (this.token) localStorage.setItem(TOKEN_KEY, this.token)
      else localStorage.removeItem(TOKEN_KEY)
      if (this.user) localStorage.setItem(USER_KEY, JSON.stringify(this.user))
      else localStorage.removeItem(USER_KEY)
    },
    async login(params: LoginParams) {
      const res = await loginApi(params)
      this.token = res.token
      this.user = res.user
      this.permissions = []
      this.persist()
    },
    /** 刷新当前用户档案（含 permissions），用于路由守卫校验 */
    async fetchProfile() {
      const profile: Profile = await getProfile()
      this.user = {
        id: profile.id,
        username: profile.username,
        role: profile.role,
        display_name: profile.display_name
      }
      this.permissions = profile.permissions || []
      this.persist()
    },
    async changePassword(params: PasswordParams) {
      await passwordApi(params)
    },
    async logoutApi() {
      try {
        await logoutApiCall()
      } catch {
        /* 后端黑名单占位，失败不影响前端清态 */
      }
    },
    clearAuth() {
      this.token = ''
      this.user = null
      this.permissions = []
      this.persist()
    }
  }
})
