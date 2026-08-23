import { reactive } from 'vue'

export type ToastItem = { id: number; kind: 'ok' | 'err' | 'info'; text: string }

const state = reactive({ items: [] as ToastItem[] })
let seq = 1

export function useToasts() {
  return state
}

export function toast(kind: ToastItem['kind'], text: string) {
  const id = seq++
  state.items.push({ id, kind, text })
  window.setTimeout(() => dismiss(id), 5000)
}

export function dismiss(id: number) {
  state.items = state.items.filter((t) => t.id !== id)
}
