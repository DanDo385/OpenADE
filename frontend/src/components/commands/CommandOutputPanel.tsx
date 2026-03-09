import { FormEvent, useState, useCallback } from 'react'
import { useMutation } from '@tanstack/react-query'
import { api } from '../../lib/api'
import type { CommandExecuteResponse } from '../../lib/api-types'
import { ErrorDisplay } from '../ErrorDisplay'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

const MAX_HISTORY = 20

export function CommandOutputPanel() {
  const [input, setInput] = useState('')
  const [confirm, setConfirm] = useState(true)
  const [lastResult, setLastResult] = useState<CommandExecuteResponse | null>(null)
  const [history, setHistory] = useState<string[]>([])

  const mutation = useMutation({
    mutationFn: () =>
      api.executeCommand({ input: input.trim(), confirm }),
    onSuccess: (data) => {
      setLastResult(data)
      setHistory((prev) => {
        const cmd = input.trim()
        if (!cmd) return prev
        const filtered = prev.filter((c) => c !== cmd)
        return [cmd, ...filtered].slice(0, MAX_HISTORY)
      })
    },
  })

  const onSubmit = useCallback(
    (e: FormEvent) => {
      e.preventDefault()
      if (!input.trim()) return
      mutation.mutate()
    },
    [input, mutation]
  )

  const onHistoryClick = useCallback((cmd: string) => {
    setInput(cmd)
  }, [])

  return (
    <Card>
      <CardHeader>
        <CardTitle>Run command</CardTitle>
        <p className="text-sm text-muted-foreground">
          Try /help, /echo hello, /date, /pwd
        </p>
      </CardHeader>
      <CardContent className="space-y-4">
        <form onSubmit={onSubmit} className="flex flex-col gap-3 sm:flex-row sm:items-end">
          <div className="flex-1 space-y-1">
            <Input
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="/help"
              className="font-mono"
            />
          </div>
          <div className="flex items-center gap-3">
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={confirm}
                onChange={(e) => setConfirm(e.target.checked)}
                className="h-4 w-4 rounded border-input"
              />
              Confirm
            </label>
            <Button
              type="submit"
              disabled={mutation.isPending || !input.trim()}
            >
              {mutation.isPending ? 'Running...' : 'Run'}
            </Button>
          </div>
        </form>

        {history.length > 0 && (
          <div className="space-y-2">
            <p className="text-sm font-medium text-muted-foreground">History</p>
            <ul className="flex flex-wrap gap-1">
              {history.map((cmd) => (
                <li key={cmd}>
                  <button
                    type="button"
                    onClick={() => onHistoryClick(cmd)}
                    className="rounded bg-muted px-2 py-1 text-xs font-mono hover:bg-accent"
                  >
                    {cmd.length > 40 ? `${cmd.slice(0, 37)}...` : cmd}
                  </button>
                </li>
              ))}
            </ul>
          </div>
        )}

        {mutation.error && (
          <ErrorDisplay error={mutation.error} title="Command failed" />
        )}

        {lastResult && (
          <div className="space-y-3 rounded-md border bg-muted/30 p-4 font-mono text-sm">
            <div>
              <p className="mb-1 font-semibold text-muted-foreground">stdout</p>
              <pre className="whitespace-pre-wrap break-words">
                {lastResult.output || '(empty)'}
              </pre>
            </div>
            {lastResult.stderr != null && lastResult.stderr !== '' && (
              <div>
                <p className="mb-1 font-semibold text-destructive">stderr</p>
                <pre className="whitespace-pre-wrap break-words text-destructive">
                  {lastResult.stderr}
                </pre>
              </div>
            )}
            <p className="text-muted-foreground">
              Exit code: {lastResult.exit_code} · {lastResult.duration_ms}ms
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
