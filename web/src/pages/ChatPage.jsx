import { useEffect, useRef, useState } from 'react'
import { Play, RefreshCw, Search, X, XCircle } from 'lucide-react'
import { api, apiStream } from '../lib/api'
import { TurnList } from '../components/turns'

export function ChatPage({ t }) {
  const [sessions, setSessions] = useState([]), [sid, setSid] = useState(''), [messages, setMessages] = useState([])
  const [prompt, setPrompt] = useState(''), [busy, setBusy] = useState(false), [err, setErr] = useState('')
  const [q, setQ] = useState(''), [lastPrompt, setLastPrompt] = useState('')
  const abortRef = useRef(null), timerRef = useRef(null)
  const TIMEOUT = 60000
  const loadSessions = async () => { try { const d = await api('/api/chat/sessions'); setSessions(d.sessions || []); if (!sid && d.sessions?.[0]) await openSession(d.sessions[0].id) } catch(e){ setErr(e.message) } }
  const openSession = async (id) => { setQ(''); try { const d = await api(`/api/chat/session/${id}`); setSid(d.id); setMessages(d.messages || []) } catch(e){ setErr(e.message) } }
  const newSession = async () => { try { const d = await api('/api/chat/session/new', { method:'POST', body:'{}' }); setSid(d.id); setMessages([]); setQ(''); await loadSessions() } catch(e){ setErr(e.message) } }
  useEffect(()=>{ loadSessions().catch(e=>setErr(e.message)) }, [])
  const cancel = () => { if (abortRef.current) { abortRef.current.abort(); abortRef.current = null } clearTimeout(timerRef.current) }
  const send = async () => {
    const text = prompt.trim()
    if (text === '/new') {
      setPrompt('')
      if (busy) { setErr('当前正在执行，完成后可使用 /new 创建新会话'); return }
      await newSession()
      return
    }
    if (!text || busy) return
    let cur = sid
    if (!cur) { try { const d = await api('/api/chat/session/new', { method:'POST', body:'{}' }); cur = d.id; setSid(cur) } catch(e){ setErr(e.message); return } }
    setPrompt(''); setBusy(true); setErr(''); setLastPrompt(text)
    const user = { id: `u-${Date.now()}`, role:'user', content:text, created_at: Math.floor(Date.now()/1000) }
    const assistant = { id: `a-${Date.now()}`, role:'assistant', content:'', created_at: Math.floor(Date.now()/1000) }
    setMessages(ms => [...ms, user, assistant])
    const ac = new AbortController(); abortRef.current = ac
    timerRef.current = setTimeout(() => { ac.abort(); setErr('请求超时(60s)，已自动取消'); setBusy(false); abortRef.current = null }, TIMEOUT)
    try {
      const res = await apiStream(`/api/chat/${cur}`, { signal: ac.signal, method:'POST', body: JSON.stringify({ prompt:text, client_user_id:user.id }) })
      const reader = res.body.getReader(), dec = new TextDecoder(); let buf = '', content = ''
      while (true) {
        const {value, done} = await reader.read(); if (done) break
        buf += dec.decode(value, {stream:true}); const lines = buf.split('\n'); buf = lines.pop() || ''
        for (const line of lines) { if (!line.trim()) continue; const ev = JSON.parse(line)
          if (ev.type === 'delta') { content += ev.delta || ''; setMessages(ms => ms.map(m => m.id === assistant.id ? {...m, content} : m)) }
          if ((ev.type === 'done' || ev.type === 'error') && ev.message) { setMessages(ms => ms.map(m => m.id === assistant.id ? ev.message : m)); if (ev.type === 'error') setErr(ev.message.content || 'error') }
        }
      }
      await loadSessions()
    } catch(e) {
      if (e.name === 'AbortError') return
      setErr(e.message); setMessages(ms => ms.map(m => m.id === assistant.id ? {...m, content:`失败：${e.message}`, error:true} : m))
    } finally { clearTimeout(timerRef.current); abortRef.current = null; setBusy(false) }
  }
  const retry = () => { if (lastPrompt) { setPrompt(lastPrompt); setLastPrompt('') } }
  const filtered = q ? messages.filter(m => m.content && m.content.toLowerCase().includes(q.toLowerCase())) : messages
  return <section className="chat-page"><div className="chat-header">
    <div className="chat-search"><Search size={14}/><input value={q} onChange={e=>setQ(e.target.value)} placeholder="搜索消息..."/><button onClick={()=>setQ('')} hidden={!q}><X size={14}/></button></div>
    <div className="chat-actions-right">
      <button className="danger" disabled={!busy} onClick={cancel} title="取消"><XCircle size={14}/>取消</button>
      <button disabled={busy || !err} onClick={retry} title="重试上次"><RefreshCw size={14}/>重试</button>
      <button disabled={busy} onClick={newSession}><Play size={14}/>新会话</button>
    </div>
  </div><main className="chat-main"><TurnList messages={filtered} empty={q ? '无匹配消息' : '选择或创建会话后开始对话'}/>
    <div className="chat-compose">
      <textarea value={prompt} onChange={e=>setPrompt(e.target.value)} onKeyDown={e=>{ if(e.key==='Enter' && e.ctrlKey) send() }} placeholder="输入给 GenericAgent 的任务，Ctrl+Enter 发送"/>
      <button disabled={busy || !prompt.trim()} onClick={send}>{busy?'执行中...':'发送'}</button>
    </div>
    {err && <p className="chat-error">{err} <button onClick={()=>setErr('')} style={{background:'none',border:'none',padding:0,cursor:'pointer',color:'var(--accent)',fontSize:'12px'}}><X size={12}/></button></p>}
  </main></section>
}