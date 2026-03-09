import { useQuery } from '@tanstack/react-query'
import { api } from '../../lib/api'
import { AgentLauncher } from './AgentLauncher'

export function AgentLibrary() {
  const { data: agents = [], isLoading, error } = useQuery({
    queryKey: ['agents'],
    queryFn: api.listAgents,
  })

  return (
    <section className="agent-library">
      <h2>Agents</h2>
      <p className="muted">Game and utility agents. Click to run.</p>

      {isLoading && <p className="muted">Loading agents...</p>}
      {error && <p className="error-text">Failed to load agents.</p>}

      {!isLoading && !error && agents.length === 0 && (
        <div className="empty-state-block">
          <p className="empty-state">No agents yet.</p>
          <p className="muted">Seed data adds Blackjack and Trivia on first run.</p>
        </div>
      )}

      {!isLoading && !error && agents.length > 0 && (
        <ul className="agent-library__items">
          {agents.map((agent) => (
            <li key={agent.id}>
              <AgentLauncher agent={agent} />
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
