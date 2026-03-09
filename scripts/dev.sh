#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [ -f "$ROOT_DIR/.env" ]; then
  # shellcheck disable=SC1090
  source "$ROOT_DIR/.env"
fi

OPENADE_PORT="${OPENADE_PORT:-8080}"
VITE_API_URL="${VITE_API_URL:-http://localhost:${OPENADE_PORT}}"

cleanup() {
  if [ -n "${BACKEND_PID:-}" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
    echo "Stopping backend (pid: $BACKEND_PID)..."
    kill "$BACKEND_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

if [ -d "$ROOT_DIR/backend" ] && [ -f "$ROOT_DIR/backend/cmd/api/main.go" ]; then
  echo "Starting backend on :$OPENADE_PORT ..."
  (
    cd "$ROOT_DIR/backend"
    OPENADE_PORT="$OPENADE_PORT" OPENADE_DB_PATH="${OPENADE_DB_PATH:-$ROOT_DIR/backend/openade.db}" go run ./cmd/api
  ) &
  BACKEND_PID=$!

  echo "Waiting for backend health endpoint..."
  for i in {1..20}; do
    if curl -fsS "http://localhost:${OPENADE_PORT}/health" >/dev/null 2>&1; then
      echo "Backend is healthy."
      break
    fi
    sleep 0.5
  done
else
  echo "Backend not found yet. Running frontend only."
fi

echo "Starting frontend with VITE_API_URL=$VITE_API_URL ..."
cd "$ROOT_DIR/frontend"
VITE_API_URL="$VITE_API_URL" pnpm run dev
