# OpenADE Implementation Parts

Detailed split for parallel execution with two primary streams and explicit handoff contracts.

---

## Piece 1: Backend (Go + SQLite + HTTP API)

**Goal:** Provide a reliable local API that serves the frontend and runs LLM operations.

### Deliverables

1. Project bootstrap
   - `backend/cmd/api/main.go`
   - `backend/go.mod`, `go.sum`
   - env support:
     - `OPENADE_PORT` (default: `8080`)
     - `OPENADE_DB_PATH` (default: `./backend/openade.db` or env-provided absolute path)

2. Database layer (`backend/internal/db`)
   - connection helper and graceful startup
   - migration runner
   - tables:
     - `conversations`
     - `messages`
     - `tasks`
     - `task_versions`
     - `runs`
     - `provider_configs`
     - `memory`

3. API handler layer (`backend/internal/handlers`)
   - `GET /health`
   - conversations CRUD + message posting
   - tasks CRUD + run + export/import
   - runs read path
   - provider config paths
   - memory paths

4. Service layer (`backend/internal/services`)
   - conversation orchestration
   - task drafting orchestration
   - task execution
   - persistence and serialization

5. LLM layer (`backend/internal/llm`)
   - adapter interface:
     - `Stream(ctx, messages, model, params) (stream, err)`
     - `Complete(ctx, prompt, model, params) (result, err)`
   - OpenAI provider implementation first
   - provider config validation
   - token usage extraction for cost reporting

6. Cross-cutting
   - response models
   - request validation and error envelopes
   - logging middleware and CORS config

### Suggested File Layout

```text
backend/
  cmd/
    api/
      main.go
  internal/
    db/
      sqlite.go
      migrations/
    handlers/
      health.go
      conversations.go
      tasks.go
      runs.go
      memory.go
      providers.go
    services/
      conversation_service.go
      task_service.go
      run_service.go
      memory_service.go
    llm/
      adapter.go
      openai.go
    model/
      types.go
```

### API Surface to Implement

| Method | Route | Purpose |
|--------|-------|---------|
| GET | `/health` | Health check |
| POST | `/api/conversations` | Create conversation |
| GET | `/api/conversations` | List conversations |
| GET | `/api/conversations/:id` | Fetch conversation + messages |
| POST | `/api/conversations/:id/messages` | Post chat message and stream reply via SSE |
| GET | `/api/providers` | Read configured providers |
| PUT | `/api/providers/:id` | Save provider config |
| POST | `/api/tasks` | Create task from conversation |
| GET | `/api/tasks` | List tasks |
| GET | `/api/tasks/:id` | Fetch task detail |
| PUT | `/api/tasks/:id` | Update task |
| POST | `/api/tasks/:id/run` | Execute task and store run |
| GET | `/api/runs` | List runs |
| GET | `/api/runs/:id` | Fetch run detail |
| POST | `/api/tasks/:id/export` | Export task bundle |
| POST | `/api/tasks/import` | Import task bundle |
| GET | `/api/memory/:task_id` | Fetch task memory |
| PUT | `/api/memory/:task_id` | Replace memory object |
| PUT | `/api/memory/:task_id/:key` | Update one memory key |

### API Contract Details to match

- Create conversation response:
  `{ "id": "uuid", "created_at": "2026-03-08T..." }`
- POST message request:
  `{ "content": "string", "model": "gpt-4o-mini", "stream": true }`
- Task create request:
  `{ "conversation_id": "uuid", "name": "Movie Picker", "template": "..." }`
- Run request:
  `{ "inputs": { "streaming_services": "Hulu,Disney+", "mood": "funny", "excluded_genre": "drama" } }`
- Run response:
  `{ "id": "uuid", "task_id": "uuid", "output": "string", "cost_usd": 0.02, "model": "gpt-4o-mini", "input_tokens": 123, "output_tokens": 45, "created_at": "..." }`
- Streaming message events:
  - `{"type":"chunk","content":"..."}`
  - `{"type":"done","message_id":"uuid","cost":{"input_tokens":123,"output_tokens":45}}`

### Done Conditions for Backend piece

- A conversation can be created and messaged from the frontend with streaming.
- A task can be drafted from a conversation and saved.
- A saved task can be run and returns output + cost.
- Memory gets saved and retrieved.
- Export/import payloads round-trip.

### Suggested Local Run

```bash
cd backend
OPENADE_PORT=8080 OPENADE_DB_PATH=./openade.db go run ./cmd/api
```

---

## Piece 2: Frontend (TypeScript + React + Vite)

**Goal:** Create a complete web app experience for chat, task management, runs, and memory.

### Deliverables

1. Core app shell
   - route-less layout with persistent sidebar and panes
   - theme state and provider status banner
   - API availability and errors

2. Chat feature set
   - conversation creation and selection
   - message streaming render with markdown support
   - "Save as Task" action from active conversation

3. Task feature set
   - 4-step task wizard
   - task editor with version-safe save
   - list + search + sort
   - run form generation from inferred variables

4. Run and memory
   - submit inputs and show live output where required
   - cost line item display
   - rerun with previous defaults
   - memory key/value inspector and update editor

5. Data and integration
   - TanStack Query setup and request cache invalidation
   - Zustand store for UI-only state
   - API client and SSE parser

6. Export/import UX
   - export JSON blob
   - import flow with basic JSON validation
   - success/failure toasts

### Suggested File Layout

```text
frontend/
  src/
    main.tsx
    App.tsx
    components/
      providers/
      chat/
      tasks/
      runs/
      ui/
    hooks/
      useConversations.ts
      useMessages.ts
      useStreaming.ts
      useRuns.ts
      useTasks.ts
    lib/
      api.ts
      api-types.ts
      store.ts
      validation.ts
      templates.ts
```

### Component Responsibilities

- `App.tsx`: app shell, top-level view switching, global error boundary.
- `MessageList` / `MessageInput`: chat render and message submission.
- `ProviderModal`: open when no provider exists and before first LLM call.
- `TaskWizard`: 4-step UI for name/template/inputs/confirm.
- `TaskLibrary`: list, filter, sort.
- `RunPanel`: render run form and output block.
- `MemoryPanel`: view/update task memory map.

### Dependency and State Plan

- **TanStack Query**
  - Query keys:
    - `["conversations"]`
    - `["conversation", id]`
    - `["tasks"]`
    - `["task", id]`
    - `["runs"]`
    - `["run", id]`
- **Zustand**
  - UI state:
    - active conversation id
    - active task id
    - right panel mode
    - current theme
    - unsaved wizard draft
- `react-markdown` for assistant output
- Optional style stack:
  - Tailwind CSS + shadcn/ui primitives

### Done Conditions for Frontend piece

- New conversation can be created and messaged.
- Streaming chunks render while backend is generating response.
- User can convert conversation into task through wizard.
- Task can be edited, saved, searched, and run with dynamic input controls.
- Export/import works from UI actions.
- Memory editing persists and updates UI.

### Suggested Local Run

```bash
cd frontend
pnpm install
pnpm dev
```

---

## Load 3: Integration + Runtime Harness (root + shared contracts)

**Goal:** Ensure both streams (backend + frontend) start, speak, recover, and stay stable without rewriting business logic.

### Deliverables

1. Runtime orchestration
   - keep `pnpm run dev`, `pnpm run dev:backend`, `pnpm run dev:frontend`, and `pnpm run health` reliable
   - unify env loading and defaults across root, backend, and frontend
2. API client and SSE hardening (`frontend/lib/api.ts`)
   - single base URL client with auth/error wrappers
   - shared SSE parser for `chunk` and `done`
3. Shared error envelope handling
   - map API error shape to consistent UI state:
     `{ "error": { "code", "message", "details" } }`
4. Provider handshake flow
   - if backend returns provider-missing, open provider modal and retry once
   - keep conversation UI usable while provider is configured
5. Backend runtime reliability
   - request logging for streaming and run endpoints
   - health endpoint includes readiness + DB accessibility

### Suggested Files

```text
scripts/
  dev.sh
  backend.sh
  healthcheck.sh
frontend/
  src/lib/api.ts
  src/lib/api-types.ts
  src/lib/store.ts
```

### Definition of Done

- one-command boot (`pnpm run dev`) reaches a stable chat flow
- missing provider error appears with actionable modal
- streaming endpoint always emits `chunk` and final `done`
- health command returns explicit service readiness

### References

- [`IMPLEMENTATION_PLAN.md`](IMPLEMENTATION_PLAN.md) – API contracts and integration notes
- [`ARCHITECTURE.md`](ARCHITECTURE.md) – runtime topology, error handling, module boundaries

---

## Load 4: QA, Acceptance Validation, and DX Documentation

**Goal:** Add concrete checks so any agent can validate progress without guessing.

### Deliverables

1. Deterministic smoke flow script in `scripts/`
   - conversation create + message stream
   - task create from conversation
   - task run + cost capture
   - export/import roundtrip
2. Fixture set for onboarding
   - sample task export JSON and expected run output format
3. Checklist documentation
   - milestone-to-acceptance mapping for backend and frontend pieces
4. Recovery UX polish
   - clear empty states, invalid import handling, and retry actions
5. Basic local observability
   - log markers for startup, stream start, run complete, and errors

### Suggested Files

```text
scripts/
  smoke.sh
  fixtures/
    movie-picker-demo.json
    exported-task-sample.json
README.md
```

### Definition of Done

- run smoke flow in under one command
- fixture import/export path is repeatable
- each acceptance item maps to a route or UI action in the docs

### References

- [`IMPLEMENTATION_PLAN.md`](IMPLEMENTATION_PLAN.md) – milestones, quick run, and cost policy
- [`IMPLEMENTATION_PARTS.md`](IMPLEMENTATION_PARTS.md) – backend/frontend done conditions
- [`ARCHITECTURE.md`](ARCHITECTURE.md) – reliability and security expectations

---

## Load 5: UI Delight + Quiz Teaching Experiences (Shadcn + React components)

**Goal:** Build a polished, interactive interface with reusable Shadcn component patterns and guided teaching sessions.

### Deliverables

1. Component system rollout
   - adopt shadcn UI primitives in shared layout and controls:
     - `Button`, `Input`, `Dialog`, `Tabs`, `Card`, `Textarea`, `Select`, `Toast`
   - create a design token strategy in CSS variables for consistent spacing and tone.
2. App shell polish
   - dashboard-first layout with quick action rail.
   - empty states, loading states, and empty conversation/task visuals.
   - animated transitions for panel switches and modal entrances.
3. Interactive quiz mode
   - create quiz session model:
     - question, choices, expected answer, explanation, difficulty, tags
   - route in UI:
     - list quiz sessions
     - run quiz
     - show progress, per-question feedback, final score
4. Slash command UI affordance
   - input parser for `/help`, `/quiz`, `/clear`, `/run`, `/export`, `/import`
   - suggestions dropdown with command descriptions.
5. "Sweetness" polish pass
   - consistent button hierarchy and visual rhythm.
   - theme toggle and high-contrast safe color profile.
   - responsive behavior for narrow screens.

### Suggested Files

```text
frontend/
  src/
    components/ui/                      # Shadcn primitives
      button.tsx
      card.tsx
      dialog.tsx
      input.tsx
      select.tsx
      tabs.tsx
      textarea.tsx
      toast.tsx
    components/
      QuizSessionPanel.tsx
      QuizRunner.tsx
      CommandPalette.tsx
      AppShell.tsx
    lib/
      quiz.ts
      theme.ts
```

### Definition of Done

- Chat and task screens use shadcn-styled shared components.
- Slash commands are discoverable before sending any message.
- A quiz can be launched and completed with score shown.
- Motion and spacing changes are visibly improved without harming usability.
- Visual regressions are manually checked at 1280, 768, 390 widths.

### References

- [`ARCHITECTURE.md`](ARCHITECTURE.md) – Interaction Surface, Security and privacy, Agent and Game Orchestration Concept
- [`IMPLEMENTATION_PLAN.md`](IMPLEMENTATION_PLAN.md) – Cost policy, state management, local verification sequence

---

## Load 6: Terminal-Safe Slash Commands + Game Agents

**Goal:** Add command-style workflows and local agent execution hooks for setup and game bootstrapping.

### Deliverables

1. Slash command runtime contracts
   - backend route `POST /api/commands/execute`.
   - frontend command pre-parser and optimistic UI states.
2. Command execution safety
   - allowlist strategy for command families:
     - `terminal`, `scaffold`, `open`, `play`.
   - command request payload with explicit confirmation flag.
3. Agent scaffold for game flows
   - `POST /api/agents` create/update agent metadata.
   - `POST /api/agents/:id/run` executes a command/script bundle.
   - output returned in one payload + optional logs stream.
4. Teaching setup templates
   - `/quiz` can generate onboarding checklist + command snippets.
   - `/agent:spin-game` triggers a pre-approved game bootstrap script bundle.
5. UX feedback loop
   - inline logs stream pane while command/agent runs.
   - completion badge with exit code and duration.

### Suggested Files

```text
backend/
  internal/
    handlers/
      commands.go
      agents.go
      quiz_sessions.go
    services/
      command_service.go
      agent_service.go
    llm/
      prompt_agent_templates.go
frontend/
  src/
    components/
      CommandPalette.tsx
      CommandOutputPanel.tsx
      AgentLauncher.tsx
    lib/
      commandBus.ts
      api.ts
```

### Definition of Done

- `/help`, `/run`, and `/quiz` commands resolve to action paths.
- command execution returns `ok`, `stdout`, `stderr`, `exit_code`, and `duration_ms`.
- game agent launch can be triggered from UI and emits log output.
- teaching session setup can call command/agent helper and continue in-app.
- dangerous commands are blocked by backend allowlist.

### References

- [`IMPLEMENTATION_PLAN.md`](IMPLEMENTATION_PLAN.md) – API contracts (extend for command/session endpoints)
- [`ARCHITECTURE.md`](ARCHITECTURE.md) – Command execution and safety, Agent concept
- [`IMPLEMENTATION_PARTS.md`](IMPLEMENTATION_PARTS.md) – Load 3 and Load 4 handoff expectations

---

## Load 7: Quiz Backend + Session Persistence (teaching engine)

**Goal:** Persist quiz sessions, scores, and optional LLM-generated quizzes so teaching flows are repeatable and auditable.

### Dependencies

- Load 5 (quiz UI surface)
- Piece 1 (backend, DB, API base)

### Deliverables

1. Quiz schema and migrations
   - `quiz_sessions` – id, title, description, created_at, updated_at
   - `quiz_questions` – id, session_id, question, choices (JSON), correct_index, explanation, difficulty, order
   - `quiz_attempts` – id, session_id, user_label, started_at, completed_at
   - `quiz_answers` – id, attempt_id, question_id, selected_index, correct, answered_at
2. Quiz API
   - `GET /api/quiz-sessions` – list sessions
   - `GET /api/quiz-sessions/:id` – session + questions
   - `POST /api/quiz-sessions` – create session (manual or from template)
   - `POST /api/quiz-sessions/:id/attempts` – start attempt
   - `POST /api/quiz-sessions/:id/attempts/:aid/answers` – submit answer
   - `GET /api/quiz-sessions/:id/attempts/:aid` – attempt + score
3. Optional LLM quiz generator
   - meta-LLM route to generate quiz from topic or conversation
4. Seed fixtures
   - onboarding quiz, CLI basics quiz

### Suggested Files

```text
backend/
  internal/
    db/migrations/...quiz...
    handlers/quiz.go
    services/quiz_service.go
  fixtures/
    quizzes/
      onboarding.json
      cli-basics.json
```

### Definition of Done

- Quiz sessions can be created and listed.
- An attempt can be started, answers submitted, and score computed.
- Quiz results persist and are viewable in UI (extends Load 5).

### References

- [`IMPLEMENTATION_PLAN.md`](IMPLEMENTATION_PLAN.md) – API contracts, data model
- [`ARCHITECTURE.md`](ARCHITECTURE.md) – Interaction Surface, Agent concept

---

## Load 8: Game Catalog + Agent Library (curated game agents)

**Goal:** Provide a library of game-ready agents and templates so users can spin up games without custom setup.

### Dependencies

- Load 6 (command execution, agent run)
- Piece 2 (task library, run panel)

### Deliverables

1. Agent catalog schema
   - `agents` – id, name, slug, description, instructions, script_bundle (JSON), enabled
   - predefined rows for starter games (e.g. blackjack, trivia, hangman)
2. Agent API
   - `GET /api/agents` – list agents
   - `GET /api/agents/:id` – agent detail
   - `POST /api/agents/:id/run` – run agent (existing or refined)
3. Game script bundles
   - safe scripts for: start, stop, scaffold project
   - bundled with templates (e.g. `play blackjack`, `play trivia`)
4. Frontend agent library
   - `AgentLibrary` – list, filter, launch
   - `AgentLauncher` – run with confirmation + output pane
5. Integration with slash commands
   - `/agent:blackjack`, `/agent:trivia` etc. resolve to catalog entries

### Suggested Files

```text
backend/
  internal/
    db/migrations/...agents...
    handlers/agents.go
    services/agent_service.go
  fixtures/
    agents/
      blackjack.json
      trivia.json
frontend/
  src/components/AgentLibrary.tsx
  src/components/AgentLauncher.tsx
  src/lib/agents.ts
```

### Definition of Done

- Agent catalog is queryable from UI.
- User can launch a predefined game agent and see output.
- Slash command `/agent:<slug>` launches catalog agent.

### References

- [`ARCHITECTURE.md`](ARCHITECTURE.md) – Agent and Game Orchestration Concept
- [`IMPLEMENTATION_PARTS.md`](IMPLEMENTATION_PARTS.md) – Load 6 command/agent contracts

---

## Load 9: Polish, Accessibility, and Command Palette (UX lift)

**Goal:** Make the app feel native-grade with shortcuts, accessibility, and a global command surface.

### Dependencies

- Load 5 (Shadcn, slash commands UI)
- Piece 2 (app shell, layout)

### Deliverables

1. Global command palette
   - `Cmd+K` / `Ctrl+K` opens palette
   - search: conversations, tasks, agents, quiz sessions, slash commands
   - execute slash commands from palette
2. Keyboard shortcuts
   - `Cmd+N` new conversation
   - `Cmd+/` focus chat input
   - `Escape` close modals / cancel
   - shortcuts documented in `/help` or settings
3. Accessibility
   - ARIA labels on interactive elements
   - focus trap in modals
   - skip-to-content
   - high-contrast / reduced-motion support
4. Animation and micro-interactions
   - subtle transitions on panel switch, toast, list items
   - loading skeletons where appropriate
5. Settings panel
   - shortcuts list, theme, reduced motion toggle

### Suggested Files

```text
frontend/
  src/
    components/
      CommandPalette.tsx
      ShortcutHelp.tsx
      SettingsPanel.tsx
    hooks/
      useKeyboardShortcuts.ts
      useCommandPalette.ts
    lib/
      shortcuts.ts
  styles/
    motion.css
```

### Definition of Done

- Command palette opens and search works across conversations, tasks, agents.
- Documented shortcuts work as specified.
- Lighthouse accessibility score acceptable; no critical a11y violations.
- Motion respects `prefers-reduced-motion` when toggled.

### References

- [`IMPLEMENTATION_PLAN.md`](IMPLEMENTATION_PLAN.md) – state management, UI layout
- [`ARCHITECTURE.md`](ARCHITECTURE.md) – Interaction Surface, Security

---

## Load 10: Desktop Packaging and Release (Tauri + distribution)

**Goal:** Ship a distributable Mac desktop app with Tauri and optional release automation.

### Dependencies

- Loads 1–4 (core flows stable)
- Load 9 (optional but recommended for polish)

### Deliverables

1. Tauri setup
   - `src-tauri/` with Cargo.toml, tauri.conf.json
   - backend subprocess spawn and health check
   - frontend served from Tauri WebView
2. Build and run scripts
   - `pnpm run tauri:dev`, `pnpm run tauri:build`
   - clear docs for local vs packaged run
3. Packaging config
   - app icon, product name, bundle ID
   - code signing placeholder (docs only for now)
4. Release docs
   - how to build for Mac
   - how to update frontend/backend in packaged app
5. Optional: keychain integration for API keys (deferred to follow-up)

### Suggested Files

```text
src-tauri/
  Cargo.toml
  tauri.conf.json
  src/
    main.rs
    lib.rs
  icons/
package.json    # tauri scripts
docs/
  RELEASE.md
  TAURI_SETUP.md
```

### Definition of Done

- `pnpm run tauri:dev` runs the app in a native window with backend.
- `pnpm run tauri:build` produces a distributable `.app` (or equivalent).
- README documents Tauri workflow for contributors.

### References

- [`ARCHITECTURE.md`](ARCHITECTURE.md) – Deployment Model, Security (keychain)
- [`IMPLEMENTATION_PLAN.md`](IMPLEMENTATION_PLAN.md) – Future (Production)
- [`README.md`](README.md) – run commands

---

## Load Summary (5–10)

| Load | Focus | Primary owner type |
|------|-------|---------------------|
| 5 | UI Delight + Quiz Teaching (Shadcn, quiz UI, slash affordances) | Frontend |
| 6 | Terminal-Safe Slash Commands + Game Agents | Backend + Frontend |
| 7 | Quiz Backend + Session Persistence | Backend |
| 8 | Game Catalog + Agent Library | Backend + Frontend |
| 9 | Polish, Accessibility, Command Palette | Frontend |
| 10 | Desktop Packaging (Tauri + Release) | DevOps / Integrator |

---

## Load 11: MCP Server Configuration + Settings UI

**Goal:** Allow users to configure MCP servers from the UI. All MCP servers must be addable, removable, and manageable through the app—no config-file editing required.

**Cross-cutting requirement:** All MCP surfaces (servers, tools, resources, prompts) must be exposed and controllable via the UI.

### Dependencies

- Load 4 (Provider settings pattern)
- Piece 2 (Settings panel)

### Deliverables

1. **MCP server schema and persistence**
   - `mcp_servers` table: id, name, transport (stdio|sse), command_or_url, args_json, env_json, enabled, created_at, updated_at
   - migrations for MCP tables

2. **Backend MCP client**
   - `internal/mcp/` – Go MCP SDK client, connect to stdio/SSE servers
   - `internal/services/mcp_service.go` – list servers, add, remove, test connection
   - API: `GET /api/mcp/servers`, `POST /api/mcp/servers`, `PUT /api/mcp/servers/:id`, `DELETE /api/mcp/servers/:id`, `POST /api/mcp/servers/:id/test`

3. **Frontend MCP Settings UI**
   - `frontend/src/components/settings/MCPServersPanel.tsx` – list configured MCP servers
   - Add server form (name, transport type, command or URL, optional args/env)
   - Enable/disable toggle per server
   - Test connection button with status feedback
   - Empty state when no servers configured

4. **Integration with Provider Settings**
   - MCP section in Settings; all MCPs manageable from UI
   - Wire into existing ProviderSettings or Settings panel

### Suggested Files

```text
backend/
  internal/
    mcp/
      client.go
      stdio.go
      sse.go
    db/migrations/...mcp...
    handlers/mcp.go
    services/mcp_service.go
frontend/
  src/
    components/settings/
      MCPServersPanel.tsx
      AddMCPServerForm.tsx
    lib/
      api-mcp.ts
```

### Definition of Done

- User can add, edit, remove, and enable/disable MCP servers from the UI.
- Test connection returns success/failure for each server.
- All MCP server configuration is stored and loaded from DB; no manual config files.

### References

- [`mcp-docs/`](mcp-docs/) – MCP protocol, transport options
- [`ARCHITECTURE.md`](ARCHITECTURE.md) – MCP integration point

---

## Load 12: MCP Tool Discovery + Task/Chat Integration

**Goal:** Surface MCP tools, resources, and prompts in the UI and make them available for chat and task execution.

### Dependencies

- Load 11 (MCP server configuration)
- Piece 2 (Chat, Tasks, Run panel)

### Deliverables

1. **Backend tool/resource discovery**
   - `GET /api/mcp/servers/:id/tools` – list tools from a server
   - `GET /api/mcp/servers/:id/resources` – list resources
   - `GET /api/mcp/servers/:id/prompts` – list prompts
   - `POST /api/mcp/tools/call` – invoke a tool (server_id, tool_name, arguments)
   - `GET /api/mcp/resources/:uri/read` – read a resource by URI

2. **Frontend MCP Tools UI**
   - `frontend/src/components/mcp/MCPToolsPanel.tsx` – list all tools across enabled servers
   - `frontend/src/components/mcp/MCPResourcesPanel.tsx` – browse resources
   - `frontend/src/components/mcp/MCPPromptsPanel.tsx` – browse prompts
   - Tool picker in Task Wizard (optional: "Use MCP tool" when defining task)
   - Inline tool-call display in chat and run output (show tool name, args, result)

3. **Chat/task integration**
   - Backend: optional MCP tool calls during chat completion or task run
   - UI: display tool calls and results in MessageList and RunPanel
   - All MCP tools discoverable and invokable from UI (even if not yet wired to LLM)

### Suggested Files

```text
backend/
  internal/
    handlers/mcp_tools.go
    services/mcp_tool_service.go
frontend/
  src/
    components/mcp/
      MCPToolsPanel.tsx
      MCPResourcesPanel.tsx
      MCPPromptsPanel.tsx
      ToolCallBadge.tsx
    hooks/
      useMCPTools.ts
      useMCPResources.ts
    lib/
      api-mcp.ts  (extend)
```

### Definition of Done

- All MCP tools from enabled servers are listed in the UI.
- User can browse resources and prompts from connected servers.
- Tool calls (if supported in chat/task flow) are visible in message and run output.
- MCP surfaces (tools, resources, prompts) are fully exposed in the UI.

### References

- [MCP Complete Reference](mcp-docs/MCP_Complete_Reference.md) – Tools, Resources, Prompts
- [MCP GitHub DeepDive](mcp-docs/MCP_GitHub_DeepDive.md) – SDKs, Inspector

---

## Load 13: MCP Registry + Full UI Surface

**Goal:** Browse and install MCP servers from the registry, and ensure every MCP primitive has a dedicated UI surface.

### Dependencies

- Load 11, Load 12 (MCP config + tools)

### Deliverables

1. **MCP Registry integration**
   - Backend: optional `GET /api/mcp/registry/search?q=` – proxy or cache from MCP Registry
   - Or frontend-only: fetch from public registry API if available
   - Display curated/recommended servers with install action

2. **Full MCP UI surface**
   - **Servers** – MCPServersPanel (Load 11): add, edit, remove, enable/disable, test
   - **Tools** – MCPToolsPanel: list, search, invoke, show schema
   - **Resources** – MCPResourcesPanel: list, read, subscribe (if supported)
   - **Prompts** – MCPPromptsPanel: list, preview, insert into chat/task
   - **Registry** – `frontend/src/components/mcp/MCPRegistryPanel.tsx`: search, install from registry
   - Navigation: Settings → MCP, or dedicated MCP section in sidebar

3. **Empty states and onboarding**
   - "No MCP servers configured" with link to registry or add form
   - "No tools available" when no servers are enabled
   - Short onboarding copy for MCP (what it is, why use it)

4. **Smoke/acceptance**
   - Extend `scripts/smoke.sh`: add MCP server, list tools, optional tool call
   - Fixture: sample stdio MCP server config for smoke

### Suggested Files

```text
frontend/
  src/
    components/mcp/
      MCPRegistryPanel.tsx
      MCPOverview.tsx
      MCPSidebarSection.tsx
    lib/
      api-mcp.ts
      mcp-registry.ts
scripts/
  smoke.sh          (extend)
  fixtures/
    mcp-server-sample.json
```

### Definition of Done

- All MCPs are added and manageable from the UI (servers, tools, resources, prompts).
- User can search and install servers from MCP Registry (or equivalent) via UI.
- Every MCP primitive has a dedicated panel/section; nothing requires config-file editing.
- Smoke script covers MCP add + tool discovery.

### References

- [MCP Registry](https://modelcontextprotocol.io/registry/about)
- [mcp-docs/](mcp-docs/)

---

## Load Summary (5–13)

| Load | Focus | Primary owner type |
|------|-------|---------------------|
| 5 | UI Delight + Quiz Teaching (Shadcn, quiz UI, slash affordances) | Frontend |
| 6 | Terminal-Safe Slash Commands + Game Agents | Backend + Frontend |
| 7 | Quiz Backend + Session Persistence | Backend |
| 8 | Game Catalog + Agent Library | Backend + Frontend |
| 9 | Polish, Accessibility, Command Palette | Frontend |
| 10 | Desktop Packaging (Tauri + Release) | DevOps / Integrator |
| **11** | **MCP Server Configuration + Settings UI** | Backend + Frontend |
| **12** | **MCP Tool Discovery + Task/Chat Integration** | Backend + Frontend |
| **13** | **MCP Registry + Full UI Surface** | Frontend |

---

## Integration Contract

Both pieces should align on:

1. Endpoint URLs and HTTP status semantics.
2. SSE event shape (`chunk`, `done`).
3. Task variable parser (`{{variable_name}}`) semantics.
4. Run output + cost keys.
5. Error payload shape:
   - `{ "error": { "code": "string", "message": "string", "details": {} } }`
6. **MCP UI requirement:** All MCP primitives (servers, tools, resources, prompts) must be configurable, discoverable, and usable from the UI—no config-file editing required. Loads 11–13 implement this.

## Local Verification Sequence

1. `pnpm run dev` from repo root.
2. Navigate to UI and create conversation.
3. Receive streaming response.
4. Save as task and complete wizard.
5. Run task and inspect cost.
6. Export task, clear local state manually (optional), import task, rerun.

## Future (Production)

Tauri can wrap the current web UI without redesigning backend contracts:

- `src-tauri/` starts backend subprocess.
- UI communicates through HTTP to localhost or controlled bridge.
- Existing REST and SSE contracts remain unchanged.
