/** 供应商模块（契约 2.2，admin） */
import { http } from '@/utils/request'
import type { PageResult, Provider, ProviderForm } from '@/types'

export interface ProviderQuery {
  page?: number
  page_size?: number
  keyword?: string
  status?: number | ''
}

/** GET /api/providers 分页 + keyword + status */
export const getProviderPage = (params: ProviderQuery) =>
  http.get<PageResult<Provider>>('/providers', { params })

/** GET /api/providers 全量启用项（模型表单下拉用；page_size 放大） */
export const getEnabledProviders = () =>
  http.get<PageResult<Provider>>('/providers', { params: { page: 1, page_size: 100, status: 1 } })

/** POST /api/providers 新建 */
export const createProvider = (data: ProviderForm) => http.post('/providers', data)

/** GET /api/providers/:id 详情 */
export const getProviderDetail = (id: number) => http.get<Provider>(`/providers/${id}`)

/** PUT /api/providers/:id 更新 */
export const updateProvider = (id: number, data: ProviderForm) => http.put(`/providers/${id}`, data)

/** DELETE /api/providers/:id 删除（后端拦截被引用删除） */
export const deleteProvider = (id: number) => http.delete(`/providers/${id}`)

/** PUT /api/providers/:id/status 启停 */
export const updateProviderStatus = (id: number, status: number) =>
  http.put(`/providers/${id}/status`, { status })
