/**
 * ECharts 按需注册（5.x）：仅注册仪表盘用到的图表与组件，控制打包体积
 */
import * as echarts from 'echarts/core'
import { BarChart, LineChart, PieChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([
  BarChart,
  LineChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
  CanvasRenderer
])

/** 创建 ECharts 实例 */
export function initEcharts(el: HTMLElement) {
  return echarts.init(el)
}

export { echarts }
export type { EChartsCoreOption as EChartsOption } from 'echarts/core'
