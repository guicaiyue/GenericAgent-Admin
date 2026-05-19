import { useEffect, useMemo, useRef, useState } from 'react'
import { Bot, Clock3, MessageSquarePlus, PanelLeftClose, PanelLeftOpen, RefreshCw, Send, Sparkles, Trash2 } from 'lucide-react'

const api = async (url, options = {}) => {
  const res = await fetch(url, { headers: { 'Content-Type': 'application/json', ...(options.headers || {}) }, ...options })
  if (!res.ok) throw new Error(await res.text())
  return res.json()
}

const fmtTime = (v) => {
  if (!v) return ''
  try { return new Date(v * 1000).toLocaleString() } catch { return '' }
}

const shortTitle = (s) => s?.title || '新会话'

export default function ChatApp() {
  const [sessions, setSessions] = useState([])
  const [sid, setSid] = useState('')
  const [messages, setMessages] = useState([])
  const [prompt, setPrompt] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const [collapsed, setCollapsed] = useState(false)
  const [notice, setNotice] = useState('')
  const endRef = useRef(null)
  const current = useMemo(() => sessions.find(s => s.id === sid), [sessions, sid])

  const loadSessions = async (prefer = sid) => {
    const d = await api('/api/chat/sessions')
    const list = d.sessions || []
    setSessions(list)
    const next = prefer || list[0]?.id || ''
    if (next) await openSession(next, false)
    return list
  }
  const openSession = async (id, refreshList = true) => {
    const d = await api(`/api/chat/session/${id}`)
    setSid(d.id); setMessages(d.messages || []); setErr('')
    if (refreshList) setSessions(xs => xs.map(x => x.id === d.id ? { ...x, title: d.title, count: d.messages?.length || x.count, updated_at: d.updated_at || x.updated_at } : x))
  }
  const newSession = async () => {
    const d = await api('/api/chat/session/new', { method:'POST', body:'{}' })
    setSid(d.id); setMessages([]); setPrompt(''); setErr(''); setNotice('已创建新会话')
    await loadSessions(d.id)
  }
  const deleteSession = async (id) => {
    if (!id || !confirm('删除这个会话？')) return
    await api(`/api/chat/session/${id}`, { method:'DELETE' })
    setNotice('会话已删除')
    setSid(''); setMessages([])
    await loadSessions('')
  }
  useEffect(()=>{ loadSessions('').catch(e=>setErr(e.message)) }, [])
  useEffect(()=>{ endRef.current?.scrollIntoView({ behavior:'smooth', block:'end' }) }, [messages, busy])

  const send = async () => {
    const text = prompt.trim()
    if (!text || busy) return
    let cur = sid
    if (!cur) {
      const d = await api('/api/chat/session/new', { method:'POST', body:'{}' })
      cur = d.id; setSid(cur)
    }
    setPrompt(''); setBusy(true); setErr(''); setNotice('')
    const ts = Math.floor(Date.now()/1000)
    const user = { id: `u-${Date.now()}`, role:'user', content:text, created_at: ts }
    const assistant = { id: `a-${Date.now()}`, role:'assistant', content:'', created_at: ts }
    setMessages(ms => [...ms, user, assistant])
    try {
      const res = await fetch(`/api/chat/${cur}`, { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ prompt:text, client_user_id:user.id }) })
      if (!res.ok) throw new Error(await res.text())
      const reader = res.body.getReader(), dec = new TextDecoder()
      let buf = '', content = ''
      while (true) {
        const {value, done} = await reader.read()
        if (done) break
        buf += dec.decode(value, {stream:true})
        const lines = buf.split('\n'); buf = lines.pop() || ''
        for (const line of lines) {
          if (!line.trim()) continue
          const ev = JSON.parse(line)
          if (ev.type === 'delta') {
            content += ev.delta || ''
            setMessages(ms => ms.map(m => m.id === assistant.id ? {...m, content} : m))
          }
          if (ev.type === 'error') {
            const msg = ev.message || '执行失败'
            setMessages(ms => ms.map(m => m.id === assistant.id ? {...m, content: content || msg, error:true} : m))
            setErr(msg)
          }
          if (ev.type === 'done' && ev.message) {
            content = ev.message.content || content
            setMessages(ms => ms.map(m => m.id === assistant.id ? ev.message : m))
          }
        }
      }
      await loadSessions(cur)
    } catch(e) {
      setErr(e.message)
      setMessages(ms => ms.map(m => m.id === assistant.id ? {...m, content:e.message, error:true} : m))
    } finally { setBusy(false) }
  }

  const examples = ['总结当前 GenericAgent 状态并指出风险', '读取最近自主任务报告并给出下一步建议', '检查模型配置是否完整']
  return <div className="chat-app">
    <aside className={`chat-dock ${collapsed ? 'collapsed' : ''}`}>
      <div className="chat-brand"><div className="logo-orb"><Bot size={22}/></div><div><b>GenericAgent Chat</b><span>Native Worker Bridge</span></div></div>
      <div className="dock-actions"><button onClick={newSession}><MessageSquarePlus size={16}/>新会话</button><button onClick={()=>loadSessions().catch(e=>setErr(e.message))}><RefreshCw size={16}/></button></div>
      <div className="session-list">{sessions.map(s => <button key={s.id} className={s.id===sid?'selected':''} onClick={()=>openSession(s.id)}><b>{shortTitle(s)}</b><span><Clock3 size={12}/>{fmtTime(s.updated_at)} · {s.count || 0} 条</span></button>)}</div>
      <div className="dock-foot"><button onClick={()=>setCollapsed(!collapsed)}>{collapsed ? <PanelLeftOpen size={16}/> : <PanelLeftClose size={16}/>}侧栏</button></div>
    </aside>
    <main className="chat-stage">
      <header className="chat-hero"><div><p className="eyebrow"><Sparkles size={15}/>GA 原生对话</p><h1>{shortTitle(current)}</h1><span>{current ? `会话 ${current.id}` : '创建或选择一个会话开始'}</span></div><div className="hero-actions"><button onClick={()=>window.location.href='/'}>返回管理台</button><button disabled={!sid} onClick={()=>deleteSession(sid)}><Trash2 size={15}/>删除</button></div></header>
      {(err || notice) && <div className={err ? 'chat-alert error' : 'chat-alert'}>{err || notice}</div>}
      <section className="conversation">
        {messages.length === 0 && <div className="welcome-card"><h2>把任务交给 GenericAgent</h2><p>独立标签页拥有更宽的上下文空间，Go 后端会按需启动 Python Worker 并流式返回结果。</p><div>{examples.map(x => <button key={x} onClick={()=>setPrompt(x)}>{x}</button>)}</div></div>}
        {messages.map(m => <article key={m.id} className={`chat-msg ${m.role} ${m.error?'error':''}`}><div className="avatar">{m.role === 'user' ? '你' : 'GA'}</div><div className="msg-card"><div className="msg-meta"><b>{m.role === 'user' ? 'User' : 'GenericAgent'}</b><span>{fmtTime(m.created_at)}</span></div><pre>{m.content || (busy && m.role === 'assistant' ? '思考中…' : '')}</pre></div></article>)}
        <div ref={endRef}/>
      </section>
      <footer className="composer"><textarea value={prompt} onChange={e=>setPrompt(e.target.value)} onKeyDown={e=>{ if(e.key==='Enter' && (e.ctrlKey || e.metaKey)) send() }} placeholder="输入任务，Ctrl/⌘ + Enter 发送"/><button disabled={busy || !prompt.trim()} onClick={send}><Send size={17}/>{busy?'执行中':'发送'}</button></footer>
    </main>
  </div>
}
