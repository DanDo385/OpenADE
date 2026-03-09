import { FormEvent, KeyboardEvent, useState } from 'react'

interface MessageInputProps {
  disabled?: boolean
  isStreaming?: boolean
  onSend: (content: string) => Promise<void>
}

export function MessageInput({ disabled = false, isStreaming = false, onSend }: MessageInputProps) {
  const [content, setContent] = useState('')
  const [isSending, setIsSending] = useState(false)

  const canSubmit = !disabled && !isSending && !isStreaming && content.trim().length > 0

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    if (!canSubmit) {
      return
    }

    const value = content.trim()
    setContent('')
    setIsSending(true)
    try {
      await onSend(value)
    } finally {
      setIsSending(false)
    }
  }

  const onKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>) => {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      if (canSubmit) {
        void submit(event)
      }
    }
  }

  return (
    <form className="message-input" onSubmit={submit}>
      <textarea
        placeholder="Ask anything..."
        value={content}
        onChange={(event) => setContent(event.target.value)}
        onKeyDown={onKeyDown}
        rows={3}
        disabled={disabled || isStreaming}
      />
      <button type="submit" className="btn btn-primary" disabled={!canSubmit}>
        {isStreaming || isSending ? 'Sending...' : 'Send'}
      </button>
    </form>
  )
}
