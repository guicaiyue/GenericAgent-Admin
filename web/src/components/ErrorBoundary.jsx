// 全局错误边界：捕获子组件渲染时未捕获的 JS 错误，防止整页白屏。
// 使用 class component（React ErrorBoundary API 只支持类组件）。

import React from 'react'

export class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error }
  }

  componentDidCatch(error, info) {
    console.error('[ErrorBoundary]', error, info.componentStack)
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null })
  }

  render() {
    if (this.state.hasError) {
      return (
        <div style={{
          display: 'flex', flexDirection: 'column', alignItems: 'center',
          justifyContent: 'center', minHeight: '100vh', gap: '16px',
          background: 'var(--bg)', color: 'var(--text)', fontFamily: 'system-ui, sans-serif',
          padding: '24px'
        }}>
          <span style={{ fontSize: '48px' }}>⚠️</span>
          <h2 style={{ margin: 0, fontSize: '20px' }}>页面渲染出错</h2>
          <p style={{ margin: 0, color: 'var(--muted)', maxWidth: '480px', textAlign: 'center', fontSize: '14px' }}>
            {this.state.error?.message || '未知错误'}
          </p>
          <button
            onClick={this.handleRetry}
            style={{
              padding: '8px 20px', background: 'var(--accent)', color: '#fff',
              border: 'none', borderRadius: '8px', cursor: 'pointer', fontSize: '14px'
            }}
          >
            重试
          </button>
        </div>
      )
    }
    return this.props.children
  }
}
