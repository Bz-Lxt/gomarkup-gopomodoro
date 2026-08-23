import { getToken } from './api'

export type WsHandler = (msg: { type: string; payload?: unknown }) => void

export function connectWS(onMsg: WsHandler): { close: () => void; send: (o: unknown) => void } {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  const url = `${proto}://${location.host}/ws?token=${encodeURIComponent(getToken())}`
  let ws: WebSocket | null = null
  let closed = false
  let timer: number | undefined

  const open = () => {
    ws = new WebSocket(url)
    ws.onmessage = (e) => {
      try { onMsg(JSON.parse(e.data)) } catch { /* ignore */ }
    }
    ws.onclose = () => {
      if (!closed) timer = window.setTimeout(open, 2000)
    }
  }
  open()
  return {
    send(o) { if (ws && ws.readyState === 1) ws.send(JSON.stringify(o)) },
    close() {
      closed = true
      if (timer) window.clearTimeout(timer)
      ws?.close()
    },
  }
}
