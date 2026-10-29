/** 用户模块（契约 2.5，admin） */
import { http } from '@/utils/request'
import type { ApiKeyItem, PageResult, UserCreateForm, UserItem, UserUpdateForm } from '@/types'

export interface UserQuery {
  page?: number
  page_size?: number
  keyword?: string
}

/** GET /api/users 分页 + keyword（不返回 password_hash） */
export const getUserPage = (params: UserQuery) =>
  http.get<PageResult<UserItem>>('/users', { params })

/** POST /api/users 新建用户（含初始密码） */
export const createUser = (data: UserCreateForm) => http.post('/users', data)

/** GET /api/users/:id 详情 */
export const getUserDetail = (id: number) => http.get<UserItem>(`/users/${id}`)

/** PUT /api/users/:id 更新 role/display_name/status（不更新密码） */
export const updateUser = (id: number, data: UserUpdateForm) => http.put(`/users/${id}`, data)

/** PUT /api/users/:id/status 启停 */
export const updateUserStatus = (id: number, status: number) =>
  http.put(`/users/${id}/status`, { status })

/** POST /api/users/:id/reset-password 管理员重置密码 */
export const resetUserPassword = (id: number, password: string) =>
  http.post(`/users/${id}/reset-password`, { password })

/** DELETE /api/users/:id 删除（后端禁止删除自己；前端同样拦截） */
export const deleteUser = (id: number) => http.delete(`/users/${id}`)

/** GET /api/users/:id/keys 该用户已分配 key 列表 */
export const getUserKeys = (id: number) =>
  http.get<ApiKeyItem[] | PageResult<ApiKeyItem>>(`/users/${id}/keys`)
