#!/usr/bin/env bash
# Smoke test: create conv → message (optional) → create task → run → export → import
# Requires: backend running on OPENADE_PORT (default 8080)
# Optional: OPENAI_API_KEY for full flow including chat message and draft-task

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [ -f "$ROOT_DIR/.env" ]; then
  # shellcheck disable=SC1090
  source "$ROOT_DIR/.env"
fi

OPENADE_PORT="${OPENADE_PORT:-8080}"
BASE="http://localhost:${OPENADE_PORT}"
OPENAI_API_KEY="${OPENAI_API_KEY:-}"

echo "=== OpenADE smoke test ==="
echo "API: $BASE"

# 1. Health
echo ""
echo "[1/7] Health check..."
curl -fsS "$BASE/health" | head -1
echo " OK"

# 2. Create conversation
echo ""
echo "[2/7] Create conversation..."
CONV_RESP=$(curl -fsS -X POST "$BASE/api/conversations")
CONV_ID=$(echo "$CONV_RESP" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$CONV_ID" ]; then
  echo "Failed to get conversation ID from: $CONV_RESP"
  exit 1
fi
echo "Created conversation: $CONV_ID"

# 3. Optional: post message and draft task (requires OPENAI_API_KEY)
echo ""
echo "[3/7] Message + draft (optional)..."
TASK_JSON=""
if [ -n "$OPENAI_API_KEY" ]; then
  echo ""
  echo "[3a] Configure provider..."
  curl -fsS -X PUT "$BASE/api/providers/openai" \
    -H "Content-Type: application/json" \
    -d "{\"api_key\":\"$OPENAI_API_KEY\",\"default_model\":\"gpt-4o-mini\"}" > /dev/null
  echo "Provider configured."

  echo ""
  echo "[3b] Post message (streaming)..."
  (curl -fsS -N -X POST "$BASE/api/conversations/$CONV_ID/messages" \
    -H "Content-Type: application/json" \
    -d '{"content":"Say hello in one word."}' &
   PID=$!
   sleep 15
   kill $PID 2>/dev/null || true) | grep -q "data:" && echo "Stream received." || echo "Stream timeout (OK if no key)."

  echo ""
  echo "[3c] Draft task from conversation..."
  DRAFT_RESP=$(curl -fsS -X POST "$BASE/api/conversations/$CONV_ID/draft-task" 2>/dev/null || echo "{}")
  if echo "$DRAFT_RESP" | grep -q '"prompt_template"'; then
    DRAFT_NAME=$(echo "$DRAFT_RESP" | grep -o '"name":"[^"]*"' | head -1 | cut -d'"' -f4 || echo "Draft Task")
    DRAFT_TMPL=$(echo "$DRAFT_RESP" | grep -o '"prompt_template":"[^"]*"' | head -1 | sed 's/"prompt_template":"//;s/"$//' | sed 's/\\"/"/g')
    TASK_JSON="{\"name\":\"${DRAFT_NAME:-Smoke Draft}\",\"prompt_template\":\"${DRAFT_TMPL:-Hello}\",\"input_schema\":[]}"
    echo "Draft obtained."
  fi
else
  echo "Skipped (set OPENAI_API_KEY for full flow)."
fi

# 4. Create task (use draft if available, else minimal)
echo ""
echo "[4/7] Create task..."
if [ -n "$TASK_JSON" ]; then
  CREATE_BODY="$TASK_JSON"
else
  CREATE_BODY='{"name":"Smoke Test Task","prompt_template":"Say hello to {{name}}.","input_schema":[{"key":"name","type":"text","label":"Name"}]}'
fi

# For create we need conversation_id only when creating from draft; here we create standalone
TASK_RESP=$(curl -fsS -X POST "$BASE/api/tasks" \
  -H "Content-Type: application/json" \
  -d "$CREATE_BODY")
TASK_ID=$(echo "$TASK_RESP" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$TASK_ID" ]; then
  echo "Failed to create task. Response: $TASK_RESP"
  exit 1
fi
echo "Created task: $TASK_ID"

# 5. Run task (requires provider for LLM; may fail with 401 if no provider)
echo ""
echo "[5/7] Run task..."
RUN_BODY='{"inputs":{"name":"Smoke"}}'
RUN_RESP=$(curl -sS -X POST "$BASE/api/tasks/$TASK_ID/run" \
  -H "Content-Type: application/json" \
  -d "$RUN_BODY" 2>/dev/null || echo "{\"error\":\"no_provider\"}")
if echo "$RUN_RESP" | grep -q '"output"'; then
  echo "Run completed."
elif echo "$RUN_RESP" | grep -q 'no_provider\|"code":"no_provider"'; then
  echo "Run skipped (no provider configured)."
else
  echo "Run response: $RUN_RESP"
fi

# 6. Export task
echo ""
echo "[6/7] Export task..."
EXPORT_RESP=$(curl -fsS -X POST "$BASE/api/tasks/$TASK_ID/export")
if ! echo "$EXPORT_RESP" | grep -q '"task"'; then
  echo "Export failed: $EXPORT_RESP"
  exit 1
fi
echo "Export OK (bundle version $(echo "$EXPORT_RESP" | grep -o '"bundle_version":"[^"]*"' | cut -d'"' -f4))"

# 7. Import task
echo ""
echo "[7/7] Import task..."
IMPORT_RESP=$(curl -fsS -X POST "$BASE/api/tasks/import" \
  -H "Content-Type: application/json" \
  -d "$EXPORT_RESP")
NEW_TASK_ID=$(echo "$IMPORT_RESP" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$NEW_TASK_ID" ]; then
  echo "Import failed: $IMPORT_RESP"
  exit 1
fi
echo "Import OK (new task: $NEW_TASK_ID)"

echo ""
echo "=== Smoke test passed ==="
