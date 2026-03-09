import { useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'

export function useRuns() {
  const { data: runs = [], isLoading } = useQuery({
    queryKey: ['runs'],
    queryFn: api.listRuns,
  })

  return {
    runs,
    isLoading,
  }
}
