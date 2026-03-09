import ReactMarkdown from 'react-markdown'
import type { Message } from '../../lib/api-types'

interface MessageListProps {
  messages: Message[]
  streamingContent: string
  isStreaming: boolean
  onSaveAsTask?: () => void
}

function formatTimestamp(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  return new Intl.DateTimeFormat(undefined, {
    hour: 'numeric',
    minute: '2-digit',
  }).format(date)
}

function MessageBubble({ message }: { message: Message }) {
  const isAssistant = message.role === 'assistant'

  return (
    <article className={`message ${isAssistant ? 'message--assistant' : 'message--user'}`}>
      <header className="message__header">
        <span>{isAssistant ? 'Assistant' : 'You'}</span>
        <span className="message__time">{formatTimestamp(message.created_at)}</span>
      </header>
      <div className="message__content">
        {isAssistant ? <ReactMarkdown>{message.content}</ReactMarkdown> : <p>{message.content}</p>}
      </div>
    </article>
  )
}

export function MessageList({ messages, streamingContent, isStreaming, onSaveAsTask }: MessageListProps) {
  return (
    <section className="message-list" aria-live="polite">
      {messages.length > 0 && onSaveAsTask && (
        <div className="message-list__actions">
          <button type="button" className="btn btn-primary" onClick={onSaveAsTask}>
            Save as Task
          </button>
        </div>
      )}
      {messages.length === 0 ? (
        <div className="message-list__empty">
          <h3>Start a conversation</h3>
          <p>Ask anything, then save useful prompts as tasks later.</p>
        </div>
      ) : (
        messages.map((message) => <MessageBubble key={message.id} message={message} />)
      )}

      {isStreaming ? (
        <article className="message message--assistant message--streaming">
          <header className="message__header">
            <span>Assistant</span>
            <span className="message__time">Streaming...</span>
          </header>
          <div className="message__content">
            {streamingContent ? <ReactMarkdown>{streamingContent}</ReactMarkdown> : <p>Thinking...</p>}
          </div>
        </article>
      ) : null}
    </section>
  )
}
