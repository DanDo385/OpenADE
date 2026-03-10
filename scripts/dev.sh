#!/usr/bin/env bash
# Combined dev: starts backend + frontend. Prefer separate terminals (dev:backend, dev:frontend).
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [ -f "$ROOT_DIR/.env" ]; then
  # shellcheck disable=SC1090
  source "$ROOT_DIR/.env"
fi

OPENADE_PORT="${OPENADE_PORT:-8080}"
OPENADE_DB_PATH="${OPENADE_DB_PATH:-$ROOT_DIR/backend/openade.db}"
[[ "$OPENADE_DB_PATH" != /* ]] && OPENADE_DB_PATH="$ROOT_DIR/$OPENADE_DB_PATH"

cleanup() {
  [ -n "${BACKEND_PID:-}" ] && kill -0 "$BACKEND_PID" 2>/dev/null && kill "$BACKEND_PID" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

echo "Starting backend on :$OPENADE_PORT ..."
(cd "$ROOT_DIR/backend" && OPENADE_PORT="$OPENADE_PORT" OPENADE_DB_PATH="$OPENADE_DB_PATH" go run ./cmd/api) &
BACKEND_PID=$!

echo "Waiting for backend..."
for i in {1..20}; do
  curl -fsS "http://localhost:${OPENADE_PORT}/health" >/dev/null 2>&1 && echo "Backend ready." && break
  sleep 0.5
done

echo "Starting frontend..."
cd "$ROOT_DIR/frontend"
VITE_API_URL= pnpm run dev
