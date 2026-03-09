import { useCallback, useRef, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { APIError, api } from '../lib/api'
import type { Conversation, Message } from '../lib/api-types'
import { conversationQueryKey, conversationsQueryKey } from './useConversations'

interface SendMessageParams {
  conversationId: string
  content: string
  model?: string
  onUnauthorized?: () => void
}

function nowISO(): string {
  return new Date().toISOString()
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException && error.name === 'AbortError'
}

function toUserFacingError(error: unknown): string {
  if (error instanceof APIError) {
    return error.message
  }
  if (error instanceof Error) {
    return error.message
  }
  return 'unexpected streaming error'
}

export function useStreaming() {
  const queryClient = useQueryClient()
  const [isStreaming, setIsStreaming] = useState(false)
  const [streamingContent, setStreamingContent] = useState('')
  const [streamError, setStreamError] = useState<string | null>(null)
  const abortControllerRef = useRef<AbortController | null>(null)

  const cancel = useCallback(() => {
    abortControllerRef.current?.abort()
  }, [])

  const sendMessage = useCallback(
    async ({ conversationId, content, model, onUnauthorized }: SendMessageParams) => {
      setIsStreaming(true)
      setStreamError(null)
      setStreamingContent('')

      const optimisticUserMessage: Message = {
        id: `temp-user-${Date.now()}`,
        conversation_id: conversationId,
        role: 'user',
        content,
        created_at: nowISO(),
      }

      queryClient.setQueryData<Conversation>(conversationQueryKey(conversationId), (current) => {
        if (!current) {
          return current
        }
        return {
          ...current,
          updated_at: nowISO(),
          messages: [...(current.messages ?? []), optimisticUserMessage],
        }
      })

      abortControllerRef.current = new AbortController()

      try {
        await api.streamConversationMessage(
          conversationId,
          { content, model },
          (event) => {
            if (event.type === 'chunk') {
              setStreamingContent((current) => current + event.content)
              return
            }
            if (event.type === 'error') {
              setStreamError(event.message)
            }
          },
          abortControllerRef.current?.signal,
        )
      } catch (error) {
        if (error instanceof APIError && error.status === 401) {
          onUnauthorized?.()
        }
        if (!isAbortError(error)) {
          setStreamError(toUserFacingError(error))
        }
        throw error
      } finally {
        abortControllerRef.current = null
        setIsStreaming(false)
        await Promise.all([
          queryClient.invalidateQueries({ queryKey: conversationQueryKey(conversationId), exact: true }),
          queryClient.invalidateQueries({ queryKey: conversationsQueryKey }),
        ])
        setStreamingContent('')
      }
    },
    [queryClient],
  )

  return {
    isStreaming,
    streamingContent,
    streamError,
    sendMessage,
    cancel,
  }
}
