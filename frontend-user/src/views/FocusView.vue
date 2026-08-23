<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { api } from '../lib/api'
import { toast } from '../lib/toast'
import { currentNoise, playNoise, setVolume, stopNoise } from '../lib/noise'
import { useSession } from '../stores/session'
import Modal from '../components/Modal.vue'

const session = useSession()
const confirm = ref(false)
const vol = ref(0.18)
const noise = ref(currentNoise())
const tick = ref(Date.now())
let iv = 0

onMounted(async () => {
  try { await session.refresh() } catch { /* ignore */ }
  iv = window.setInterval(() => {
    tick.value = Date.now()
    if (session.view?.session?.state === 'running') {
      session.view.remaining_ms = Math.max(0, session.view.remaining_ms - 1000)
    }
  }, 1000)
})
onUnmounted(() => window.clearInterval(iv))

const remain = computed(() => {
  void tick.value
  return session.view?.remaining_ms ?? 25 * 60 * 1000
})
const mm = computed(() => String(Math.floor(remain.value / 60000)).padStart(2, '0'))
const ss = computed(() => String(Math.floor((remain.value % 60000) / 1000)).padStart(2, '0'))
const pct = computed(() => {
  const total = session.view?.session?.focus_duration_ms || 1500000
  return 1 - remain.value / total
})
const state = computed(() => session.view?.session?.state || 'idle')

const ring = computed(() => {
  const r = 120
  const c = 2 * Math.PI * r
  return { c, dash: c * pct.value }
})

async function startDemo() {
  try {
    const projects = await api.projects()
    const tasks = await api.tasks(projects[0].id)
    const t = tasks.find((x) => x.kanban_column !== 'done') || tasks[0]
    if (!t) { toast('err', '没有可专注的任务'); return }
    session.apply(await api.startPomo(t.id))
    session.flash('start')
  } catch (e: unknown) {
    toast('err', e instanceof Error ? e.message : '启动失败')
  }
}

async function pause() {
  if (!session.view?.session) return
  session.apply(await api.pausePomo(session.view.session.id))
  session.flash('pause')
}
async function resume() {
  if (!session.view?.session) return
  session.apply(await api.resumePomo(session.view.session.id))
  session.flash('start')
}
async function abort() {
  if (!session.view?.session) return
  session.apply(await api.abortPomo(session.view.session.id))
  session.flash('abort')
  confirm.value = false
  toast('info', '本枚番茄已放弃，不计入燃尽')
}

function toggleNoise(k: 'rain' | 'cafe' | 'white') {
  if (noise.value === k) { stopNoise(); noise.value = null; return }
  playNoise(k)
  noise.value = k
}
</script>

<template>
  <div class="w-full min-h-[calc(100vh-64px)] flex flex-col items-center justify-center px-4 py-10"
    :class="{
      'animate-[pulse_1s_ease]': session.fx === 'start',
      'saturate-50': state === 'paused' || session.fx === 'pause',
      'contrast-150': session.fx === 'complete',
    }">
    <p class="font-mono text-xs text-fog mb-6 tracking-[0.3em]">
      {{ state === 'running' ? 'FOCUS LOCKED' : state === 'paused' ? 'HOLDING PATTERN' : state === 'completed' ? 'SETTLED' : 'STANDBY' }}
    </p>
    <div class="relative" :class="session.fx === 'abort' ? 'animate-bounce' : ''">
      <svg width="300" height="300" viewBox="0 0 300 300">
        <circle cx="150" cy="150" r="120" fill="none" stroke="#24352c" stroke-width="10" />
        <circle cx="150" cy="150" r="120" fill="none" stroke="#c6ff3d" stroke-width="10"
          stroke-linecap="round" transform="rotate(-90 150 150)"
          :stroke-dasharray="`${ring.dash} ${ring.c}`" />
      </svg>
      <div class="absolute inset-0 flex flex-col items-center justify-center">
        <div class="font-mono text-7xl tracking-tight">{{ mm }}:{{ ss }}</div>
        <p class="text-fog text-sm mt-2">服务端权威倒计时</p>
      </div>
    </div>

    <div class="flex flex-wrap gap-3 mt-10">
      <button v-if="state === 'idle' || !session.view?.session" class="bg-acid text-bg px-6 py-2 rounded-full font-semibold" @click="startDemo">开始</button>
      <button v-if="state === 'running'" class="border border-amber text-amber px-6 py-2 rounded-full" @click="pause">暂停</button>
      <button v-if="state === 'paused'" class="bg-acid text-bg px-6 py-2 rounded-full" @click="resume">继续</button>
      <button v-if="state === 'running' || state === 'paused'" class="border border-rose text-rose px-6 py-2 rounded-full" @click="confirm = true">放弃</button>
    </div>

    <div class="mt-10 card p-4 w-full max-w-lg">
      <p class="text-xs text-fog mb-3">白噪音（本地合成，无外网音频）</p>
      <div class="flex gap-2">
        <button class="px-3 py-1.5 rounded-full border text-sm" :class="noise==='rain' ? 'border-acid text-acid' : 'border-line'" @click="toggleNoise('rain')">雨声</button>
        <button class="px-3 py-1.5 rounded-full border text-sm" :class="noise==='cafe' ? 'border-acid text-acid' : 'border-line'" @click="toggleNoise('cafe')">咖啡馆</button>
        <button class="px-3 py-1.5 rounded-full border text-sm" :class="noise==='white' ? 'border-acid text-acid' : 'border-line'" @click="toggleNoise('white')">白噪</button>
      </div>
      <input class="w-full mt-3" type="range" min="0" max="1" step="0.01" v-model.number="vol" @input="setVolume(vol)" />
    </div>

    <p v-if="session.grace" class="mt-6 text-amber text-sm">连接中断，剩余 {{ session.grace }}s 内重连可恢复</p>
    <p v-if="state === 'completed'" class="mt-6 font-display text-3xl text-acid">SETTLED · 已结算 1 点</p>
  </div>

  <Modal v-if="confirm" title="确认放弃这枚番茄？" @close="confirm = false">
    <p class="text-sm text-fog mb-4">放弃不扣燃尽，但会推高专注废弃率。</p>
    <div class="flex justify-end gap-2">
      <button @click="confirm = false">取消</button>
      <button class="bg-rose text-white px-4 py-2 rounded-lg" @click="abort">确认放弃</button>
    </div>
  </Modal>
</template>
