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

# --- Dependency checks and helpers ---
check_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    return 1
  fi
  return 0
}

ensure_docker_available() {
  if ! check_command docker; then
    echo "ERROR: 'docker' command not found. Install Docker: https://docs.docker.com/get-docker/" >&2
    exit 2
  fi
  if ! docker info >/dev/null 2>&1; then
    echo "ERROR: Docker daemon not running or you lack permissions to talk to it." >&2
    echo "Start the Docker service or ensure your user is in the 'docker' group." >&2
    exit 3
  fi
}

detect_compose() {
  if check_command docker-compose; then
    COMPOSE_CMD="docker-compose"
  elif check_command docker && docker compose version >/dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
  else
    echo "ERROR: neither 'docker-compose' nor the 'docker compose' plugin is available." >&2
    echo "Install Docker Compose or enable the Compose plugin: https://docs.docker.com/compose/" >&2
    exit 4
  fi
}

run_compose() {
  # wrapper to support both 'docker-compose' and 'docker compose'
  if [ "${COMPOSE_CMD:-}" = "docker-compose" ]; then
    docker-compose "$@"
  else
    # shellcheck disable=SC2086
    docker compose "$@"
  fi
}

ensure_docker_available
detect_compose

echo "Starting services with compose using $ENV_FILE"
run_compose --env-file "$ENV_FILE" up --build -d

echo "Tailing app logs (press Ctrl-C to exit)"
run_compose logs -f app
