import React, { useEffect, useState } from 'react'
import { api } from '../lib/api'
import { confirmDanger } from '../lib/danger'
import { EntryList } from '../components/common'

// Simple markdown renderer without external deps
function renderMarkdown(text) {
  if (!text) return null
  const lines = text.split('\n')
  return lines.map((line, i) => {
    // Code block
    if (line.startsWith('```')) return <pre key={i} style={{background:'#1e1e1e',padding:8,borderRadius:4,margin:'4px 0',overflow:'auto'}}>{line.slice(3)}</pre>
    // Headers
    if (line.startsWith('#### ')) return <h4 key={i} style={{margin:'8px 0 4px',fontSize:13,color:'#e0e0e0'}}>{line.slice(5)}</h4>
    if (line.startsWith('### ')) return <h3 key={i} style={{margin:'10px 0 4px',fontSize:14,color:'#e0e0e0'}}>{line.slice(4)}</h3>
    if (line.startsWith('## ')) return <h2 key={i} style={{margin:'12px 0 4px',fontSize:15,color:'#a0c0ff'}}>{line.slice(3)}</h2>
    if (line.startsWith('# ')) return <h1 key={i} style={{margin:'14px 0 6px',fontSize:16,color:'#80a0ff'}}>{line.slice(2)}</h1>
    // List items
    if (line.match(/^[-*] /)) return <div key={i} style={{paddingLeft:12,margin:'2px 0'}}>• {line.slice(2)}</div>
    if (line.match(/^\d+\. /)) {
      const m = line.match(/^(\d+)\. (.*)/)
      return <div key={i} style={{paddingLeft:12,margin:'2px 0'}}>{m[1]}. {m[2]}</div>
    }
    // Horizontal rule
    if (line.match(/^---+$/)) return <hr key={i} style={{borderColor:'#333',margin:'8px 0'}}/>
    // Empty line
    if (!line.trim()) return <div key={i} style={{height:4}}/>
    // Regular paragraph with inline formatting
    const formatted = line
      .replace(/`([^`]+)`/g, '<code style="background:#2a2a2a;padding:1px 4px;border-radius:3px;font-size:12px">$1</code>')
      .replace(/\*\*([^*]+)\*\*/g, '<b>$1</b>')
      .replace(/\*([^*]+)\*/g, '<i>$1</i>')
      .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" style="color:#80a0ff">$1</a>')
    return <p key={i} style={{margin:'2px 0',lineHeight:1.5,fontSize:13}} dangerouslySetInnerHTML={{__html: formatted}}/>
  })
}

export default function WikiPage({ t, cfg, api: apiFn }) {
  const [status, setStatus] = useState(null)
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState('')
  const [selectedFile, setSelectedFile] = useState('')
  const [fileContent, setFileContent] = useState('')
  const [llms, setLLMs] = useState([])
  const [selectedLLM, setSelectedLLM] = useState(0)
  const [ingestLog, setIngestLog] = useState('')

  const refresh = async () => {
    try {
      const d = await (apiFn || api)('/api/wiki/status')
      setStatus(d)
      return d
    } catch (e) {
      setMsg(e.message)
      return null
    }
  }

  const loadLLMs = async () => {
    try {
      const d = await (apiFn || api)('/api/chat/state')
      setLLMs(d.llms || [])
      if (d.llm_no !== undefined) setSelectedLLM(d.llm_no)
    } catch (e) { /* ignore */ }
  }

  useEffect(() => {
    refresh()
    loadLLMs()
  }, [])

  const handleSync = async () => {
    if (!confirmDanger('wiki-sync', t?.confirmSync || '同步 GA memory 到 wiki/raw？')) return
    setBusy(true); setMsg('')
    try {
      const d = await (apiFn || api)('/api/wiki/sync', { dangerous: true, method: 'POST', body: '{}' })
      setStatus(d)
      setMsg(`同步完成：新增 ${d.added}，修改 ${d.modified}，删除 ${d.removed}`)
    } catch (e) { setMsg(`同步失败：${e.message}`) }
    finally { setBusy(false) }
  }

  const handleIngest = async () => {
    if (!confirmDanger('wiki-ingest', t?.confirmIngest || '使用 AI 重建 wiki index？')) return
    setBusy(true); setMsg(''); setIngestLog('')
    try {
      const d = await (apiFn || api)(`/api/wiki/ingest?llm_no=${selectedLLM}`, { dangerous: true, method: 'POST', body: '{}' })
      setStatus(d)
      setMsg('Wiki 重建任务已启动，请刷新状态查看进度')
    } catch (e) { setMsg(`重建失败：${e.message}`) }
    finally { setBusy(false) }
  }

  const handleFileClick = async (file) => {
    setSelectedFile(file)
    try {
      const d = await (apiFn || api)(`/api/wiki/file?path=${encodeURIComponent(file)}`)
      setFileContent(d.content || d)
    } catch (e) { setFileContent(`加载失败：${e.message}`) }
  }

  const files = status?.files || []
  const rawCount = status?.raw_count ?? 0
  const indexSize = status?.index_size ?? 0

  return (
    <div className="wiki-page">
      {/* Status bar */}
      <div style={{display:'flex',gap:8,flexWrap:'wrap',marginBottom:12,fontSize:12}}>
        <span>Wiki 目录: <code>{cfg?.wiki_dir || '未配置'}</code></span>
        <span>raw 文件: <b>{rawCount}</b></span>
        <span>index 大小: <b>{indexSize > 0 ? `${(indexSize/1024).toFixed(1)} KB` : '-'}</b></span>
      </div>

      {/* Action buttons */}
      <div style={{display:'flex',gap:8,marginBottom:12,flexWrap:'wrap'}}>
        <button disabled={busy} onClick={handleSync}>
          ↺ {(t?.syncBtn) || '同步 memory → raw'}
        </button>
        <button disabled={busy} onClick={handleIngest}>
          ⚙ {(t?.rebuildBtn) || '重建 wiki'}
        </button>
        {llms.length > 0 && (
          <select value={selectedLLM} onChange={e => setSelectedLLM(Number(e.target.value))} style={{fontSize:12}}>
            {llms.map(l => <option key={l.index} value={l.index}>{l.label || l.name}</option>)}
          </select>
        )}
        <button disabled={busy} onClick={refresh}>↻ 刷新</button>
      </div>

      {msg && <p style={{fontSize:12,color: msg.includes('失败') || msg.includes('错误') ? '#ff6b6b' : '#80ff80',margin:'4px 0'}}>{msg}</p>}

      {/* File browser + preview */}
      <div style={{display:'grid',gridTemplateColumns:'280px 1fr',gap:12,minHeight:300}}>
        {/* File list */}
        <div style={{border:'1px solid #333',borderRadius:6,padding:8,overflow:'auto',maxHeight:500}}>
          <div style={{fontSize:11,marginBottom:6,color:'#888'}}>{t?.uploadedFiles || '文件列表'}</div>
          {files.length === 0
            ? <p style={{fontSize:12,color:'#666'}}>暂无文件</p>
            : files.map(f => (
              <button key={f.path} onClick={() => handleFileClick(f.path)}
                style={{
                  display:'block',width:'100%',textAlign:'left',padding:'4px 6px',marginBottom:2,
                  background: selectedFile === f.path ? '#2a3a5a' : 'transparent',
                  border:'none',borderRadius:4,cursor:'pointer',fontSize:12,color:'#c0d0e0'
                }}>
                {f.path.split('/').pop()}
                <small style={{color:'#666',marginLeft:6}}>{f.size > 1024 ? `${(f.size/1024).toFixed(0)}K` : f.size}b</small>
              </button>
            ))
          }
        </div>

        {/* File preview */}
        <div style={{border:'1px solid #333',borderRadius:6,padding:12,overflow:'auto',maxHeight:500,background:'#141820'}}>
          {selectedFile
            ? <>
                <div style={{fontSize:11,color:'#888',marginBottom:8,borderBottom:'1px solid #2a2a2a',paddingBottom:6}}>
                  {selectedFile}
                </div>
                <div style={{fontSize:12,color:'#d0d8e0',lineHeight:1.6}}>
                  {renderMarkdown(fileContent)}
                </div>
              </>
            : <p style={{fontSize:12,color:'#555'}}>← 点击左侧文件预览内容</p>
          }
        </div>
      </div>
    </div>
  )
}
