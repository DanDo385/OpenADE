#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [ -f "$ROOT_DIR/.env" ]; then
  # shellcheck disable=SC1090
  source "$ROOT_DIR/.env"
fi

OPENADE_PORT="${OPENADE_PORT:-8080}"
VITE_API_URL="${VITE_API_URL:-http://localhost:${OPENADE_PORT}}"

# Resolve OPENADE_DB_PATH to absolute so it works regardless of backend cwd
OPENADE_DB_PATH="${OPENADE_DB_PATH:-$ROOT_DIR/backend/openade.db}"
if [[ "$OPENADE_DB_PATH" != /* ]]; then
  OPENADE_DB_PATH="$ROOT_DIR/$OPENADE_DB_PATH"
fi

cleanup() {
  if [ -n "${BACKEND_PID:-}" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
    echo "Stopping backend (pid: $BACKEND_PID)..."
    kill "$BACKEND_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

# Free ports if previous runs left processes
for port in 8080 5173; do
  if lsof -ti :"$port" >/dev/null 2>&1; then
    echo "Port $port in use, stopping previous process..."
    lsof -ti :"$port" | xargs kill -9 2>/dev/null || true
    sleep 1
  fi
done

if [ -d "$ROOT_DIR/backend" ] && [ -f "$ROOT_DIR/backend/cmd/api/main.go" ]; then
  echo "Starting backend on :$OPENADE_PORT ..."
  (
    cd "$ROOT_DIR/backend"
    OPENADE_PORT="$OPENADE_PORT" OPENADE_DB_PATH="$OPENADE_DB_PATH" go run ./cmd/api
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
echo ""
echo "  → Open http://localhost:5173 in your browser (first run may take ~60s)"
echo ""
cd "$ROOT_DIR/frontend"
# Node 25 hangs Vite; pnpm dlx node@22 ensures it works
# Use empty VITE_API_URL so frontend uses Vite proxy (same-origin, avoids CORS)
VITE_API_URL= pnpm dlx node@22 ./node_modules/vite/bin/vite.js --host --port 5173 --strictPort
