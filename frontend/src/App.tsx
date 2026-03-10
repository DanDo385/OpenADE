import { Component, type ErrorInfo, type ReactNode, useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { APIError, API_BASE_URL, api } from './lib/api'
import { useUIStore } from './lib/store'
import { useConversations } from './hooks/useConversations'
import { useMessages } from './hooks/useMessages'
import { useStreaming } from './hooks/useStreaming'
import { useTasks } from './hooks/useTasks'
import { useRuns } from './hooks/useRuns'
import { ProviderModal } from './components/ProviderModal'
import { ProviderSettings } from './components/settings/ProviderSettings'
import { MCPServerSettings } from './components/settings/MCPServerSettings'
import { ConversationList } from './components/chat/ConversationList'
import { MessageList } from './components/chat/MessageList'
import { MessageInput } from './components/chat/MessageInput'
import { TaskLibrary } from './components/tasks/TaskLibrary'
import { TaskWizard } from './components/tasks/TaskWizard'
import { TaskEditor } from './components/tasks/TaskEditor'
import { RunPanel } from './components/tasks/RunPanel'
import { SchedulePanel } from './components/tasks/SchedulePanel'
import { MemoryPanel } from './components/memory/MemoryPanel'
import { ExportImport } from './components/tasks/ExportImport'
import { RunDetail } from './components/runs/RunDetail'
import { ErrorDisplay } from './components/ErrorDisplay'
import { ObjectivePanel } from './components/chat/ObjectivePanel'
import { AgentLibrary } from './components/agents/AgentLibrary'
import { CommandOutputPanel } from './components/commands/CommandOutputPanel'

interface ErrorBoundaryState {
  hasError: boolean
  message: string
}

class AppErrorBoundary extends Component<{ children: ReactNode }, ErrorBoundaryState> {
  override state: ErrorBoundaryState = {
    hasError: false,
    message: '',
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, message: error.message }
  }

  override componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Unhandled app error:', error, errorInfo)
  }

  override render() {
    if (this.state.hasError) {
      return (
        <main className="fatal-error">
          <h1>App crashed</h1>
          <p>{this.state.message || 'Unexpected error in UI tree.'}</p>
          <button type="button" className="btn btn-primary" onClick={() => window.location.reload()}>
            Reload
          </button>
        </main>
      )
    }

    return this.props.children
  }
}

function AppShell() {
  const {
    activePanel,
    setActivePanel,
    activeConversationId,
    setActiveConversationId,
    activeTaskId,
    setActiveTaskId,
    activeRunId,
    setActiveRunId,
    theme,
    toggleTheme,
    providerModalOpen,
    setProviderModalOpen,
    taskWizardOpen,
    taskWizardConversationId,
    setTaskWizardOpen,
  } = useUIStore()

  const {
    conversations,
    isLoading: isConversationsLoading,
    createConversation,
    isCreating,
    deleteConversation,
  } = useConversations()

  const [taskSearch, setTaskSearch] = useState('')
  const { tasks, isLoading: isTasksLoading, deleteTask } = useTasks(taskSearch || undefined)
  const { runs, isLoading: isRunsLoading } = useRuns()

  const { messages, isLoading: isMessagesLoading } = useMessages(activeConversationId)

  const { isStreaming, streamingContent, streamError, sendMessage, cancel } = useStreaming()

  const providersQuery = useQuery({
    queryKey: ['providers'],
    queryFn: api.listProviders,
  })

  const providerConfigured = (providersQuery.data ?? []).some((provider) => provider.configured)
  const providerMissing = providersQuery.isSuccess && !providerConfigured
  const showProviderModal = providerModalOpen || providerMissing

  useEffect(() => {
    document.documentElement.dataset.theme = theme
  }, [theme])

  useEffect(() => {
    if (activeConversationId && conversations.some((c) => c.id === activeConversationId)) {
      return
    }
    setActiveConversationId(conversations[0]?.id ?? null)
  }, [activeConversationId, conversations, setActiveConversationId])

  useEffect(() => {
    if (activeTaskId && tasks.some((t) => t.id === activeTaskId)) {
      return
    }
    setActiveTaskId(tasks[0]?.id ?? null)
  }, [activeTaskId, tasks, setActiveTaskId])

  const onCreateConversation = async () => {
    try {
      const conversation = await createConversation()
      setActiveConversationId(conversation.id)
    } catch (error) {
      console.error('Failed to create conversation:', error)
    }
  }

  const onDeleteConversation = async (id: string) => {
    const approved = window.confirm('Delete this conversation?')
    if (!approved) {
      return
    }

    try {
      await deleteConversation(id)
    } catch (error) {
      console.error('Failed to delete conversation:', error)
    }
  }

  const onSendMessage = async (content: string) => {
    if (providerMissing) {
      setProviderModalOpen(true)
      return
    }

    let targetConversationId = activeConversationId

    if (!targetConversationId) {
      const conversation = await createConversation()
      targetConversationId = conversation.id
      setActiveConversationId(conversation.id)
    }

    try {
      await sendMessage({
        conversationId: targetConversationId,
        content,
        onUnauthorized: () => setProviderModalOpen(true),
      })
    } catch (error) {
      if (error instanceof APIError && error.status === 401) {
        return
      }
      console.error('Streaming failed:', error)
    }
  }

  const chatBlocked = providersQuery.isLoading || providerMissing

  const onDeleteTask = async (id: string) => {
    if (!window.confirm('Delete this task?')) return
    try {
      await deleteTask(id)
      if (activeTaskId === id) setActiveTaskId(null)
    } catch (e) {
      console.error('Failed to delete task:', e)
    }
  }

  const selectedTask = activeTaskId ? tasks.find((t) => t.id === activeTaskId) : null

  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="topbar__brand">
          <h1>OpenADE</h1>
          <small>API: {API_BASE_URL || '(proxy)'}</small>
        </div>

        <nav className="topbar__panels" aria-label="Panel selector">
          <button
            type="button"
            className={`btn ${activePanel === 'chat' ? 'btn-primary' : ''}`}
            onClick={() => setActivePanel('chat')}
          >
            Chat
          </button>
          <button
            type="button"
            className={`btn ${activePanel === 'tasks' ? 'btn-primary' : ''}`}
            onClick={() => setActivePanel('tasks')}
          >
            Tasks
          </button>
          <button
            type="button"
            className={`btn ${activePanel === 'runs' ? 'btn-primary' : ''}`}
            onClick={() => setActivePanel('runs')}
          >
            Runs
          </button>
          <button
            type="button"
            className={`btn ${activePanel === 'agents' ? 'btn-primary' : ''}`}
            onClick={() => setActivePanel('agents')}
          >
            Agents
          </button>
          <button
            type="button"
            className={`btn ${activePanel === 'commands' ? 'btn-primary' : ''}`}
            onClick={() => setActivePanel('commands')}
          >
            Commands
          </button>
          <button
            type="button"
            className={`btn ${activePanel === 'settings' ? 'btn-primary' : ''}`}
            onClick={() => setActivePanel('settings')}
          >
            Settings
          </button>
        </nav>

        <div className="topbar__actions">
          <button type="button" className="btn" onClick={() => setProviderModalOpen(true)}>
            Provider
          </button>
          <button type="button" className="btn" onClick={toggleTheme}>
            {theme === 'dark' ? 'Light' : 'Dark'}
          </button>
        </div>
      </header>

      <div
        className={
          ['runs', 'agents', 'commands', 'settings'].includes(activePanel)
            ? 'workspace workspace--full'
            : 'workspace'
        }
      >
        {activePanel === 'chat' && (
          <ConversationList
            conversations={conversations}
            activeConversationId={activeConversationId}
            isLoading={isConversationsLoading}
            onCreate={onCreateConversation}
            onSelect={setActiveConversationId}
            onDelete={onDeleteConversation}
          />
        )}

        {activePanel === 'tasks' && (
          <TaskLibrary
            tasks={tasks}
            activeTaskId={activeTaskId}
            isLoading={isTasksLoading}
            searchQuery={taskSearch}
            onSearchChange={setTaskSearch}
            onSelectTask={setActiveTaskId}
            onDeleteTask={onDeleteTask}
            onCreateTask={() => setTaskWizardOpen(true, null)}
          />
        )}

        <main className="main-panel">
          {providersQuery.error && (
            <ErrorDisplay
              error={providersQuery.error}
              title="Provider error"
              onRetry={() => providersQuery.refetch()}
            />
          )}

          {activePanel === 'chat' && (
            <>
              <ObjectivePanel conversationId={activeConversationId} />
              {isMessagesLoading ? <p className="muted">Loading conversation...</p> : null}
              <MessageList
                messages={messages}
                streamingContent={streamingContent}
                isStreaming={isStreaming}
                onSaveAsTask={
                  activeConversationId && messages.length > 0
                    ? () => setTaskWizardOpen(true, activeConversationId)
                    : undefined
                }
              />
              {streamError ? <p className="error-text">{streamError}</p> : null}
              <MessageInput
                disabled={chatBlocked || isCreating}
                isStreaming={isStreaming}
                onSend={onSendMessage}
              />
              {isStreaming ? (
                <button type="button" className="btn" onClick={cancel}>
                  Cancel stream
                </button>
              ) : null}
            </>
          )}

          {activePanel === 'tasks' && (
            <section className="tasks-detail">
              {!selectedTask ? (
                <div className="empty-state-block">
                  <p className="empty-state">Select a task</p>
                  <p className="muted">Choose a task from the list to view memory and export/import.</p>
                </div>
              ) : (
                <>
                  <RunPanel task={selectedTask} />
                  <SchedulePanel task={selectedTask} />
                  <TaskEditor task={selectedTask} />
                  <MemoryPanel taskId={selectedTask.id} taskName={selectedTask.name} />
                  <ExportImport tasks={tasks} onTaskImported={(t) => setActiveTaskId(t.id)} />
                </>
              )}
            </section>
          )}

          {activePanel === 'runs' && (
            <section className="runs-panel">
              <h2>Run history</h2>
              {isRunsLoading ? <p className="muted">Loading runs...</p> : null}
              {!isRunsLoading && runs.length === 0 && (
                <div className="empty-state-block">
                  <p className="empty-state">No runs yet.</p>
                  <p className="muted">Run a task to see history.</p>
                </div>
              )}
              {!isRunsLoading && runs.length > 0 && (
                <>
                  <ul className="runs-list">
                    {runs.map((r) => (
                      <li
                        key={r.id}
                        className={activeRunId === r.id ? 'runs-list__item--active' : ''}
                      >
                        <button
                          type="button"
                          className="runs-list__trigger"
                          onClick={() => setActiveRunId(activeRunId === r.id ? null : r.id)}
                        >
                          <span>{r.task_id.slice(0, 8)}…</span>
                          <span>{r.status}</span>
                          <span>${r.cost_usd.toFixed(4)}</span>
                          <span>{new Date(r.created_at).toLocaleString()}</span>
                        </button>
                      </li>
                    ))}
                  </ul>
                  <RunDetail runId={activeRunId} />
                </>
              )}
            </section>
          )}

          {activePanel === 'agents' && (
            <section className="agents-panel">
              <AgentLibrary />
            </section>
          )}

          {activePanel === 'commands' && (
            <section className="commands-panel">
              <CommandOutputPanel />
            </section>
          )}

          {activePanel === 'settings' && (
            <section className="settings-panel space-y-6">
              <ProviderSettings />
              <MCPServerSettings />
            </section>
          )}
        </main>
      </div>

      <ProviderModal
        open={showProviderModal}
        blocking={providerMissing}
        onClose={() => setProviderModalOpen(false)}
      />

      <TaskWizard
        open={taskWizardOpen}
        conversationId={taskWizardConversationId}
        onClose={() => setTaskWizardOpen(false)}
        onSaved={(taskId) => {
          setActivePanel('tasks')
          setActiveTaskId(taskId)
        }}
      />
    </div>
  )
}

export default function App() {
  return (
    <AppErrorBoundary>
      <AppShell />
    </AppErrorBoundary>
  )
}
