#!/usr/bin/env bash
set -euo pipefail

# Stop and remove docker-compose services for local testing
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

ENV_FILE=".env"
if [ ! -f "$ENV_FILE" ]; then
  if [ -f ".env.example" ]; then
    echo ".env not found — copying from .env.example to .env"
    cp .env.example .env
    echo "Created .env from .env.example"
  else
    echo "WARNING: .env not found and .env.example not present. Proceeding without --env-file." >&2
    ENV_FILE=""
  fi
fi

if [ -n "$ENV_FILE" ]; then
  echo "Stopping services using $ENV_FILE"
  docker-compose --env-file "$ENV_FILE" down
else
  echo "Stopping services (no env file)"
  docker-compose down
fi

echo "Removing unused images and volumes (optional)"
docker image prune -f || true
docker volume prune -f || true

echo "Stopped and cleaned up." 
