export interface ErrorDisplayProps {
  error: unknown
  /** Optional title override */
  title?: string
  /** Optional retry callback */
  onRetry?: () => void
}

function getErrorMessage(error: unknown): string {
  if (error && typeof error === 'object' && 'message' in error && typeof (error as { message: unknown }).message === 'string') {
    return (error as { message: string }).message
  }
  if (error instanceof Error) {
    return error.message
  }
  return String(error ?? 'An error occurred')
}

function getErrorCode(error: unknown): string | undefined {
  if (error && typeof error === 'object' && 'code' in error && typeof (error as { code: unknown }).code === 'string') {
    return (error as { code: string }).code
  }
  return undefined
}

function getStatusCode(error: unknown): number | undefined {
  if (error && typeof error === 'object' && 'status' in error && typeof (error as { status: unknown }).status === 'number') {
    return (error as { status: number }).status
  }
  return undefined
}

/** Shared error display that maps API error shape { error: { code, message } } to UI */
export function ErrorDisplay({ error, title = 'Error', onRetry }: ErrorDisplayProps) {
  if (!error) {
    return null
  }

  const message = getErrorMessage(error)
  const code = getErrorCode(error)
  const status = getStatusCode(error)

  return (
    <div className="error-display" role="alert">
      <h3 className="error-display__title">{title}</h3>
      <p className="error-display__message">{message}</p>
      {(code || status) && (
        <p className="error-display__meta muted">
          {code && <span>Code: {code}</span>}
          {code && status && ' · '}
          {status && <span>HTTP {status}</span>}
        </p>
      )}
      {onRetry && (
        <button type="button" className="btn btn-primary" onClick={onRetry}>
          Retry
        </button>
      )}
    </div>
  )
}
