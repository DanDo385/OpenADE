import type {
  Agent,
  AgentRunRequest,
  AgentRunResponse,
  ChatStreamEvent,
  CommandExecuteRequest,
  CommandExecuteResponse,
  Conversation,
  CreateMCPServerRequest,
  CreateMessageRequest,
  CreateScheduleRequest,
  CreateTaskRequest,
  ErrorResponse,
  ExportBundle,
  MCPServer,
  Objective,
  ProviderSaveRequest,
  ProviderSaveResponse,
  ProviderSummary,
  Run,
  RunTaskRequest,
  Schedule,
  SetMemoryRequest,
  Task,
  TaskDraft,
  UpdateMCPServerRequest,
  UpdateScheduleRequest,
  UpdateTaskRequest,
  UpsertObjectiveRequest,
} from './api-types'

const DEFAULT_API_BASE_URL = 'http://localhost:8080'

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, '')
}

export const API_BASE_URL = trimTrailingSlash(import.meta.env.VITE_API_URL ?? DEFAULT_API_BASE_URL)

export class APIError extends Error {
  readonly status: number
  readonly code: string
  readonly details?: unknown

  constructor(status: number, code: string, message: string, details?: unknown) {
    super(message)
    this.name = 'APIError'
    this.status = status
    this.code = code
    this.details = details
  }
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function isErrorResponse(value: unknown): value is ErrorResponse {
  if (!isObject(value)) {
    return false
  }
  const error = value.error
  if (!isObject(error)) {
    return false
  }
  return typeof error.code === 'string' && typeof error.message === 'string'
}

async function parseResponseBody(response: Response): Promise<unknown> {
  const contentType = response.headers.get('content-type') ?? ''
  if (contentType.includes('application/json')) {
    return response.json()
  }

  const text = await response.text()
  if (!text) {
    return null
  }

  try {
    return JSON.parse(text)
  } catch {
    return text
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
  })

  const payload = await parseResponseBody(response)

  if (!response.ok) {
    if (isErrorResponse(payload)) {
      throw new APIError(response.status, payload.error.code, payload.error.message, payload.error.details)
    }
    throw new APIError(response.status, 'request_failed', `request failed with status ${response.status}`)
  }

  return payload as T
}

function parseSSEEvent(rawEvent: string): ChatStreamEvent | null {
  const lines = rawEvent.split('\n')
  const dataLines: string[] = []

  for (const line of lines) {
    if (line.startsWith('data:')) {
      dataLines.push(line.slice(5).trimStart())
    }
  }

  if (dataLines.length === 0) {
    return null
  }

  try {
    const parsed = JSON.parse(dataLines.join('\n')) as ChatStreamEvent
    if (parsed && typeof parsed === 'object' && 'type' in parsed) {
      return parsed
    }
    return null
  } catch {
    return { type: 'error', message: 'invalid stream payload' }
  }
}

async function handleStreamResponse(
  response: Response,
  onEvent: (event: ChatStreamEvent) => void,
): Promise<void> {
  if (!response.body) {
    throw new APIError(500, 'stream_unavailable', 'stream response body is unavailable')
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) {
      break
    }

    buffer += decoder.decode(value, { stream: true })
    const frames = buffer.split('\n\n')
    buffer = frames.pop() ?? ''

    for (const frame of frames) {
      const event = parseSSEEvent(frame)
      if (event) {
        onEvent(event)
      }
    }
  }

  const rest = decoder.decode()
  if (rest) {
    buffer += rest
  }

  if (buffer.trim() !== '') {
    const event = parseSSEEvent(buffer)
    if (event) {
      onEvent(event)
    }
  }
}

export const api = {
  health: () => request<{ status: string }>('/health', { method: 'GET' }),

  createConversation: () => request<Conversation>('/api/conversations', { method: 'POST' }),
  listConversations: () => request<Conversation[]>('/api/conversations', { method: 'GET' }),
  getConversation: (id: string) =>
    request<Conversation>(`/api/conversations/${encodeURIComponent(id)}`, { method: 'GET' }),
  deleteConversation: (id: string) =>
    request<{ ok: boolean }>(`/api/conversations/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  draftTaskFromConversation: (conversationId: string) =>
    request<TaskDraft>(`/api/conversations/${encodeURIComponent(conversationId)}/draft-task`, {
      method: 'POST',
    }),
  getObjective: (conversationId: string) =>
    request<Objective>(`/api/conversations/${encodeURIComponent(conversationId)}/objective`, {
      method: 'GET',
    }),
  upsertObjective: (conversationId: string, body: UpsertObjectiveRequest) =>
    request<Objective>(`/api/conversations/${encodeURIComponent(conversationId)}/objective`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),
  exportObjectiveMarkdown: async (conversationId: string): Promise<string> => {
    const response = await fetch(
      `${API_BASE_URL}/api/conversations/${encodeURIComponent(conversationId)}/objective/export`,
      { method: 'GET' }
    )
    if (!response.ok) {
      const payload = await response.json().catch(() => ({}))
      if (payload?.error?.message) {
        throw new APIError(response.status, payload.error.code ?? 'export_failed', payload.error.message)
      }
      throw new APIError(response.status, 'export_failed', `export failed with status ${response.status}`)
    }
    return response.text()
  },

  streamConversationMessage: async (
    conversationId: string,
    body: CreateMessageRequest,
    onEvent: (event: ChatStreamEvent) => void,
    signal?: AbortSignal,
  ): Promise<void> => {
    const response = await fetch(
      `${API_BASE_URL}/api/conversations/${encodeURIComponent(conversationId)}/messages`,
      {
        method: 'POST',
        headers: {
          Accept: 'text/event-stream',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(body),
        signal,
      },
    )

    if (!response.ok) {
      const payload = await parseResponseBody(response)
      if (isErrorResponse(payload)) {
        throw new APIError(
          response.status,
          payload.error.code,
          payload.error.message,
          payload.error.details,
        )
      }
      throw new APIError(response.status, 'stream_failed', `stream failed with status ${response.status}`)
    }

    await handleStreamResponse(response, onEvent)
  },

  listProviders: () => request<ProviderSummary[]>('/api/providers', { method: 'GET' }),
  saveProvider: (id: string, body: ProviderSaveRequest) =>
    request<ProviderSaveResponse>(`/api/providers/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),

  createTask: (body: CreateTaskRequest) =>
    request<Task>('/api/tasks', { method: 'POST', body: JSON.stringify(body) }),
  listTasks: (query?: string) => {
    const suffix = query ? `?q=${encodeURIComponent(query)}` : ''
    return request<Task[]>(`/api/tasks${suffix}`, { method: 'GET' })
  },
  getTask: (id: string) => request<Task>(`/api/tasks/${encodeURIComponent(id)}`, { method: 'GET' }),
  updateTask: (id: string, body: UpdateTaskRequest) =>
    request<Task>(`/api/tasks/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),
  deleteTask: (id: string) =>
    request<{ ok: boolean }>(`/api/tasks/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  runTask: (id: string, body: RunTaskRequest) =>
    request<Run>(`/api/tasks/${encodeURIComponent(id)}/run`, {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  exportTask: (id: string) =>
    request<ExportBundle>(`/api/tasks/${encodeURIComponent(id)}/export`, { method: 'POST' }),
  importTask: (body: ExportBundle) =>
    request<Task>('/api/tasks/import', { method: 'POST', body: JSON.stringify(body) }),

  listSchedules: (taskId?: string) => {
    const suffix = taskId ? `?task_id=${encodeURIComponent(taskId)}` : ''
    return request<Schedule[]>(`/api/schedules${suffix}`, { method: 'GET' })
  },
  createSchedule: (body: CreateScheduleRequest) =>
    request<Schedule>('/api/schedules', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  updateSchedule: (id: string, body: UpdateScheduleRequest) =>
    request<Schedule>(`/api/schedules/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),
  deleteScheduleById: (id: string) =>
    request<{ ok: boolean }>(`/api/schedules/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    }),

  getSchedule: (taskId: string) =>
    request<Schedule>(`/api/tasks/${encodeURIComponent(taskId)}/schedule`, { method: 'GET' }),
  upsertSchedule: (taskId: string, body: CreateScheduleRequest | UpdateScheduleRequest) =>
    request<Schedule>(`/api/tasks/${encodeURIComponent(taskId)}/schedule`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),
  deleteSchedule: (taskId: string) =>
    request<{ ok: boolean }>(`/api/tasks/${encodeURIComponent(taskId)}/schedule`, {
      method: 'DELETE',
    }),

  listRuns: () => request<Run[]>('/api/runs', { method: 'GET' }),
  getRun: (id: string) => request<Run>(`/api/runs/${encodeURIComponent(id)}`, { method: 'GET' }),

  getMemory: (taskId: string) =>
    request<{ entries: Record<string, string> }>(`/api/memory/${encodeURIComponent(taskId)}`, {
      method: 'GET',
    }),
  setMemory: (taskId: string, body: SetMemoryRequest) =>
    request<{ ok: boolean }>(`/api/memory/${encodeURIComponent(taskId)}`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),
  setMemoryKey: (taskId: string, key: string, value: string) =>
    request<{ ok: boolean }>(
      `/api/memory/${encodeURIComponent(taskId)}/${encodeURIComponent(key)}`,
      {
        method: 'PUT',
        body: JSON.stringify({ value }),
      },
    ),

  // Commands (Load 6)
  executeCommand: (body: CommandExecuteRequest) =>
    request<CommandExecuteResponse>('/api/commands/execute', {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  // Agents (Load 6, 8)
  listAgents: () => request<Agent[]>('/api/agents', { method: 'GET' }),
  getAgent: (id: string) =>
    request<Agent>(`/api/agents/${encodeURIComponent(id)}`, { method: 'GET' }),
  runAgent: (id: string, body?: AgentRunRequest) =>
    request<AgentRunResponse>(`/api/agents/${encodeURIComponent(id)}/run`, {
      method: 'POST',
      body: JSON.stringify(body ?? {}),
    }),

  // MCP Servers
  listMCPServers: () => request<MCPServer[]>('/api/mcp/servers', { method: 'GET' }),
  createMCPServer: (body: CreateMCPServerRequest) =>
    request<MCPServer>('/api/mcp/servers', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  updateMCPServer: (id: string, body: UpdateMCPServerRequest) =>
    request<MCPServer>(`/api/mcp/servers/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),
  deleteMCPServer: (id: string) =>
    request<{ ok: boolean }>(`/api/mcp/servers/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    }),
  testMCPServer: (id: string) =>
    request<{ ok: boolean; message?: string }>(
      `/api/mcp/servers/${encodeURIComponent(id)}/test`,
      { method: 'POST' }
    ),
}
