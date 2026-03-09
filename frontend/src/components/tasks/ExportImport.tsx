import { FormEvent, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { ExportBundle, Task } from '../../lib/api-types'
import { api } from '../../lib/api'
import { ErrorDisplay } from '../ErrorDisplay'

interface ExportImportProps {
  tasks: Task[]
  onTaskImported?: (task: Task) => void
}

export function ExportImport({ tasks, onTaskImported }: ExportImportProps) {
  const queryClient = useQueryClient()
  const [exportTaskId, setExportTaskId] = useState('')
  const [importJson, setImportJson] = useState('')
  const [exportResult, setExportResult] = useState<string | null>(null)
  const [importError, setImportError] = useState<string | null>(null)
  const [importSuccess, setImportSuccess] = useState<string | null>(null)

  const exportMutation = useMutation({
    mutationFn: (id: string) => api.exportTask(id),
    onSuccess: (bundle: ExportBundle) => {
      setExportResult(JSON.stringify(bundle, null, 2))
    },
  })

  const importMutation = useMutation({
    mutationFn: (bundle: ExportBundle) => api.importTask(bundle),
    onSuccess: (task) => {
      setImportError(null)
      setImportSuccess(`Task "${task.name}" imported successfully.`)
      setImportJson('')
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
      onTaskImported?.(task)
      setTimeout(() => setImportSuccess(null), 3000)
    },
    onError: (err) => {
      setImportError(err instanceof Error ? err.message : String(err))
    },
  })

  const onExport = (e: FormEvent) => {
    e.preventDefault()
    if (!exportTaskId) return
    setExportResult(null)
    exportMutation.mutate(exportTaskId)
  }

  const onImport = (e: FormEvent) => {
    e.preventDefault()
    setImportError(null)
    try {
      const parsed = JSON.parse(importJson) as ExportBundle
      if (!parsed?.task?.name) {
        setImportError('Invalid bundle: task.name is required')
        return
      }
      importMutation.mutate(parsed)
    } catch {
      setImportError('Invalid JSON')
    }
  }

  const copyExport = () => {
    if (exportResult) {
      navigator.clipboard.writeText(exportResult)
    }
  }

  return (
    <section className="export-import">
      <h3>Export / Import</h3>

      <div className="export-import__export">
        <h4>Export task</h4>
        {tasks.length === 0 ? (
          <p className="empty-state">No tasks to export. Create a task first.</p>
        ) : (
          <form onSubmit={onExport}>
            <select
              value={exportTaskId}
              onChange={(e) => setExportTaskId(e.target.value)}
              aria-label="Select task to export"
            >
              <option value="">Select a task...</option>
              {tasks.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.name}
                </option>
              ))}
            </select>
            <button type="submit" className="btn btn-primary" disabled={exportMutation.isPending || !exportTaskId}>
              {exportMutation.isPending ? 'Exporting...' : 'Export'}
            </button>
          </form>
        )}

        {exportMutation.error && (
          <ErrorDisplay error={exportMutation.error} title="Export failed" />
        )}

        {exportResult && (
          <div className="export-import__result">
            <div className="export-import__actions">
              <button type="button" className="btn" onClick={copyExport}>
                Copy to clipboard
              </button>
            </div>
            <pre className="export-import__json">{exportResult}</pre>
          </div>
        )}
      </div>

      <div className="export-import__import">
        <h4>Import task</h4>
        <form onSubmit={onImport}>
          <textarea
            value={importJson}
            onChange={(e) => setImportJson(e.target.value)}
            placeholder='Paste export JSON (e.g. {"bundle_version":"0.1","task":{...}})...'
            rows={6}
            aria-label="Import JSON"
          />
          {importError && <p className="error-text">{importError}</p>}
          {importSuccess && <p className="success-text">{importSuccess}</p>}
          <button type="submit" className="btn btn-primary" disabled={importMutation.isPending || !importJson.trim()}>
            {importMutation.isPending ? 'Importing...' : 'Import'}
          </button>
        </form>
      </div>
    </section>
  )
}
