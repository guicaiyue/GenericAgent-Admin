import React from 'react'
import { createRoot } from 'react-dom/client'
import App from './App.jsx'
import ChatApp from './ChatApp.jsx'
import './style.css'

const Root = window.location.pathname.replace(/\/+$/, '') === '/chat' ? ChatApp : App
createRoot(document.getElementById('root')).render(<Root />)
