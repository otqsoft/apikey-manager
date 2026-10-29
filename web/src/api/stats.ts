/** 统计模块（契约 2.8 + 第 6 节 v2 统计接口） */
import { http } from '@/utils/request'
import type { BreakdownItem, OverviewResult, RankItem, TrendItem } from '@/types'

/** GET /api/stats/overview；admin 返回全局七字段，user 返回本人五字段 */
export const getOverview = () => http.get<OverviewResult>('/stats/overview')

/** 6.1 GET /api/stats/rankings?dim=model|provider&metric=calls|tokens|cost&limit=10 */
export interface RankingQuery {
  dim: 'model' | 'provider'
  metric: 'calls' | 'tokens' | 'cost'
  limit?: number
}
export const getRankings = (params: RankingQuery) =>
  http.get<RankItem[]>('/stats/rankings', { params: { limit: 10, ...params } })

/** 6.2 GET /api/stats/trend?days=30（取值 1~90；无数据日期补 0，数组按日期升序） */
export const getTrend = (days: number) => http.get<TrendItem[]>('/stats/trend', { params: { days } })

/** 6.3 GET /api/stats/breakdown 固定三项 Token 构成 */
export const getBreakdown = () => http.get<BreakdownItem[]>('/stats/breakdown')
