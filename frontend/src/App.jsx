import React, { useState, useRef, useEffect } from 'react'
import {
  MessageSquare, Plus, Trash2, Send, Menu, X, User, Copy, Check, Search, Sun, Moon, Square, Zap
} from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { vscDarkPlus, oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'
import {
  listConversations, createConversation, getConversation,
  deleteConversation, sendMessageStream, getModels
} from './api'

const PURPOSE_RECOMMENDATIONS = [
  { purpose: 'General Chat',       model: 'meta/llama-4-maverick-17b-128e-instruct' },
  { purpose: 'Coding',             model: 'qwen/qwen2.5-coder-32b-instruct' },
  { purpose: 'Reasoning',          model: 'deepseek-ai/deepseek-r1' },
  { purpose: 'Agentic workflows',  model: 'moonshotai/kimi-k2-instruct' },
  { purpose: 'Large context',      model: 'nvidia/nemotron-3-ultra-550b-a55b' },
  { purpose: 'Fast & lightweight', model: 'google/gemma-3-12b-it' },
  { purpose: 'Vision',             model: 'meta/llama-3.2-90b-vision-instruct' },
  { purpose: 'RAG',                model: 'nvidia/nemotron-mini-4b-instruct' },
  { purpose: 'OCR',                model: 'nvidia/nemotron-ocr-v1' },
  { purpose: 'Translation',        model: 'nvidia/riva-translate-4b-instruct-v1.1' },
]

export default function App() {
  const [conversations, setConversations] = useState([])
  const [activeConv, setActiveConv] = useState(null)
  const [messages, setMessages] = useState([])
  const [input, setInput] = useState('')
  const [streaming, setStreaming] = useState(false)
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [loadingConv, setLoadingConv] = useState(false)
  const [modelCategories, setModelCategories] = useState([])
  const [selectedModel, setSelectedModel] = useState('')
  const [modelDropdownOpen, setModelDropdownOpen] = useState(false)
  const [modelSearch, setModelSearch] = useState('')
  const [selectedPurpose, setSelectedPurpose] = useState('')
  const [purposeDropdownOpen, setPurposeDropdownOpen] = useState(false)
  const [theme, setTheme] = useState(() => {
    const saved = localStorage.getItem('theme')
    if (saved) return saved
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  })
  const messagesEndRef = useRef(null)
  const inputRef = useRef(null)
  const abortControllerRef = useRef(null)

  useEffect(() => {
    const root = document.documentElement
    if (theme === 'dark') {
      root.classList.add('dark')
    } else {
      root.classList.remove('dark')
    }
    localStorage.setItem('theme', theme)
  }, [theme])

  const toggleTheme = () => {
    setTheme(prev => prev === 'dark' ? 'light' : 'dark')
  }

  useEffect(() => {
    loadConversations()
    loadModels()
  }, [])

  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [messages])

  const loadConversations = async () => {
    try {
      const convs = await listConversations()
      setConversations(convs)
    } catch (err) {
      console.error('Failed to load conversations:', err)
    }
  }

  const loadModels = async () => {
    try {
      const categories = await getModels()
      setModelCategories(categories)
      const first = PURPOSE_RECOMMENDATIONS[0]
      if (first) {
        setSelectedPurpose(first.purpose)
        setSelectedModel(first.model)
      }
    } catch (err) {
      console.error('Failed to load models:', err)
    }
  }

  const handleSelectPurpose = (purposeName) => {
    setSelectedPurpose(purposeName)
    setPurposeDropdownOpen(false)
    const rec = PURPOSE_RECOMMENDATIONS.find(r => r.purpose === purposeName)
    if (rec) {
      setSelectedModel(rec.model)
    }
  }

  const recommendedModel = PURPOSE_RECOMMENDATIONS.find(r => r.purpose === selectedPurpose)?.model

  const handleNewChat = async () => {
    try {
      const conv = await createConversation('New Chat')
      setConversations([conv, ...conversations])
      setActiveConv(conv)
      setMessages([])
      inputRef.current?.focus()
    } catch (err) {
      console.error('Failed to create conversation:', err)
    }
  }

  const handleSelectConv = async (conv) => {
    setLoadingConv(true)
    try {
      const data = await getConversation(conv.id)
      setActiveConv(conv)
      setMessages(data.messages || [])
    } catch (err) {
      console.error('Failed to load conversation:', err)
    } finally {
      setLoadingConv(false)
    }
  }

  const handleDeleteConv = async (e, id) => {
    e.stopPropagation()
    try {
      await deleteConversation(id)
      setConversations(conversations.filter(c => c.id !== id))
      if (activeConv?.id === id) {
        setActiveConv(null)
        setMessages([])
      }
    } catch (err) {
      console.error('Failed to delete conversation:', err)
    }
  }

  const handleSend = async () => {
    if (!input.trim() || streaming) return

    let convId = activeConv?.id

    // Create conversation if none active
    if (!convId) {
      try {
        const conv = await createConversation(input.slice(0, 40) || 'New Chat')
        setConversations([conv, ...conversations])
        setActiveConv(conv)
        convId = conv.id
      } catch (err) {
        console.error('Failed to create conversation:', err)
        return
      }
    }

    const userMessage = { role: 'user', content: input }
    const assistantMessage = { role: 'assistant', content: '' }
    setMessages(prev => [...prev, userMessage, assistantMessage])
    setInput('')
    setStreaming(true)

    const controller = new AbortController()
    abortControllerRef.current = controller

    await sendMessageStream(
      convId,
      input,
      selectedModel,
      (token) => {
        setMessages(prev => {
          const updated = [...prev]
          updated[updated.length - 1] = {
            ...updated[updated.length - 1],
            content: updated[updated.length - 1].content + token,
          }
          return updated
        })
      },
      () => {
        setStreaming(false)
        abortControllerRef.current = null
        loadConversations()
      },
      (error) => {
        setMessages(prev => {
          const updated = [...prev]
          updated[updated.length - 1] = {
            ...updated[updated.length - 1],
            content: `Error: ${error}`,
          }
          return updated
        })
        setStreaming(false)
        abortControllerRef.current = null
      },
      controller.signal
    )
  }

  const handleStop = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
    }
  }

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="flex h-screen overflow-hidden" style={{ backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}>
      {/* Sidebar */}
      {sidebarOpen && (
        <div className="w-72 flex flex-col" style={{ backgroundColor: 'var(--bg-sidebar)', borderRight: '1px solid var(--border-primary)' }}>
          <div className="p-3" style={{ borderBottom: '1px solid var(--border-primary)' }}>
            <button
              onClick={handleNewChat}
              className="w-full flex items-center gap-2 px-3 py-2.5 rounded-lg transition-colors text-sm font-medium"
              style={{ backgroundColor: 'var(--bg-tertiary)' }}
              onMouseEnter={e => e.currentTarget.style.backgroundColor = 'var(--bg-hover)'}
              onMouseLeave={e => e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)'}
            >
              <Plus size={18} />
              New Chat
            </button>
          </div>

          <div className="flex-1 overflow-y-auto p-2 space-y-1">
            {conversations.length === 0 ? (
              <p className="text-sm text-center mt-8" style={{ color: 'var(--text-muted)' }}>No conversations yet</p>
            ) : (
              conversations.map(conv => (
                <div
                  key={conv.id}
                  onClick={() => handleSelectConv(conv)}
                  className="group flex items-center gap-2 px-3 py-2.5 rounded-lg cursor-pointer transition-colors"
                  style={{
                    backgroundColor: activeConv?.id === conv.id ? 'var(--bg-tertiary)' : 'transparent',
                  }}
                  onMouseEnter={e => { if (activeConv?.id !== conv.id) e.currentTarget.style.backgroundColor = 'var(--bg-hover)' }}
                  onMouseLeave={e => { if (activeConv?.id !== conv.id) e.currentTarget.style.backgroundColor = 'transparent' }}
                >
                  <MessageSquare size={16} className="shrink-0" style={{ color: 'var(--text-tertiary)' }} />
                  <span className="flex-1 truncate text-sm">{conv.title}</span>
                  <button
                    onClick={(e) => handleDeleteConv(e, conv.id)}
                    className="opacity-0 group-hover:opacity-100 hover:text-red-400 transition-all"
                    style={{ color: 'var(--text-muted)' }}
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              ))
            )}
          </div>

          <div className="p-3" style={{ borderTop: '1px solid var(--border-primary)' }}>
            <div className="flex items-center gap-2">
              <img src="/logo_short.png" alt="ntoricGPT" className="h-6 w-auto" />
            </div>
          </div>
        </div>
      )}

      {/* Main chat area */}
      <div className="flex-1 flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3" style={{ borderBottom: '1px solid var(--border-primary)' }}>
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="p-1.5 rounded-lg transition-colors"
            onMouseEnter={e => e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)'}
            onMouseLeave={e => e.currentTarget.style.backgroundColor = 'transparent'}
          >
            {sidebarOpen ? <X size={20} /> : <Menu size={20} />}
          </button>
          <h1 className="text-sm font-medium" style={{ color: 'var(--text-secondary)' }}>
            {activeConv?.title || 'ntoricGPT'}
          </h1>
          <div className="flex items-center gap-2">
            <button
              onClick={toggleTheme}
              className="p-1.5 rounded-lg transition-colors"
              onMouseEnter={e => e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)'}
              onMouseLeave={e => e.currentTarget.style.backgroundColor = 'transparent'}
              title="Toggle theme"
            >
              {theme === 'dark' ? <Sun size={18} /> : <Moon size={18} />}
            </button>
            <div className="relative">
              <button
                onClick={() => setModelDropdownOpen(!modelDropdownOpen)}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg transition-colors text-xs font-medium"
                style={{ backgroundColor: 'var(--bg-tertiary)', color: 'var(--text-secondary)' }}
                onMouseEnter={e => e.currentTarget.style.backgroundColor = 'var(--bg-hover)'}
                onMouseLeave={e => e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)'}
              >
                <img src="/logo_short.png" alt="" className="h-4 w-auto" />
                <span className="max-w-[160px] truncate">{selectedModel || 'Select model'}</span>
              </button>
              {modelDropdownOpen && (
                <>
                  <div
                    className="fixed inset-0 z-10"
                    onClick={() => { setModelDropdownOpen(false); setModelSearch('') }}
                  />
                  <div className="absolute right-0 top-full mt-1 w-80 rounded-lg shadow-xl z-20" style={{ backgroundColor: 'var(--bg-dropdown)', border: '1px solid var(--border-secondary)' }}>
                    <div className="p-2 relative" style={{ borderBottom: '1px solid var(--border-secondary)' }}>
                      <Search size={12} className="absolute left-4 top-1/2 -translate-y-1/2" style={{ color: 'var(--text-muted)' }} />
                      <input
                        type="text"
                        value={modelSearch}
                        onChange={e => setModelSearch(e.target.value)}
                        placeholder="Search models..."
                        autoFocus
                        className="w-full rounded-md pl-7 pr-3 py-2 text-xs outline-none focus:ring-1 focus:ring-[#0d9488]"
                        style={{ backgroundColor: 'var(--bg-dropdown-input)', color: 'var(--text-primary)' }}
                      />
                    </div>
                    <div className="max-h-64 overflow-y-auto">
                      {(() => {
                        const search = modelSearch.toLowerCase()
                        const filtered = modelCategories
                          .map(cat => ({
                            ...cat,
                            models: cat.models.filter(m => m.toLowerCase().includes(search))
                          }))
                          .filter(cat => cat.models.length > 0)

                        if (filtered.length === 0) {
                          return <p className="text-xs text-center py-4" style={{ color: 'var(--text-muted)' }}>No models found</p>
                        }

                        const totalUnique = new Set(
                          filtered.flatMap(cat => cat.models)
                        ).size

                        return (
                          <>
                            {filtered.map(cat => (
                              <div key={cat.name}>
                                <div
                                  className="px-3 py-1.5 text-[10px] font-semibold uppercase tracking-wide sticky top-0"
                                  style={{ color: 'var(--text-muted)', backgroundColor: 'var(--bg-dropdown)' }}
                                >
                                  {cat.name}
                                </div>
                                {cat.models.map(model => (
                                  <button
                                    key={model}
                                    onClick={() => {
                                      setSelectedModel(model)
                                      setModelDropdownOpen(false)
                                      setModelSearch('')
                                    }}
                                    className="w-full text-left px-3 py-2.5 text-xs transition-colors flex items-center gap-2"
                                    style={{ color: selectedModel === model ? '#0d9488' : 'var(--text-secondary)' }}
                                    onMouseEnter={e => e.currentTarget.style.backgroundColor = 'var(--bg-hover)'}
                                    onMouseLeave={e => e.currentTarget.style.backgroundColor = 'transparent'}
                                  >
                                    {model === recommendedModel && (
                                      <span
                                        className="text-[9px] font-semibold px-1.5 py-0.5 rounded shrink-0"
                                        style={{ backgroundColor: 'rgba(13, 148, 136, 0.15)', color: '#0d9488' }}
                                      >
                                        RECOMMENDED
                                      </span>
                                    )}
                                    <span className="truncate">{model}</span>
                                  </button>
                                ))}
                              </div>
                            ))}
                          </>
                        )
                      })()}
                    </div>
                    <div className="p-2 text-center" style={{ borderTop: '1px solid var(--border-secondary)' }}>
                      <span className="text-[10px]" style={{ color: 'var(--text-muted)' }}>
                        {(() => {
                          const all = modelCategories.flatMap(c => c.models)
                          return new Set(all).size
                        })()} models available
                      </span>
                    </div>
                  </div>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto">
          {messages.length === 0 ? (
            <div className="h-full flex flex-col items-center justify-center text-center px-4">
              <img
                src={theme === 'dark' ? '/logo_dark.png' : '/logo_white.png'}
                alt="ntoricGPT"
                className="h-28 w-auto mb-4"
              />
              <h2 className="text-2xl font-semibold mb-2">ntoricGPT</h2>
              <p className="text-sm max-w-md" style={{ color: 'var(--text-muted)' }}>
                Start a conversation by typing a message below. Powered by ntoric API.
              </p>
            </div>
          ) : (
            <div className="max-w-3xl mx-auto py-4 px-4 space-y-6">
              {messages.map((msg, i) => (
                <MessageBubble key={i} message={msg} streaming={streaming && i === messages.length - 1} theme={theme} />
              ))}
              <div ref={messagesEndRef} />
            </div>
          )}
        </div>

        {/* Input */}
        <div className="p-4" style={{ borderTop: '1px solid var(--border-primary)' }}>
          <div className="max-w-3xl mx-auto">
            {/* Purpose selector */}
            <div className="flex items-center gap-2 mb-2 flex-wrap">
              <span className="text-[11px] font-medium" style={{ color: 'var(--text-muted)' }}>Purpose:</span>
              <div className="relative">
                <button
                  onClick={() => setPurposeDropdownOpen(!purposeDropdownOpen)}
                  className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg transition-colors text-[11px] font-medium"
                  style={{ backgroundColor: 'var(--bg-tertiary)', color: 'var(--text-secondary)' }}
                  onMouseEnter={e => e.currentTarget.style.backgroundColor = 'var(--bg-hover)'}
                  onMouseLeave={e => e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)'}
                >
                  <Zap size={11} style={{ color: '#0d9488' }} />
                  <span>{selectedPurpose || 'Select purpose'}</span>
                </button>
                {purposeDropdownOpen && (
                  <>
                    <div
                      className="fixed inset-0 z-10"
                      onClick={() => setPurposeDropdownOpen(false)}
                    />
                    <div className="absolute left-0 bottom-full mb-1 w-72 rounded-lg shadow-xl z-20" style={{ backgroundColor: 'var(--bg-dropdown)', border: '1px solid var(--border-secondary)' }}>
                      <div className="px-3 py-1.5 text-[10px] font-semibold uppercase tracking-wide" style={{ color: 'var(--text-muted)' }}>
                        Recommended free models
                      </div>
                      <div className="max-h-56 overflow-y-auto">
                        {PURPOSE_RECOMMENDATIONS.map(rec => (
                          <button
                            key={rec.purpose}
                            onClick={() => handleSelectPurpose(rec.purpose)}
                            className="w-full text-left px-3 py-2.5 text-xs transition-colors flex items-center gap-2"
                            style={{ color: selectedPurpose === rec.purpose ? '#0d9488' : 'var(--text-secondary)' }}
                            onMouseEnter={e => e.currentTarget.style.backgroundColor = 'var(--bg-hover)'}
                            onMouseLeave={e => e.currentTarget.style.backgroundColor = 'transparent'}
                          >
                            <Zap size={11} style={{ color: selectedPurpose === rec.purpose ? '#0d9488' : 'var(--text-muted)' }} />
                            <div className="flex flex-col">
                              <span className="font-medium">{rec.purpose}</span>
                              <span className="text-[10px]" style={{ color: 'var(--text-muted)' }}>{rec.model}</span>
                            </div>
                          </button>
                        ))}
                      </div>
                    </div>
                  </>
                )}
              </div>
              {selectedModel && (
                <span className="text-[11px]" style={{ color: 'var(--text-muted)' }}>
                  → {selectedModel}
                </span>
              )}
            </div>
            <div className="relative">
            <textarea
              ref={inputRef}
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Send a message..."
              rows={1}
              className="w-full rounded-2xl px-4 py-3 pr-12 resize-none outline-none focus:ring-1 focus:ring-[#0d9488] text-sm"
              style={{ minHeight: '48px', maxHeight: '200px', backgroundColor: 'var(--bg-input)', color: 'var(--text-primary)' }}
              onInput={e => {
                e.target.style.height = 'auto'
                e.target.style.height = Math.min(e.target.scrollHeight, 200) + 'px'
              }}
            />
            {streaming ? (
              <button
                onClick={handleStop}
                className="absolute right-3 bottom-3 p-2 rounded-lg transition-colors"
                style={{ backgroundColor: 'var(--bg-tertiary)', color: 'var(--text-primary)' }}
                title="Stop generating"
              >
                <Square size={16} fill="currentColor" />
              </button>
            ) : (
              <button
                onClick={handleSend}
                disabled={!input.trim()}
                className="absolute right-3 bottom-3 p-2 rounded-lg bg-[#0d9488] hover:bg-[#14b8a6] disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
              >
                <Send size={16} className="text-black" />
              </button>
            )}
            </div>
          </div>
          <p className="text-center text-xs mt-2" style={{ color: 'var(--text-muted)' }}>
            ntoricGPT can make mistakes. Check important info.
          </p>
        </div>
      </div>
    </div>
  )
}

function CodeBlock({ code, language, theme }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = () => {
    navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="code-block-wrapper">
      <div className="code-block-header">
        <span className="code-lang-label">{language || 'text'}</span>
        <button
          onClick={handleCopy}
          className="code-copy-btn"
          title="Copy code"
        >
          {copied ? <Check size={14} /> : <Copy size={14} />}
          {copied ? 'Copied!' : 'Copy'}
        </button>
      </div>
      <SyntaxHighlighter
        language={language || 'text'}
        style={theme === 'dark' ? vscDarkPlus : oneLight}
        customStyle={{
          margin: 0,
          borderRadius: '0 0 8px 8px',
          fontSize: '0.85rem',
          background: 'var(--bg-code)',
        }}
        codeTagProps={{
          style: { fontFamily: "'Fira Code', 'Courier New', monospace" }
        }}
      >
        {code}
      </SyntaxHighlighter>
    </div>
  )
}

function MessageBubble({ message, streaming, theme }) {
  const isUser = message.role === 'user'

  return (
    <div className={`flex gap-3 ${isUser ? 'flex-row-reverse' : ''}`}>
      <div className={`w-8 h-8 rounded-full flex items-center justify-center shrink-0`}
        style={{ backgroundColor: isUser ? 'var(--avatar-user)' : 'rgba(13, 148, 136, 0.1)' }}
      >
        {isUser ? <User size={16} /> : <img src="/logo_short.png" alt="" className="h-4 w-auto" />}
      </div>
      <div className={`flex-1 ${isUser ? 'text-right' : ''}`}>
        <div className={`${isUser ? 'inline-block' : 'block w-full'} text-left rounded-2xl px-4 py-3`}
          style={{
            backgroundColor: isUser ? 'var(--bg-tertiary)' : 'transparent',
            color: isUser ? 'var(--text-primary)' : 'var(--text-secondary)',
          }}
        >
          {isUser ? (
            <p className="text-sm whitespace-pre-wrap">{message.content}</p>
          ) : (
            <div className="markdown-content text-sm">
              {message.content ? (
                <ReactMarkdown
                components={{
                  pre: ({ children }) => <>{children}</>,
                  code({ className, children, ...props }) {
                    const match = /language-(\w+)/.exec(className || '')
                    const code = String(children).replace(/\n$/, '')
                    if (match) {
                      return <CodeBlock code={code} language={match[1]} theme={theme} />
                    }
                    return <code className={className} {...props}>{children}</code>
                  },
                }}
              >
                {message.content}
              </ReactMarkdown>
              ) : streaming ? (
                <div className="flex gap-1 items-center py-1">
                  <span className="typing-dot w-2 h-2 rounded-full" style={{ backgroundColor: 'var(--text-muted)' }}></span>
                  <span className="typing-dot w-2 h-2 rounded-full" style={{ backgroundColor: 'var(--text-muted)' }}></span>
                  <span className="typing-dot w-2 h-2 rounded-full" style={{ backgroundColor: 'var(--text-muted)' }}></span>
                </div>
              ) : null}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
