/**
 * axios 实例封装（对齐契约第 0 节）：
 * - 请求拦截：自动携带 Authorization: Bearer <token>
 * - 响应拦截：解包 {code,msg,data}；code=0 返回 data；401 清登录态跳登录；403 提示无权限；其余 code 弹错误提示
 */
import axios, { type AxiosRequestConfig, type AxiosError } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'
import { useUserStore } from '@/store/user'

/**
 * 解析后端接口地址（优先级从高到低）：
 * 1. 运行时全局变量 window.__AK_API_BASE__：可在 index.html 里注入，改完刷新即生效，无需重新构建
 *    <script>window.__AK_API_BASE__ = 'http://192.168.1.10:8090/api'</script>
 * 2. 构建期环境变量 VITE_API_BASE
 * 3. Electron 桌面端：页面以 file: 协议加载，没有代理，直连同机后端
 * 4. 浏览器部署（nginx 等）：同源相对路径 /api，由 nginx 反代到后端 8090
 */
const DEFAULT_DESKTOP_API = 'http://127.0.0.1:8090/api'

function resolveBaseURL(): string {
  const runtime = (window as unknown as Record<string, unknown>).__AK_API_BASE__
  if (typeof runtime === 'string' && runtime.length > 0) {
    return runtime
  }
  const injected = (import.meta.env as Record<string, unknown>).VITE_API_BASE
  if (typeof injected === 'string' && injected.length > 0) {
    return injected
  }
  if (typeof window !== 'undefined' && window.location.protocol === 'file:') {
    return DEFAULT_DESKTOP_API
  }
  return '/api'
}

const API_BASE = resolveBaseURL()

// 启动时打印一次实际使用的接口地址，便于排查「请求打到哪儿去了」
if (typeof console !== 'undefined') {
  console.info('[api] baseURL =', API_BASE)
}

const request = axios.create({
  baseURL: API_BASE,
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
      // 带上可操作的排查方向：没有状态码 = 请求根本没拿到响应（后端未启动 / 跨域被拦）
      // 404·502·504 = 拿到了 nginx 的响应，说明静态托管生效但反向代理或后端有问题
      let hint = ''
      if (status === 404 || status === 502 || status === 503 || status === 504) {
        hint = `请求失败（HTTP ${status}）：请确认 nginx 已把 /api/ 反向代理到后端，且后端服务已启动`
      } else if (!status) {
        hint = `无法连接后端服务（${API_BASE}）：请检查后端是否启动、接口地址是否正确；跨域部署需后端开启 CORS`
      }
      ElMessage.error(msg || hint || error.message || '请求失败，请稍后重试')
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
