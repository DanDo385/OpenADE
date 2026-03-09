import { FormEvent, useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { CreateTaskRequest, InputField } from '../../lib/api-types'
import { api } from '../../lib/api'
import { extractVariables, variablesToInputSchema } from '../../lib/templates'
import { ErrorDisplay } from '../ErrorDisplay'

const STEPS = ['name', 'template', 'inputs', 'confirm'] as const

interface TaskWizardProps {
  open: boolean
  conversationId: string | null
  onClose: () => void
  onSaved: (taskId: string) => void
}

export function TaskWizard({ open, conversationId, onClose, onSaved }: TaskWizardProps) {
  const queryClient = useQueryClient()
  const [stepIndex, setStepIndex] = useState(0)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [promptTemplate, setPromptTemplate] = useState('')
  const [inputSchema, setInputSchema] = useState<InputField[]>([])

  const {
    data: draft,
    isLoading: isDraftLoading,
    isError: isDraftError,
    refetch: refetchDraft,
  } = useQuery({
    queryKey: ['draft', conversationId],
    queryFn: () => api.draftTaskFromConversation(conversationId!),
    enabled: open && !!conversationId,
    retry: false,
  })

  useEffect(() => {
    if (draft) {
      setName(draft.name || '')
      setDescription(draft.description || '')
      setPromptTemplate(draft.prompt_template || '')
      setInputSchema(draft.input_schema || [])
    }
  }, [draft])

  useEffect(() => {
    if (!open) {
      setStepIndex(0)
      if (!conversationId) {
        setName('')
        setDescription('')
        setPromptTemplate('')
        setInputSchema([])
      }
    }
  }, [open, conversationId])

  const createMutation = useMutation({
    mutationFn: (body: CreateTaskRequest) => api.createTask(body),
    onSuccess: (task) => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
      onSaved(task.id)
      onClose()
    },
  })

  const handleTemplateChange = (value: string) => {
    setPromptTemplate(value)
    const vars = extractVariables(value)
    setInputSchema((prev) => variablesToInputSchema(vars, prev))
  }

  const step = STEPS[stepIndex]

  const onNext = () => {
    if (stepIndex < STEPS.length - 1) setStepIndex(stepIndex + 1)
  }

  const onBack = () => {
    if (stepIndex > 0) setStepIndex(stepIndex - 1)
  }

  const onSubmit = (e: FormEvent) => {
    e.preventDefault()
    createMutation.mutate({
      conversation_id: conversationId ?? undefined,
      name: name.trim(),
      description: description.trim(),
      prompt_template: promptTemplate.trim(),
      input_schema: inputSchema,
    })
  }

  const canProceed =
    (step === 'name' && name.trim()) ||
    (step === 'template' && promptTemplate.trim()) ||
    step === 'inputs' ||
    step === 'confirm'

  if (!open) return null

  return (
    <div className="modal-backdrop" role="presentation" onClick={() => onClose()}>
      <div
        className="modal modal--wizard"
        role="dialog"
        aria-modal="true"
        aria-labelledby="task-wizard-title"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 id="task-wizard-title">Save as Task</h2>

        {conversationId && (
          <div className="task-wizard__draft">
            {isDraftLoading && <p className="muted">Generating draft from conversation...</p>}
            {isDraftError && (
              <p className="muted">
                Could not generate draft (provider may be missing). You can fill in manually.
              </p>
            )}
            {!isDraftLoading && !draft && !isDraftError && (
              <button
                type="button"
                className="btn"
                onClick={() => refetchDraft()}
              >
                Generate draft from conversation
              </button>
            )}
          </div>
        )}

        <form onSubmit={step === 'confirm' ? onSubmit : (e) => { e.preventDefault(); onNext() }}>
          <div className="task-wizard__steps">
            {STEPS.map((s, i) => (
              <span
                key={s}
                className={`task-wizard__step ${i === stepIndex ? 'task-wizard__step--active' : ''} ${i < stepIndex ? 'task-wizard__step--done' : ''}`}
              >
                {i + 1}. {s}
              </span>
            ))}
          </div>

          {step === 'name' && (
            <div className="task-wizard__field">
              <label>
                Task name
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. Movie Recommender"
                  autoFocus
                />
              </label>
              <label>
                Description (optional)
                <input
                  type="text"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Short description"
                />
              </label>
            </div>
          )}

          {step === 'template' && (
            <div className="task-wizard__field">
              <label>
                Prompt template (use {'{{variable}}'} for inputs)
                <textarea
                  value={promptTemplate}
                  onChange={(e) => handleTemplateChange(e.target.value)}
                  rows={6}
                  placeholder="e.g. Recommend a movie. I have {{streaming_services}}. Mood: {{mood}}."
                />
              </label>
              {inputSchema.length > 0 && (
                <p className="muted">Variables found: {inputSchema.map((f) => f.key).join(', ')}</p>
              )}
            </div>
          )}

          {step === 'inputs' && (
            <div className="task-wizard__field">
              <p className="muted">Adjust input fields if needed.</p>
              {inputSchema.map((field, i) => (
                <div key={field.key} className="task-wizard__input-row">
                  <input
                    type="text"
                    value={field.key}
                    readOnly
                    className="task-wizard__input-key"
                  />
                  <input
                    type="text"
                    value={field.label}
                    onChange={(e) => {
                      const next = [...inputSchema]
                      next[i] = { ...next[i], label: e.target.value }
                      setInputSchema(next)
                    }}
                    placeholder="Label"
                  />
                  <select
                    value={field.type}
                    onChange={(e) => {
                      const next = [...inputSchema]
                      next[i] = { ...next[i], type: e.target.value as InputField['type'] }
                      setInputSchema(next)
                    }}
                  >
                    <option value="text">text</option>
                    <option value="select">select</option>
                    <option value="number">number</option>
                    <option value="boolean">boolean</option>
                  </select>
                </div>
              ))}
            </div>
          )}

          {step === 'confirm' && (
            <div className="task-wizard__confirm">
              <p><strong>{name || 'Untitled'}</strong></p>
              <pre className="task-wizard__preview">{promptTemplate || '(empty)'}</pre>
              <p className="muted">Inputs: {inputSchema.map((f) => f.key).join(', ') || 'none'}</p>
            </div>
          )}

          {createMutation.error && (
            <ErrorDisplay error={createMutation.error} title="Save failed" />
          )}

          <div className="task-wizard__actions">
            <button type="button" className="btn" onClick={stepIndex === 0 ? onClose : onBack}>
              {stepIndex === 0 ? 'Cancel' : 'Back'}
            </button>
            {step === 'confirm' ? (
              <button
                type="submit"
                className="btn btn-primary"
                disabled={createMutation.isPending || !name.trim() || !promptTemplate.trim()}
              >
                {createMutation.isPending ? 'Saving...' : 'Save task'}
              </button>
            ) : (
              <button type="submit" className="btn btn-primary" disabled={!canProceed}>
                Next
              </button>
            )}
          </div>
        </form>
      </div>
    </div>
  )
}
