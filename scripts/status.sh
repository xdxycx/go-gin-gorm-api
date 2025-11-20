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

echo "Docker Compose services (ps):"
docker-compose --env-file .env ps || docker-compose ps

echo "\nBrief app container logs (last 100 lines):"
docker-compose --env-file .env logs --tail=100 app || docker-compose logs --tail=100 app

echo "\nOpen ports (host view):"
docker ps --format 'table {{.ID}}	{{.Names}}	{{.Status}}	{{.Ports}}'
