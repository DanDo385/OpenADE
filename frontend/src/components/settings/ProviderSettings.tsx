import { FormEvent, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { APIError, api } from '../../lib/api'
import { ErrorDisplay } from '../ErrorDisplay'

function getErrorMessage(error: unknown): string {
  if (error instanceof APIError) {
    return error.message
  }
  if (error instanceof Error) {
    return error.message
  }
  return 'Failed to save provider configuration'
}

export function ProviderSettings() {
  const queryClient = useQueryClient()
  const [apiKey, setAPIKey] = useState('')
  const [baseURL, setBaseURL] = useState('')
  const [defaultModel, setDefaultModel] = useState('gpt-4o-mini')
  const [submitError, setSubmitError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const { data: providers = [], isLoading, error } = useQuery({
    queryKey: ['providers'],
    queryFn: api.listProviders,
  })

  const mutation = useMutation({
    mutationFn: () =>
      api.saveProvider('openai', {
        api_key: apiKey.trim(),
        base_url: baseURL.trim() || undefined,
        default_model: defaultModel.trim() || undefined,
      }),
    onSuccess: async () => {
      setSubmitError(null)
      setSuccessMessage('Provider saved successfully.')
      setAPIKey('')
      await queryClient.invalidateQueries({ queryKey: ['providers'] })
      setTimeout(() => setSuccessMessage(null), 3000)
    },
    onError: (err) => {
      setSubmitError(getErrorMessage(err))
    },
  })

  const onSubmit = (event: FormEvent) => {
    event.preventDefault()
    if (!apiKey.trim()) {
      setSubmitError('API key is required')
      return
    }
    mutation.mutate()
  }

  const configured = providers.some((p) => p.provider === 'openai' && p.configured)
  const openaiProvider = providers.find((p) => p.provider === 'openai')

  return (
    <section className="provider-settings">
      <h2>Provider settings</h2>
      <p className="muted">Configure your LLM provider API key to enable chat and task runs.</p>

      {isLoading && <p className="muted">Loading provider status...</p>}
      {error && <ErrorDisplay error={error} title="Provider error" onRetry={() => queryClient.invalidateQueries({ queryKey: ['providers'] })} />}

      {!isLoading && (
        <div className="provider-settings__status">
          <p>
            OpenAI: <strong>{configured ? 'Configured' : 'Not configured'}</strong>
            {openaiProvider?.default_model && (
              <span className="muted"> · Model: {openaiProvider.default_model}</span>
            )}
          </p>
        </div>
      )}

      <form onSubmit={onSubmit} className="provider-form">
        <label>
          API Key
          <input
            type="password"
            value={apiKey}
            onChange={(e) => setAPIKey(e.target.value)}
            placeholder="sk-..."
            autoComplete="off"
          />
        </label>

        <label>
          Base URL (optional)
          <input
            type="url"
            value={baseURL}
            onChange={(e) => setBaseURL(e.target.value)}
            placeholder="https://api.openai.com/v1"
          />
        </label>

        <label>
          Default model
          <input
            type="text"
            value={defaultModel}
            onChange={(e) => setDefaultModel(e.target.value)}
            placeholder="gpt-4o-mini"
          />
        </label>

        {submitError && <p className="error-text">{submitError}</p>}
        {successMessage && <p className="success-text">{successMessage}</p>}

        <button type="submit" className="btn btn-primary" disabled={mutation.isPending}>
          {mutation.isPending ? 'Saving...' : 'Save provider'}
        </button>
      </form>
    </section>
  )
}
