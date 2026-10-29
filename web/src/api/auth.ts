/** 认证模块（契约 2.1 / 8.1） */
import { http } from '@/utils/request'
import type { LoginParams, LoginResult, PasswordParams, Profile, RegisterReq, RegisterResult } from '@/types'

/** POST /api/auth/login 登录（公开接口） */
export const login = (data: LoginParams) => http.post<LoginResult>('/auth/login', data)

/** POST /api/auth/register 注册（公开接口，不签发 token；契约 8.1） */
export const register = (data: RegisterReq) => http.post<RegisterResult>('/auth/register', data)

/** POST /api/auth/logout 退出（前端清登录态为主，后端黑名单占位） */
export const logout = () => http.post('/auth/logout')

/** GET /api/auth/profile 当前用户信息 + permissions */
export const getProfile = () => http.get<Profile>('/auth/profile')

/** PUT /api/auth/password 修改本人密码 */
export const changePassword = (data: PasswordParams) => http.put('/auth/password', data)
