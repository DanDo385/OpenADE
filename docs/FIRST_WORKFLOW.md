# First Real Workflow In This Repo

This is a concrete starter workflow you can build today with the current OpenADE feature set.

## Workflow Name

`OpenADE Weekly Feature Review`

## What It Does

Once a week, you use OpenADE to turn rough feature thoughts into a structured implementation brief for the OpenADE project itself.

This workflow is intentionally simple because the current product does not yet have:

- a full workflow graph builder
- deep tool chaining inside task execution
- background scheduling when the app is closed

## Why This Is A Good First Workflow

It uses features that already exist:

1. `Chat` to explore the idea
2. `Objective` to define intent and success
3. `Task` to package the capability
4. `Run` to execute the packaged prompt
5. `Schedule` to repeat it weekly

## Step 1: Create The Objective

In `Chat`, create a conversation for feature review work.

Set the `Objective` to something like:

- `Title`: `Weekly OpenADE Feature Review`
- `Goal`: convert rough feature ideas into buildable implementation briefs
- `Constraints`: current architecture only, avoid fake capabilities, note product gaps
- `Tools required`: `chat, objective, task, schedule`
- `Success criteria`: each output has summary, implementation work, risks, and next steps

## Step 2: Shape The Capability In Chat

Use a prompt like:

```text
I want a reusable capability that turns rough OpenADE feature ideas into implementation briefs.

The output should include:
- summary
- user value
- backend work
- frontend work
- architecture concerns
- risks
- recommended next step

Keep the output practical and aligned with the current repo.
```

Keep iterating until the model produces something you would actually act on.

## Step 3: Save It As A Task

Use `Save as Task`.

Recommended task shape:

- `Name`: `OpenADE Feature Brief Builder`
- `Description`: `Turn rough OpenADE feature ideas into implementation briefs`
- `Prompt template`: use variables such as:
  - `{{feature_idea}}`
  - `{{user_value}}`
  - `{{constraints}}`

Example template:

```text
You are helping design features for OpenADE.

Feature idea:
{{feature_idea}}

User value:
{{user_value}}

Constraints:
{{constraints}}

Write a concise implementation brief with:
- Summary
- User value
- Backend work
- Frontend work
- Risks
- Recommended next step
```

## Step 4: Run It

In the `Run` panel, try:

- `feature_idea`: `Add an MCP tool browser and invoke UI in Settings`
- `user_value`: `Users can discover and test tools without leaving the app`
- `constraints`: `Use the current Go backend and React UI, avoid adding a workflow engine yet`

Inspect the output.

If it is weak:

1. edit the task
2. tighten the prompt
3. run again

## Step 5: Schedule It

In the `Schedule` panel, set a weekly cron like:

- `0 10 * * 1`

Optional timezone:

- `America/New_York`

This gives you a repeating planning task while the app is open.

## Step 6: Use MCP When Ready

After the basic workflow is stable:

1. open `Settings`
2. add an MCP server
3. test the server
4. inspect its tools
5. call a tool manually from the MCP UI

That does not yet make the scheduled task invoke MCP automatically, but it lets you build the tool layer that future workflows will depend on.

## What To Build Next

Once this workflow feels useful, the next logical upgrades are:

1. make task execution aware of MCP tools
2. add a workflow layer that chains tasks and tool calls
3. add a dedicated schedules/automations panel
4. support background automation beyond "app must be open"
