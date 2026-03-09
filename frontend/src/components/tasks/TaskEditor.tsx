import { FormEvent, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { InputField, Task, UpdateTaskRequest } from '../../lib/api-types'
import { api } from '../../lib/api'
import { extractVariables, variablesToInputSchema } from '../../lib/templates'
import { ErrorDisplay } from '../ErrorDisplay'

interface TaskEditorProps {
  task: Task
  onSaved?: () => void
}

export function TaskEditor({ task, onSaved }: TaskEditorProps) {
  const queryClient = useQueryClient()
  const [name, setName] = useState(task.name)
  const [description, setDescription] = useState(task.description)
  const [promptTemplate, setPromptTemplate] = useState(task.prompt_template)
  const [inputSchema, setInputSchema] = useState<InputField[]>(task.input_schema ?? [])

  const updateMutation = useMutation({
    mutationFn: (body: UpdateTaskRequest) => api.updateTask(task.id, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
      queryClient.invalidateQueries({ queryKey: ['task', task.id] })
      onSaved?.()
    },
  })

  const handleTemplateChange = (value: string) => {
    setPromptTemplate(value)
    const vars = extractVariables(value)
    setInputSchema((prev) => variablesToInputSchema(vars, prev))
  }

  const onSubmit = (e: FormEvent) => {
    e.preventDefault()
    updateMutation.mutate({
      name: name.trim() || undefined,
      description: description.trim() || undefined,
      prompt_template: promptTemplate.trim() || undefined,
      input_schema: inputSchema,
    })
  }

  return (
    <section className="task-editor">
      <h3>Edit task</h3>
      <form onSubmit={onSubmit} className="task-editor__form">
        <label>
          Name
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Task name"
          />
        </label>
        <label>
          Description (optional)
          <input
            type="text"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Description"
          />
        </label>
        <label>
          Prompt template
          <textarea
            value={promptTemplate}
            onChange={(e) => handleTemplateChange(e.target.value)}
            rows={5}
            placeholder="Use {{variable}} for inputs"
          />
        </label>
        <div className="task-editor__inputs">
          <p className="muted">Input fields</p>
          {inputSchema.map((field, i) => (
            <div key={field.key} className="task-editor__input-row">
              <input
                type="text"
                value={field.key}
                readOnly
                className="task-editor__input-key"
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
        {updateMutation.error && <ErrorDisplay error={updateMutation.error} title="Update failed" />}
        <button
          type="submit"
          className="btn btn-primary"
          disabled={updateMutation.isPending || !name.trim() || !promptTemplate.trim()}
        >
          {updateMutation.isPending ? 'Saving...' : 'Save changes'}
        </button>
      </form>
    </section>
  )
}
