import { defineStore } from 'pinia'
import { api, type SessionView } from '../lib/api'

export const useSession = defineStore('session', {
  state: () => ({
    view: null as SessionView | null,
    fx: '' as '' | 'start' | 'pause' | 'complete' | 'abort',
    grace: 0,
  }),
  actions: {
    apply(v: SessionView | null) {
      this.view = v && v.session ? v : null
    },
    async refresh() {
      const v = await api.activePomo()
      this.apply(v as SessionView)
    },
    flash(kind: typeof this.fx) {
      this.fx = kind
      window.setTimeout(() => { this.fx = '' }, 900)
    },
  },
})
