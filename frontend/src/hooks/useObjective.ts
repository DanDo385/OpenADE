import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { APIError, api } from '../lib/api'
import type { Objective, UpsertObjectiveRequest } from '../lib/api-types'

export function useObjective(conversationId: string | null) {
  const queryClient = useQueryClient()

  const query = useQuery({
    queryKey: ['objective', conversationId],
    queryFn: async (): Promise<Objective | null> => {
      if (!conversationId) return null
      try {
        return await api.getObjective(conversationId)
      } catch (e) {
        if (e instanceof APIError && e.status === 404) return null
        throw e
      }
    },
    enabled: !!conversationId,
    retry: false,
  })

  const upsertMutation = useMutation({
    mutationFn: (body: UpsertObjectiveRequest) => {
      if (!conversationId) throw new Error('No conversation selected')
      return api.upsertObjective(conversationId, body)
    },
    onSuccess: () => {
      if (conversationId) {
        queryClient.invalidateQueries({ queryKey: ['objective', conversationId] })
      }
    },
  })

  return {
    objective: query.data ?? null,
    isLoading: query.isLoading,
    error: query.error,
    upsert: upsertMutation.mutateAsync,
    isUpserting: upsertMutation.isPending,
  }
}
