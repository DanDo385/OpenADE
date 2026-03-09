import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import cronstrue from 'cronstrue'
import type { Task } from '../../lib/api-types'
import { APIError, api } from '../../lib/api'
import { ErrorDisplay } from '../ErrorDisplay'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'

function formatCronPreview(expr: string): string {
  if (!expr.trim()) return ''
  try {
    return cronstrue.toString(expr.trim())
  } catch {
    return 'Invalid cron expression'
  }
}

function formatDateTime(iso: string | null | undefined): string {
  if (!iso) return '—'
  try {
    const d = new Date(iso)
    return isNaN(d.getTime()) ? '—' : d.toLocaleString()
  } catch {
    return '—'
  }
}

interface SchedulePanelProps {
  task: Task
}

export function SchedulePanel({ task }: SchedulePanelProps) {
  const queryClient = useQueryClient()
  const [cronExpr, setCronExpr] = useState('')
  const [timezone, setTimezone] = useState('')
  const [hasEdited, setHasEdited] = useState(false)

  const { data: schedule, isLoading, error } = useQuery({
    queryKey: ['schedule', task.id],
    queryFn: async () => {
      try {
        return await api.getSchedule(task.id)
      } catch (e) {
        if (e instanceof APIError && e.status === 404) return null
        throw e
      }
    },
    retry: false,
  })

  const upsertMutation = useMutation({
    mutationFn: () =>
      api.upsertSchedule(task.id, {
        cron_expr: cronExpr.trim(),
        timezone: timezone.trim() || undefined,
        enabled: schedule?.enabled ?? true,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['schedule', task.id] })
      setHasEdited(false)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: () => api.deleteSchedule(task.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['schedule', task.id] })
      setCronExpr('')
      setTimezone('')
      setHasEdited(false)
    },
  })

  const toggleMutation = useMutation({
    mutationFn: (enabled: boolean) =>
      api.upsertSchedule(task.id, {
        cron_expr: schedule?.cron_expr ?? cronExpr.trim(),
        timezone: schedule?.timezone ?? (timezone.trim() || undefined),
        enabled,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['schedule', task.id] })
    },
  })

  useEffect(() => {
    if (!hasEdited) {
      if (schedule) {
        setCronExpr(schedule.cron_expr)
        setTimezone(schedule.timezone ?? '')
      } else if (!isLoading) {
        setCronExpr('')
        setTimezone('')
      }
    }
  }, [schedule, isLoading, hasEdited])

  const cronPreview = formatCronPreview(hasEdited ? cronExpr : (schedule?.cron_expr ?? cronExpr))
  const is404 = error instanceof APIError && error.status === 404
  const showError = error && !is404

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle>Schedule</CardTitle>
        {schedule && (
          <button
            type="button"
            onClick={() => toggleMutation.mutate(!schedule.enabled)}
            disabled={toggleMutation.isPending}
            className={cn(
              'rounded px-2 py-1 text-sm font-medium',
              schedule.enabled
                ? 'bg-primary/20 text-primary'
                : 'bg-muted text-muted-foreground'
            )}
          >
            {schedule.enabled ? 'On' : 'Off'}
          </button>
        )}
      </CardHeader>
      <CardContent className="space-y-4">
        <p className="text-sm text-muted-foreground">
          Run this task on a cron schedule. In v1, scheduled runs only fire while the app and backend are open.
        </p>

        {showError && (
          <ErrorDisplay
            error={error}
            title="Failed to load schedule"
            onRetry={() => queryClient.invalidateQueries({ queryKey: ['schedule', task.id] })}
          />
        )}

        {schedule && (
          <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
            <span>Last run: {formatDateTime(schedule.last_run_at)}</span>
            <span>Next run: {formatDateTime(schedule.next_run_at)}</span>
          </div>
        )}

        {isLoading && !schedule && <p className="text-sm text-muted-foreground">Loading...</p>}

        <div className="space-y-3">
          <div className="grid gap-2">
            <label className="text-sm font-medium">Cron expression</label>
            <Input
              value={hasEdited ? cronExpr : (schedule?.cron_expr ?? cronExpr)}
              onChange={(e) => {
                setCronExpr(e.target.value)
                setHasEdited(true)
              }}
              placeholder="0 9 * * *"
              className="font-mono"
            />
            {cronPreview && (
              <p
                className={cn(
                  'text-sm',
                  cronPreview === 'Invalid cron expression'
                    ? 'text-destructive'
                    : 'text-muted-foreground'
                )}
              >
                {cronPreview}
              </p>
            )}
          </div>
          <div className="grid gap-2">
            <label className="text-sm font-medium">Timezone (optional)</label>
            <Input
              value={hasEdited ? timezone : (schedule?.timezone ?? timezone)}
              onChange={(e) => {
                setTimezone(e.target.value)
                setHasEdited(true)
              }}
              placeholder="America/New_York"
            />
          </div>
          <div className="flex gap-2">
            {hasEdited && (
              <Button
                size="sm"
                onClick={() => upsertMutation.mutate()}
                disabled={upsertMutation.isPending || !cronExpr.trim()}
              >
                {upsertMutation.isPending ? 'Saving...' : schedule ? 'Update' : 'Create'}
              </Button>
            )}
            {schedule && !hasEdited && (
              <Button
                variant="destructive"
                size="sm"
                onClick={() => {
                  if (window.confirm('Remove schedule for this task?')) {
                    deleteMutation.mutate()
                  }
                }}
                disabled={deleteMutation.isPending}
              >
                Remove
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
