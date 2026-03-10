# OpenADE Architecture & Connections

## Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              BROWSER                                         │
│  http://localhost:5173/                                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  React + Vite + TanStack Query + Zustand                              │   │
│  │  - SPA: index.html → main.tsx → App.tsx                               │   │
│  │  - Fetches: /api/*, /health (same-origin in dev via Vite proxy)       │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────┬──────────────────────────────────┘
                                           │ HTTP (fetch)
                                           │ Same-origin in dev: requests to
                                           │ :5173/api/* → Vite proxies to :8080
                                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  VITE DEV SERVER (port 5173)                                                 │
│  - Serves static: index.html, /src/*                                         │
│  - Proxies /api, /health → http://127.0.0.1:8080                             │
└──────────────────────────────────────────┬──────────────────────────────────┘
                                           │ proxy
                                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  GO BACKEND (port 8080)                                                      │
│  - Chi router, CORS for localhost:5173                                       │
│  - REST: /health, /api/conversations, /api/tasks, /api/providers, etc.       │
│  - SSE: POST /api/conversations/:id/messages (streaming chat)                │
└──────────────────────────────────────────┬──────────────────────────────────┘
                                           │
                                           ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  SQLITE (backend/openade.db)                                                 │
│  - Conversations, tasks, runs, providers, schedules, agents, MCP config      │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Languages & Parts

| Part       | Language   | Port | Entry                    | Purpose                          |
|-----------|------------|------|--------------------------|----------------------------------|
| Backend   | Go         | 8080 | backend/cmd/api/main.go  | REST API, SSE streaming, SQLite  |
| Frontend  | TypeScript | 5173 | frontend/src/main.tsx    | React SPA, Vite dev server       |
| Database  | SQLite     | —    | backend/openade.db       | Persistence                      |

## Connection: Frontend ↔ Backend

**In development (separate terminals):**

1. **Backend** runs `go run ./cmd/api` → listens on `:8080`
2. **Frontend** runs `vite` → listens on `:5173`, proxies `/api` and `/health` to `http://127.0.0.1:8080`
3. Frontend `api.ts` uses **relative URLs** in dev (`''`) → `fetch('/api/providers')` hits `localhost:5173/api/providers` → Vite proxies to backend

**Flow:**
```
Browser fetch('/api/providers')
  → Request to http://localhost:5173/api/providers (same origin)
  → Vite proxy forwards to http://127.0.0.1:8080/api/providers
  → Go backend handles, returns JSON
```

**If proxy fails:** Frontend gets 502/504 from proxy. Backend must be running first.

**Streaming:** Chat uses `fetch` with `Accept: text/event-stream`; same proxy path, SSE streams through.

## Recommended Dev Workflow (Separate Terminals)

1. **Terminal 1:** `pnpm run dev:backend` — starts Go on :8080
2. **Terminal 2:** `pnpm run dev:frontend` — starts Vite on :5173
3. Open `http://localhost:5173/` in a **new** browser tab

No combined script, no port killing, no Node version switching. Each process runs in its own terminal.

## Why the Frontend Might Not Render

1. **Backend not running** → Proxy returns 502 → API fails
2. **Wrong URL** → Use `http://localhost:5173/` (include port)
3. **Vite 504 Outdated Request** → Close old tabs, open a fresh one after server is ready
4. **Node 25** → Use Node 20 or 22 (`nvm use 22`)

## npm vs pnpm

Both work. The project uses pnpm for lockfile consistency. Switching to npm:
- `package-lock.json` instead of `pnpm-lock.yaml`
- `npm run dev` instead of `pnpm run dev`
- No functional difference for this architecture

## HTTP vs WebSocket

Current: **HTTP + SSE** (Server-Sent Events for chat streaming). No WebSockets.

- REST: standard `fetch` for CRUD
- Chat: POST with `Accept: text/event-stream`, streamed response

WebSockets would require backend and frontend changes. HTTP/SSE is sufficient for this app and avoids extra complexity.
