export const NAV_ITEMS = ['overview','chat','control','files','tasks','bbs','pets','memory','channels','autonomous','goals','models','settings','logs']
export const ROUTE_TABS = NAV_ITEMS
export const TASK_SUB_TABS = ['services','scheduled','runs','reports']

const TAB_ALIASES = {
  '': 'overview',
  home: 'overview',
  index: 'overview',
  task: 'tasks',
  tasks: 'tasks',
  tmwebdriver: 'control',
  tmwd: 'control',
  config: 'settings',
  pet: 'pets',
  pets: 'pets',
  desktop_pet: 'pets',
  hatch_pet: 'pets',
}

const TASK_ROUTE_ALIASES = {
  '': 'services',
  service: 'services',
  services: 'services',
  schedule: 'scheduled',
  scheduled: 'scheduled',
  runs: 'runs',
  reports: 'reports',
}

const baseURL = () => (import.meta.env?.BASE_URL || '/').replace(/\/$/, '')

const routePartsFrom = (pathname = '/', hash = '') => {
  const rawHash = (hash || '').replace(/^#\/?/, '').split('/').filter(Boolean)
  if (rawHash.length) return rawHash
  const base = baseURL()
  let path = pathname || '/'
  if (base && base !== '/' && path.startsWith(base)) path = path.slice(base.length) || '/'
  return path.replace(/^\/+|\/+$/g, '').split('/').filter(Boolean)
}

const routeParts = () => routePartsFrom(window.location.pathname, window.location.hash)

const parseRouteFromParts = (parts) => {
  const rawFirst = parts[0] || ''
  const directTaskSubTab = TASK_ROUTE_ALIASES[rawFirst]
  const first = directTaskSubTab && rawFirst !== '' ? 'tasks' : (TAB_ALIASES[rawFirst] || rawFirst)
  const tab = ROUTE_TABS.includes(first) ? first : 'overview'
  const rawSub = tab === 'tasks' ? (parts[1] || (directTaskSubTab ? rawFirst : '')) : ''
  const sub = TASK_ROUTE_ALIASES[rawSub] || rawSub
  const taskSubTab = tab === 'tasks' && TASK_SUB_TABS.includes(sub) ? sub : 'services'
  return { tab, taskSubTab }
}

const parseRouteFromPath = (pathname = '/', hash = '') => parseRouteFromParts(routePartsFrom(pathname, hash))

const isKnownRouteParts = (parts) => {
  const rawFirst = parts[0] || ''
  if (!rawFirst) return true
  if (TASK_ROUTE_ALIASES[rawFirst]) return true
  const first = TAB_ALIASES[rawFirst] || rawFirst
  if (!ROUTE_TABS.includes(first)) return false
  if (first !== 'tasks') return true
  const rawSub = parts[1] || ''
  return !rawSub || Boolean(TASK_ROUTE_ALIASES[rawSub]) || TASK_SUB_TABS.includes(rawSub)
}

const isKnownRoutePath = (pathname = '/', hash = '') => isKnownRouteParts(routePartsFrom(pathname, hash))

export const parseRoute = () => parseRouteFromParts(routeParts())

export const buildRoute = (tab, taskSubTab = 'services') => {
  const safeTab = ROUTE_TABS.includes(tab) ? tab : 'overview'
  const suffix = safeTab === 'tasks' ? `/${TASK_SUB_TABS.includes(taskSubTab) ? taskSubTab : 'services'}` : ''
  const base = baseURL()
  return `${base}/${safeTab}${suffix}`.replace(/\/+/g, '/')
}

export const currentRoute = () => `${window.location.pathname}${window.location.search}${window.location.hash}`

export const normalizeAdminReturnRoute = (raw, fallback = buildRoute('overview')) => {
  if (!raw) return fallback
  const text = String(raw).trim()
  if (!text || /^[a-z][a-z0-9+.-]*:/i.test(text) || text.startsWith('//')) return fallback
  try {
    const url = new URL(text, window.location.origin)
    if (url.origin !== window.location.origin) return fallback
    const candidate = `${url.pathname}${url.search}${url.hash}`
    if (!isKnownRoutePath(url.pathname, url.hash)) return fallback
    const route = parseRouteFromPath(url.pathname, url.hash)
    if (route.tab === 'chat') return fallback
    return candidate || fallback
  } catch (_) {
    return fallback
  }
}

export const buildChatRoute = (from = currentRoute()) => {
  const base = baseURL()
  const safeFrom = normalizeAdminReturnRoute(from, buildRoute('overview'))
  return `${base}/chat?from=${encodeURIComponent(safeFrom)}`.replace(/\/+/g, '/')
}

export const chatReturnRoute = (fallback = buildRoute('overview')) => {
  try {
    const from = new URLSearchParams(window.location.search || '').get('from')
    const remembered = window.sessionStorage?.getItem('ga-admin-chat-return') || ''
    const next = normalizeAdminReturnRoute(from || remembered, fallback)
    window.sessionStorage?.setItem('ga-admin-chat-return', next)
    return next
  } catch (_) {
    return fallback
  }
}
