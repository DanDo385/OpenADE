# Behavior Layer Implementation Plan

This plan plugs a stronger behavior/runtime layer into the current OpenADE loop:

`chat -> objective -> task -> run`

It does not replace the existing flow. It makes that flow more uniform, more inspectable, and more reusable.

## Why This Plan

Today OpenADE already has:

- chat
- objective
- task creation
- runs
- schedules
- MCP server config
- command execution

What it does not yet have is a consistent way to show:

- what guidance the model received
- what skills were attached
- what context was injected
- what tools were available
- how the result was graded
- how to improve the next run

This plan adds that layer without throwing away the current app.

## North Star

Every serious OpenADE session should be inspectable as one stack:

1. objective
2. soul
3. attached skills
4. injected context
5. tools available
6. prompt assembly
7. schema
8. output
9. reflection
10. grade

That same stack should be visible whether the user is looking at:

- a skill
- a task
- an agent
- a run

## Uniform Page Shape

Create one shared page pattern for capability objects. It can back Tasks, Agents, and Skills with the same layout.

Suggested sections:

1. `Overview`
2. `Instructions`
3. `Context`
4. `Tools`
5. `Schema`
6. `Prompt Assembly`
7. `Runs`
8. `Reflection`
9. `Files`

This is the key uniformity rule:

- every capability page shows the markdown instructions
- every capability page shows the context injection sources
- every capability page shows the tools available
- every capability page shows the schema/rubric/grade
- every capability page shows the exact assembled prompt used in runs

## Filesystem Layout

Use a markdown-first library on disk. Keep it human-readable. Keep JSON only where structure helps.

Example:

```text
library/
  souls/
    default/
      SOUL.md
      metadata.json
      rubric.md
  skills/
    feature-brief/
      SKILL.md
      metadata.json
      schema.json
      rubric.md
      prompts/
        system.md
        planner.md
      context/
        sources.json
      examples/
        input.md
        output.md
      notes/
        failures.md
```

Rules:

- one soul per folder
- one skill per folder
- markdown is the primary authoring format
- folders can grow subdirectories as skills become richer
- the app should render folder content as a collection, not as one flat blob

## Runtime Concepts

### Objective

The mission brief for a conversation or build session.

### Soul

Global behavior and standards:

- tone
- initiative
- rigor
- error correction behavior
- taste
- preferred working style

### Skill

A focused capability pack:

- markdown instructions
- examples
- tool hints
- context sources
- schema
- rubric

### Task

A reusable execution wrapper built from chat. In the new model, a task is not just a prompt template. It is a prompt plus attached behavior and context.

### Agent

A runtime that can orchestrate multiple skills and tools.

### Run

One execution record with prompt assembly, context, output, reflection, and grade.

## Prompt Assembly Order

Use one consistent assembly order across chat, task runs, and agent runs:

1. objective
2. soul
3. selected skills
4. task or agent base prompt
5. context injections
6. user inputs
7. memory

The assembled result should be visible in the UI under `Prompt Assembly`.

This is a product requirement, not a backend detail. If the user cannot see what the model was given, the system becomes too opaque to tune.

## How This Plugs Into The Current Flow

### Chat

Add:

- active soul selector
- attached skill chips
- tool visibility
- prompt assembly preview

The conversation runtime should use objective + soul + skills in the system/instruction layer.

### Objective

Keep the current objective panel. Expand it later only as needed. It remains the mission brief.

### Task

When a conversation is saved as a task, also persist:

- selected soul id
- attached skill ids
- context source references
- schema reference
- rubric reference

This makes the task a reusable behavior bundle, not only a prompt.

### Run

Every run should capture:

- assembled prompt
- soul used
- skills used
- context injected
- tools available
- output
- reflection
- grade

## UI Implementation Plan

### Phase 1: Uniform Capability Frame

Goal: make Tasks and Agents readable the same way before adding more behavior.

Frontend:

- build a shared `CapabilityPage` layout
- use it first for Task detail
- use it second for Agent detail
- add tabs or sections for:
  - Instructions
  - Context
  - Tools
  - Prompt Assembly
  - Runs

Definition of done:

- Task pages and Agent pages share the same information structure
- the user can inspect guidance and structure instead of only editing raw fields

### Phase 2: Soul Loader

Goal: add one active soul to the current flow.

Backend:

- scan `library/souls/*`
- index soul metadata
- load `SOUL.md`
- expose list/get APIs

Frontend:

- add a soul selector in Chat
- show soul on Task and Run pages

Definition of done:

- one soul can be selected in chat
- it influences chat and task generation
- it is visible on the task and run pages

### Phase 3: Skill Loader

Goal: attach one or more skills to conversations and tasks.

Backend:

- scan `library/skills/*`
- index metadata, schema, rubric, and context source declarations
- add conversation-skill and task-skill linkage

Frontend:

- skill library page
- skill chips in Chat
- skill list on Task pages

Definition of done:

- the user can attach a skill to a conversation
- attached skills flow into saved tasks
- each skill has its own page showing markdown, context, schema, rubric, and examples

### Phase 4: Prompt Assembly Viewer

Goal: show exactly what the model received.

Backend:

- centralize assembly in one service
- return assembly breakdown with runs

Frontend:

- add `Prompt Assembly` panel on task runs and agent runs
- render:
  - objective
  - soul
  - skills
  - context injections
  - final prompt

Definition of done:

- every run can be inspected as assembled parts, not just final output

### Phase 5: Reflection And Grade

Goal: turn runs into learning loops.

Backend:

- add post-run reflection generation
- add rubric-based grading
- store result summaries

Frontend:

- show `Reflection` panel on Run detail
- show `Grade` badge and rubric view

Definition of done:

- each run includes a short reflection and a grade
- the user can see how to improve the next run

## Starter Toolkit Phase

This phase is mandatory before expecting users to build durable tools.

### Required Starter Toolkit

1. `Shell / terminal`
2. `Filesystem`
3. `Git`
4. `Fetch / web`
5. `Browser automation`
6. `Secrets`

### Terminal Access

Do not rely on the current minimal command panel as the final answer.

Implement:

- session-based PTY backend
- frontend terminal surface
- safe approval model
- command history and output capture

This is the missing bridge between prompts and real system action.

### MCP Starter Pack

Ship with a curated default set, not an empty settings screen.

Suggested baseline:

- filesystem
- git
- fetch/web
- browser automation
- memory
- sequential thinking

### Secrets

Keep the current provider abstraction and add a real first-party secrets path:

- env provider for local dev
- 1Password CLI/provider next

### External Integrations To Stage In

These should be treated as toolkit modules, not as one-off hacks:

1. `1Password CLI`
2. `AgentMail`
3. `Browser automation` for agent browsing
4. `Vercel deployment / browser-agent style workflows`

The rule is simple:

- first build the generic toolkit slots
- then plug these integrations into those slots

## Backend Work

### New Services

- `library_service`
- `soul_service`
- `skill_service`
- `prompt_assembly_service`
- `reflection_service`
- `grade_service`
- `terminal_service`

### Data To Add

Prefer DB indexing plus on-disk source files.

Add:

- souls
- skills
- conversation_skills
- task_skills
- agent_skills
- run_reflections
- run_grades
- context_sources

Keep the markdown, examples, and support files on disk in `library/`.

## Frontend Work

### New Screens / Panels

- Soul selector in Chat
- Skill library
- Skill detail page
- Uniform capability page
- Prompt assembly panel
- Reflection panel
- Grade display
- Terminal panel

### Existing Screens To Extend

- Chat
- Objective panel
- Task editor
- Run detail
- Agents

## Delivery Order

1. uniform capability page
2. soul loader
3. skill loader
4. prompt assembly viewer
5. reflection and grade
6. terminal access
7. curated MCP starter pack
8. 1Password / AgentMail / browser automation modules

## First Milestone

Keep the first milestone small and real:

1. one global soul
2. skill library loader
3. attach skills to chat
4. save attached skills into tasks
5. show prompt assembly on run detail

If that works, OpenADE becomes meaningfully more agentic without changing its basic shape.

## Success Criteria

This plan is working when:

- the user can see exactly what guidance was given to the model
- the user can attach souls and skills without leaving the current loop
- tasks become reusable behavior bundles instead of only prompt templates
- every run can be inspected, reflected on, and graded
- the app ships with enough baseline tools that reusable capability building is actually practical
