# OpenADE Status Update and Detailed Plan (Rest of Project)

Last updated: 2026-03-09

## 1) Canonical references

All execution and decisions below are anchored to these docs:

- [ARCHITECTURE.md](ARCHITECTURE.md)
- [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md)
- [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md)
- [README.md](README.md)
- Tauri runtime constraints: [src-tauri/src/lib.rs](src-tauri/src/lib.rs)

## 2) Where we are right now

### 2.1 Verified runtime status

- Backend compiles: `go build -o openade-backend ./cmd/api` in `backend/`.
- Backend package checks run: `go test ./...` (no test files yet).
- Frontend compiles: `pnpm run build` in `frontend/`.
- Local API smoke check passed for non-LLM paths:
  - `GET /health`
  - `POST /api/conversations`
  - `GET /api/conversations/:id`
  - `POST /api/tasks`
  - `GET /api/tasks`
  - `PUT /api/memory/:task_id`
  - `GET /api/memory/:task_id`
  - `GET /api/runs`

### 2.2 Milestone status vs `IMPLEMENTATION_PLAN.md`

- Milestone 0 (Core chat + infra): **mostly complete**
  - Backend health, conversations/messages, provider config, SSE chat, migrations are implemented.
  - Frontend chat shell, streaming renderer, provider modal, TanStack Query + Zustand are implemented.
- Milestone 1 (Task capture/editing): **partially complete**
  - Backend task draft/meta-LLM + task CRUD/versioning is implemented.
  - Frontend wizard/editor components exist, but are not fully wired into the top-level app flow.
- Milestone 2 (Task execution/history): **partially complete**
  - Backend run execution and run history APIs are implemented.
  - Frontend run components/hooks exist, but are not fully integrated in `App.tsx`.
- Milestone 3 (Memory + task management): **partially complete**
  - Backend export/import + memory APIs are implemented.
  - Frontend memory/export/import/settings components exist, but are not fully integrated in the shell.

### 2.3 Load status vs `IMPLEMENTATION_PARTS.md`

- Piece 1 (Backend): **substantially complete** for v1 core routes.
- Piece 2 (Frontend): **chat/foundation complete**, tasks/runs/memory features are present but integration is incomplete.
- Load 3 (Integration harness): **partially complete** (scripts exist, SSE parser exists, provider handshake exists).
- Load 4 (QA + acceptance validation): **not complete** (no deterministic smoke script/fixture set in repo yet).
- Loads 5-12 (quiz, agents, command runtime, MCP, polish): **not started as integrated product features**.

## 3) Confirmed implementation evidence (code)

### Backend

- Server boot, middleware, CORS, graceful shutdown:
  - [backend/cmd/api/main.go](backend/cmd/api/main.go)
- Route registration (including `draft-task`, delete routes, runs, memory):
  - [backend/internal/handlers/server.go](backend/internal/handlers/server.go)
- Chat streaming SSE (`chunk` + `done`):
  - [backend/internal/handlers/conversations.go](backend/internal/handlers/conversations.go)
- Schema + migrations (`schema_version`, `input_schema_json`, run status/tokens/cost, memory PK upsert semantics):
  - [backend/internal/db/migrations.go](backend/internal/db/migrations.go)
- Task draft extraction, CRUD, versions, export/import:
  - [backend/internal/services/task_service.go](backend/internal/services/task_service.go)
- Task runs + persisted execution metadata:
  - [backend/internal/services/run_service.go](backend/internal/services/run_service.go)
- Memory read/write/upsert:
  - [backend/internal/services/memory_service.go](backend/internal/services/memory_service.go)

### Frontend

- App foundation and chat flow:
  - [frontend/src/App.tsx](frontend/src/App.tsx)
  - [frontend/src/lib/api.ts](frontend/src/lib/api.ts)
  - [frontend/src/lib/store.ts](frontend/src/lib/store.ts)
  - [frontend/src/components/chat/ConversationList.tsx](frontend/src/components/chat/ConversationList.tsx)
  - [frontend/src/components/chat/MessageList.tsx](frontend/src/components/chat/MessageList.tsx)
  - [frontend/src/components/chat/MessageInput.tsx](frontend/src/components/chat/MessageInput.tsx)
  - [frontend/src/hooks/useStreaming.ts](frontend/src/hooks/useStreaming.ts)
- Task/run/memory/settings components exist and are buildable:
  - [frontend/src/components/tasks/TaskWizard.tsx](frontend/src/components/tasks/TaskWizard.tsx)
  - [frontend/src/components/tasks/TaskEditor.tsx](frontend/src/components/tasks/TaskEditor.tsx)
  - [frontend/src/components/tasks/RunPanel.tsx](frontend/src/components/tasks/RunPanel.tsx)
  - [frontend/src/components/tasks/ExportImport.tsx](frontend/src/components/tasks/ExportImport.tsx)
  - [frontend/src/components/memory/MemoryPanel.tsx](frontend/src/components/memory/MemoryPanel.tsx)
  - [frontend/src/components/settings/ProviderSettings.tsx](frontend/src/components/settings/ProviderSettings.tsx)
  - [frontend/src/components/runs/RunDetail.tsx](frontend/src/components/runs/RunDetail.tsx)

## 4) Critical drift to fix first (doc + runtime contract)

These are blockers for predictable collaboration and should be resolved before major new features:

1. **Port contract drift (`8080` vs `38473`)**
   - Tauri default is `38473` and passes `OPENADE_PORT` to backend.
   - Web scripts and docs still default to `8080`.
   - A single explicit strategy is required (recommended in Phase A below).

2. **Schema/API doc drift vs implemented backend**
   - Docs in `IMPLEMENTATION_PLAN.md` still show older fields.
   - Current backend schema includes richer fields (`title`, `input_schema_json`, run status/error/tokens/duration/prompt_final).

3. **Duplicate sections in `IMPLEMENTATION_PLAN.md`**
   - The file currently includes duplicated plan content, which increases ambiguity during implementation.

## 5) Detailed implementation plan for the rest of the project

## Phase A: Contract and doc reconciliation (required first)

Reference docs:
- [ARCHITECTURE.md](ARCHITECTURE.md)
- [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md)
- [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md)
- [src-tauri/src/lib.rs](src-tauri/src/lib.rs)

Deliverables:

1. Decide and document port strategy:
   - Option A (recommended): backend default `38473` for Tauri parity, with explicit web override examples.
   - Option B: keep web default `8080` but document dual-profile clearly for web and Tauri.
2. Update all port-related docs and scripts to match the chosen strategy.
3. Align `IMPLEMENTATION_PLAN.md` schema and route tables with current backend implementation.
4. Remove duplicate sections in `IMPLEMENTATION_PLAN.md`.
5. Add a short "Contract Changelog" section so future drift is obvious.

Definition of done:

- `README.md`, `.env.example`, `frontend/.env.example`, `scripts/*.sh`, `ARCHITECTURE.md`, `IMPLEMENTATION_PLAN.md`, `IMPLEMENTATION_PARTS.md` agree on environment defaults and API contract language.

## Phase B: Complete Milestones 1-3 integration in frontend shell

Reference docs:
- Milestones in [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md)
- Frontend deliverables in [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md)
- Module boundaries in [ARCHITECTURE.md](ARCHITECTURE.md)

Deliverables:

1. Integrate task and run panels into `App.tsx` panel navigation.
2. Add "Save as Task" action in chat flow to open `TaskWizard` with active conversation id.
3. Wire `TaskLibrary` selection to `TaskEditor`, `RunPanel`, and `MemoryPanel`.
4. Add runs panel with `useRuns` and `RunDetail` integration.
5. Integrate provider settings into a settings panel route/view.
6. Integrate export/import panel for task lifecycle completion.
7. Ensure all query invalidations are correct after create/update/delete/run/import.
8. Add UX handling for 401/no-provider across all task/run entrypoints (not just chat).

Definition of done:

- End-to-end flow in browser works: chat -> save as task -> edit -> run -> see output/cost -> inspect run history -> edit memory -> export/import.

## Phase C: Load 3 and 4 hardening (integration harness + QA)

Reference docs:
- Load 3/4 in [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md)
- Reliability/error model in [ARCHITECTURE.md](ARCHITECTURE.md)

Deliverables:

1. Add deterministic smoke script in `scripts/smoke.sh`:
   - health, create conversation, task create, run (mock or real provider), export/import, memory write/read.
2. Add fixtures under `scripts/fixtures/`.
3. Add explicit readiness details to health response (including DB open state).
4. Add structured log markers around stream start/done, run start/done/fail.
5. Add minimal integration checks for non-LLM routes (Go tests around handlers/services).

Definition of done:

- One command validates baseline acceptance criteria and outputs pass/fail clearly.

## Phase D: Playground expansion (Loads 5-10)

Reference docs:
- Loads 5-10 in [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md)
- Interaction and safety in [ARCHITECTURE.md](ARCHITECTURE.md)

Execution order:

1. **Load 5**: UI polish + quiz teaching UI scaffolding.
2. **Load 6**: terminal-safe slash commands and command execution safety controls.
3. **Load 7**: quiz backend persistence and scoring.
4. **Load 8**: agent catalog and curated game agents.
5. **Load 9**: command palette, shortcuts, accessibility pass.
6. **Load 10**: Tauri packaging/release hardening.

Hard gate before Load 6:

- Security review of command allowlist and explicit confirmation semantics.

## Phase E: MCP platform capability (Loads 11-12)

Reference docs:
- Loads 11-12 in [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md)
- MCP references in [mcp-docs/](mcp-docs/)
- Integration points in [ARCHITECTURE.md](ARCHITECTURE.md)

Deliverables:

1. MCP server configuration UI + DB persistence.
2. Backend MCP client and health/test endpoints.
3. Tool/resource/prompt discovery endpoints.
4. MCP tool/resource invocation UI and chat/task integration.
5. Result provenance and basic usage tracing in run records.

Definition of done:

- Users can configure MCP servers in-app and successfully call tools/resources from UI workflows.

## 6) Immediate next sprint (recommended)

1. Execute Phase A fully (contract/doc reconciliation).
2. Execute Phase B fully (wire existing frontend components into a complete Milestone 1-3 UX).
3. Add Phase C smoke harness and fixture set.

This gives a stable, documented baseline before adding advanced playground features (quiz, agents, command runtime, MCP).

## 7) Working rule for this repo (playground discipline)

For each feature PR, require:

1. A "Docs referenced" section citing exact sections/files used.
2. A "Contract impact" section listing any API/schema/env changes.
3. A "Validation run" section listing commands executed and outcomes.

This keeps the project as a reliable playground for LLM-driven delivery while preserving contract clarity.
