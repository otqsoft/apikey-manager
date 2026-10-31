<template>
  <div class="dashboard">
    <!-- 欢迎卡：按小时问候 + 刷新 -->
    <el-card shadow="never" class="welcome-card">
      <div class="welcome">
        <div>
          <h3>{{ greeting() }}，{{ store.displayName }}</h3>
          <p>{{ store.isAdmin ? '全局运营概览（管理员口径）' : '我的使用概览（个人口径）' }} · 数据随对话自动聚合</p>
        </div>
        <el-button type="primary" plain :icon="Refresh" :loading="loading" @click="loadAll">刷新统计</el-button>
      </div>
    </el-card>

    <!-- 统计卡：admin 七项 / user 五项（GET /api/stats/overview） -->
    <el-row :gutter="16">
      <el-col v-for="card in statCards" :key="card.key" :xs="24" :sm="12" :md="8" :lg="6" :xl="4" class="card-col">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon" :style="{ background: card.bg }">
            <el-icon :size="24"><component :is="card.icon" /></el-icon>
          </div>
          <div class="stat-body">
            <div class="stat-label">{{ card.label }}</div>
            <div class="stat-value num">{{ card.value }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表栅格 -->
    <el-row :gutter="16">
      <!-- ① Top 榜单：维度 + 指标切换 -->
      <el-col :xs="24" :md="24" :lg="14" class="card-col">
        <ChartCard title="调用 Top 榜单" :empty="rankEmpty" :loading="loading" height="340px" :option="rankOption">
          <template #extra>
            <el-radio-group v-model="dim" size="small" @change="loadRankings">
              <el-radio-button value="model">模型</el-radio-button>
              <el-radio-button value="provider">供应商</el-radio-button>
            </el-radio-group>
            <el-radio-group v-model="metric" size="small" @change="loadRankings">
              <el-radio-button value="calls">调用次数</el-radio-button>
              <el-radio-button value="tokens">Token</el-radio-button>
              <el-radio-button value="cost">费用</el-radio-button>
            </el-radio-group>
          </template>
        </ChartCard>
      </el-col>

      <!-- ② 供应商调用占比 -->
      <el-col :xs="24" :md="12" :lg="10" class="card-col">
        <ChartCard title="供应商调用占比" :empty="shareEmpty" :loading="loading" height="340px" :option="shareOption" />
      </el-col>

      <!-- ③ 近 N 天趋势：柱=费用，折线=Token -->
      <el-col :xs="24" :md="24" :lg="14" class="card-col">
        <ChartCard title="用量趋势" :empty="trendEmpty" :loading="loading" height="340px" :option="trendOption">
          <template #extra>
            <el-radio-group v-model="days" size="small" @change="loadTrend">
              <el-radio-button :value="7">近 7 天</el-radio-button>
              <el-radio-button :value="30">近 30 天</el-radio-button>
            </el-radio-group>
          </template>
        </ChartCard>
      </el-col>

      <!-- ④ Token 用量构成 -->
      <el-col :xs="24" :md="12" :lg="10" class="card-col">
        <ChartCard title="Token 用量构成" :empty="breakdownEmpty" :loading="loading" height="340px" :option="breakdownOption" />
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { useUserStore } from '@/store/user'
import { getOverview, getRankings, getTrend, getBreakdown } from '@/api/stats'
import type { BreakdownItem, OverviewAdmin, OverviewUser, RankItem, TrendItem } from '@/types'
import type { EChartsOption } from '@/utils/echarts'
import ChartCard from '@/components/ChartCard.vue'
import { formatInt, formatMoney, formatMoneyCompact } from '@/utils/format'

const store = useUserStore()
const loading = ref(false)
let toastGuard = false

/** 接口失败统一静默兜底：单个图表空态，最多一条提示，不阻塞页面 */
function failToast(msg: string) {
  if (toastGuard) return
  toastGuard = true
  ElMessage.warning(msg)
}

// ---------------- 问候语 ----------------
function greeting(): string {
  const h = new Date().getHours()
  if (h >= 5 && h < 12) return '上午好'
  if (h >= 12 && h < 14) return '中午好'
  if (h >= 14 && h < 18) return '下午好'
  if (h >= 18 && h < 24) return '晚上好'
  return '夜深好'
}

// ---------------- 概览统计卡（契约 2.8） ----------------
const adminData = ref<OverviewAdmin>({
  provider_count: 0,
  model_count: 0,
  key_count: 0,
  user_count: 0,
  today_tokens: 0,
  today_cost: 0,
  total_cost: 0
})
const userData = ref<OverviewUser>({
  my_key_count: 0,
  my_session_count: 0,
  today_tokens: 0,
  today_cost: 0,
  total_cost: 0
})

async function loadOverview() {
  try {
    const res = await getOverview()
    if (store.isAdmin) Object.assign(adminData.value, res)
    else Object.assign(userData.value, res)
  } catch {
    failToast('概览统计加载失败，稍后可点击刷新')
  }
}

const statCards = computed(() => {
    // 普通用户：供应商/模型为全局共享资源显示真实总数；Key 与会话为个人口径
    if (!store.isAdmin) {
      return [
        { key: 'provider', label: '供应商数', value: formatInt(userData.value.provider_count), icon: 'Connection', bg: 'linear-gradient(135deg,#dbeafe,#eff6ff)' },
        { key: 'model', label: '模型数', value: formatInt(userData.value.model_count), icon: 'Files', bg: 'linear-gradient(135deg,#dcfce7,#f0fdf4)' },
        { key: 'myKey', label: 'Key 数', value: formatInt(userData.value.my_key_count), icon: 'Key', bg: 'linear-gradient(135deg,#fef3c7,#fffbeb)' },
        // { key: 'mySession', label: '我的会话数', value: formatInt(userData.value.my_session_count), icon: 'ChatDotRound', bg: 'linear-gradient(135deg,#fee2e2,#fef2f2)' },
        { key: 'todayTokens', label: '今日Token', value: formatInt(userData.value.today_tokens), icon: 'DataLine', bg: 'linear-gradient(135deg,#e0f2fe,#f0f9ff)' },
        { key: 'todayCost', label: '今日费用', value: formatMoney(userData.value.today_cost), icon: 'Money', bg: 'linear-gradient(135deg,#dbeafe,#eff6ff)' },
        { key: 'totalCost', label: '累计费用', value: formatMoneyCompact(userData.value.total_cost), icon: 'Wallet', bg: 'linear-gradient(135deg,#dcfce7,#f0fdf4)' }
      ]
    }
    return [
      { key: 'provider', label: '供应商数', value: formatInt(adminData.value.provider_count), icon: 'Connection', bg: 'linear-gradient(135deg,#dbeafe,#eff6ff)' },
      { key: 'model', label: '模型数', value: formatInt(adminData.value.model_count), icon: 'Files', bg: 'linear-gradient(135deg,#dcfce7,#f0fdf4)' },
      { key: 'key', label: 'Key 数', value: formatInt(adminData.value.key_count), icon: 'Key', bg: 'linear-gradient(135deg,#fef3c7,#fffbeb)' },
      // { key: 'user', label: '用户数', value: formatInt(adminData.value.user_count), icon: 'User', bg: 'linear-gradient(135deg,#fee2e2,#fef2f2)' },
      { key: 'todayTokens', label: '今日Token', value: formatInt(adminData.value.today_tokens), icon: 'DataLine', bg: 'linear-gradient(135deg,#e0f2fe,#f0f9ff)' },
      { key: 'todayCost', label: '今日费用', value: formatMoney(adminData.value.today_cost), icon: 'Money', bg: 'linear-gradient(135deg,#dbeafe,#eff6ff)' },
      { key: 'totalCost', label: '累计费用', value: formatMoneyCompact(adminData.value.total_cost), icon: 'Wallet', bg: 'linear-gradient(135deg,#dcfce7,#f0fdf4)' }
    ]
})

// ---------------- ① Top 榜单（6.1） ----------------
const dim = ref<'model' | 'provider'>('model')
const metric = ref<'calls' | 'tokens' | 'cost'>('calls')
const rankData = ref<RankItem[]>([])
const rankEmpty = computed(() => rankData.value.length === 0)

async function loadRankings() {
  try {
    const res = await getRankings({ dim: dim.value, metric: metric.value, limit: 10 })
    rankData.value = res || []
  } catch {
    rankData.value = []
    failToast('榜单数据加载失败')
  }
}

const metricOf = (d: RankItem) => (metric.value === 'calls' ? d.calls : metric.value === 'tokens' ? d.tokens : d.cost)

const rankOption = computed<EChartsOption>(() => ({
  grid: { left: 8, right: 30, top: 8, bottom: 8, containLabel: true },
  tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
  xAxis: {
    type: 'value',
    splitLine: { lineStyle: { color: '#eef2f7' } },
    axisLabel: { color: '#64748b' }
  },
  yAxis: {
    type: 'category',
    inverse: true,
    data: rankData.value.map((d) => d.name),
    axisTick: { show: false },
    axisLine: { lineStyle: { color: '#cbd5e1' } },
    axisLabel: { color: '#334155' }
  },
  series: [
    {
      type: 'bar',
      data: rankData.value.map((d) => metricOf(d)),
      barWidth: 13,
      itemStyle: { color: '#2563eb', borderRadius: [0, 4, 4, 0] },
      label: { show: true, position: 'right', color: '#475569', fontSize: 11 }
    }
  ]
}))

// ---------------- ② 供应商调用占比（6.1 dim=provider metric=calls） ----------------
const shareData = ref<RankItem[]>([])
const shareEmpty = computed(() => shareData.value.length === 0)
const totalCalls = computed(() => shareData.value.reduce((s, d) => s + d.calls, 0))

async function loadShare() {
  try {
    const res = await getRankings({ dim: 'provider', metric: 'calls', limit: 20 })
    shareData.value = res || []
  } catch {
    shareData.value = []
    failToast('供应商占比加载失败')
  }
}

const shareOption = computed<EChartsOption>(() => ({
  color: ['#2563eb', '#60a5fa', '#7dd3fc', '#14b8a6', '#f59e0b', '#94a3b8'],
  tooltip: { trigger: 'item', formatter: '{b}：{c} 次（{d}%）' },
  legend: { bottom: 0, textStyle: { color: '#475569', fontSize: 11 }, itemWidth: 10, itemHeight: 10 },
  title: {
    text: formatInt(totalCalls.value),
    subtext: '总调用次数',
    left: 'center',
    top: '36%',
    textStyle: { fontSize: 20, color: '#1f2937' },
    subtextStyle: { color: '#94a3b8', fontSize: 11 }
  },
  series: [
    {
      type: 'pie',
      radius: ['54%', '74%'],
      center: ['50%', '44%'],
      avoidLabelOverlap: true,
      label: { show: false },
      itemStyle: { borderColor: '#fff', borderWidth: 2 },
      data: shareData.value.map((d) => ({ name: d.name, value: d.calls }))
    }
  ]
}))

// ---------------- ③ 近 N 天趋势（6.2） ----------------
const days = ref<number>(7)
const trendData = ref<TrendItem[]>([])
const trendEmpty = computed(() => trendData.value.length === 0)

async function loadTrend() {
  try {
    const res = await getTrend(days.value)
    trendData.value = res || []
  } catch {
    trendData.value = []
    failToast('趋势数据加载失败')
  }
}

const trendOption = computed<EChartsOption>(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: ['费用(元)', 'Token 数'], top: 0, textStyle: { color: '#475569' } },
  grid: { left: 8, right: 8, top: 34, bottom: 4, containLabel: true },
  xAxis: {
    type: 'category',
    data: trendData.value.map((t) => t.date.slice(5)),
    axisLabel: { color: '#64748b' }
  },
  yAxis: [
    { type: 'value', name: '费用(元)', axisLabel: { color: '#64748b' }, splitLine: { lineStyle: { color: '#eef2f7' } } },
    { type: 'value', name: 'Token', axisLabel: { color: '#64748b' }, splitLine: { show: false } }
  ],
  series: [
    {
      name: '费用(元)',
      type: 'bar',
      data: trendData.value.map((t) => Number(t.cost.toFixed(4))),
      itemStyle: { color: '#93c5fd', borderRadius: [3, 3, 0, 0] },
      barWidth: 12
    },
    {
      name: 'Token 数',
      type: 'line',
      yAxisIndex: 1,
      smooth: true,
      data: trendData.value.map((t) => t.tokens),
      itemStyle: { color: '#2563eb' },
      lineStyle: { width: 2 },
      symbolSize: 6
    }
  ]
}))

// ---------------- ④ Token 用量构成（6.3） ----------------
const breakdownData = ref<BreakdownItem[]>([])
const breakdownEmpty = computed(() => breakdownData.value.length === 0)
const breakdownTokens = computed(() => breakdownData.value.reduce((s, d) => s + Number(d.tokens || 0), 0))

async function loadBreakdown() {
  try {
    const res = await getBreakdown()
    breakdownData.value = res || []
  } catch {
    breakdownData.value = []
    failToast('用量构成加载失败')
  }
}

const breakdownOption = computed<EChartsOption>(() => ({
  color: ['#2563eb', '#60a5fa', '#f59e0b'],
  tooltip: { trigger: 'item', formatter: '{b}：{c} Token（{d}%）' },
  legend: { bottom: 0, textStyle: { color: '#475569', fontSize: 11 }, itemWidth: 10, itemHeight: 10 },
  title: {
    text: formatInt(breakdownTokens.value),
    subtext: '总 Token',
    left: 'center',
    top: '36%',
    textStyle: { fontSize: 20, color: '#1f2937' },
    subtextStyle: { color: '#94a3b8', fontSize: 11 }
  },
  series: [
    {
      type: 'pie',
      radius: ['54%', '74%'],
      center: ['50%', '44%'],
      label: { show: false },
      itemStyle: { borderColor: '#fff', borderWidth: 2 },
      data: breakdownData.value.map((d) => ({ name: d.name, value: d.tokens }))
    }
  ]
}))

// ---------------- 统一加载 ----------------
async function loadAll() {
  loading.value = true
  toastGuard = false
  await Promise.allSettled([loadOverview(), loadRankings(), loadShare(), loadTrend(), loadBreakdown()])
  loading.value = false
}

onMounted(loadAll)
watch(dim, loadRankings)
watch(metric, loadRankings)
</script>

<style scoped>
/* 统一无边框轻量卡片 */
.dashboard :deep(.el-card) {
  border: none;
  border-radius: 12px;
}
.welcome-card {
  margin-bottom: 16px;
  border-radius: 12px;
}
.welcome {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.welcome h3 {
  margin: 0 0 6px;
  font-size: 19px;
  color: #1f2937;
}
.welcome p {
  margin: 0;
  color: #94a3b8;
  font-size: 13px;
}
.card-col {
  margin-bottom: 16px;
}
.stat-card {
  border-radius: 12px;
  transition: box-shadow 0.2s, transform 0.2s;
}
.stat-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 20px rgba(37, 99, 235, 0.12);
}
.stat-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 100%;
}
.stat-icon {
  width: 44px;
  height: 44px;
  border-radius: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #2563eb;
  flex-shrink: 0;
}
.stat-body {
  flex: 1;
  min-width: 0;
}
.stat-label {
  color: #94a3b8;
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.stat-value {
  font-size: 21px;
  font-weight: 700;
  color: #1f2937;
  margin-top: 4px;
}
.num {
  font-variant-numeric: tabular-nums;
}
</style>
