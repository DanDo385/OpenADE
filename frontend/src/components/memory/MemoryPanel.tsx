import { FormEvent, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../../lib/api'
import { ErrorDisplay } from '../ErrorDisplay'

interface MemoryPanelProps {
  taskId: string
  taskName?: string
}

export function MemoryPanel({ taskId, taskName }: MemoryPanelProps) {
  const queryClient = useQueryClient()
  const [newKey, setNewKey] = useState('')
  const [newValue, setNewValue] = useState('')
  const [editKey, setEditKey] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')

  const { data, isLoading, error } = useQuery({
    queryKey: ['memory', taskId],
    queryFn: () => api.getMemory(taskId),
    enabled: !!taskId,
  })

  const setAllMutation = useMutation({
    mutationFn: (entries: Record<string, string>) => api.setMemory(taskId, { entries }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['memory', taskId] })
      setNewKey('')
      setNewValue('')
    },
  })

  const setKeyMutation = useMutation({
    mutationFn: ({ key, value }: { key: string; value: string }) =>
      api.setMemoryKey(taskId, key, value),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['memory', taskId] })
      setEditKey(null)
      setEditValue('')
    },
  })

  const entries = data?.entries ?? {}

  const onAdd = (e: FormEvent) => {
    e.preventDefault()
    if (!newKey.trim()) return
    const next = { ...entries, [newKey.trim()]: newValue.trim() }
    setAllMutation.mutate(next)
  }

  const onEdit = (key: string) => {
    setEditKey(key)
    setEditValue(entries[key] ?? '')
  }

  const onSaveEdit = (e: FormEvent) => {
    e.preventDefault()
    if (!editKey) return
    setKeyMutation.mutate({ key: editKey, value: editValue })
  }

  const onRemove = (key: string) => {
    const next = { ...entries }
    delete next[key]
    setAllMutation.mutate(next)
  }

  const keys = Object.keys(entries)

  return (
    <section className="memory-panel">
      <h3>{taskName ? `Memory: ${taskName}` : 'Task memory'}</h3>
      <p className="muted">Key-value store scoped to this task. Use for run context.</p>

      {isLoading && <p className="muted">Loading memory...</p>}
      {error && <ErrorDisplay error={error} title="Memory error" onRetry={() => queryClient.invalidateQueries({ queryKey: ['memory', taskId] })} />}

      {!isLoading && !error && (
        <>
          {keys.length === 0 && (
            <p className="empty-state">No memory entries yet. Add one below.</p>
          )}

          {keys.length > 0 && (
            <ul className="memory-list">
              {keys.map((k) => (
                <li key={k} className="memory-item">
                  {editKey === k ? (
                    <form onSubmit={onSaveEdit} className="memory-item__edit">
                      <input
                        type="text"
                        value={editKey}
                        readOnly
                        className="memory-item__key"
                        aria-label="Key"
                      />
                      <input
                        type="text"
                        value={editValue}
                        onChange={(e) => setEditValue(e.target.value)}
                        placeholder="Value"
                        aria-label="Value"
                      />
                      <button type="submit" className="btn btn-primary">Save</button>
                      <button type="button" className="btn" onClick={() => setEditKey(null)}>Cancel</button>
                    </form>
                  ) : (
                    <>
                      <span className="memory-item__key">{k}</span>
                      <span className="memory-item__value">{entries[k]}</span>
                      <button type="button" className="btn" onClick={() => onEdit(k)}>Edit</button>
                      <button type="button" className="btn" onClick={() => onRemove(k)}>Remove</button>
                    </>
                  )}
                </li>
              ))}
            </ul>
          )}

          <form onSubmit={onAdd} className="memory-add-form">
            <input
              type="text"
              value={newKey}
              onChange={(e) => setNewKey(e.target.value)}
              placeholder="Key"
              aria-label="New key"
            />
            <input
              type="text"
              value={newValue}
              onChange={(e) => setNewValue(e.target.value)}
              placeholder="Value"
              aria-label="New value"
            />
            <button type="submit" className="btn btn-primary" disabled={setAllMutation.isPending || !newKey.trim()}>
              Add
            </button>
          </form>
        </>
      )}
    </section>
  )
}
