import { useQuery } from '@tanstack/react-query'
import ReactMarkdown from 'react-markdown'
import { api } from '../../lib/api'
import { ErrorDisplay } from '../ErrorDisplay'

interface RunDetailProps {
  runId: string | null
}

export function RunDetail({ runId }: RunDetailProps) {
  const { data: run, isLoading, error } = useQuery({
    queryKey: ['run', runId],
    queryFn: () => api.getRun(runId!),
    enabled: !!runId,
  })

  if (!runId) return null
  if (isLoading) return <p className="muted">Loading run...</p>
  if (error) return <ErrorDisplay error={error} title="Run not found" />
  if (!run) return null

  return (
    <div className="run-detail">
      <h4>Run output</h4>
      <div className="run-detail__output">
        {run.output ? (
          <ReactMarkdown>{run.output}</ReactMarkdown>
        ) : (
          <p className="muted">(no output)</p>
        )}
      </div>
      <div className="run-detail__meta">
        <span>Status: {run.status}</span>
        {run.model && <span>Model: {run.model}</span>}
        <span>Cost: ${run.cost_usd.toFixed(4)}</span>
        <span>Tokens: {run.input_tokens} in / {run.output_tokens} out</span>
        <span>{new Date(run.created_at).toLocaleString()}</span>
      </div>
      {run.error && <p className="error-text">{run.error}</p>}
    </div>
  )
}
