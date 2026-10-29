/** 模型模块（契约 2.3，admin） */
import { http } from '@/utils/request'
import type { ModelForm, ModelItem, PageResult } from '@/types'

export interface ModelQuery {
  page?: number
  page_size?: number
  provider_id?: number | ''
  keyword?: string
}

/** GET /api/models 分页 + provider_id + keyword；列表可带 provider_name */
export const getModelPage = (params: ModelQuery) =>
  http.get<PageResult<ModelItem>>('/models', { params })

/** POST /api/models 新建 */
export const createModel = (data: ModelForm) => http.post('/models', data)

/** GET /api/models/:id 详情 */
export const getModelDetail = (id: number) => http.get<ModelItem>(`/models/${id}`)

/** PUT /api/models/:id 更新 */
export const updateModel = (id: number, data: ModelForm) => http.put(`/models/${id}`, data)

/** DELETE /api/models/:id 删除 */
export const deleteModel = (id: number) => http.delete(`/models/${id}`)

/** PUT /api/models/:id/status 启停 */
export const updateModelStatus = (id: number, status: number) =>
  http.put(`/models/${id}/status`, { status })
