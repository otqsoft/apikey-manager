/**
 * axios 实例封装（对齐契约第 0 节）：
 * - 请求拦截：自动携带 Authorization: Bearer <token>
 * - 响应拦截：解包 {code,msg,data}；code=0 返回 data；401 清登录态跳登录；403 提示无权限；其余 code 弹错误提示
 */
import axios, { type AxiosRequestConfig, type AxiosError } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'
import { useUserStore } from '@/store/user'

const request = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// 请求拦截：注入 Bearer token
request.interceptors.request.use((config) => {
  const token = localStorage.getItem('ak_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

/** 401 统一处理：清空登录态并跳转登录页 */
function handleUnauthorized(msg?: string) {
  const store = useUserStore()
  store.clearAuth()
  ElMessage.error(msg || '登录已过期，请重新登录')
  if (router.currentRoute.value.path !== '/login') {
    router.replace({ path: '/login', query: { redirect: router.currentRoute.value.fullPath } })
  }
}

// 响应拦截：统一解包与错误处理
request.interceptors.response.use(
  (response) => {
    const res = response.data
    // 非标准包（理论上不存在）直接透传
    if (res === null || typeof res !== 'object' || !('code' in res)) {
      return res
    }
    if (res.code === 0) {
      return res.data
    }
    if (res.code === 401) {
      handleUnauthorized(res.msg)
      return Promise.reject(new Error(res.msg || '未登录或登录已过期'))
    }
    if (res.code === 403) {
      ElMessage.error(res.msg || '无权限执行该操作')
      return Promise.reject(new Error(res.msg || '无权限'))
    }
    ElMessage.error(res.msg || '请求失败，请稍后重试')
    return Promise.reject(new Error(res.msg || '请求失败'))
  },
  (error: AxiosError<{ code?: number; msg?: string }>) => {
    const status = error.response?.status
    const msg = error.response?.data?.msg
    if (status === 401) {
      handleUnauthorized(msg)
    } else if (status === 403) {
      ElMessage.error(msg || '无权限执行该操作')
    } else if (status === 400) {
      ElMessage.error(msg || '参数错误，请检查表单')
    } else if (status === 500) {
      ElMessage.error(msg || '服务端错误，请稍后重试')
    } else {
      ElMessage.error(msg || error.message || '网络异常，请检查后端服务是否启动')
    }
    return Promise.reject(error)
  }
)

/** 对外暴露带泛型的简写方法（拦截器已解包 data） */
export const http = {
  get<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return request.get(url, config) as unknown as Promise<T>
  },
  post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    return request.post(url, data, config) as unknown as Promise<T>
  },
  put<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    return request.put(url, data, config) as unknown as Promise<T>
  },
  delete<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return request.delete(url, config) as unknown as Promise<T>
  }
}

export default request
