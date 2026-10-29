/** 通用格式化工具 */

/** 金额（人民币元）展示：统一 ¥ + 两位小数 */
export function formatMoney(n: number | undefined | null): string {
  const v = Number(n || 0)
  return `¥${v.toFixed(2)}`
}

/** 紧凑金额（仪表盘大卡片用）：过万显示 ¥x.xx万，否则千分位两位小数 */
export function formatMoneyCompact(n: number | undefined | null): string {
  const v = Number(n || 0)
  if (Math.abs(v) >= 10000) return `¥${(v / 10000).toFixed(2)}万`
  return `¥${v.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

/** 整数千分位（Token 数等计数展示） */
export function formatInt(n: number | undefined | null): string {
  return Number(n || 0).toLocaleString('zh-CN')
}

/**
 * 上下文长度：原始 Token 数 -> 百万 Token（M）展示
 * 128000 -> 0.13M；1000000 -> 1.00M；0 -> 0.00M
 */
export function formatContext(n: number | undefined | null): string {
  return `${(Number(n || 0) / 1e6).toFixed(2)}M`
}

/** ISO 时间字符串 -> 本地展示 yyyy-MM-dd HH:mm */
export function formatDateTime(iso?: string | null): string {
  if (!iso) return '-'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  const pad = (x: number) => String(x).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

/** 日期截取 YYYY-MM-DD */
export function formatDate(iso?: string | null): string {
  if (!iso) return '永久'
  return iso.slice(0, 10)
}
