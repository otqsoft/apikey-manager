<template>
  <el-card shadow="never" class="chart-card">
    <div class="chart-head">
      <span class="chart-title">{{ title }}</span>
      <div class="chart-extra">
        <slot name="extra" />
      </div>
    </div>
    <div v-loading="loading" class="chart-body" :style="{ height }">
      <el-empty v-if="empty" description="暂无调用数据" :image-size="64" class="chart-empty" />
      <div v-show="!empty" ref="bodyRef" class="chart-inner" />
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { initEcharts } from '@/utils/echarts'
import type { EChartsOption } from '@/utils/echarts'

const props = withDefaults(
  defineProps<{
    title: string
    loading?: boolean
    height?: string
    /** ECharts option；为空数组/无数据时配合 empty 展示空态 */
    option?: EChartsOption | null
    /** 是否空数据（空态展示「暂无调用数据」） */
    empty?: boolean
  }>(),
  {
    loading: false,
    height: '320px',
    option: null,
    empty: false
  }
)

const bodyRef = ref<HTMLElement>()
let chart: ReturnType<typeof initEcharts> | null = null
let ro: ResizeObserver | null = null

function ensureChart() {
  if (chart || !bodyRef.value || props.empty) return
  chart = initEcharts(bodyRef.value)
  if (props.option) chart.setOption(props.option)
  ro = new ResizeObserver(() => chart && chart.resize())
  ro.observe(bodyRef.value)
}

function renderOption() {
  if (props.empty) return
  ensureChart()
  if (chart && props.option) chart.setOption(props.option, true)
}

onMounted(async () => {
  await nextTick()
  renderOption()
})

watch(
  () => [props.option, props.empty],
  async () => {
    await nextTick()
    renderOption()
  }
)

onUnmounted(() => {
  ro?.disconnect()
  chart?.dispose()
  chart = null
})
</script>

<style scoped>
.chart-card {
  border-radius: 12px;
  border: 1px solid #e8edf3;
}
.chart-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}
.chart-title {
  font-size: 15px;
  font-weight: 600;
  color: #1f2937;
}
.chart-extra {
  display: flex;
  align-items: center;
  gap: 8px;
}
.chart-body {
  position: relative;
}
.chart-inner {
  width: 100%;
  height: 100%;
}
.chart-empty {
  padding: 30px 0;
}
</style>
