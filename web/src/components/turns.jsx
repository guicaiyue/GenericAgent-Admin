import { Copy } from 'lucide-react'
export function TurnBubble({ message, onCopy }) {
  const m = message || {}
  return <div className={`bubble ${m.role || 'assistant'} ${m.type || ''} ${m.error?'error':''}`}>
    <div className="role">{m.title || m.role || 'assistant'}</div>
    <div className="content">{m.content || ''}</div>
    {onCopy && m.content && <button className="copy-msg-btn" title="复制" onClick={() => onCopy(m.content)}><Copy size={12}/></button>}
  </div>
}

export function TurnList({ messages, empty, className = '', onCopy }) {
  const items = messages || []
  return <div className={`chat-messages turn-list ${className}`}>
    {items.length===0 && <div className="empty-chat">{empty}</div>}
    {items.map((m, i) => <TurnBubble key={m.id || i} message={m} onCopy={onCopy} />)}
  </div>
}
