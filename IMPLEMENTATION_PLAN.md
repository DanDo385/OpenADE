# OpenADE Implementation Plan

## Goal and Success Criteria

Deliver a shippable web-first MVP with the movie-picker experience working end-to-end:
chat → save as task → task wizard → run with inputs → view output + cost → rerun later.

The implementation is considered complete when:

- A user can start both services with one command.
- Chat messages stream in the browser from `/api/conversations/:id/messages`.
- Conversation can be turned into a task with detected variables.
- A saved task can be executed repeatedly with persisted outputs and cost.
- Task export/import and task history are available in UI.

## Delivery Strategy

- **Local-first**: All user data in local SQLite plus local API keys in local config.
- **Slice-based milestones**: Every milestone is independently runnable.
- **Web-first runtime**: TypeScript frontend runs in browser via Vite during development.
- **Desktop later**: Tauri wrapping is intentionally postponed until functionality is stable.

## Demo Walkthrough

1. Open OpenADE in the browser (`http://localhost:5173`).
2. Create a new conversation by typing:
   "Recommend a movie for tonight. I have Netflix and Max. I'm in the mood for something suspenseful but not horror."
3. App streams back a response with 3 recommendations and concise rationale.
4. Click **Save as Task**.
5. Wizard opens with draft:
   - Name: "Movie Recommender"
   - Prompt template:
     `Recommend a movie for tonight. I have {{streaming_services}}. I'm in the mood for {{mood}} but not {{excluded_genre}}.`
   - Variables discovered: `streaming_services`, `mood`, `excluded_genre`
6. User edits and saves as "Friday Movie Picker".
7. Task appears in Task Library.
8. User opens task, fills variables: `Hulu + Disney+`, `funny`, `drama`.
9. Clicks **Run**.
10. App shows streamed output for chat/LLM tasks and non-streamed final output for task runs, plus runtime cost.
11. User returns later, reruns the same task with changed inputs.
12. User can export and import the task package.

## System Context (MVP Scope)

```mermaid
flowchart TD
  U[User] --> FE[Frontend (React + Vite)]
  FE -->|HTTP/JSON + SSE| API[Go Backend]
  API --> DB[SQLite]
  API --> LLM[LLM Adapter]
  LLM --> PROV[Provider APIs: OpenAI]
  API -->|JSON| FE
```

## Milestones and Definition of Done

### Milestone 0 — Core chat and infra

- Backend
  - Health endpoint
  - Conversations and messages API
  - Provider config persistence
  - LLM request/response path with streaming for chat
  - SQLite initialization and migrations
- Frontend
  - Conversation list and chat UI
  - Provider modal for API key setup
  - Streaming chat rendering
  - Dark mode baseline
- Environment/run wiring in scripts and README

**Definition of Done**

- `pnpm run dev` starts services.
- Chat can be created and persisted.
- A user-visible conversation stream appears end-to-end.
- API key can be added/loaded from backend.

### Milestone 1 — Task capture and editing

- Backend
  - Meta-LLM route to extract task template candidates
  - CRUD endpoints for tasks and versions
- Frontend
  - Save-as-task action from chat
  - 4-step wizard (name/template/input list/confirm)
  - Template variable parser (`{{variable_name}}`)
  - Task edit flow with manual override

**Definition of Done**

- A conversation turns into a reusable task with at least one extracted variable.
- Task draft can be manually corrected before save.
- Updated task is persisted and re-openable.

### Milestone 2 — Task execution and history

- Backend
  - Task run endpoint
- Frontend
  - Run panel with variable input form
  - Render run output and recorded token/cost info
  - Run history list and detail view

**Definition of Done**

- Task execution returns output and cost metadata.
- Run records are searchable and visible in history.

### Milestone 3 — Memory and task management

- Backend
  - Export/import endpoints
  - Memory key-value store endpoint per task
- Frontend
  - Task library list/search/sort
  - Task export/import UI
  - Provider settings and memory editing

**Definition of Done**

- User can export one task and re-import it.
- Memory values persist across task runs.
- Task list supports text search and last-used/created sorting.

## Repository Structure

```text
backend/
  cmd/api/main.go
  go.mod
  go.sum
  internal/
    db/                 # SQLite setup and migrations
    handlers/           # HTTP handlers for each API route
    services/           # Domain services and orchestration
    llm/                # Adapter and providers
frontend/
  index.html
  vite.config.ts
  package.json
  src/
    App.tsx
    components/
    hooks/
    lib/
```

## Data Model (v1)

### conversations
- `id` UUID
- `created_at` ISO8601 datetime
- `updated_at` ISO8601 datetime

### messages
- `id` UUID
- `conversation_id` UUID fk
- `role` string (`user` or `assistant`)
- `content` text
- `created_at` ISO8601 datetime

### tasks
- `id` UUID
- `name` string
- `description` text
- `prompt_template` text
- `output_style` string (`markdown` currently)
- `version` integer
- `created_at` ISO8601 datetime
- `updated_at` ISO8601 datetime

### task_versions
- `id` UUID
- `task_id` UUID fk
- `version` integer
- `snapshot` JSON (full serialized task body)
- `created_at` ISO8601 datetime

### runs
- `id` UUID
- `task_id` UUID fk
- `version` integer
- `inputs` JSON
- `output` text
- `cost` decimal
- `model` string
- `created_at` ISO8601 datetime

### provider_configs
- `id` UUID
- `provider` string (`openai`)
- `config` JSON (encrypted at-rest later; plaintext in v1)

### memory
- `task_id` UUID fk
- `key` string
- `value` text

## API Contracts (web-facing)

### Base path and behavior
- Backend base URL: `http://localhost:8080` (configurable via `OPENADE_PORT` and `VITE_API_URL`)
- JSON request/response; SSE for streaming chat endpoint.

| Method | Route | Body | Response |
|--------|-------|------|----------|
| GET | `/health` | none | `200` + `{ "status":"ok" }` |
| POST | `/api/conversations` | none | Conversation `{ "id", ... }` |
| GET | `/api/conversations` | none | `[{...}, ...]` |
| GET | `/api/conversations/:id` | none | `{ id, messages: [...] }` |
| POST | `/api/conversations/:id/messages` | `{ "content": "..." }` | SSE stream (`role`, `content`, `done`) |
| GET | `/api/providers` | none | `[{ provider, configured }]` |
| PUT | `/api/providers/:id` | `{ "api_key": "..." }` | `{ "provider": "openai", "configured": true }` |
| POST | `/api/tasks` | `{ "conversation_id":"...", "name":"...", "template":"..." }` | Task object |
| GET | `/api/tasks` | query params `?q=` optional | task array |
| GET | `/api/tasks/:id` | none | task object + latest version |
| PUT | `/api/tasks/:id` | patch object | updated task object |
| POST | `/api/tasks/:id/run` | `{ "inputs": { ... } }` | `{ "run_id": "...", "output": "...", "cost": 0.02, "model":"gpt-4o-mini" }` |
| GET | `/api/runs` | none | run array |
| GET | `/api/runs/:id` | none | full run object |
| POST | `/api/tasks/:id/export` | none | `{ "task": {...}, "versions": [...], "memory": {...} }` |
| POST | `/api/tasks/import` | exported payload | imported task object |
| GET | `/api/memory/:task_id` | none | `{ "entries": {...} }` |
| PUT | `/api/memory/:task_id` | `{ "entries": {...} }` | `{ "ok": true }` |
| PUT | `/api/memory/:task_id/:key` | `{ "value": "..." }` | `{ "ok": true }` |

### Streaming contract (chat endpoint)

- `POST /api/conversations/:id/messages` responds with `text/event-stream`.
- Each SSE frame has JSON:
  - `{"type":"chunk","content":"..."}`
  - `{"type":"done","message_id":"...","cost":{"prompt":123,"completion":45,"total":168}}`
- Frontend appends `chunk.content` incrementally.

## State Management (Frontend)

- **Server state**: TanStack Query
  - conversation list/details
  - current task list
  - run history
- **UI state**: Zustand
  - active conversation
  - active panel (`chat`, `library`, `wizard`, `run`)
  - theme (`light` / `dark`)
  - unsaved wizard form values

## Error Handling Model

- `401` from provider route: show "API key missing" and open provider modal.
- `422` on validation: show inline form errors.
- Stream interruptions: append system note and keep composer usable.
- Backend failure: show retry action and last known safe UI state.

## Cost Display Policy

- Show cost only after run completion from backend token counts.
- Render cost note before run:
  - `Provider price is approximate and displayed after execution.`
- Track three values:
  - `input_tokens`
  - `output_tokens`
  - `total_usd`

## Task Organization (v1)

- Flat list with:
  - search query filter
  - sort order by `last_used_at` or `created_at`
- No tags, folders, or category taxonomy in MVP.

## Integration Notes

- **Default dev ports**:
  - API: `8080`
  - Frontend: `5173`
- Dev command runs both via root script (`pnpm run dev`).
- Backend is expected to create local DB file at `OPENADE_DB_PATH`.

## What Is Out of Scope

- Non-Mac desktop packaging in this milestone (Tauri is optional later).
- Card/table output rendering modes.
- Full version diff/compare UI for task revisions.
- Multi-provider provider-switching UX in M0.

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| LLM provider variance | Use provider adapter interface and contract tests |
| SQLite migration drift | Centralized migration scripts and idempotent startup initialization |
| Streaming edge cases | Keep deterministic state machine for partial chunks and done event |
| Scope creep | Milestone gating and acceptance criteria before adding features |
# OpenADE Implementation Plan

## Goal and Success Criteria

Deliver a shippable MVP with the movie-picker demo working end-to-end: chat → save as task → wizard → run with inputs → output + cost → rerun later.

## Delivery Strategy

- **Local-first** – All data on the user's machine
- **Slice-based** – Each milestone produces a runnable build

## Demo Walkthrough

1. User opens OpenADE, sees empty chat
2. Types: "Recommend a movie for tonight. I have Netflix and Max. I'm in the mood for something suspenseful but not horror."
3. Model responds with 3 suggestions with brief rationales
4. User clicks "Save as Task"
5. System shows wizard:
   - Draft name: "Movie Recommender"
   - Draft template: "Recommend a movie for tonight. I have {{streaming_services}}. I'm in the mood for {{mood}} but not {{excluded_genre}}."
   - Suggested inputs: streaming_services (multi-select), mood (select), excluded_genre (select)
6. User tweaks the name to "Friday Movie Picker", saves
7. Task appears in Task Library
8. User opens it, fills in: Hulu + Disney+, "funny", "drama"
9. Clicks Run, sees streaming response
10. Sees output + cost ($0.02) + model used
11. Next Friday, opens "Friday Movie Picker" again, changes inputs, reruns
12. User can export/import tasks

## Milestones

### Milestone 0

- Go backend (SQLite, API routes, provider adapter)
- Chat UI (streaming)
- Provider modal for API key setup
- Dark mode (shadcn/ui or equivalent)

### Milestone 1

- Meta-LLM draft generation from conversation
- Wizard UI (4 steps: name, template, inputs, confirm)
- Task persistence
- Template rendering with `{{variable_name}}`

### Milestone 2

- Task runs with full execution flow
- Cost tracking
- Run history

### Milestone 3

- Task library (list, search, sort)
- Export/import
- Memory store
- Provider settings UI

## File Structure

```
backend/           # Go
  cmd/
  internal/
    handlers/      # API handlers
    services/
    db/            # SQLite
    llm/           # LLM adapter
frontend/          # TypeScript (Vite + React)
  src/
    components/
    lib/
```

## Database Schema

- **conversations** – id, created_at, updated_at
- **messages** – id, conversation_id, role, content, created_at
- **tasks** – id, name, description, prompt_template, output_style, version, created_at, updated_at
- **task_versions** – id, task_id, version, snapshot (JSON), created_at
- **runs** – id, task_id, version, inputs (JSON), output, cost, model, created_at
- **provider_configs** – id, provider, config (JSON)
- **memory** – task_id, key, value

## API Routes (Go HTTP Handlers)

| Method | Route | Description |
|--------|-------|-------------|
| GET | /health | Health check |
| POST | /api/conversations | Create conversation |
| GET | /api/conversations | List conversations |
| GET | /api/conversations/:id | Get conversation |
| POST | /api/conversations/:id/messages | Add message, get model response (streaming) |
| POST | /api/tasks | Create task (e.g. from conversation) |
| GET | /api/tasks | List tasks |
| GET | /api/tasks/:id | Get task |
| PUT | /api/tasks/:id | Update task |
| POST | /api/tasks/:id/run | Run task |
| GET | /api/runs | List runs |
| GET | /api/runs/:id | Get run |
| POST | /api/tasks/:id/export | Export task |
| POST | /api/tasks/import | Import task |
| GET | /api/memory/:task_id | Get memory for task |
| PUT | /api/memory/:task_id | Set memory for task |
| PUT | /api/memory/:task_id/:key | Set memory key |

## State Management (Frontend)

- **TanStack Query** – Server state (conversations, tasks, runs)
- **Zustand or React Context** – UI state (current panel, selected conversation, editor state)

## Cost Display

- Show cost **after** execution based on actual token counts from the API
- Show a "~$X per 1K tokens" note before run
- Do not predict total cost before execution – it is misleading

## Task Organization

- v1: Flat list, text search, sort by last-used or created
- No folders, tags, or categories in v1

## Quick Run

Future UX: one-click rerun for tasks with no inputs or with remembered defaults. Defer to a later milestone.

## What NOT to Build

- Cards/tables output (plain text + markdown only)
- Web hosting or serverless
- Full version comparison UI
- Folders or tags for tasks
- Desktop app wrapper (Tauri can be added later for production)

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| LLM API quirks | Abstract behind model adapter; support multiple providers |
| SQLite edge cases | Use mature driver (modernc.org/sqlite); test migrations |
| Yak-shaving | Stick to milestones; cut scope before adding new tech |
