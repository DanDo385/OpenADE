export interface ErrorDetail {
  code: string
  message: string
  details?: unknown
}

export interface ErrorResponse {
  error: ErrorDetail
}

export type MessageRole = 'user' | 'assistant' | 'system'

export interface Message {
  id: string
  conversation_id: string
  role: MessageRole
  content: string
  created_at: string
}

export interface Conversation {
  id: string
  title: string
  created_at: string
  updated_at: string
  messages?: Message[]
}

export interface InputField {
  key: string
  type: 'text' | 'select' | 'multi_select' | 'number' | 'boolean'
  label: string
  options?: string[]
  default?: string
}

export interface Task {
  id: string
  name: string
  description: string
  prompt_template: string
  input_schema: InputField[]
  output_style: string
  version: number
  created_at: string
  updated_at: string
}

export interface TaskVersion {
  id: string
  task_id: string
  version: number
  snapshot: string
  created_at: string
}

export interface Run {
  id: string
  task_id: string
  task_version: number
  input_values: Record<string, unknown>
  prompt_final: string
  output: string
  status: 'running' | 'completed' | 'failed'
  error?: string
  model: string
  input_tokens: number
  output_tokens: number
  cost_usd: number
  duration_ms: number
  created_at: string
}

export interface ProviderSummary {
  id: string
  provider: string
  configured: boolean
  default_model?: string
  base_url?: string
}

export interface ProviderSaveRequest {
  api_key: string
  base_url?: string
  default_model?: string
}

export interface ProviderSaveResponse {
  provider: string
  configured: boolean
  default_model?: string
}

export interface TaskDraft {
  name: string
  description: string
  prompt_template: string
  input_schema: InputField[]
}

export interface ExportBundle {
  bundle_version: string
  task: Task
  versions?: TaskVersion[]
  memory?: Record<string, string>
}

export interface SetMemoryRequest {
  entries?: Record<string, string>
  value?: string
}

export interface CreateMessageRequest {
  content: string
  model?: string
}

export interface CreateTaskRequest {
  conversation_id?: string
  name: string
  description?: string
  prompt_template: string
  input_schema?: InputField[]
  output_style?: string
}

export interface UpdateTaskRequest {
  name?: string
  description?: string
  prompt_template?: string
  input_schema?: InputField[]
  output_style?: string
}

export interface RunTaskRequest {
  inputs: Record<string, unknown>
  model?: string
}

export interface CostInfo {
  prompt: number
  completion: number
  total: number
}

export interface StreamChunkEvent {
  type: 'chunk'
  content: string
}

export interface StreamDoneEvent {
  type: 'done'
  message_id: string
  cost: CostInfo
}

export interface StreamErrorEvent {
  type: 'error'
  message: string
}

export type ChatStreamEvent = StreamChunkEvent | StreamDoneEvent | StreamErrorEvent
