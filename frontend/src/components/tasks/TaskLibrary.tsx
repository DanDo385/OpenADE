import type { Task } from '../../lib/api-types'

interface TaskLibraryProps {
  tasks: Task[]
  activeTaskId: string | null
  isLoading: boolean
  searchQuery?: string
  onSearchChange?: (q: string) => void
  onSelectTask: (id: string) => void
  onDeleteTask?: (id: string) => void
  onCreateTask?: () => void
}

function formatUpdatedAt(updatedAt: string): string {
  const date = new Date(updatedAt)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat(undefined, {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  }).format(date)
}

export function TaskLibrary({
  tasks,
  activeTaskId,
  isLoading,
  searchQuery = '',
  onSearchChange,
  onSelectTask,
  onDeleteTask,
  onCreateTask,
}: TaskLibraryProps) {
  return (
    <aside className="task-library">
      <div className="task-library__header">
        <h2>Tasks</h2>
        {onCreateTask && (
          <button type="button" className="btn btn-primary" onClick={onCreateTask}>
            New
          </button>
        )}
      </div>
      {onSearchChange && (
        <input
          type="search"
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          placeholder="Search tasks..."
          className="task-library__search"
          aria-label="Search tasks"
        />
      )}

      {isLoading && <p className="muted">Loading tasks...</p>}

      {!isLoading && tasks.length === 0 && (
        <div className="empty-state-block">
          <p className="empty-state">No tasks yet.</p>
          <p className="muted">Save a conversation as a task, or import one.</p>
        </div>
      )}

      {!isLoading && tasks.length > 0 && (
        <ul className="task-library__items">
          {tasks.map((task) => {
            const isActive = task.id === activeTaskId
            return (
              <li key={task.id}>
                <button
                  type="button"
                  className={`task-item ${isActive ? 'task-item--active' : ''}`}
                  onClick={() => onSelectTask(task.id)}
                >
                  <span className="task-item__name">{task.name}</span>
                  <span className="task-item__meta">{formatUpdatedAt(task.updated_at)}</span>
                </button>
                {onDeleteTask && (
                  <button
                    type="button"
                    className="task-item__delete"
                    onClick={() => onDeleteTask(task.id)}
                    aria-label="Delete task"
                    title="Delete task"
                  >
                    Delete
                  </button>
                )}
              </li>
            )
          })}
        </ul>
      )}
    </aside>
  )
}
