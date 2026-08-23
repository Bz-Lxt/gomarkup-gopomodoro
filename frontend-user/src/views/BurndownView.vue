<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import * as echarts from 'echarts'
import { api, type BurndownChart, type Metrics, type Milestone, type Project } from '../lib/api'
import { toast } from '../lib/toast'
import { connectWS } from '../lib/ws'

const projects = ref<Project[]>([])
const pid = ref('')
const milestones = ref<Milestone[]>([])
const mid = ref('')
const gran = ref('day')
const chart = ref<BurndownChart | null>(null)
const metrics = ref<Metrics | null>(null)
const el = ref<HTMLDivElement | null>(null)
let inst: echarts.ECharts | null = null
let sock: ReturnType<typeof connectWS> | null = null

onMounted(async () => {
  try {
    projects.value = await api.projects()
    pid.value = projects.value[0]?.id || ''
    await loadMs()
  } catch (e: unknown) {
    toast('err', e instanceof Error ? e.message : '加载失败')
  }
  sock = connectWS((msg) => {
    if (msg.type === 'burndown.update' && mid.value) void loadChart()
  })
})

watch(pid, loadMs)
watch([mid, gran], loadChart)

async function loadMs() {
  if (!pid.value) return
  milestones.value = await api.milestones(pid.value)
  mid.value = milestones.value[0]?.id || ''
}

async function loadChart() {
  if (!mid.value) return
  chart.value = await api.burndown(mid.value, gran.value)
  metrics.value = await api.metrics(mid.value)
  sock?.send({ type: 'subscribe', milestone_id: mid.value })
  await nextTick()
  draw()
}

function draw() {
  if (!el.value || !chart.value) return
  if (!inst) inst = echarts.init(el.value)
  inst.setOption({
    backgroundColor: 'transparent',
    tooltip: { trigger: 'axis' },
    legend: { data: ['理想燃尽', '真实燃尽'], textStyle: { color: '#8aa396' } },
    grid: { left: 40, right: 20, top: 40, bottom: 40 },
    xAxis: { type: 'category', data: mergeX(), axisLine: { lineStyle: { color: '#24352c' } }, axisLabel: { color: '#8aa396' } },
    yAxis: { type: 'value', name: '剩余点数', axisLabel: { color: '#8aa396' }, splitLine: { lineStyle: { color: '#24352c' } } },
    series: [
      { name: '理想燃尽', type: 'line', data: chart.value.ideal.y, smooth: false, symbol: 'none', lineStyle: { type: 'dashed', color: '#3ee0c4', width: 2 } },
      {
        name: '真实燃尽', type: 'line', data: alignActual(), smooth: false,
        lineStyle: { color: '#c6ff3d', width: 3 }, itemStyle: { color: '#c6ff3d' },
        markPoint: {
          data: (chart.value.scope_marks || []).map((m) => ({ name: '范围变更', coord: [m.at.slice(0, 10), m.remaining], value: 'SCOPE' })),
          itemStyle: { color: '#ffb020' },
        },
      },
    ],
  }, { notMerge: false })
}

function mergeX() {
  return chart.value?.ideal.x || []
}

function alignActual() {
  if (!chart.value) return []
  const map = new Map<string, number>()
  chart.value.actual.x.forEach((x, i) => map.set(x.slice(0, 10), chart.value!.actual.y[i]))
  let last = chart.value.baseline_points
  return (chart.value.ideal.x || []).map((d) => {
    if (map.has(d)) last = map.get(d)!
    return last
  })
}
</script>

<template>
  <div class="w-full p-4 md:p-6 space-y-4">
    <div class="flex flex-wrap items-end gap-3">
      <div>
        <p class="text-xs text-fog">项目</p>
        <select v-model="pid" class="bg-bg border border-line rounded-lg px-3 py-2">
          <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
      </div>
      <div>
        <p class="text-xs text-fog">里程碑</p>
        <select v-model="mid" class="bg-bg border border-line rounded-lg px-3 py-2 min-w-[16rem]">
          <option v-for="m in milestones" :key="m.id" :value="m.id">{{ m.title }}</option>
        </select>
      </div>
      <div>
        <p class="text-xs text-fog">粒度</p>
        <select v-model="gran" class="bg-bg border border-line rounded-lg px-3 py-2">
          <option value="day">日</option>
          <option value="week">周</option>
        </select>
      </div>
    </div>

    <div v-if="metrics" class="grid grid-cols-2 lg:grid-cols-5 gap-3">
      <div class="card p-4"><p class="text-xs text-fog">今日番茄</p><p class="font-mono text-3xl text-acid">{{ metrics.today_completed }}</p></div>
      <div class="card p-4"><p class="text-xs text-fog">本周番茄</p><p class="font-mono text-3xl">{{ metrics.week_completed }}</p></div>
      <div class="card p-4"><p class="text-xs text-fog">日均产能</p><p class="font-mono text-3xl">{{ metrics.avg_daily_velocity.toFixed(1) }}</p></div>
      <div class="card p-4"><p class="text-xs text-fog">专注废弃率</p><p class="font-mono text-3xl text-rose">{{ (metrics.abort_rate * 100).toFixed(0) }}%</p></div>
      <div class="card p-4"><p class="text-xs text-fog">预测完成日</p><p class="font-mono text-xl">{{ metrics.predicted_done_on || '—' }}</p></div>
    </div>

    <div class="card p-3 h-[520px] w-full">
      <div ref="el" class="w-full h-full" />
      <p v-if="!chart" class="text-fog text-sm p-8">选择里程碑以渲染双折线。</p>
    </div>
  </div>
</template>
