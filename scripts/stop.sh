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

# Dependency checks
check_command() { command -v "$1" >/dev/null 2>&1; }
if ! check_command docker; then
  echo "ERROR: 'docker' command not found. Install Docker: https://docs.docker.com/get-docker/" >&2
  exit 2
fi
if check_command docker-compose; then
  COMPOSE_CMD="docker-compose"
elif check_command docker && docker compose version >/dev/null 2>&1; then
  COMPOSE_CMD="docker compose"
else
  echo "ERROR: neither 'docker-compose' nor 'docker compose' available. Install Docker Compose." >&2
  exit 3
fi

run_compose() {
  if [ "$COMPOSE_CMD" = "docker-compose" ]; then
    docker-compose "$@"
  else
    docker compose "$@"
  fi
}

if [ -n "$ENV_FILE" ]; then
  echo "Stopping services using $ENV_FILE"
  run_compose --env-file "$ENV_FILE" down
else
  echo "Stopping services (no env file)"
  run_compose down
fi

echo "Removing unused images and volumes (optional)"
docker image prune -f || true
docker volume prune -f || true

echo "Stopped and cleaned up." 
