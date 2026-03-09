import { useState, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../../lib/api'
import type { MCPServer, MCPTransport } from '../../lib/api-types'
import { ErrorDisplay } from '../ErrorDisplay'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { cn } from '@/lib/utils'

const DEFAULT_ARGS = '[]'
const DEFAULT_ENV = '{}'

function safeJson(value: string, fallback: string): string {
  if (!value.trim()) return fallback
  try {
    JSON.parse(value)
    return value.trim()
  } catch {
    return fallback
  }
}

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

function serverToForm(s: MCPServer): ServerFormState {
  return {
    name: s.name,
    transport: s.transport,
    command_or_url: s.command_or_url,
    args_json: s.args_json || DEFAULT_ARGS,
    env_json: s.env_json || DEFAULT_ENV,
    enabled: s.enabled,
  }
}

export function MCPServerSettings() {
  const queryClient = useQueryClient()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [form, setForm] = useState<ServerFormState>(emptyForm)
  const [formError, setFormError] = useState<string | null>(null)

  const { data: servers = [], isLoading, error } = useQuery({
    queryKey: ['mcp-servers'],
    queryFn: api.listMCPServers,
    retry: false,
  })

  const createMutation = useMutation({
    mutationFn: api.createMCPServer,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mcp-servers'] })
      closeDialog()
    },
    onError: () => setFormError('Failed to create server'),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, body }: { id: string; body: Parameters<typeof api.updateMCPServer>[1] }) =>
      api.updateMCPServer(id, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mcp-servers'] })
      closeDialog()
    },
    onError: () => setFormError('Failed to update server'),
  })

  const deleteMutation = useMutation({
    mutationFn: api.deleteMCPServer,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['mcp-servers'] }),
  })

  const testMutation = useMutation({
    mutationFn: api.testMCPServer,
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

  const openEdit = useCallback((s: MCPServer) => {
    setForm(serverToForm(s))
    setEditingId(s.id)
    setFormError(null)
    setDialogOpen(true)
  }, [])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setFormError(null)
    if (!form.name.trim()) {
      setFormError('Name is required')
      return
    }
    if (!form.command_or_url.trim()) {
      setFormError('Command or URL is required')
      return
    }
    const argsJson = safeJson(form.args_json, DEFAULT_ARGS)
    const envJson = safeJson(form.env_json, DEFAULT_ENV)
    const body = {
      name: form.name.trim(),
      transport: form.transport,
      command_or_url: form.command_or_url.trim(),
      args_json: argsJson,
      env_json: envJson,
      enabled: form.enabled,
    }
    if (editingId) {
      updateMutation.mutate({ id: editingId, body })
    } else {
      createMutation.mutate(body)
    }
  }

  const toggleEnabled = useCallback(
    (s: MCPServer) => {
      updateMutation.mutate({
        id: s.id,
        body: { enabled: !s.enabled },
      })
    },
    [updateMutation]
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
          Configure MCP servers for tools, resources, and prompts. Requires backend MCP support (Piece C).
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
          <ul className="space-y-2">
            {servers.map((s) => (
              <li
                key={s.id}
                className={cn(
                  'flex items-center justify-between rounded-md border p-3',
                  !s.enabled && 'opacity-60'
                )}
              >
                <div className="min-w-0 flex-1">
                  <p className="font-medium">{s.name}</p>
                  <p className="truncate text-sm text-muted-foreground">
                    {s.transport} · {s.command_or_url}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => testMutation.mutate(s.id)}
                    disabled={testMutation.isPending}
                  >
                    {testMutation.isPending ? 'Testing...' : 'Test'}
                  </Button>
                  <button
                    type="button"
                    onClick={() => toggleEnabled(s)}
                    className={cn(
                      'rounded px-2 py-1 text-sm',
                      s.enabled
                        ? 'bg-primary/20 text-primary'
                        : 'bg-muted text-muted-foreground'
                    )}
                  >
                    {s.enabled ? 'On' : 'Off'}
                  </button>
                  <Button variant="outline" size="sm" onClick={() => openEdit(s)}>
                    Edit
                  </Button>
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() => {
                      if (window.confirm(`Delete "${s.name}"?`)) {
                        deleteMutation.mutate(s.id)
                      }
                    }}
                    disabled={deleteMutation.isPending}
                  >
                    Delete
                  </Button>
                </div>
              </li>
            ))}
          </ul>
        )}

        {testMutation.data && (
          <p className="text-sm text-muted-foreground">
            Test: {testMutation.data.ok ? 'Success' : testMutation.data.message ?? 'Failed'}
          </p>
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
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                  placeholder="my-mcp-server"
                />
              </div>
              <div className="grid gap-2">
                <label className="text-sm font-medium">Transport</label>
                <select
                  value={form.transport}
                  onChange={(e) =>
                    setForm((f) => ({ ...f, transport: e.target.value as MCPTransport }))
                  }
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                >
                  <option value="stdio">stdio</option>
                  <option value="sse">sse</option>
                </select>
              </div>
              <div className="grid gap-2">
                <label className="text-sm font-medium">
                  Command (stdio) or URL (sse)
                </label>
                <Input
                  value={form.command_or_url}
                  onChange={(e) =>
                    setForm((f) => ({ ...f, command_or_url: e.target.value }))
                  }
                  placeholder="npx -y @modelcontextprotocol/server-filesystem"
                />
              </div>
              <div className="grid gap-2">
                <label className="text-sm font-medium">Args (JSON array)</label>
                <Textarea
                  value={form.args_json}
                  onChange={(e) => setForm((f) => ({ ...f, args_json: e.target.value }))}
                  placeholder='["--allow-read", "/path"]'
                  rows={2}
                  className="font-mono text-sm"
                />
              </div>
              <div className="grid gap-2">
                <label className="text-sm font-medium">Env (JSON key-value)</label>
                <Textarea
                  value={form.env_json}
                  onChange={(e) => setForm((f) => ({ ...f, env_json: e.target.value }))}
                  placeholder='{"API_KEY":"..."}'
                  rows={2}
                  className="font-mono text-sm"
                />
              </div>
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="enabled"
                  checked={form.enabled}
                  onChange={(e) =>
                    setForm((f) => ({ ...f, enabled: e.target.checked }))
                  }
                  className="h-4 w-4 rounded"
                />
                <label htmlFor="enabled" className="text-sm">
                  Enabled
                </label>
              </div>
              {formError && (
                <p className="text-sm text-destructive">{formError}</p>
              )}
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
