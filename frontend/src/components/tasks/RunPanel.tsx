import { FormEvent, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import ReactMarkdown from 'react-markdown'
import type { InputField, Task } from '../../lib/api-types'
import { api } from '../../lib/api'
import { ErrorDisplay } from '../ErrorDisplay'

interface RunPanelProps {
  task: Task
}

function InputFieldControl({
  field,
  value,
  onChange,
}: {
  field: InputField
  value: string
  onChange: (v: string) => void
}) {
  if (field.type === 'boolean') {
    return (
      <select value={value} onChange={(e) => onChange(e.target.value)}>
        <option value="true">Yes</option>
        <option value="false">No</option>
      </select>
    )
  }
  if (field.type === 'select' && field.options?.length) {
    return (
      <select value={value} onChange={(e) => onChange(e.target.value)}>
        <option value="">Select...</option>
        {field.options.map((o) => (
          <option key={o} value={o}>
            {o}
          </option>
        ))}
      </select>
    )
  }
  if (field.type === 'number') {
    return (
      <input
        type="number"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={field.label}
      />
    )
  }
  return (
    <input
      type="text"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={field.label}
    />
  )
}

export function RunPanel({ task }: RunPanelProps) {
  const queryClient = useQueryClient()
  const [inputs, setInputs] = useState<Record<string, string>>(() => {
    const init: Record<string, string> = {}
    for (const f of task.input_schema || []) {
      init[f.key] = f.default ?? ''
    }
    return init
  })

  const runMutation = useMutation({
    mutationFn: () =>
      api.runTask(task.id, {
        inputs: Object.fromEntries(
          Object.entries(inputs).map(([k, v]) => [k, v === '' ? undefined : v])
        ),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['runs'] })
    },
  })

  const run = runMutation.data

  const onSubmit = (e: FormEvent) => {
    e.preventDefault()
    runMutation.mutate()
  }

  const schema = task.input_schema ?? []

  return (
    <section className="run-panel">
      <h3>Run: {task.name}</h3>

      <form onSubmit={onSubmit} className="run-panel__form">
        {schema.map((field) => (
          <label key={field.key}>
            {field.label}
            <InputFieldControl
              field={field}
              value={inputs[field.key] ?? ''}
              onChange={(v) => setInputs((prev) => ({ ...prev, [field.key]: v }))}
            />
          </label>
        ))}

        {runMutation.error && (
          <ErrorDisplay
            error={runMutation.error}
            title="Run failed"
            onRetry={() => runMutation.reset()}
          />
        )}

        <button
          type="submit"
          className="btn btn-primary"
          disabled={runMutation.isPending}
        >
          {runMutation.isPending ? 'Running...' : 'Run'}
        </button>
      </form>

      {run && (
        <div className="run-panel__output">
          <h4>Output</h4>
          <div className="run-panel__output-body">
            {run.output ? <ReactMarkdown>{run.output}</ReactMarkdown> : <p className="muted">(empty)</p>}
          </div>
          <div className="run-panel__meta">
            <span>Status: {run.status}</span>
            {run.model && <span>Model: {run.model}</span>}
            {run.cost_usd > 0 && <span>Cost: ${run.cost_usd.toFixed(4)}</span>}
            {run.input_tokens > 0 && (
              <span>Tokens: {run.input_tokens} in / {run.output_tokens} out</span>
            )}
          </div>
        </div>
      )}
    </section>
  )
}
