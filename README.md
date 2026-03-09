# OpenADE

OpenADE is a capability studio and MCP host for experimenting with LLMs. It is a design-time environment for building, testing, packaging, and evolving AI capabilities through chat, tools, skills, workflows, and interactive UI.

## Vision

OpenADE is not just a task app. The goal is to give models a flexible environment where they can explore ideas, call tools, generate interfaces, package working patterns into reusable capabilities, and eventually export those capabilities into other runtimes.

## Core Concepts

- `Objective` — A lightweight design brief for a conversation or build session.
- `Task` — A reusable unit of work with inputs, outputs, and execution history.
- `Skill` — A packaged capability made of prompts, tools, templates, and behavior.
- `Workflow` — A composition of tasks and skills into a larger process.
- `Run` — One execution of a task, skill, or workflow.
- `Automation` — A scheduled or triggered execution that runs over time.
- `Tool` — A safe callable capability, built in or provided through MCP.
- `Command` — A shell-like action that requires explicit approval.

## Commands, Tools, Agents

- `Commands` are shell-like and higher risk; they should require explicit user approval.
- `Tools` are safe callables; they can be built in or exposed by MCP servers.
- `Agents` orchestrate tools, tasks, and workflows toward a larger goal.

## Architecture

- Backend: Go API with Chi and SQLite, default local API port `8080`
- Frontend: React + Vite with Zustand and TanStack Query, default dev port `5173`
- Desktop wrapper: Tauri shell around the local frontend and backend

## Run

```bash
pnpm install
cd frontend && pnpm install && cd ..
cp .env.example .env
cp frontend/.env.example frontend/.env
pnpm run dev
```

```bash
pnpm run dev:backend   # backend only
pnpm run dev:frontend  # frontend only
pnpm run health
```

## Roadmap

- `1A` Foundation: README/product framing, ontology, objective model, agents and commands visible in the shell
- `1B` Component protocol: safe model-driven UI generation for chat without arbitrary code execution
- `2` MCP host: MCP server config, discovery, invocation, and secrets-provider abstraction
- `3` Scheduling: automations, recurring runs, and skill suggestion flows
- `4` Polish/export: Shadcn migration, export formats, MCP apps, and runtime handoff

## Current Limits

- Scheduled jobs in v1 run only while the app and backend are open.

## Direction

OpenADE should feel like a place to chat with models, give them tools, inspect what they do, shape that behavior into reusable skills, and package the best results into portable capabilities.
