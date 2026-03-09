import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { CreateTaskRequest, UpdateTaskRequest } from '../lib/api-types'
import { api } from '../lib/api'

export function useTasks(query?: string) {
  const queryClient = useQueryClient()

  const { data: tasks = [], isLoading } = useQuery({
    queryKey: ['tasks', query ?? ''],
    queryFn: () => api.listTasks(query),
  })

  const createMutation = useMutation({
    mutationFn: (body: CreateTaskRequest) => api.createTask(body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateTaskRequest }) =>
      api.updateTask(id, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
      queryClient.invalidateQueries({ queryKey: ['task'] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteTask(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] })
      queryClient.invalidateQueries({ queryKey: ['task'] })
    },
  })

  return {
    tasks,
    isLoading,
    createTask: createMutation.mutateAsync,
    isCreating: createMutation.isPending,
    updateTask: updateMutation.mutateAsync,
    deleteTask: deleteMutation.mutateAsync,
  }
}
