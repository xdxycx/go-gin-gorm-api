#!/usr/bin/env bash
set -euo pipefail

# Show docker-compose status and basic health info for the app
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

ENV_FILE=".env"
if [ ! -f "$ENV_FILE" ] && [ -f ".env.example" ]; then
  # don't create .env here, just warn
  echo ".env not found — using .env.example values only for display purposes"
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

echo "Docker Compose services (ps):"
run_compose --env-file .env ps || run_compose ps

echo "\nBrief app container logs (last 100 lines):"
run_compose --env-file .env logs --tail=100 app || run_compose logs --tail=100 app

echo "\nOpen ports (host view):"
docker ps --format 'table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Ports}}'
