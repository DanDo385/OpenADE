# OpenADE

A local-first "chat-to-task" tool for repeatable LLM workflows. Chat with an LLM to explore a use case, then save the conversation as a reusable task with templated inputs.

**Tech stack:** Go backend + TypeScript (React + Vite) frontend. Tauri is optional for later desktop packaging.

- [ARCHITECTURE.md](ARCHITECTURE.md) – design, modules, key decisions
- [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) – milestones, API routes, demo walkthrough
- [IMPLEMENTATION_PARTS.md](IMPLEMENTATION_PARTS.md) – split work assignments
- [PROJECT_STATUS_AND_REST_IMPLEMENTATION_PLAN.md](PROJECT_STATUS_AND_REST_IMPLEMENTATION_PLAN.md) – live status + detailed remaining plan

## Local Development (Web-First)

```bash
# 1) Install deps
pnpm install
cd frontend && pnpm install && cd ..

# 2) Copy env defaults (optional)
cp .env.example .env
cp frontend/.env.example frontend/.env

# 3) Start backend + frontend together
pnpm run dev
```

Useful commands:

```bash
pnpm run dev:backend   # backend only
pnpm run dev:frontend  # frontend only
pnpm run health        # checks backend /health
```

## Optional Tauri Packaging Later

When web app behavior is stable, you can package with Tauri:

```bash
pnpm run tauri:dev
pnpm run tauri:build
```

## Commit and Push

Run the script with an optional commit message:

```bash
./commit-and-push.sh "Your commit message"
```

If no message is given, it defaults to `Update`.
