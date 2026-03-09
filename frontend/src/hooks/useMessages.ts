import { useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'
import { conversationQueryKey } from './useConversations'

export function useMessages(conversationId: string | null) {
  const query = useQuery({
    queryKey: conversationId ? conversationQueryKey(conversationId) : ['conversation', 'none'],
    queryFn: () => {
      if (!conversationId) {
        throw new Error('conversation id is required')
      }
      return api.getConversation(conversationId)
    },
    enabled: Boolean(conversationId),
  })

  return {
    ...query,
    messages: query.data?.messages ?? [],
  }
}
