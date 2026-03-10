#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [ -f "$ROOT_DIR/.env" ]; then
  # shellcheck disable=SC1090
  source "$ROOT_DIR/.env"
fi

OPENADE_PORT="${OPENADE_PORT:-8080}"
OPENADE_DB_PATH="${OPENADE_DB_PATH:-$ROOT_DIR/backend/openade.db}"
[[ "$OPENADE_DB_PATH" != /* ]] && OPENADE_DB_PATH="$ROOT_DIR/$OPENADE_DB_PATH"

cd "$ROOT_DIR/backend"
exec env OPENADE_PORT="$OPENADE_PORT" OPENADE_DB_PATH="$OPENADE_DB_PATH" go run ./cmd/api
