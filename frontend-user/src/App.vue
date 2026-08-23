<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { clearToken, getToken } from './lib/api'
import { connectWS } from './lib/ws'
import { useSession } from './stores/session'
import { dismiss, useToasts } from './lib/toast'
import type { SessionView } from './lib/api'

const route = useRoute()
const router = useRouter()
const session = useSession()
const toasts = useToasts()
let sock: ReturnType<typeof connectWS> | null = null

onMounted(async () => {
  if (!getToken()) return
  try { await session.refresh() } catch { /* ignore */ }
  sock = connectWS((msg) => {
    if (msg.type === 'session.state') session.apply(msg.payload as SessionView)
    if (msg.type === 'session.tick') {
      const p = msg.payload as { remaining_ms: number; state: string }
      if (session.view) session.view.remaining_ms = p.remaining_ms
    }
    if (msg.type === 'grace') {
      session.grace = (msg.payload as { remaining_s: number }).remaining_s
    }
  })
  if (session.view?.session?.resume_token) {
    sock.send({ type: 'hello', resume_token: session.view.session.resume_token })
  }
})

onUnmounted(() => sock?.close())

function logout() {
  clearToken()
  router.push('/login')
}
</script>

<template>
  <div class="noise" />
  <div class="min-h-screen w-full">
    <header v-if="route.path !== '/login'" class="sticky top-0 z-20 border-b border-line bg-bg/80 backdrop-blur">
      <div class="w-full px-5 py-3 flex items-center justify-between gap-4">
        <div class="flex items-center gap-3">
          <span class="font-display text-xl tracking-tight text-acid">GoPomodoro</span>
          <span class="hidden sm:inline text-xs text-fog font-mono">MINI POMODORO SCRUM</span>
        </div>
        <nav class="flex items-center gap-2 text-sm">
          <router-link to="/" class="px-3 py-1.5 rounded-full border border-line hover:border-acid">看板</router-link>
          <router-link to="/focus" class="px-3 py-1.5 rounded-full border border-line hover:border-acid">番茄钟</router-link>
          <router-link to="/burndown" class="px-3 py-1.5 rounded-full border border-line hover:border-acid">燃尽</router-link>
          <button class="px-3 py-1.5 text-fog" @click="logout">退出</button>
        </nav>
      </div>
    </header>
    <router-view />
    <div class="fixed right-4 top-20 z-50 flex flex-col gap-2 w-80 max-w-[calc(100vw-2rem)]">
      <div v-for="t in toasts.items" :key="t.id"
        class="card px-4 py-3 text-sm flex justify-between gap-3"
        :class="t.kind === 'err' ? 'border-rose' : t.kind === 'ok' ? 'border-acid' : ''">
        <span>{{ t.text }}</span>
        <button class="text-fog" @click="dismiss(t.id)">×</button>
      </div>
    </div>
  </div>
</template>
