/** 日志管理模块（仅 admin） */
import { http } from '@/utils/request'
import type { LoginLogItem, OperationLogItem, PageResult } from '@/types'

export interface OperationLogQuery {
  page?: number
  page_size?: number
  keyword?: string
  module?: string
  status?: number
}

export interface LoginLogQuery {
  page?: number
  page_size?: number
  keyword?: string
  status?: number
}

/** GET /api/logs/operation 操作日志分页 */
export const getOperationLogs = (params: OperationLogQuery) =>
  http.get<PageResult<OperationLogItem>>('/logs/operation', { params })

/** GET /api/logs/login 登录日志分页 */
export const getLoginLogs = (params: LoginLogQuery) =>
  http.get<PageResult<LoginLogItem>>('/logs/login', { params })
