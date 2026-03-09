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

## Browser Start

- Start with the browser version first. It is the fastest way to explore the product.
- `pnpm run dev` starts:
  - the Go backend on `http://localhost:8080`
  - the Vite frontend on `http://localhost:5173`
- Open `http://localhost:5173` in your browser.
- The default local env values live in [.env.example](/Users/danmagro/Desktop/Code/open-ade/.env.example) and [frontend/.env.example](/Users/danmagro/Desktop/Code/open-ade/frontend/.env.example).
- For a concrete onboarding path, see [docs/FIRST_30_MINUTES.md](/Users/danmagro/Desktop/Code/open-ade/docs/FIRST_30_MINUTES.md).

## First Use

1. Open the app in the browser.
2. Go to `Settings`, or click the `Provider` button in the top bar.
3. Enter an API key, optional base URL, and a default model.
4. Save the provider config.
5. Go back to `Chat` and start a conversation.

Without a provider, chat and task runs will not work.

## Chat Flow

1. Create or select a conversation in the left sidebar.
2. Open the `Objective` panel and write:
   - title
   - goal
   - constraints
   - tools required
   - success criteria
3. Chat with the model normally.
4. Use chat to explore the capability until the prompt pattern is stable.
5. Click `Save as Task` when the conversation becomes reusable.

The `Objective` is the design brief for the conversation. The chat is where you shape the behavior.

## Task Flow

1. Use `Save as Task` from chat, or create one directly from the `Tasks` panel.
2. In the task wizard:
   - confirm the task name
   - refine the prompt template
   - review generated input fields
3. Save the task.
4. Open the task in the `Tasks` panel.
5. Run it with real inputs.
6. Edit the task until the output is reliable.

This is the current core loop for turning conversation experiments into reusable capabilities.

## Workflow Flow

There is not yet a full workflow canvas or DAG builder. Right now, the practical workflow loop is:

1. Use `Chat` to explore an idea.
2. Capture intent in `Objective`.
3. Save the result as a `Task`.
4. Use `Run` to test it repeatedly.
5. Add a `Schedule` if it should recur.
6. Add MCP servers in `Settings` when the task needs external tools.
7. Use `Commands` for support actions like:
   - `/summarize <conversation_id>`
   - `/list-runs`
   - `/inspect-run <run_id>`
   - `/objective <conversation_id>`

For now, a workflow is a composition of chat, tasks, runs, schedules, MCP configuration, and commands.

## MCP and Tools

- MCP servers are configured in `Settings`.
- Start with local `stdio` MCP servers first.
- Once configured, MCP becomes the tool layer for future task and workflow expansion.
- Commands are not the same thing as tools:
  - `Commands` are shell-like and explicit
  - `Tools` are safe callables, built in or MCP-backed
  - `Agents` orchestrate tools, tasks, and workflows

## Good First Project

Try a simple recurring capability first:

1. Chat goal: "Summarize incoming issues into a short daily digest."
2. Save that conversation as a task.
3. Add inputs like repo, date range, and output format.
4. Run the task manually until the prompt is good.
5. Add a schedule.
6. Later, add MCP-backed GitHub or browser tooling.

For a concrete example built around this repo, see [docs/FIRST_WORKFLOW.md](/Users/danmagro/Desktop/Code/open-ade/docs/FIRST_WORKFLOW.md).

## Roadmap

- `1A` Foundation: README/product framing, ontology, objective model, agents and commands visible in the shell
- `1B` Component protocol: safe model-driven UI generation for chat without arbitrary code execution
- `2` MCP host: MCP server config, discovery, invocation, and secrets-provider abstraction
- `3` Scheduling: automations, recurring runs, and skill suggestion flows
- `4` Polish/export: Shadcn migration, export formats, MCP apps, and runtime handoff

## Current Limits

- Scheduled jobs in v1 run only while the app and backend are open.
- The browser version is the best place to start; Tauri is not required for the initial loop.
- There is no full workflow builder yet.
- MCP server configuration exists, but the broader MCP UI and workflow integration are still evolving.

## Direction

OpenADE should feel like a place to chat with models, give them tools, inspect what they do, shape that behavior into reusable skills, and package the best results into portable capabilities.
