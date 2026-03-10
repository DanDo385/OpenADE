#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [ -f "$ROOT_DIR/.env" ]; then
  # shellcheck disable=SC1090
  source "$ROOT_DIR/.env"
fi

if [ -f "$ROOT_DIR/frontend/.env" ]; then
  # shellcheck disable=SC1090
  source "$ROOT_DIR/frontend/.env"
fi

cd "$ROOT_DIR/frontend"
exec env VITE_API_URL="${VITE_API_URL:-}" pnpm exec vite --host --port 5173 --strictPort
