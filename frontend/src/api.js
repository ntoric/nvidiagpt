const API_BASE = '/api'

export async function getModels() {
  const res = await fetch(`${API_BASE}/models`)
  if (!res.ok) throw new Error('Failed to fetch models')
  const data = await res.json()
  return data.models || []
}

export async function listConversations() {
  const res = await fetch(`${API_BASE}/conversations`)
  if (!res.ok) throw new Error('Failed to list conversations')
  return res.json()
}

export async function createConversation(title = 'New Chat') {
  const res = await fetch(`${API_BASE}/conversations`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title }),
  })
  if (!res.ok) throw new Error('Failed to create conversation')
  return res.json()
}

export async function getConversation(id) {
  const res = await fetch(`${API_BASE}/conversations/${id}`)
  if (!res.ok) throw new Error('Failed to get conversation')
  return res.json()
}

export async function deleteConversation(id) {
  const res = await fetch(`${API_BASE}/conversations/${id}`, {
    method: 'DELETE',
  })
  if (!res.ok) throw new Error('Failed to delete conversation')
}

export async function sendMessageStream(conversationId, message, model, onToken, onDone, onError) {
  try {
    const res = await fetch(`${API_BASE}/conversations/${conversationId}/chat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message, model }),
    })

    if (!res.ok) {
      const text = await res.text()
      onError(text || 'Request failed')
      return
    }

    const reader = res.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (!line.startsWith('data: ')) continue
        const data = line.slice(6).trim()
        if (data === '[DONE]') {
          onDone()
          return
        }
        try {
          const parsed = JSON.parse(data)
          if (parsed.error) {
            onError(parsed.error)
          } else if (parsed.content) {
            onToken(parsed.content)
          }
        } catch {
          // skip non-JSON lines
        }
      }
    }
    onDone()
  } catch (err) {
    onError(err.message)
  }
}
