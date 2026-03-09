# First 30 Minutes In OpenADE

This walkthrough gets you from a fresh checkout to a usable chat, a saved task, and a scheduled automation in the browser UI.

## 0-5 Minutes: Start The App

1. Install dependencies:
   ```bash
   pnpm install
   cd frontend && pnpm install && cd ..
   ```
2. Create local env files:
   ```bash
   cp .env.example .env
   cp frontend/.env.example frontend/.env
   ```
3. Start the browser version:
   ```bash
   pnpm run dev
   ```
4. Open `http://localhost:5173`.
5. If you want to confirm the backend is alive:
   ```bash
   pnpm run health
   ```

## 5-10 Minutes: Configure A Provider

1. In the app, open `Settings`.
2. In `Provider settings`, enter:
   - your API key
   - optional base URL
   - a default model such as `gpt-4o-mini`
3. Click `Save provider`.

Without a provider, chat and task execution will not work.

## 10-15 Minutes: Start A Real Capability Chat

Use this exact starter prompt in `Chat`:

```text
I want to create a reusable OpenADE capability called "OpenADE Feature Brief Builder".

Goal:
- Take a rough feature idea for OpenADE
- Turn it into a concise implementation brief
- Produce sections for summary, user outcome, backend work, frontend work, risks, and next steps

Constraints:
- Keep the output concrete
- Prefer changes that match the current OpenADE architecture
- Be honest about current limitations

Ask me 3 short clarification questions first, then produce a reusable prompt pattern.
```

Then:
1. Open the `Objective` panel.
2. Fill it with:
   - `Title`: `OpenADE Feature Brief Builder`
   - `Goal`: turn rough feature ideas into implementation briefs
   - `Constraints`: concrete, architecture-aware, honest about limits
   - `Tools required`: `chat, objective, task wizard`
   - `Success criteria`: output is reusable and specific enough to implement
3. Answer the model’s clarification questions.

## 15-20 Minutes: Save The Chat As A Task

1. Click `Save as Task`.
2. In the task wizard:
   - confirm the generated name, or use `OpenADE Feature Brief Builder`
   - keep or refine the description
   - refine the prompt template so it uses variables such as `{{feature_idea}}`, `{{user_outcome}}`, and `{{constraints}}`
3. In the inputs step, give the fields readable labels.
4. Save the task.
5. The app will switch to `Tasks` and select the new task.

## 20-25 Minutes: Run The Task

In the `Run` panel for the task, enter inputs like:

- `feature_idea`: `Add a visual MCP tool browser with invoke support`
- `user_outcome`: `Users can inspect available tools and call them without leaving Settings`
- `constraints`: `Keep it simple and aligned with the current OpenADE UI`

Click `Run`.

What you should look for:
1. The output should be markdown.
2. It should separate backend and frontend work.
3. It should mention real project constraints, not generic product fluff.

If the output is weak:
1. Edit the task prompt.
2. Run it again.
3. Repeat until it behaves predictably.

## 25-30 Minutes: Add A Schedule

If the task is useful as a recurring planning aid:

1. Open the `Schedule` panel under the selected task.
2. Set a cron such as:
   - `0 9 * * 1` for every Monday at 9:00
3. Optionally set a timezone such as `America/New_York`.
4. Save the schedule.

Important limitation:
- In v1, scheduled jobs only run while the app and backend are open.

## What You Have At The End

By the end of this walkthrough, you will have:

1. a configured browser-based OpenADE setup
2. a real capability chat with an explicit objective
3. a saved reusable task
4. at least one run to inspect and refine
5. an optional recurring schedule

That is the current core OpenADE loop: chat -> objective -> task -> run -> schedule.
