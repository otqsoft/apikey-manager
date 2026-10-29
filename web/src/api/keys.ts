/** API-KEY 模块（契约 2.4；管理端 admin，对话端用 /me/keys） */
import { http } from '@/utils/request'
import type { ApiKeyForm, ApiKeyItem, MyKeyItem, PageResult, UserItem } from '@/types'

export interface KeyQuery {
  page?: number
  page_size?: number
  provider_id?: number | ''
  keyword?: string
}

/** GET /api/keys 管理列表（掩码 key） */
export const getKeyPage = (params: KeyQuery) => http.get<PageResult<ApiKeyItem>>('/keys', { params })

/** POST /api/keys 新建 */
export const createKey = (data: ApiKeyForm) => http.post('/keys', data)

/** GET /api/keys/:id 详情（返回完整 api_key，用于复制明文） */
export const getKeyDetail = (id: number) => http.get<ApiKeyItem>(`/keys/${id}`)

/** PUT /api/keys/:id 更新 */
export const updateKey = (id: number, data: ApiKeyForm) => http.put(`/keys/${id}`, data)

/** DELETE /api/keys/:id 删除（级联清理分配关系） */
export const deleteKey = (id: number) => http.delete(`/keys/${id}`)

/** PUT /api/keys/:id/status 启停 */
export const updateKeyStatus = (id: number, status: number) =>
  http.put(`/keys/${id}/status`, { status })

/** GET /api/keys/:id/users 已分配用户（user_ids + 用户简表） */
export interface KeyUsersResult {
  user_ids: number[]
  users: UserItem[]
}
export const getKeyUsers = (id: number) => http.get<KeyUsersResult>(`/keys/${id}/users`)

/** PUT /api/keys/:id/users 全量覆盖分配关系 */
export const saveKeyUsers = (id: number, user_ids: number[]) =>
  http.put(`/keys/${id}/users`, { user_ids })

/** GET /api/me/keys 登录用户可用 key（含 provider 与该供应商启用模型） */
export const getMyKeys = () => http.get<MyKeyItem[]>('/me/keys')
