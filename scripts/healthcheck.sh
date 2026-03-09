#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [ -f "$ROOT_DIR/.env" ]; then
  # shellcheck disable=SC1090
  source "$ROOT_DIR/.env"
fi

OPENADE_PORT="${OPENADE_PORT:-8080}"
URL="http://localhost:${OPENADE_PORT}/health"

echo "Checking $URL"
curl -fsS "$URL" && echo "\nOK"
