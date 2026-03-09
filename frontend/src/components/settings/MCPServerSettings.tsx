import type { FormEvent } from 'react'
import { useCallback, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api, APIError } from '../../lib/api'
import type {
  MCPServer,
  MCPServerTestResponse,
  MCPToolInfo,
  MCPTransport,
} from '../../lib/api-types'
import { ErrorDisplay } from '../ErrorDisplay'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { cn } from '@/lib/utils'

const DEFAULT_ARGS = '[]'
const DEFAULT_ENV = '{}'

interface ServerFormState {
  name: string
  transport: MCPTransport
  command_or_url: string
  args_json: string
  env_json: string
  enabled: boolean
}

const emptyForm: ServerFormState = {
  name: '',
  transport: 'stdio',
  command_or_url: '',
  args_json: DEFAULT_ARGS,
  env_json: DEFAULT_ENV,
  enabled: true,
}

function serverToForm(server: MCPServer): ServerFormState {
  return {
    name: server.name,
    transport: server.transport,
    command_or_url: server.command_or_url,
    args_json: JSON.stringify(server.args ?? [], null, 2),
    env_json: JSON.stringify(server.env ?? {}, null, 2),
    enabled: server.enabled,
  }
}

function parseArgsJson(value: string): string[] {
  const parsed = JSON.parse(value.trim() || DEFAULT_ARGS)
  if (!Array.isArray(parsed) || parsed.some((item) => typeof item !== 'string')) {
    throw new Error('Args must be a JSON array of strings')
  }
  return parsed
}

function parseEnvJson(value: string): Record<string, string> {
  const parsed = JSON.parse(value.trim() || DEFAULT_ENV)
  if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
    throw new Error('Env must be a JSON object of string values')
  }
  const env: Record<string, string> = {}
  for (const [key, val] of Object.entries(parsed)) {
    if (typeof val !== 'string') {
      throw new Error('Env values must be strings')
    }
    env[key] = val
  }
  return env
}

function getErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof APIError) return error.message
  if (error instanceof Error) return error.message
  return fallback
}

function toolKey(serverId: string, toolName: string): string {
  return `${serverId}:${toolName}`
}

export function MCPServerSettings() {
  const queryClient = useQueryClient()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [form, setForm] = useState<ServerFormState>(emptyForm)
  const [formError, setFormError] = useState<string | null>(null)
  const [activeServerId, setActiveServerId] = useState<string | null>(null)
  const [toolArgs, setToolArgs] = useState<Record<string, string>>({})
  const [toolResults, setToolResults] = useState<Record<string, string>>({})
  const [toolErrors, setToolErrors] = useState<Record<string, string>>({})
  const [testResults, setTestResults] = useState<Record<string, MCPServerTestResponse>>({})

  const { data: servers = [], isLoading, error } = useQuery({
    queryKey: ['mcp-servers'],
    queryFn: api.listMCPServers,
    retry: false,
  })

  const toolsQuery = useQuery({
    queryKey: ['mcp-tools', activeServerId],
    queryFn: () => api.listMCPTools(activeServerId!),
    enabled: !!activeServerId,
    retry: false,
  })

  const createMutation = useMutation({
    mutationFn: api.createMCPServer,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mcp-servers'] })
      closeDialog()
    },
    onError: (err) => setFormError(getErrorMessage(err, 'Failed to create server')),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, body }: { id: string; body: Parameters<typeof api.updateMCPServer>[1] }) =>
      api.updateMCPServer(id, body),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['mcp-servers'] })
      if (activeServerId === variables.id) {
        queryClient.invalidateQueries({ queryKey: ['mcp-tools', variables.id] })
      }
      closeDialog()
    },
    onError: (err) => setFormError(getErrorMessage(err, 'Failed to update server')),
  })

  const deleteMutation = useMutation({
    mutationFn: api.deleteMCPServer,
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ['mcp-servers'] })
      if (activeServerId === id) {
        setActiveServerId(null)
      }
    },
  })

  const testMutation = useMutation({
    mutationFn: api.testMCPServer,
    onSuccess: (result, id) => {
      setTestResults((prev) => ({ ...prev, [id]: result }))
    },
  })

  const callMutation = useMutation({
    mutationFn: ({
      serverId,
      toolName,
      args,
    }: {
      serverId: string
      toolName: string
      args: Record<string, unknown>
    }) =>
      api.callMCPTool({
        server_id: serverId,
        tool_name: toolName,
        arguments: args,
      }),
    onSuccess: (result, variables) => {
      const key = toolKey(variables.serverId, variables.toolName)
      setToolErrors((prev) => ({ ...prev, [key]: '' }))
      setToolResults((prev) => ({
        ...prev,
        [key]: JSON.stringify(result, null, 2),
      }))
    },
    onError: (err, variables) => {
      const key = toolKey(variables.serverId, variables.toolName)
      setToolErrors((prev) => ({ ...prev, [key]: getErrorMessage(err, 'Failed to call tool') }))
    },
  })

  const closeDialog = useCallback(() => {
    setDialogOpen(false)
    setEditingId(null)
    setForm(emptyForm)
    setFormError(null)
  }, [])

  const openCreate = useCallback(() => {
    setForm(emptyForm)
    setEditingId(null)
    setFormError(null)
    setDialogOpen(true)
  }, [])

  const openEdit = useCallback((server: MCPServer) => {
    setForm(serverToForm(server))
    setEditingId(server.id)
    setFormError(null)
    setDialogOpen(true)
  }, [])

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault()
    setFormError(null)

    if (!form.name.trim()) {
      setFormError('Name is required')
      return
    }
    if (!form.command_or_url.trim()) {
      setFormError('Command or URL is required')
      return
    }

    let args: string[]
    let env: Record<string, string>
    try {
      args = parseArgsJson(form.args_json)
      env = parseEnvJson(form.env_json)
    } catch (error) {
      setFormError(getErrorMessage(error, 'Invalid server configuration'))
      return
    }

    const body = {
      name: form.name.trim(),
      transport: form.transport,
      command_or_url: form.command_or_url.trim(),
      args,
      env,
      enabled: form.enabled,
    }

    if (editingId) {
      updateMutation.mutate({ id: editingId, body })
    } else {
      createMutation.mutate(body)
    }
  }

  const toggleEnabled = useCallback(
    (server: MCPServer) => {
      updateMutation.mutate({
        id: server.id,
        body: { enabled: !server.enabled },
      })
    },
    [updateMutation],
  )

  const toggleTools = useCallback(
    (serverId: string) => {
      setActiveServerId((current) => (current === serverId ? null : serverId))
      setToolResults({})
      setToolErrors({})
    },
    [],
  )

  const handleToolCall = useCallback(
    (serverId: string, toolName: string) => {
      const key = toolKey(serverId, toolName)
      let parsedArgs: Record<string, unknown>
      try {
        const parsed = JSON.parse(toolArgs[key] || '{}')
        if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
          throw new Error('Tool arguments must be a JSON object')
        }
        parsedArgs = parsed as Record<string, unknown>
      } catch (error) {
        setToolErrors((prev) => ({
          ...prev,
          [key]: getErrorMessage(error, 'Tool arguments must be valid JSON'),
        }))
        return
      }

      callMutation.mutate({ serverId, toolName, args: parsedArgs })
    },
    [callMutation, toolArgs],
  )

  const isPending = createMutation.isPending || updateMutation.isPending

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle>MCP Servers</CardTitle>
        <Button onClick={openCreate} size="sm">
          Add server
        </Button>
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-sm text-muted-foreground">
          Configure MCP servers, inspect available tools, and call them directly from the UI. Tool discovery and invocation currently require `stdio` transport.
        </p>

        {error && (
          <ErrorDisplay
            error={error}
            title="Failed to load MCP servers"
            onRetry={() => queryClient.invalidateQueries({ queryKey: ['mcp-servers'] })}
          />
        )}

        {isLoading && <p className="text-sm text-muted-foreground">Loading...</p>}

        {!isLoading && !error && servers.length === 0 && (
          <p className="text-sm text-muted-foreground">No MCP servers configured.</p>
        )}

        {!isLoading && !error && servers.length > 0 && (
          <ul className="space-y-3">
            {servers.map((server) => {
              const isActive = activeServerId === server.id
              const testResult = testResults[server.id]

              return (
                <li
                  key={server.id}
                  className={cn(
                    'rounded-md border p-3',
                    !server.enabled && 'opacity-60',
                  )}
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="min-w-0 flex-1">
                      <p className="font-medium">{server.name}</p>
                      <p className="truncate text-sm text-muted-foreground">
                        {server.transport} · {server.command_or_url}
                      </p>
                      {server.args.length > 0 && (
                        <p className="mt-1 text-xs text-muted-foreground">
                          Args: {JSON.stringify(server.args)}
                        </p>
                      )}
                    </div>

                    <div className="flex flex-wrap items-center gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => testMutation.mutate(server.id)}
                        disabled={testMutation.isPending}
                      >
                        {testMutation.isPending ? 'Testing...' : 'Test'}
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => toggleTools(server.id)}
                        disabled={server.transport !== 'stdio'}
                      >
                        {isActive ? 'Hide tools' : 'Browse tools'}
                      </Button>
                      <button
                        type="button"
                        onClick={() => toggleEnabled(server)}
                        className={cn(
                          'rounded px-2 py-1 text-sm',
                          server.enabled
                            ? 'bg-primary/20 text-primary'
                            : 'bg-muted text-muted-foreground',
                        )}
                      >
                        {server.enabled ? 'On' : 'Off'}
                      </button>
                      <Button variant="outline" size="sm" onClick={() => openEdit(server)}>
                        Edit
                      </Button>
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => {
                          if (window.confirm(`Delete "${server.name}"?`)) {
                            deleteMutation.mutate(server.id)
                          }
                        }}
                        disabled={deleteMutation.isPending}
                      >
                        Delete
                      </Button>
                    </div>
                  </div>

                  {testResult && (
                    <p className="mt-2 text-sm text-muted-foreground">
                      Test: {testResult.ok ? `${testResult.message} (${testResult.tool_count ?? 0} tools)` : testResult.message}
                    </p>
                  )}

                  {server.transport !== 'stdio' && (
                    <p className="mt-2 text-xs text-muted-foreground">
                      Tool discovery and calling are currently implemented for `stdio` servers only.
                    </p>
                  )}

                  {isActive && (
                    <div className="mt-4 space-y-3 border-t pt-4">
                      <div className="flex items-center justify-between">
                        <h4 className="text-sm font-semibold">Tools</h4>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => toolsQuery.refetch()}
                          disabled={toolsQuery.isFetching}
                        >
                          {toolsQuery.isFetching ? 'Loading...' : 'Refresh'}
                        </Button>
                      </div>

                      {toolsQuery.isLoading && <p className="text-sm text-muted-foreground">Loading tools...</p>}

                      {toolsQuery.error && (
                        <ErrorDisplay
                          error={toolsQuery.error}
                          title="Failed to load tools"
                          onRetry={() => toolsQuery.refetch()}
                        />
                      )}

                      {!toolsQuery.isLoading && !toolsQuery.error && (toolsQuery.data?.length ?? 0) === 0 && (
                        <p className="text-sm text-muted-foreground">No tools exposed by this server.</p>
                      )}

                      {!toolsQuery.isLoading && !toolsQuery.error && (toolsQuery.data?.length ?? 0) > 0 && (
                        <ul className="space-y-3">
                          {(toolsQuery.data ?? []).map((tool: MCPToolInfo) => {
                            const key = toolKey(server.id, tool.name)
                            const result = toolResults[key]
                            const errorMessage = toolErrors[key]

                            return (
                              <li key={key} className="rounded-md border bg-muted/20 p-3">
                                <div className="space-y-2">
                                  <div>
                                    <p className="font-medium">{tool.name}</p>
                                    {tool.description && (
                                      <p className="text-sm text-muted-foreground">{tool.description}</p>
                                    )}
                                  </div>

                                  <details>
                                    <summary className="cursor-pointer text-sm text-muted-foreground">
                                      Input schema
                                    </summary>
                                    <pre className="mt-2 overflow-auto rounded bg-background p-3 text-xs">
                                      {JSON.stringify(tool.input_schema ?? {}, null, 2)}
                                    </pre>
                                  </details>

                                  <div className="grid gap-2">
                                    <label className="text-sm font-medium">Arguments (JSON object)</label>
                                    <Textarea
                                      value={toolArgs[key] ?? '{}'}
                                      onChange={(event) =>
                                        setToolArgs((prev) => ({ ...prev, [key]: event.target.value }))
                                      }
                                      rows={4}
                                      className="font-mono text-sm"
                                      placeholder="{}"
                                    />
                                  </div>

                                  <div className="flex items-center gap-2">
                                    <Button
                                      size="sm"
                                      onClick={() => handleToolCall(server.id, tool.name)}
                                      disabled={callMutation.isPending}
                                    >
                                      {callMutation.isPending ? 'Calling...' : 'Call tool'}
                                    </Button>
                                  </div>

                                  {errorMessage && (
                                    <p className="text-sm text-destructive">{errorMessage}</p>
                                  )}

                                  {result && (
                                    <pre className="overflow-auto rounded bg-background p-3 text-xs">
                                      {result}
                                    </pre>
                                  )}
                                </div>
                              </li>
                            )
                          })}
                        </ul>
                      )}
                    </div>
                  )}
                </li>
              )
            })}
          </ul>
        )}
      </CardContent>

      <Dialog
        open={dialogOpen}
        onOpenChange={(open) => {
          if (!open) closeDialog()
          else setDialogOpen(true)
        }}
      >
        <DialogContent
          onPointerDownOutside={() => closeDialog()}
          onEscapeKeyDown={() => closeDialog()}
        >
          <form onSubmit={handleSubmit}>
            <DialogHeader>
              <DialogTitle>{editingId ? 'Edit MCP server' : 'Add MCP server'}</DialogTitle>
            </DialogHeader>

            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <label className="text-sm font-medium">Name</label>
                <Input
                  value={form.name}
                  onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))}
                  placeholder="my-mcp-server"
                />
              </div>

              <div className="grid gap-2">
                <label className="text-sm font-medium">Transport</label>
                <select
                  value={form.transport}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      transport: event.target.value as MCPTransport,
                    }))
                  }
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                >
                  <option value="stdio">stdio</option>
                  <option value="sse">sse</option>
                </select>
              </div>

              <div className="grid gap-2">
                <label className="text-sm font-medium">Command (stdio) or URL (sse)</label>
                <Input
                  value={form.command_or_url}
                  onChange={(event) =>
                    setForm((current) => ({ ...current, command_or_url: event.target.value }))
                  }
                  placeholder="npx"
                />
              </div>

              <div className="grid gap-2">
                <label className="text-sm font-medium">Args (JSON array of strings)</label>
                <Textarea
                  value={form.args_json}
                  onChange={(event) =>
                    setForm((current) => ({ ...current, args_json: event.target.value }))
                  }
                  placeholder='["-y", "@modelcontextprotocol/server-filesystem", "/path"]'
                  rows={3}
                  className="font-mono text-sm"
                />
              </div>

              <div className="grid gap-2">
                <label className="text-sm font-medium">Env (JSON object)</label>
                <Textarea
                  value={form.env_json}
                  onChange={(event) =>
                    setForm((current) => ({ ...current, env_json: event.target.value }))
                  }
                  placeholder='{"API_KEY":"..."}'
                  rows={3}
                  className="font-mono text-sm"
                />
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="enabled"
                  checked={form.enabled}
                  onChange={(event) =>
                    setForm((current) => ({ ...current, enabled: event.target.checked }))
                  }
                  className="h-4 w-4 rounded"
                />
                <label htmlFor="enabled" className="text-sm">
                  Enabled
                </label>
              </div>

              {formError && <p className="text-sm text-destructive">{formError}</p>}
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={closeDialog}>
                Cancel
              </Button>
              <Button type="submit" disabled={isPending}>
                {isPending ? 'Saving...' : editingId ? 'Update' : 'Create'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </Card>
  )
}
