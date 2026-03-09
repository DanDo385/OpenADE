import { FormEvent, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { APIError, api } from '../lib/api'

interface ProviderModalProps {
  open: boolean
  blocking: boolean
  onClose: () => void
}

function getErrorMessage(error: unknown): string {
  if (error instanceof APIError) {
    return error.message
  }
  if (error instanceof Error) {
    return error.message
  }
  return 'failed to save provider configuration'
}

export function ProviderModal({ open, blocking, onClose }: ProviderModalProps) {
  const queryClient = useQueryClient()
  const [apiKey, setAPIKey] = useState('')
  const [baseURL, setBaseURL] = useState('')
  const [defaultModel, setDefaultModel] = useState('gpt-4o-mini')
  const [submitError, setSubmitError] = useState<string | null>(null)

  const mutation = useMutation({
    mutationFn: () =>
      api.saveProvider('openai', {
        api_key: apiKey.trim(),
        base_url: baseURL.trim() || undefined,
        default_model: defaultModel.trim() || undefined,
      }),
    onSuccess: async () => {
      setSubmitError(null)
      setAPIKey('')
      await queryClient.invalidateQueries({ queryKey: ['providers'] })
      onClose()
    },
    onError: (error) => {
      setSubmitError(getErrorMessage(error))
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

  if (!open) {
    return null
  }

  return (
    <div className="modal-backdrop" role="presentation">
      <div className="modal" role="dialog" aria-modal="true" aria-labelledby="provider-modal-title">
        <h2 id="provider-modal-title">Configure OpenAI provider</h2>
        <p className="muted">Add a provider API key to enable chat and task runs.</p>

        <form onSubmit={onSubmit} className="provider-form">
          <label>
            API Key
            <input
              type="password"
              value={apiKey}
              onChange={(event) => setAPIKey(event.target.value)}
              placeholder="sk-..."
              autoComplete="off"
            />
          </label>

          <label>
            Base URL (optional)
            <input
              type="url"
              value={baseURL}
              onChange={(event) => setBaseURL(event.target.value)}
              placeholder="https://api.openai.com/v1"
            />
          </label>

          <label>
            Default model
            <input
              type="text"
              value={defaultModel}
              onChange={(event) => setDefaultModel(event.target.value)}
              placeholder="gpt-4o-mini"
            />
          </label>

          {submitError ? <p className="error-text">{submitError}</p> : null}

          <div className="provider-form__actions">
            {!blocking ? (
              <button type="button" className="btn" onClick={onClose} disabled={mutation.isPending}>
                Cancel
              </button>
            ) : null}
            <button type="submit" className="btn btn-primary" disabled={mutation.isPending}>
              {mutation.isPending ? 'Saving...' : 'Save provider'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
