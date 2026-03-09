import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import type { Conversation } from '../lib/api-types'

export const conversationsQueryKey = ['conversations'] as const

export function conversationQueryKey(id: string): readonly ['conversation', string] {
  return ['conversation', id] as const
}

export function useConversations() {
  const queryClient = useQueryClient()

  const listQuery = useQuery({
    queryKey: conversationsQueryKey,
    queryFn: api.listConversations,
  })

  const createMutation = useMutation({
    mutationFn: api.createConversation,
    onSuccess: (conversation) => {
      queryClient.setQueryData<Conversation[]>(conversationsQueryKey, (current) => {
        if (!current) {
          return [conversation]
        }
        return [conversation, ...current]
      })
      queryClient.setQueryData(conversationQueryKey(conversation.id), conversation)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteConversation(id),
    onSuccess: (_result, id) => {
      queryClient.setQueryData<Conversation[]>(conversationsQueryKey, (current) =>
        (current ?? []).filter((conversation) => conversation.id !== id),
      )
      queryClient.removeQueries({ queryKey: conversationQueryKey(id), exact: true })
    },
  })

  return {
    conversations: listQuery.data ?? [],
    isLoading: listQuery.isLoading,
    isFetching: listQuery.isFetching,
    error: listQuery.error,
    createConversation: createMutation.mutateAsync,
    isCreating: createMutation.isPending,
    createError: createMutation.error,
    deleteConversation: deleteMutation.mutateAsync,
    isDeleting: deleteMutation.isPending,
    deleteError: deleteMutation.error,
    refetch: listQuery.refetch,
  }
}
