import { useEffect, useMemo, useState } from 'react'
import { Activity, Bot, Brain, CalendarClock, CheckCircle2, Eye, FileCode2, FolderCog, GitBranch, Globe2, Play, RefreshCw, Save, Server, ShieldAlert, SlidersHorizontal, Square, Terminal, UploadCloud, XCircle } from 'lucide-react'

const api = async (url, options = {}) => {
  const res = await fetch(url, { headers: { 'Content-Type': 'application/json' }, ...options })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.detail || `${res.status} ${res.statusText}`)
  }
  return res.json()
}

const I18N = {
  zh: {
    appName: 'GA Admin', tagline: 'GenericAgent 生命周期控制面', root: 'GenericAgent 根目录', save: '保存', refresh: '刷新', busy: '执行中', ready: '就绪', error: '错误', unknown: '未知', empty: '暂无', missing: '缺失', bytes: '字节', enabled: '启用', disabled: '停用', start: '启动', stop: '停止', running: '运行中', stopped: '已停止', language: '语言', show: '显示', hide: '隐藏',
    nav: { overview: '总览', tasks: '任务', memory: '记忆', channels: '通道', autonomous: '自主进化', schedule: '定时', models: '模型', logs: '日志' },
    desc: { overview: '从 GA 的功能域理解并接管生命周期，而不是只做进程启动器。', tasks: '普通会话、任务文件、批处理入口与任务型服务。', memory: '分层记忆、SOP 与工具能力索引。', channels: '桌面、TUI、Web、IM Bot 等前端入口。', autonomous: '反思、自主运行、Goal Mode 与团队 Worker。', schedule: 'sche_tasks JSON 定时任务与完成报告。', models: '读取/预览/写回 GA mykey.py 模型配置。', logs: '进程状态与输出日志。' },
    cards: { processes: '进程', running: '运行中', stopped: '已停止', memoryLayers: '记忆层', sopTools: 'SOP/工具', schedule: '定时任务', enabledTasks: '已启用', reports: '报告', coreFiles: '核心文件', channels: '通道文件', reflect: '反思脚本', health: 'GA 健康' },
    lists: { serviceGroups: '服务域', coreFiles: '核心文件', reflect: 'Reflect / Autonomous', frontends: '前端 / 通道', memory: '记忆层级', sop: 'SOP 与工具', taskServices: '任务服务', frontendServices: '前端服务', reflectServices: '反思服务', discoveredChannels: '发现的通道文件', reflectScripts: '反思脚本', scheduledTasks: '定时任务', recentReports: '最近报告', processes: '进程', generatedPreview: '生成预览', riskHints: '接管提示' },
    hints: { rootSaved: 'GA 根目录已保存', taskToggled: '任务状态已更新', modelsSaved: 'mykey.py 已备份并写回', savedSecret: '已保存；输入新值可替换', secret: 'API Key / Token', noFrontend: '未发现前端服务', noReflect: '未发现 reflect 服务', noTasks: '暂无 sche_tasks/*.json', noLogs: '暂无日志', previewHelp: '点击“预览”查看配置；点击“写回 mykey.py”会先备份再覆盖 GA 的 mykey.py。', modelSource: '来源', secretHidden: '已隐藏真实密钥', addProfile: '新增 Profile', preview: '预览', writeMykey: '写回 mykey.py', rootHint: '此 Admin 会按 GA 根目录读取 memory、sche_tasks、frontends、reflect 与 mykey.py。', scheduleHint: '启停定时任务会直接修改 sche_tasks JSON 的 enabled 字段。', modelHint: '模型写回前会调用后端备份 mykey.py；真实密钥不会在 UI 回显。' },
    fields: { varName: '变量名', type: '类型', name: '名称', model: '模型', apiBase: 'API Base', apiKey: 'API Key', stream: '流式', maxRetries: '最大重试', readTimeout: '读取超时', reasoning: '推理强度' },
  },
  en: {
    appName: 'GA Admin', tagline: 'Lifecycle control plane for GenericAgent', root: 'GenericAgent Root', save: 'Save', refresh: 'Refresh', busy: 'Busy', ready: 'Ready', error: 'Error', unknown: 'Unknown', empty: 'Empty', missing: 'Missing', bytes: 'bytes', enabled: 'Enable', disabled: 'Disable', start: 'Start', stop: 'Stop', running: 'Running', stopped: 'Stopped', language: 'Language', show: 'Show', hide: 'Hide',
    nav: { overview: 'Overview', tasks: 'Tasks', memory: 'Memory', channels: 'Channels', autonomous: 'Autonomous', schedule: 'Schedule', models: 'Models', logs: 'Logs' },
    desc: { overview: 'Manage GA by native capability domains, not just as a process launcher.', tasks: 'Normal sessions, task files, batch entries and task-oriented services.', memory: 'Layered memory, SOPs and tool capability index.', channels: 'Desktop, TUI, Web and IM bot frontends.', autonomous: 'Reflection, autonomous runs, Goal Mode and team workers.', schedule: 'sche_tasks JSON schedules and completion reports.', models: 'Read, preview and write back GA mykey.py model profiles.', logs: 'Process status and output logs.' },
    cards: { processes: 'Processes', running: 'Running', stopped: 'Stopped', memoryLayers: 'Memory Layers', sopTools: 'SOP/Tools', schedule: 'Schedules', enabledTasks: 'Enabled', reports: 'Reports', coreFiles: 'Core Files', channels: 'Channels', reflect: 'Reflect Scripts', health: 'GA Health' },
    lists: { serviceGroups: 'Service Domains', coreFiles: 'Core Files', reflect: 'Reflect / Autonomous', frontends: 'Frontends / Channels', memory: 'Memory Layers', sop: 'SOPs & Tools', taskServices: 'Task Services', frontendServices: 'Frontend Services', reflectServices: 'Reflect Services', discoveredChannels: 'Discovered Channel Files', reflectScripts: 'Reflect Scripts', scheduledTasks: 'Scheduled Tasks', recentReports: 'Recent Reports', processes: 'Processes', generatedPreview: 'Generated Preview', riskHints: 'Lifecycle Notes' },
    hints: { rootSaved: 'GA root saved', taskToggled: 'Task status updated', modelsSaved: 'mykey.py backed up and written', savedSecret: 'Saved; enter a new value to replace', secret: 'API Key / Token', noFrontend: 'No frontend service found', noReflect: 'No reflect service found', noTasks: 'No sche_tasks/*.json found', noLogs: 'No logs', previewHelp: 'Click Preview to inspect generated config. Write Back will backup and overwrite GA mykey.py.', modelSource: 'Source', secretHidden: 'real secrets are hidden', addProfile: 'Add Profile', preview: 'Preview', writeMykey: 'Write mykey.py', rootHint: 'Admin reads memory, sche_tasks, frontends, reflect and mykey.py from the configured GA root.', scheduleHint: 'Toggling schedules directly updates the enabled field in sche_tasks JSON.', modelHint: 'The backend backs up mykey.py before writing; real secrets are never echoed to UI.' },
    fields: { varName: 'Variable', type: 'Type', name: 'Name', model: 'Model', apiBase: 'API Base', apiKey: 'API Key', stream: 'Stream', maxRetries: 'Max Retries', readTimeout: 'Read Timeout', reasoning: 'Reasoning' },
  },
}

const emptyProfile = (i = 0) => ({ enabled: true, var_name: `native_oai_config${i}`, type: 'native_oai', name: 'new-model', apibase: '', model: '', apikey: '', fake_cc_system_prompt: null, thinking_type: '', reasoning_effort: '', stream: true, max_retries: 3, read_timeout: 300, connect_timeout: null, user_agent: '', api_mode: '', extra: {} })

function SecretInput({ value, onChange, t }) {
  const [show, setShow] = useState(false)
  const display = value === '***SET***' ? '' : (value || '')
  return <div className="secret-row"><input type={show ? 'text' : 'password'} value={display} onChange={(e) => onChange(e.target.value)} placeholder={value === '***SET***' ? t.hints.savedSecret : t.hints.secret} /><button type="button" className="ghost" onClick={() => setShow(!show)}><Eye size={14}/>{show ? t.hide : t.show}</button></div>
}

function Card({ title, value, sub, icon: Icon, tone = '' }) {
  return <div className={`stat ${tone}`}><div><span>{title}</span><strong>{value}</strong><small>{sub}</small></div>{Icon && <Icon size={28}/>}</div>
}

function MiniList({ title, items = [], empty, render }) {
  return <div className="panel"><div className="panel-title">{title}</div><div className="mini-list">{items.length ? items.map((it, idx) => <div className="mini-row" key={`${it.path || it.name || it.id || idx}`}>{render ? render(it, idx) : <><b>{it.name}</b><span>{it.path}</span></>}</div>) : <div className="empty">{empty}</div>}</div></div>
}

export default function App() {
  const [lang, setLang] = useState(() => { const saved = localStorage.getItem('ga-admin-lang'); return saved || ((navigator.language || '').toLowerCase().startsWith('zh') ? 'zh' : 'en') })
  const t = I18N[lang] || I18N.zh
  const [tab, setTab] = useState('overview')
  const [config, setConfig] = useState(null)
  const [gaRoot, setGaRoot] = useState('')
  const [services, setServices] = useState([])
  const [selected, setSelected] = useState('')
  const [logs, setLogs] = useState([])
  const [summary, setSummary] = useState({ total: 0, running: 0, stopped: 0 })
  const [inventory, setInventory] = useState(null)
  const [gaHealth, setGaHealth] = useState(null)
  const [schedule, setSchedule] = useState(null)
  const [profiles, setProfiles] = useState([])
  const [modelSource, setModelSource] = useState(null)
  const [preview, setPreview] = useState('')
  const [busy, setBusy] = useState(false)
  const [notice, setNotice] = useState('')
  const [err, setErr] = useState('')

  const run = async (fn, okMsg = '') => { setBusy(true); setErr(''); try { await fn(); if (okMsg) setNotice(okMsg) } catch (e) { setErr(e.message || String(e)) } finally { setBusy(false) } }
  const loadConfig = async () => { const data = await api('/api/config'); setConfig(data); setGaRoot(data.ga_root || '') }
  const refreshServices = async () => { const data = await api('/api/services'); setServices(data.services || []); setSummary(data.summary || {}) }
  const loadLogs = async (name = selected) => { if (!name) return; const data = await api(`/api/services/logs?name=${encodeURIComponent(name)}`); setLogs(data.lines || []) }
  const refreshGA = async () => { const [inv, health, tasks] = await Promise.all([api('/api/ga/inventory'), api('/api/ga/health').catch(() => null), api('/api/schedule/tasks').catch(() => null)]); setInventory(inv); setGaHealth(health); setSchedule(tasks) }
  const loadModels = async () => { const data = await api('/api/models/raw'); setProfiles(data.profiles || []); setModelSource(data.source || null); setPreview('') }
  const refresh = async () => { await Promise.all([loadConfig(), refreshServices(), refreshGA(), loadModels()]) }

  useEffect(() => { refresh() }, [])
  useEffect(() => { localStorage.setItem('ga-admin-lang', lang) }, [lang])
  useEffect(() => { if (!selected && services.length) setSelected(services[0].name) }, [services, selected])

  const svcByKind = useMemo(() => services.reduce((acc, s) => { const k = s.kind || 'other'; (acc[k] ||= []).push(s); return acc }, {}), [services])
  const saveRoot = () => run(async () => { const data = await api('/api/config', { method: 'PUT', body: JSON.stringify({ ga_root: gaRoot }) }); setConfig(data); await refresh() }, t.hints.rootSaved)
  const start = (name) => run(async () => { await api('/api/services/start', { method: 'POST', body: JSON.stringify({ name }) }); await refreshServices(); await loadLogs(name) }, `${name} ${t.running}`)
  const stop = (name) => run(async () => { await api('/api/services/stop', { method: 'POST', body: JSON.stringify({ name }) }); await refreshServices(); await loadLogs(name) }, `${name} ${t.stopped}`)
  const toggleTask = (task) => run(async () => { await api('/api/schedule/toggle', { method: 'POST', body: JSON.stringify({ id: task.id, enabled: !task.enabled }) }); await refreshGA() }, t.hints.taskToggled)
  const patchProfile = (idx, patch) => setProfiles((arr) => arr.map((p, i) => i === idx ? { ...p, ...patch } : p))
  const saveModels = () => run(async () => { await api('/api/models/raw/save', { method: 'POST', body: JSON.stringify({ profiles }) }); await loadModels() }, t.hints.modelsSaved)
  const previewModels = () => run(async () => { const data = await api('/api/models/raw/preview', { method: 'POST', body: JSON.stringify({ profiles }) }); setPreview(data.preview || '') })

  const nav = [['overview', Activity], ['tasks', Bot], ['memory', Brain], ['channels', Terminal], ['autonomous', GitBranch], ['schedule', CalendarClock], ['models', SlidersHorizontal], ['logs', FileCode2]]
  const ServiceRow = ({ svc }) => <div className={`service-card ${svc.running ? 'running' : ''}`} onClick={() => setSelected(svc.name)}><div><b>{svc.name}</b><span>{svc.command?.join(' ')}</span></div><em>{svc.running ? t.running.toUpperCase() : t.stopped.toUpperCase()}</em><div className="svc-actions">{svc.running ? <button onClick={(e) => { e.stopPropagation(); stop(svc.name) }}><Square size={14}/>{t.stop}</button> : <button onClick={(e) => { e.stopPropagation(); start(svc.name) }}><Play size={14}/>{t.start}</button>}</div></div>
  const sectionTitle = <header><div><h2>{t.nav[tab]}</h2><p>{t.desc[tab]}</p></div><div className="badges"><span>{config?.ga_root || t.unknown}</span><span className={busy ? 'err' : 'ok'}>{busy ? t.busy : t.ready}</span>{err && <span className="err">{err}</span>}{notice && <span className="ok">{notice}</span>}</div></header>

  return <div className="app">
    <aside className="sidebar">
      <div className="brand"><Server size={34}/><div><h1>{t.appName}</h1><p>{t.tagline}</p></div></div>
      <div className="lang-switch"><Globe2 size={15}/><span>{t.language}</span><select value={lang} onChange={(e) => setLang(e.target.value)}><option value="zh">中文</option><option value="en">English</option></select></div>
      <div className="root-box"><label>{t.root}</label><div><input value={gaRoot} onChange={(e) => setGaRoot(e.target.value)} /><button onClick={saveRoot} disabled={busy}><Save size={14}/>{t.save}</button></div></div>
      <nav>{nav.map(([id, Icon]) => <button key={id} className={tab === id ? 'active' : ''} onClick={() => setTab(id)}><Icon size={18}/>{t.nav[id]}</button>)}</nav>
      <button className="refresh" onClick={() => run(refresh)} disabled={busy}><RefreshCw size={16}/>{t.refresh}</button>
    </aside>

    <main className="main">
      {sectionTitle}
      {tab === 'overview' && <section><div className="stats"><Card title={t.cards.processes} value={summary.total || 0} sub={`${summary.running || 0} ${t.running} · ${summary.stopped || 0} ${t.stopped}`} icon={Server}/><Card title={t.cards.memoryLayers} value={inventory?.memory_layers?.length || 0} sub={t.cards.sopTools} icon={Brain}/><Card title={t.cards.schedule} value={schedule?.summary?.total || 0} sub={`${schedule?.summary?.enabled || 0} ${t.cards.enabledTasks}`} icon={CalendarClock}/><Card title={t.cards.health} value={gaHealth?.ok ? 'OK' : '?'} sub={gaHealth?.detail || 'agentmain'} icon={gaHealth?.ok ? CheckCircle2 : XCircle} tone={gaHealth?.ok ? 'good' : 'warn'}/></div><div className="grid2"><MiniList title={t.lists.serviceGroups} empty={t.empty} items={Object.entries(svcByKind).map(([name, arr]) => ({ name, path: `${arr.length} ${t.cards.processes}` }))}/><MiniList title={t.lists.coreFiles} empty={t.empty} items={inventory?.core_files || []} render={(f) => <><b>{f.exists ? '✓' : '×'} {f.path}</b><span>{f.exists ? `${f.size || 0} ${t.bytes}` : t.missing}</span></>}/><MiniList title={t.lists.riskHints} empty={t.empty} items={[{name:t.hints.rootHint,path:''},{name:t.hints.scheduleHint,path:''},{name:t.hints.modelHint,path:''}]} render={(h)=><><b><ShieldAlert size={14}/> {h.name}</b></>}/></div></section>}
      {tab === 'tasks' && <section><div className="grid2"><div className="panel"><div className="panel-title">{t.lists.taskServices}</div>{(svcByKind.task || []).map(s => <ServiceRow key={s.name} svc={s}/>)}{!(svcByKind.task || []).length && <div className="empty">{t.empty}</div>}</div><MiniList title={t.lists.coreFiles} empty={t.empty} items={inventory?.core_files || []} render={(f) => <><b>{f.exists ? '✓' : '×'} {f.path}</b><span>{f.exists ? `${f.size || 0} ${t.bytes}` : t.missing}</span></>} /></div></section>}
      {tab === 'memory' && <section><div className="grid2"><MiniList title={t.lists.memory} empty={t.empty} items={inventory?.memory_layers || []}/><MiniList title={t.lists.sop} empty={t.empty} items={inventory?.sops || []}/></div></section>}
      {tab === 'channels' && <section><div className="grid2"><div className="panel"><div className="panel-title">{t.lists.frontendServices}</div>{(svcByKind.frontend || []).map(s => <ServiceRow key={s.name} svc={s}/>)}{!(svcByKind.frontend || []).length && <div className="empty">{t.hints.noFrontend}</div>}</div><MiniList title={t.lists.discoveredChannels} empty={t.empty} items={inventory?.frontends || []} render={(f) => <><b>{f.name}</b><span>{f.kind} · {f.path}</span></>}/></div></section>}
      {tab === 'autonomous' && <section><div className="grid2"><div className="panel"><div className="panel-title">{t.lists.reflectServices}</div>{(svcByKind.reflect || []).map(s => <ServiceRow key={s.name} svc={s}/>)}{!(svcByKind.reflect || []).length && <div className="empty">{t.hints.noReflect}</div>}</div><MiniList title={t.lists.reflectScripts} empty={t.empty} items={inventory?.reflect || []}/></div></section>}
      {tab === 'schedule' && <section><div className="stats"><Card title="Total" value={schedule?.summary?.total || 0} sub="schedule json" icon={CalendarClock}/><Card title={t.cards.enabledTasks} value={schedule?.summary?.enabled || 0} sub="active tasks" icon={CheckCircle2}/><Card title={t.cards.reports} value={schedule?.reports?.length || 0} sub="sche_tasks/done" icon={FileCode2}/></div><div className="panel"><div className="panel-title">{t.lists.scheduledTasks}</div>{(schedule?.tasks || []).map(task => <div className="task-row" key={task.id}><div><b>{task.id}</b><span>{task.schedule || '-'} · {task.repeat || '-'} · max delay {task.max_delay_hours || 6}h</span><p>{task.prompt}</p></div><button onClick={() => toggleTask(task)}>{task.enabled ? t.disabled : t.enabled}</button></div>)}{!(schedule?.tasks || []).length && <div className="empty">{t.hints.noTasks}</div>}</div><MiniList title={t.lists.recentReports} empty={t.empty} items={schedule?.reports || []}/></section>}
      {tab === 'models' && <section><div className="model-top"><div><h3>{t.nav.models}</h3><p>{t.hints.modelSource}: {modelSource?.path || 'mykey.py'} · {t.hints.secretHidden}</p></div><div className="actions"><button onClick={() => setProfiles([...profiles, emptyProfile(profiles.length)])}>{t.hints.addProfile}</button><button onClick={previewModels}><Eye size={14}/>{t.hints.preview}</button><button onClick={saveModels}><UploadCloud size={14}/>{t.hints.writeMykey}</button></div></div><div className="models-layout"><div className="profiles">{profiles.map((p, idx) => <div className="profile" key={idx}><div className="profile-head"><b>#{idx + 1} {p.name || p.var_name}</b><label><input type="checkbox" checked={!!p.enabled} onChange={(e) => patchProfile(idx, { enabled: e.target.checked })}/> enabled</label></div><div className="form-grid"><label>{t.fields.varName}<input value={p.var_name || ''} onChange={(e) => patchProfile(idx, { var_name: e.target.value })}/></label><label>{t.fields.type}<input value={p.type || ''} onChange={(e) => patchProfile(idx, { type: e.target.value })}/></label><label>{t.fields.name}<input value={p.name || ''} onChange={(e) => patchProfile(idx, { name: e.target.value })}/></label><label>{t.fields.model}<input value={p.model || ''} onChange={(e) => patchProfile(idx, { model: e.target.value })}/></label><label className="span2">{t.fields.apiBase}<input value={p.apibase || ''} onChange={(e) => patchProfile(idx, { apibase: e.target.value })}/></label><label className="span2">{t.fields.apiKey}<SecretInput value={p.apikey} onChange={(v) => patchProfile(idx, { apikey: v })} t={t}/></label><label>{t.fields.stream}<select value={String(!!p.stream)} onChange={(e) => patchProfile(idx, { stream: e.target.value === 'true' })}><option value="true">true</option><option value="false">false</option></select></label><label>{t.fields.maxRetries}<input type="number" value={p.max_retries ?? 3} onChange={(e) => patchProfile(idx, { max_retries: Number(e.target.value) })}/></label><label>{t.fields.readTimeout}<input type="number" value={p.read_timeout ?? 300} onChange={(e) => patchProfile(idx, { read_timeout: Number(e.target.value) })}/></label><label>{t.fields.reasoning}<input value={p.reasoning_effort || ''} onChange={(e) => patchProfile(idx, { reasoning_effort: e.target.value })}/></label></div></div>)}</div><div className="panel preview"><div className="panel-title"><FileCode2 size={18}/> {t.lists.generatedPreview}</div><pre>{preview || t.hints.previewHelp}</pre></div></div></section>}
      {tab === 'logs' && <section><div className="workspace"><div className="panel"><div className="panel-title">{t.lists.processes}</div>{services.map(s => <ServiceRow key={s.name} svc={s}/>)}</div><div className="panel log-panel"><div className="panel-title">Logs · {selected}</div><pre>{logs.join('\n') || t.hints.noLogs}</pre></div></div></section>}
    </main>
  </div>
}
