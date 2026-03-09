import { FormEvent, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { api } from '../../lib/api'
import { ErrorDisplay } from '../ErrorDisplay'

export function CommandOutputPanel() {
  const [input, setInput] = useState('')
  const [lastOutput, setLastOutput] = useState<string | null>(null)

  const mutation = useMutation({
    mutationFn: () =>
      api.executeCommand({ input: input.trim(), confirm: true }),
    onSuccess: (data) => {
      setLastOutput(
        `Output:\n${data.output}\nExit code: ${data.exit_code}\nDuration: ${data.duration_ms}ms`,
      )
    },
  })

  const onSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (!input.trim()) return
    mutation.mutate()
  }

  return (
    <section className="command-output-panel">
      <h2>Run command</h2>
      <p className="muted">Try /help, /echo hello, /date, /pwd</p>

      <form onSubmit={onSubmit}>
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="/help"
          className="command-output-panel__input"
        />
        <button
          type="submit"
          className="btn btn-primary"
          disabled={mutation.isPending || !input.trim()}
        >
          {mutation.isPending ? 'Running...' : 'Run'}
        </button>
      </form>

      {mutation.error && <ErrorDisplay error={mutation.error} title="Command failed" />}

      {lastOutput && (
        <pre className="command-output-panel__output">{lastOutput}</pre>
      )}
    </section>
  )
}
