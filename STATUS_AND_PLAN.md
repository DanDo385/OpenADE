# OpenADE: Status & Implementation Plan

> **Purpose.** OpenADE is a playground for exploring what LLMs can do: chat-to-task workflows, repeatable prompts, cost tracking, and extensible tooling. This document summarizes current status and outlines the detailed plan for the rest of the project. All work should reference the canonical docs: [ARCHITECTURE.md](ARCHITECTURE.md), [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md), [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md), and [mcp-docs/](mcp-docs/).

---

## Doc Reference Quick Links

| Doc | Purpose |
|-----|---------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | Design principles, runtime topology, data model, error handling, MCP/agent concepts |
| [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) | Milestones 0–3, API contracts, data model, demo walkthrough |
| [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md) | Piece 1–2, Loads 3–13, work splits, file layouts |
| [mcp-docs/MCP_Complete_Reference.md](mcp-docs/MCP_Complete_Reference.md) | MCP protocol, tools, resources, prompts |
| [mcp-docs/MCP_GitHub_DeepDive.md](mcp-docs/MCP_GitHub_DeepDive.md) | MCP SDKs, registry, community servers |

---

## Current Status Summary

### Milestones vs Reality (per [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md))

| Milestone | Backend | Frontend | Status |
|-----------|---------|----------|--------|
| **M0** – Core chat & infra | ✅ | ✅ | Complete |
| **M1** – Task capture and editing | ✅ | ✅ | Complete |
| **M2** – Task execution and history | ✅ | ✅ | Complete |
| **M3** – Memory and task management | ✅ | ✅ | Complete |

### Piece 1: Backend (per [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md))

| Area | Status | Notes |
|------|--------|------|
| Project bootstrap | ✅ | `backend/cmd/api/main.go`, env support |
| Database layer | ✅ | SQLite, migrations, all tables |
| API handlers | ✅ | Health, conversations, tasks, runs, providers, memory |
| Service layer | ✅ | Conversation, task, run, memory, provider |
| LLM layer | ✅ | OpenAI adapter, streaming, token/cost extraction |
| Cross-cutting | ✅ | Error envelopes, CORS, logging |

**API surface:** All routes from [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) implemented.

### Piece 2: Frontend (per [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md))

| Area | Status | Notes |
|------|--------|------|
| Core app shell | ✅ | Topbar, panels (chat/tasks/runs/settings) |
| Chat feature set | ✅ | Conversation list, message streaming, "Save as Task" |
| Task feature set | ✅ | Task wizard (4 steps), task library, task editor, run panel |
| Run and memory | ✅ | Run form, output + cost, run history + detail, memory panel |
| Data and integration | ✅ | TanStack Query, Zustand, API client, SSE parser |
| Export/import UX | ✅ | Export task, import bundle |

**Components:** `ConversationList`, `MessageList`, `MessageInput`, `ProviderModal`, `TaskLibrary`, `TaskWizard`, `TaskEditor`, `RunPanel`, `RunDetail`, `MemoryPanel`, `ExportImport`, `ProviderSettings`, `ErrorDisplay`.

### Loads 3–4: Integration & QA (per [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md))

| Area | Status | Notes |
|------|--------|------|
| Runtime orchestration | ✅ | `pnpm run dev`, scripts |
| API client + SSE | ✅ | Base URL, error handling, chunk/done parsing |
| Error envelope | ✅ | `ErrorDisplay`, API error mapping |
| Provider handshake | ✅ | 401 → provider modal |
| Smoke script | ✅ | `scripts/smoke.sh` – conv → task → run → export → import |
| Fixtures | ✅ | `scripts/fixtures/movie-picker-demo.json` |

### Loads 5–13: Not Started

| Load | Focus | Status |
|------|-------|--------|
| 5 | UI Delight + Quiz Teaching (Shadcn, quiz UI, slash commands) | ❌ |
| 6 | Terminal-Safe Slash Commands + Game Agents | ❌ |
| 7 | Quiz Backend + Session Persistence | ❌ |
| 8 | Game Catalog + Agent Library | ❌ |
| 9 | Polish, Accessibility, Command Palette | ❌ |
| 10 | Desktop Packaging (Tauri) | ❌ |
| 11 | MCP Server Configuration + Settings UI | ❌ |
| 12 | MCP Tool Discovery + Task/Chat Integration | ❌ |
| 13 | MCP Registry + Full UI Surface | ❌ |

---

## What Works Today (LLM Playground Baseline)

OpenADE currently supports the full **movie-picker demo** from [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md):

1. **Chat** – Create conversation, stream LLM responses, markdown rendering
2. **Save as Task** – Wizard with optional meta-LLM draft, `{{variable}}` parsing
3. **Task Library** – List, search, select, edit, delete
4. **Run** – Variable form, run task, view output + cost
5. **Run History** – List runs, click to see full output
6. **Memory** – Per-task key/value store
7. **Export/Import** – Task bundle round-trip

**Run:** `pnpm run dev` → `http://localhost:5173`  
**Smoke:** `bash scripts/smoke.sh` (optionally `OPENAI_API_KEY` for full flow)

---

## Detailed Implementation Plan for the Rest of the Project

The following plan is structured by Load and references [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md) and [ARCHITECTURE.md](ARCHITECTURE.md). Treat this as the master checklist for extending the LLM playground.

---

### Phase A: Polish & Interaction Surface (Loads 5, 9)

**Goal:** Sharpen UX and add discoverability so the playground feels intentional. See [ARCHITECTURE.md](ARCHITECTURE.md) – Interaction Surface.

#### Load 5: UI Delight + Quiz Teaching

| # | Deliverable | Files / Reference | Notes |
|---|-------------|-------------------|-------|
| 5.1 | Shadcn UI primitives | `frontend/src/components/ui/` | Button, Input, Dialog, Tabs, Card, Textarea, Select, Toast. Per [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md) Load 5. |
| 5.2 | Design tokens | `frontend/src/index.css` | CSS variables for spacing, colors. |
| 5.3 | Empty/loading states | Existing components | Already present; refine per Load 5. |
| 5.4 | Quiz session model (UI) | `QuizSessionPanel.tsx`, `QuizRunner.tsx` | Question, choices, answer, explanation, score. [ARCHITECTURE.md](ARCHITECTURE.md) – Interactive teaching sessions. |
| 5.5 | Slash command UI | MessageInput, suggestions | `/help`, `/quiz`, `/clear`, `/run`, `/export`, `/import`. [ARCHITECTURE.md](ARCHITECTURE.md) – Slash command parser. |

#### Load 9: Polish, Accessibility, Command Palette

| # | Deliverable | Files / Reference | Notes |
|---|-------------|-------------------|-------|
| 9.1 | Command palette | `CommandPalette.tsx` | `Cmd+K` / `Ctrl+K`, search conversations/tasks/agents. |
| 9.2 | Keyboard shortcuts | `useKeyboardShortcuts.ts` | `Cmd+N` new conv, `Cmd+/` focus input, `Escape` close. |
| 9.3 | Accessibility | Global | ARIA, focus trap, skip-to-content, `prefers-reduced-motion`. |
| 9.4 | Settings panel | Extend `ProviderSettings` | Shortcuts list, theme, reduced-motion toggle. |

**Dependencies:** Load 5 before Load 9 (per [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md)).

---

### Phase B: Commands & Agents (Loads 6, 7, 8)

**Goal:** Add command-style workflows and local agents. See [ARCHITECTURE.md](ARCHITECTURE.md) – Command Execution and Safety, Agent and Game Orchestration Concept.

#### Load 6: Terminal-Safe Slash Commands + Game Agents

| # | Deliverable | Files / Reference | Notes |
|---|-------------|-------------------|-------|
| 6.1 | `POST /api/commands/execute` | `handlers/commands.go`, `command_service.go` | Allowlist: `terminal`, `scaffold`, `open`, `play`. Payload: `{ input, confirm }`. Response: `{ ok, output, exit_code, duration_ms }`. [ARCHITECTURE.md](ARCHITECTURE.md) – Frontend and Data Contract Extensions. |
| 6.2 | `POST /api/agents`, `POST /api/agents/:id/run` | `handlers/agents.go`, `agent_service.go` | Agent metadata, script bundle. [ARCHITECTURE.md](ARCHITECTURE.md) – Agent and Game Orchestration Concept. |
| 6.3 | Frontend command parser | MessageInput, `commandBus.ts` | Parse `/help`, `/run`, `/quiz`; route to backend. |
| 6.4 | Command output panel | `CommandOutputPanel.tsx` | Inline logs, exit code, duration. |

#### Load 7: Quiz Backend + Session Persistence

| # | Deliverable | Files / Reference | Notes |
|---|-------------|-------------------|-------|
| 7.1 | Quiz schema | `db/migrations` | `quiz_sessions`, `quiz_questions`, `quiz_attempts`, `quiz_answers`. [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md) Load 7. |
| 7.2 | Quiz API | `handlers/quiz.go`, `quiz_service.go` | List/create sessions, start attempt, submit answers, get score. |
| 7.3 | Meta-LLM quiz generator | Optional | Generate quiz from topic or conversation. |
| 7.4 | Fixtures | `fixtures/quizzes/` | onboarding.json, cli-basics.json. |

**Dependencies:** Load 5 (quiz UI); Piece 1 (backend). [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md).

#### Load 8: Game Catalog + Agent Library

| # | Deliverable | Files / Reference | Notes |
|---|-------------|-------------------|-------|
| 8.1 | Agent catalog schema | `db/migrations` | `agents` table. [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md) Load 8. |
| 8.2 | `GET /api/agents`, `GET /api/agents/:id` | `handlers/agents.go` | List, detail. |
| 8.3 | Game fixtures | `fixtures/agents/` | blackjack.json, trivia.json. |
| 8.4 | AgentLibrary, AgentLauncher | Frontend | List, filter, launch, output pane. |
| 8.5 | Slash integration | `/agent:blackjack`, etc. | Resolve to catalog entries. |

**Dependencies:** Load 6, Piece 2. [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md).

---

### Phase C: MCP Integration (Loads 11, 12, 13)

**Goal:** Expose MCP servers, tools, resources, and prompts in the UI. See [ARCHITECTURE.md](ARCHITECTURE.md) – MCP; [mcp-docs/](mcp-docs/).

#### Load 11: MCP Server Configuration + Settings UI

| # | Deliverable | Files / Reference | Notes |
|---|-------------|-------------------|-------|
| 11.1 | `mcp_servers` table | `db/migrations` | id, name, transport, command_or_url, args_json, env_json, enabled. [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md) Load 11. |
| 11.2 | Go MCP client | `internal/mcp/` | Stdio + SSE transport. [mcp-docs/MCP_Complete_Reference.md](mcp-docs/MCP_Complete_Reference.md). |
| 11.3 | MCP API | `GET/POST/PUT/DELETE /api/mcp/servers`, `POST /api/mcp/servers/:id/test` | CRUD, test connection. |
| 11.4 | MCPServersPanel | `components/settings/MCPServersPanel.tsx` | List, add, edit, enable/disable, test. |

#### Load 12: MCP Tool Discovery + Task/Chat Integration

| # | Deliverable | Files / Reference | Notes |
|---|-------------|-------------------|-------|
| 12.1 | `GET /api/mcp/servers/:id/tools`, `resources`, `prompts` | Backend | List tools, resources, prompts. [mcp-docs/MCP_Complete_Reference.md](mcp-docs/MCP_Complete_Reference.md) – Core Primitives. |
| 12.2 | `POST /api/mcp/tools/call` | Backend | Invoke tool. |
| 12.3 | MCPToolsPanel, MCPResourcesPanel, MCPPromptsPanel | Frontend | Browse, invoke. |
| 12.4 | Tool call display | MessageList, RunPanel | Show tool calls in chat and run output. |

#### Load 13: MCP Registry + Full UI Surface

| # | Deliverable | Files / Reference | Notes |
|---|-------------|-------------------|-------|
| 13.1 | MCP Registry integration | Frontend or backend | Search, install from registry. [mcp-docs/MCP_Complete_Reference.md](mcp-docs/MCP_Complete_Reference.md) – MCP Registry. |
| 13.2 | MCPRegistryPanel | Frontend | Search, install. |
| 13.3 | Full MCP navigation | Settings / sidebar | Servers, tools, resources, prompts, registry. |
| 13.4 | Smoke extension | `scripts/smoke.sh` | Add MCP server, list tools. |

**Integration contract (per [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md)):** All MCP primitives must be configurable and discoverable from the UI; no config-file editing.

---

### Phase D: Distribution (Load 10)

**Goal:** Ship as a desktop app. See [ARCHITECTURE.md](ARCHITECTURE.md) – Deployment Model; [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) – Future (Production).

#### Load 10: Desktop Packaging and Release

| # | Deliverable | Files / Reference | Notes |
|---|-------------|-------------------|-------|
| 10.1 | Tauri setup | `src-tauri/` | Cargo.toml, tauri.conf.json, backend subprocess. [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md) Load 10. |
| 10.2 | Scripts | `pnpm run tauri:dev`, `tauri:build` | Per [README.md](README.md). |
| 10.3 | Packaging | App icon, bundle ID | Code signing placeholder. |
| 10.4 | Docs | `docs/RELEASE.md`, `TAURI_SETUP.md` | Build and packaging instructions. |

**Dependencies:** Loads 1–4 stable; Load 9 recommended.

---

## Recommended Execution Order

1. **Load 5** – UI polish, quiz surface, slash affordances (enables Loads 7, 9).
2. **Load 9** – Command palette, shortcuts, accessibility (quick UX win).
3. **Load 6** – Commands + agents backend (enables Load 8).
4. **Load 7** – Quiz backend (completes quiz story).
5. **Load 8** – Game catalog (completes agent story).
6. **Load 11** – MCP config (enables 12, 13).
7. **Load 12** – MCP tools in tasks/chat.
8. **Load 13** – MCP registry + full UI.
9. **Load 10** – Tauri packaging.

---

## LLM Playground Use Cases (Reference)

OpenADE is intended as a playground for:

- **Repeatable prompts** – Chat → save as task → run with different inputs.
- **Cost awareness** – Token and cost tracking per run.
- **Meta-LLM workflows** – Draft extraction, quiz generation.
- **Tool integration** – MCP servers as LLM tooling.
- **Teaching flows** – Quiz sessions, agents, slash commands.

When implementing, prioritize flows that demonstrate these use cases and keep the demo walkthrough in [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) working.

---

## Changelog

- **2026-03-09** – Initial status + plan. M0–M3 complete. Loads 5–13 and Load 10 planned.
