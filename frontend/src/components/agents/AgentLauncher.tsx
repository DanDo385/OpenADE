import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import type { Agent } from '../../lib/api-types'
import { api } from '../../lib/api'
import { ErrorDisplay } from '../ErrorDisplay'

interface AgentLauncherProps {
  agent: Agent
}

export function AgentLauncher({ agent }: AgentLauncherProps) {
  const [output, setOutput] = useState<string | null>(null)

  const runMutation = useMutation({
    mutationFn: () => api.runAgent(agent.id, {}),
    onSuccess: (data) => {
      setOutput(data.output)
    },
  })

  return (
    <div className="agent-launcher">
      <div className="agent-launcher__header">
        <h3>{agent.name}</h3>
        <p className="muted">{agent.description}</p>
        <button
          type="button"
          className="btn btn-primary"
          onClick={() => runMutation.mutate()}
          disabled={runMutation.isPending || !agent.enabled}
        >
          {runMutation.isPending ? 'Running...' : 'Run'}
        </button>
      </div>

      {runMutation.error && <ErrorDisplay error={runMutation.error} title="Run failed" />}

      {output && (
        <div className="agent-launcher__output">
          <pre>{output}</pre>
        </div>
      )}
    </div>
  )
}
