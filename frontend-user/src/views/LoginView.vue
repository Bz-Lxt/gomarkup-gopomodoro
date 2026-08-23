<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, setToken } from '../lib/api'
import { toast } from '../lib/toast'

const router = useRouter()
const mode = ref<'login' | 'register'>('login')
const form = reactive({ email: 'geek@gopomodoro.dev', password: 'pomodoro123', display_name: '极客番茄' })
const errors = reactive<Record<string, string>>({})
const loading = ref(false)

function validate() {
  Object.keys(errors).forEach((k) => delete errors[k])
  if (!/^\S+@\S+\.\S+$/.test(form.email)) errors.email = '请输入有效邮箱'
  if (form.password.length < 8) errors.password = '密码至少 8 位'
  if (mode.value === 'register' && !form.display_name.trim()) errors.display_name = '必填'
  return Object.keys(errors).length === 0
}

async function submit() {
  if (!validate()) {
    toast('err', '请先修正表单错误')
    return
  }
  loading.value = true
  try {
    const data = mode.value === 'login'
      ? await api.login(form.email, form.password)
      : await api.register(form.email, form.password, form.display_name)
    setToken(data.auth.token)
    toast('ok', '已进入座舱')
    router.push('/')
  } catch (e: unknown) {
    toast('err', e instanceof Error ? e.message : '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen w-full flex items-center justify-center px-4">
    <div class="card w-full max-w-md p-8">
      <p class="font-mono text-xs text-acid mb-2">SESSION // AUTH</p>
      <h1 class="font-display text-3xl mb-1">量化时间座舱</h1>
      <p class="text-fog text-sm mb-8">用真实番茄钟驱动里程碑燃尽线。</p>
      <form class="space-y-4" @submit.prevent="submit">
        <label class="block text-sm">
          邮箱 *
          <input v-model="form.email" class="mt-1 w-full bg-bg border border-line rounded-lg px-3 py-2" />
          <p v-if="errors.email" class="text-rose text-xs mt-1">{{ errors.email }}</p>
        </label>
        <label class="block text-sm">
          密码 *
          <input v-model="form.password" type="password" class="mt-1 w-full bg-bg border border-line rounded-lg px-3 py-2" />
          <p v-if="errors.password" class="text-rose text-xs mt-1">{{ errors.password }}</p>
        </label>
        <label v-if="mode === 'register'" class="block text-sm">
          显示名 *
          <input v-model="form.display_name" class="mt-1 w-full bg-bg border border-line rounded-lg px-3 py-2" />
          <p v-if="errors.display_name" class="text-rose text-xs mt-1">{{ errors.display_name }}</p>
        </label>
        <button class="w-full bg-acid text-bg font-semibold py-2.5 rounded-lg" :disabled="loading">
          {{ loading ? '同步中…' : (mode === 'login' ? '进入看板' : '创建账号') }}
        </button>
      </form>
      <button class="mt-4 text-sm text-cyan" @click="mode = mode === 'login' ? 'register' : 'login'">
        {{ mode === 'login' ? '没有账号？注册' : '已有账号？登录' }}
      </button>
    </div>
  </div>
</template>
