#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEBUG_LOG_PATH="/Users/danmagro/Desktop/Code/open-ade/.cursor/debug-73e4d6.log"
RUN_ID="pre-fix"

if [ -f "$ROOT_DIR/.env" ]; then
  # shellcheck disable=SC1090
  source "$ROOT_DIR/.env"
fi

OPENADE_PORT="${OPENADE_PORT:-8080}"
OPENADE_DB_PATH="${OPENADE_DB_PATH:-$ROOT_DIR/backend/openade.db}"
[[ "$OPENADE_DB_PATH" != /* ]] && OPENADE_DB_PATH="$ROOT_DIR/$OPENADE_DB_PATH"

# #region agent log
printf '{"sessionId":"73e4d6","runId":"%s","hypothesisId":"H9","location":"scripts/backend.sh:17","message":"backend_script_start","data":{"pid":%d,"port":"%s"},"timestamp":%s}\n' \
  "$RUN_ID" "$$" "$OPENADE_PORT" "$(date +%s%3N)" >> "$DEBUG_LOG_PATH"
# #endregion

trap '
  code=$?
  # #region agent log
  printf "{\"sessionId\":\"73e4d6\",\"runId\":\"%s\",\"hypothesisId\":\"H9\",\"location\":\"scripts/backend.sh:25\",\"message\":\"backend_script_exit\",\"data\":{\"pid\":%d,\"exitCode\":%d},\"timestamp\":%s}\n" \
    "'"$RUN_ID"'" "$$" "$code" "$(date +%s%3N)" >> "'"$DEBUG_LOG_PATH"'"
  # #endregion
' EXIT

cd "$ROOT_DIR/backend"
OPENADE_PORT="$OPENADE_PORT" OPENADE_DB_PATH="$OPENADE_DB_PATH" go run ./cmd/api
