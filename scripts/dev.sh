#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [ -f "$ROOT_DIR/.env" ]; then
  # shellcheck disable=SC1090
  source "$ROOT_DIR/.env"
fi

OPENADE_PORT="${OPENADE_PORT:-8080}"

cleanup() {
  if [ -n "${BACKEND_PID:-}" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
    kill "$BACKEND_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

echo "Starting backend API on http://localhost:${OPENADE_PORT} ..."
bash "$ROOT_DIR/scripts/backend.sh" &
BACKEND_PID=$!

echo "Waiting for backend health check..."
for i in {1..20}; do
  if curl -fsS "http://localhost:${OPENADE_PORT}/health" >/dev/null 2>&1; then
    echo "Backend is ready."
    break
  fi
  sleep 0.5
  if [ "$i" -eq 20 ]; then
    echo "Backend did not become ready on http://localhost:${OPENADE_PORT}/health" >&2
    exit 1
  fi
done

echo "Starting frontend web app on http://localhost:5173 ..."
exec bash "$ROOT_DIR/scripts/frontend.sh"
