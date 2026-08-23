<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, type Milestone, type Project, type Task } from '../lib/api'
import { toast } from '../lib/toast'
import { useSession } from '../stores/session'
import Modal from '../components/Modal.vue'

const router = useRouter()
const session = useSession()
const loading = ref(true)
const projects = ref<Project[]>([])
const projectId = ref('')
const milestones = ref<Milestone[]>([])
const filterMid = ref('')
const tasks = ref<Task[]>([])
const showTask = ref(false)
const showMs = ref(false)
const confirmAbort = ref(false)
const formErr = ref('')
const taskForm = ref({ title: '', estimated_pomodoros: 3, milestone_id: '', kanban_column: 'todo' })
const msForm = ref({ title: '', start_date: '2026-08-23', due_date: '2026-09-15', baseline_points: 10 })

const cols = [
  { key: 'backlog', label: 'Backlog' },
  { key: 'todo', label: 'Todo' },
  { key: 'in_progress', label: 'In Progress' },
  { key: 'done', label: 'Done' },
]

const visible = computed(() =>
  filterMid.value ? tasks.value.filter((t) => t.milestone_id === filterMid.value) : tasks.value,
)

onMounted(load)

watch(projectId, () => { filterMid.value = ''; reloadBoard() })

async function load() {
  loading.value = true
  try {
    projects.value = await api.projects()
    if (!projects.value.length) {
      const p = await api.createProject('我的工作区')
      projects.value = [p]
    }
    projectId.value = projects.value[0].id
    await reloadBoard()
  } catch (e: unknown) {
    toast('err', e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

async function reloadBoard() {
  if (!projectId.value) return
  milestones.value = await api.milestones(projectId.value)
  tasks.value = await api.tasks(projectId.value)
}

function byCol(col: string) {
  return visible.value.filter((t) => t.kanban_column === col).sort((a, b) => a.sort_order - b.sort_order)
}

function riskColor(r: string) {
  if (r === 'overdue' || r === 'at_risk') return 'bg-rose'
  if (r === 'tight') return 'bg-amber'
  if (r === 'done') return 'bg-cyan'
  return 'bg-acid'
}

let dragId = ''
function onDragStart(id: string) { dragId = id }
async function onDrop(col: string) {
  const t = tasks.value.find((x) => x.id === dragId)
  if (!t) return
  const prev = t.kanban_column
  t.kanban_column = col
  try {
    const items = cols.flatMap((c) =>
      byCol(c.key).map((x, i) => ({ id: x.id, kanban_column: c.key, sort_order: i })),
    )
    await api.reorder(items)
  } catch (e: unknown) {
    t.kanban_column = prev
    toast('err', e instanceof Error ? e.message : '拖拽失败')
  }
}

async function createTask() {
  formErr.value = ''
  if (!taskForm.value.title.trim()) {
    formErr.value = '标题必填'
    return
  }
  if (taskForm.value.estimated_pomodoros < 1) {
    formErr.value = '估点须为正整数'
    return
  }
  try {
    await api.createTask(projectId.value, {
      ...taskForm.value,
      milestone_id: taskForm.value.milestone_id || filterMid.value || undefined,
    })
    showTask.value = false
    taskForm.value.title = ''
    await reloadBoard()
    toast('ok', '任务已入列')
  } catch (e: unknown) {
    formErr.value = e instanceof Error ? e.message : '创建失败'
  }
}

async function createMs() {
  formErr.value = ''
  if (!msForm.value.title.trim()) { formErr.value = '标题必填'; return }
  if (msForm.value.due_date < msForm.value.start_date) { formErr.value = '截止日期不得早于开始日'; return }
  try {
    await api.createMilestone(projectId.value, msForm.value)
    showMs.value = false
    await reloadBoard()
    toast('ok', '里程碑已冻结基线')
  } catch (e: unknown) {
    formErr.value = e instanceof Error ? e.message : '创建失败'
  }
}

async function startFocus(t: Task) {
  try {
    const v = await api.startPomo(t.id)
    session.apply(v)
    session.flash('start')
    router.push('/focus')
  } catch (e: unknown) {
    toast('err', e instanceof Error ? e.message : '无法开始')
  }
}

function formatDue(s: string) {
  return (s || '').slice(0, 10)
}
</script>

<template>
  <div class="w-full min-h-[calc(100vh-64px)] flex flex-col lg:flex-row">
    <aside class="w-full lg:w-80 shrink-0 border-b lg:border-b-0 lg:border-r border-line p-4 space-y-3">
      <div class="flex items-center justify-between">
        <h2 class="font-display text-lg">里程碑</h2>
        <button class="text-xs text-acid" @click="showMs = true">+ 新建</button>
      </div>
      <select v-model="projectId" class="w-full bg-bg border border-line rounded-lg px-3 py-2 text-sm">
        <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
      </select>
      <div v-if="loading" class="space-y-2">
        <div v-for="i in 3" :key="i" class="h-16 rounded-xl bg-line/40 animate-pulse" />
      </div>
      <p v-else-if="!milestones.length" class="text-fog text-sm py-8">还没有里程碑。先冻结一条带起止日的基线。</p>
      <button v-for="m in milestones" :key="m.id" class="w-full text-left card p-3 hover:border-acid transition"
        :class="filterMid === m.id ? 'border-acid' : ''" @click="filterMid = filterMid === m.id ? '' : m.id">
        <div class="flex items-center gap-2">
          <span class="w-2 h-2 rounded-full" :class="riskColor(m.risk)" />
          <span class="font-medium text-sm">{{ m.title }}</span>
        </div>
        <p class="text-xs text-fog mt-1 font-mono">截止 {{ formatDue(m.due_date) }}</p>
        <div class="mt-2 h-1.5 bg-bg rounded-full overflow-hidden">
          <div class="h-full bg-acid" :style="{ width: `${Math.min(100, 100 - (m.remaining_points / Math.max(m.baseline_points, 1)) * 100)}%` }" />
        </div>
        <p class="text-xs text-fog mt-1">剩余 {{ m.remaining_points }} / 基线 {{ m.baseline_points }} 点</p>
      </button>
    </aside>

    <main class="flex-1 p-4 overflow-x-auto">
      <div class="flex items-center justify-between mb-4">
        <h1 class="font-display text-2xl">敏捷看板</h1>
        <button class="bg-acid text-bg text-sm font-semibold px-4 py-2 rounded-lg" @click="showTask = true">+ 任务</button>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-3 min-w-0">
        <section v-for="c in cols" :key="c.key" class="card min-h-[420px] p-3"
          @dragover.prevent @drop="onDrop(c.key)">
          <h3 class="font-mono text-xs text-fog mb-3 tracking-widest">{{ c.label }}</h3>
          <article v-for="t in byCol(c.key)" :key="t.id" draggable="true" @dragstart="onDragStart(t.id)"
            class="mb-2 rounded-xl border border-line bg-bg/60 p-3 cursor-grab hover:border-acid">
            <p class="text-sm font-medium">{{ t.title }}</p>
            <p class="font-mono text-xs text-acid mt-2">{{ t.consumed_pomodoros }}/{{ t.estimated_pomodoros }} 🍅</p>
            <button class="mt-2 text-xs text-cyan" @click="startFocus(t)">开始专注</button>
          </article>
          <p v-if="!byCol(c.key).length" class="text-fog text-xs pt-8 text-center">空列</p>
        </section>
      </div>
    </main>
  </div>

  <Modal v-if="showTask" title="新任务" @close="showTask = false">
    <div class="space-y-3 text-sm">
      <input v-model="taskForm.title" placeholder="标题 *" class="w-full bg-bg border border-line rounded-lg px-3 py-2" />
      <input v-model.number="taskForm.estimated_pomodoros" type="number" min="1" class="w-full bg-bg border border-line rounded-lg px-3 py-2" />
      <select v-model="taskForm.milestone_id" class="w-full bg-bg border border-line rounded-lg px-3 py-2">
        <option value="">不绑定里程碑</option>
        <option v-for="m in milestones" :key="m.id" :value="m.id">{{ m.title }}</option>
      </select>
      <p v-if="formErr" class="text-rose text-xs">{{ formErr }}</p>
      <div class="flex justify-end gap-2">
        <button class="px-3 py-2" @click="showTask = false">取消</button>
        <button class="bg-acid text-bg px-4 py-2 rounded-lg" @click="createTask">保存</button>
      </div>
    </div>
  </Modal>

  <Modal v-if="showMs" title="新里程碑" @close="showMs = false">
    <div class="space-y-3 text-sm">
      <input v-model="msForm.title" placeholder="标题 *" class="w-full bg-bg border border-line rounded-lg px-3 py-2" />
      <input v-model="msForm.start_date" type="date" class="w-full bg-bg border border-line rounded-lg px-3 py-2" />
      <input v-model="msForm.due_date" type="date" class="w-full bg-bg border border-line rounded-lg px-3 py-2" />
      <input v-model.number="msForm.baseline_points" type="number" min="0" class="w-full bg-bg border border-line rounded-lg px-3 py-2" />
      <p v-if="formErr" class="text-rose text-xs">{{ formErr }}</p>
      <div class="flex justify-end gap-2">
        <button @click="showMs = false">取消</button>
        <button class="bg-acid text-bg px-4 py-2 rounded-lg" @click="createMs">冻结基线</button>
      </div>
    </div>
  </Modal>

  <Modal v-if="confirmAbort" title="确认放弃？" @close="confirmAbort = false">
    <p class="text-sm text-fog mb-4">放弃的番茄钟不计入燃尽扣减，只会计入废弃率。</p>
    <div class="flex justify-end gap-2">
      <button @click="confirmAbort = false">继续专注</button>
    </div>
  </Modal>
</template>
