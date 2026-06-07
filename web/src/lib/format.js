export const emptyProfile = (idx = 0) => ({ var_name: `MODEL_${idx + 1}`, type: 'openai', name: '', model: '', apibase: '', apikey: '', stream: true, max_retries: 3, read_timeout: 300, reasoning_effort: '', enabled: true })
export const safeJson = (v) => JSON.stringify(v ?? {}, null, 2)
export const group = (items, pred) => (items || []).filter(pred)

export const formatDuration = (seconds) => {
  const n = Math.max(0, Number(seconds) || 0)
  const h = Math.floor(n / 3600), m = Math.floor((n % 3600) / 60), s = Math.floor(n % 60)
  return h ? `${h}h ${m}m` : (m ? `${m}m ${s}s` : `${s}s`)
}

export const formatGoalTime = (value) => value ? new Date(value).toLocaleString() : '-'

export const formatBytes = (bytes) => {
  const n = Math.max(0, Number(bytes) || 0)
  if (n >= 1024 * 1024) return `${(n / (1024 * 1024)).toFixed(n >= 10 * 1024 * 1024 ? 0 : 1)} MiB`
  if (n >= 1024) return `${(n / 1024).toFixed(n >= 10 * 1024 ? 0 : 1)} KiB`
  return `${n} B`
}

export const outputLineCount = (text) => {
  if (!text) return 0
  const normalized = String(text).replace(/\r?\n$/, '')
  return normalized ? normalized.split(/\r?\n/).length : 0
}

const fallbackCopyText = (value) => {
  if (typeof document === 'undefined' || !document.execCommand) throw new Error('Clipboard copy is not supported')
  const el = document.createElement('textarea')
  const selection = document.getSelection?.()
  const selectedRange = selection?.rangeCount ? selection.getRangeAt(0) : null
  el.value = value
  el.setAttribute('readonly', '')
  el.style.position = 'fixed'
  el.style.top = '0'
  el.style.left = '-9999px'
  el.style.opacity = '0'
  document.body.appendChild(el)
  try {
    el.focus({ preventScroll: true })
    el.select()
    const copied = document.execCommand('copy')
    if (!copied) throw new Error('Clipboard copy failed')
  } finally {
    document.body.removeChild(el)
    if (selection && selectedRange) {
      selection.removeAllRanges()
      selection.addRange(selectedRange)
    }
  }
}

export const copyText = async (text) => {
  const value = text || ''
  if (!value) return false
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(value)
      return true
    } catch {
      // 非安全上下文/权限拒绝时继续尝试传统同步复制，避免按钮静默失效。
    }
  }
  fallbackCopyText(value)
  return true
}

const clampPercent = (value) => Math.max(0, Math.min(100, Number.isFinite(value) ? value : 0))

export const goalTurnPercent = (g) => {
  const serverPct = Number(g?.turn_percent)
  if (Number.isFinite(serverPct)) return clampPercent(serverPct)
  const maxTurns = Number(g?.max_turns || 0)
  if (!maxTurns) return 0
  return clampPercent((Number(g?.turns_used || 0) / maxTurns) * 100)
}

export const goalBudgetPercent = (g) => {
  const serverPct = Number(g?.budget_percent)
  if (Number.isFinite(serverPct)) return clampPercent(serverPct)
  const elapsed = Number(g?.elapsed_seconds || 0)
  const remaining = Number(g?.remaining_seconds || 0)
  const total = elapsed + Math.max(0, remaining)
  if (!total) return 0
  return clampPercent((elapsed / total) * 100)
}
