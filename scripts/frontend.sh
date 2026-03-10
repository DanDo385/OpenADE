#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEBUG_LOG_PATH="/Users/danmagro/Desktop/Code/open-ade/.cursor/debug-73e4d6.log"
RUN_ID="pre-fix"

# #region agent log
printf '{"sessionId":"73e4d6","runId":"%s","hypothesisId":"H10","location":"scripts/frontend.sh:9","message":"frontend_script_start","data":{"pid":%d,"node":"%s"},"timestamp":%s}\n' \
  "$RUN_ID" "$$" "$(node -v 2>/dev/null || echo unknown)" "$(date +%s%3N)" >> "$DEBUG_LOG_PATH"
# #endregion

trap '
  code=$?
  # #region agent log
  printf "{\"sessionId\":\"73e4d6\",\"runId\":\"%s\",\"hypothesisId\":\"H10\",\"location\":\"scripts/frontend.sh:16\",\"message\":\"frontend_script_exit\",\"data\":{\"pid\":%d,\"exitCode\":%d},\"timestamp\":%s}\n" \
    "'"$RUN_ID"'" "$$" "$code" "$(date +%s%3N)" >> "'"$DEBUG_LOG_PATH"'"
  # #endregion
' EXIT

cd "$ROOT_DIR/frontend"
VITE_API_URL= pnpm dlx node@22 ./node_modules/vite/bin/vite.js --host --port 5173 --strictPort
