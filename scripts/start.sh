#!/usr/bin/env bash
set -euo pipefail

# Simple helper to ensure .env exists and start docker-compose for local testing.
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

ENV_FILE=".env"
if [ ! -f "$ENV_FILE" ]; then
  if [ -f ".env.example" ]; then
    echo ".env not found — copying from .env.example to .env"
    cp .env.example .env
    echo "Please review and edit .env before using in production. Continuing with copied .env..."
  else
    echo "ERROR: .env not found and .env.example not present. Create one and retry." >&2
    exit 1
  fi
fi

echo "Starting services with docker-compose using $ENV_FILE"
docker-compose --env-file "$ENV_FILE" up --build -d

echo "Tailing app logs (press Ctrl-C to exit)"
docker-compose logs -f app
