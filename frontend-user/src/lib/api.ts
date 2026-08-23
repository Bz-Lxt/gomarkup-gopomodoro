export type Envelope<T> = { ok: boolean; data?: T; error?: { code: string; message: string; details?: Record<string, unknown> } }

const TOKEN_KEY = 'gopomo.token'

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setToken(t: string) {
  localStorage.setItem(TOKEN_KEY, t)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

export class ApiError extends Error {
  code: string
  constructor(code: string, message: string) {
    super(message)
    this.code = code
  }
}

export async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  headers.set('Content-Type', 'application/json')
  const tok = getToken()
  if (tok) headers.set('Authorization', `Bearer ${tok}`)
  const res = await fetch(path, { ...init, headers })
  const body = (await res.json().catch(() => ({}))) as Envelope<T>
  if (!res.ok || !body.ok) {
    throw new ApiError(body.error?.code || 'E_INTERNAL', body.error?.message || '请求失败')
  }
  return body.data as T
}

export const api = {
  login: (email: string, password: string) =>
    request<{ user: User; auth: { token: string; expires_at: string } }>('/api/v1/auth/login', {
      method: 'POST', body: JSON.stringify({ email, password }),
    }),
  register: (email: string, password: string, display_name: string) =>
    request<{ user: User; auth: { token: string } }>('/api/v1/auth/register', {
      method: 'POST', body: JSON.stringify({ email, password, display_name }),
    }),
  me: () => request<User>('/api/v1/me'),
  projects: () => request<Project[]>('/api/v1/projects'),
  createProject: (name: string) =>
    request<Project>('/api/v1/projects', { method: 'POST', body: JSON.stringify({ name }) }),
  milestones: (pid: string) => request<Milestone[]>(`/api/v1/projects/${pid}/milestones`),
  createMilestone: (pid: string, body: Partial<Milestone> & { start_date: string; due_date: string; title: string; baseline_points: number }) =>
    request<Milestone>(`/api/v1/projects/${pid}/milestones`, { method: 'POST', body: JSON.stringify(body) }),
  tasks: (pid: string, mid?: string) =>
    request<Task[]>(`/api/v1/projects/${pid}/tasks${mid ? `?milestone_id=${mid}` : ''}`),
  createTask: (pid: string, body: Record<string, unknown>) =>
    request<Task>(`/api/v1/projects/${pid}/tasks`, { method: 'POST', body: JSON.stringify(body) }),
  patchTask: (id: string, body: Record<string, unknown>) =>
    request<Task>(`/api/v1/tasks/${id}`, { method: 'PATCH', body: JSON.stringify(body) }),
  reorder: (items: { id: string; kanban_column: string; sort_order: number }[]) =>
    request('/api/v1/tasks/reorder', { method: 'POST', body: JSON.stringify({ items }) }),
  startPomo: (task_id: string) =>
    request<SessionView>('/api/v1/pomodoros', { method: 'POST', body: JSON.stringify({ task_id }) }),
  activePomo: () => request<SessionView | { session: null }>('/api/v1/pomodoros/active'),
  pausePomo: (id: string) => request<SessionView>(`/api/v1/pomodoros/${id}/pause`, { method: 'POST' }),
  resumePomo: (id: string) => request<SessionView>(`/api/v1/pomodoros/${id}/resume`, { method: 'POST' }),
  abortPomo: (id: string) => request<SessionView>(`/api/v1/pomodoros/${id}/abort`, { method: 'POST', body: JSON.stringify({ reason: 'user' }) }),
  burndown: (id: string, gran = 'day') => request<BurndownChart>(`/api/v1/milestones/${id}/burndown?granularity=${gran}`),
  metrics: (id: string) => request<Metrics>(`/api/v1/milestones/${id}/metrics`),
}

export type User = { id: string; email: string; display_name: string }
export type Project = { id: string; name: string; description: string }
export type Milestone = {
  id: string; project_id: string; title: string; start_date: string; due_date: string
  baseline_points: number; remaining_points: number; risk: string; status: string
}
export type Task = {
  id: string; project_id: string; milestone_id?: string | null; title: string
  estimated_pomodoros: number; consumed_pomodoros: number; kanban_column: string; sort_order: number
}
export type SessionView = {
  session: {
    id: string; task_id: string; state: string; resume_token: string; focus_duration_ms: number
  } | null
  remaining_ms: number
  grace_left_s: number
}
export type BurndownChart = {
  remaining_points: number
  baseline_points: number
  ideal: { x: string[]; y: number[] }
  actual: { x: string[]; y: number[] }
  scope_marks: { at: string; remaining: number }[]
}
export type Metrics = {
  today_completed: number; week_completed: number; abort_rate: number
  avg_daily_velocity: number; predicted_done_on?: string
}
