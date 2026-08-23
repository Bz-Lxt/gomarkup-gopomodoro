type Kind = 'rain' | 'cafe' | 'white'

let ctx: AudioContext | null = null
let node: AudioBufferSourceNode | null = null
let gain: GainNode | null = null
let current: Kind | null = null

function ensure() {
  if (!ctx) ctx = new AudioContext()
  if (!gain) {
    gain = ctx.createGain()
    gain.gain.value = 0.18
    gain.connect(ctx.destination)
  }
  return ctx
}

function buffer(kind: Kind, ac: AudioContext): AudioBuffer {
  const len = ac.sampleRate * 3
  const buf = ac.createBuffer(1, len, ac.sampleRate)
  const d = buf.getChannelData(0)
  for (let i = 0; i < len; i++) {
    const n = Math.random() * 2 - 1
    if (kind === 'white') d[i] = n * 0.35
    else if (kind === 'rain') d[i] = (n * 0.22 + (i % 180 === 0 ? n * 0.6 : 0))
    else d[i] = n * 0.12 + Math.sin(i / 90) * 0.04 + (i % 1400 < 40 ? n * 0.25 : 0)
  }
  return buf
}

export function playNoise(kind: Kind) {
  const ac = ensure()
  stopNoise()
  const src = ac.createBufferSource()
  src.buffer = buffer(kind, ac)
  src.loop = true
  src.connect(gain!)
  src.start()
  node = src
  current = kind
  void ac.resume()
}

export function stopNoise() {
  if (node) {
    try { node.stop() } catch { /* ignore */ }
    node.disconnect()
    node = null
  }
  current = null
}

export function setVolume(v: number) {
  if (gain) gain.gain.value = Math.max(0, Math.min(1, v))
}

export function currentNoise() {
  return current
}
