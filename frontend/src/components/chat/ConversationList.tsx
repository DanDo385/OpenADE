import type { Conversation } from '../../lib/api-types'

interface ConversationListProps {
  conversations: Conversation[]
  activeConversationId: string | null
  isLoading: boolean
  onCreate: () => void
  onSelect: (id: string) => void
  onDelete: (id: string) => void
}

function formatTitle(conversation: Conversation): string {
  const title = conversation.title.trim()
  if (title) {
    return title
  }
  return 'Untitled conversation'
}

function formatUpdatedAt(updatedAt: string): string {
  const date = new Date(updatedAt)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  return new Intl.DateTimeFormat(undefined, {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  }).format(date)
}

export function ConversationList({
  conversations,
  activeConversationId,
  isLoading,
  onCreate,
  onSelect,
  onDelete,
}: ConversationListProps) {
  return (
    <aside className="conversation-list">
      <div className="conversation-list__header">
        <h2>Conversations</h2>
        <button type="button" className="btn btn-primary" onClick={onCreate}>
          New
        </button>
      </div>

      {isLoading ? <p className="muted">Loading conversations...</p> : null}

      {!isLoading && conversations.length === 0 ? (
        <div className="empty-state-block">
          <p className="empty-state">No conversations yet.</p>
          <p className="muted">Click New to start a conversation.</p>
        </div>
      ) : null}

      <ul className="conversation-list__items">
        {conversations.map((conversation) => {
          const isActive = conversation.id === activeConversationId
          return (
            <li key={conversation.id}>
              <button
                type="button"
                className={`conversation-item ${isActive ? 'conversation-item--active' : ''}`}
                onClick={() => onSelect(conversation.id)}
              >
                <span className="conversation-item__title">{formatTitle(conversation)}</span>
                <span className="conversation-item__meta">{formatUpdatedAt(conversation.updated_at)}</span>
              </button>
              <button
                type="button"
                className="conversation-item__delete"
                onClick={() => onDelete(conversation.id)}
                aria-label="Delete conversation"
                title="Delete conversation"
              >
                Delete
              </button>
            </li>
          )
        })}
      </ul>
    </aside>
  )
}
