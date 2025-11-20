# Deployment Guide

This document provides concise deployment steps for local development and simple production testing using Docker Compose. It duplicates and expands the README's deployment section into a standalone guide.

Prerequisites
- Docker
- Docker Compose
- A copy of `.env.example` as `.env` with real credentials

4) Stop and remove containers
````
# Deployment Guide

This document provides concise deployment steps for local development and simple production testing using Docker Compose. It duplicates and expands the README's deployment section into a standalone guide.

Prerequisites
- Docker
- Docker Compose
- A copy of `.env.example` as `.env` with real credentials

1) Prepare environment

```bash
# copy example env and edit values
cp .env.example .env
# edit .env to set MYSQL_PASSWORD, MYSQL_ROOT_PASSWORD, and any other secrets
```

2) Start services with Docker Compose

```bash
docker-compose --env-file .env up --build -d

# follow logs; wait until migrations complete and app prints listening message
docker-compose logs -f app
```

3) Verify the app

```bash
curl http://localhost:8080/
# expected JSON health response
```

4) Stop and remove containers

```bash
docker-compose down
```

5) Override env values temporarily

```bash
# Example: start with different dynamic SQL limits
DYNAMIC_MAX_ROWS=500 DYNAMIC_QUERY_TIMEOUT_SECONDS=10 docker-compose --env-file .env up --build -d
```

Security notes
- Do NOT commit `.env` with real secrets. Use `.env.example` as a template. For production, inject secrets via orchestration (Kubernetes secrets, Docker secrets, or CI/CD secret management).
- The project currently logs audit records to the database; secure access to the `audits` table and avoid exposing audit data publicly.

Advanced
- Consider adding a systemd unit, Kubernetes manifests, or a CI/CD pipeline for repeatable deployments.

Scripts (local helpers)

The repository includes helper scripts under `scripts/` to simplify local testing and quick verification with Docker Compose:

- `scripts/start.sh`: Ensures `.env` exists (copies from `.env.example` if missing), runs `docker-compose --env-file .env up --build -d`, and tails `app` logs.
- `scripts/stop.sh`: Stops the Compose stack and performs lightweight cleanup (`docker-compose down`, `docker image prune -f`, `docker volume prune -f`).
- `scripts/status.sh`: Shows `docker-compose ps` output, the last ~100 lines of `app` logs and host-side `docker ps` port mappings.

Usage examples

```bash
# Prepare env (only first time or when you need to change secrets)
cp .env.example .env
# Start services and follow logs
./scripts/start.sh
# In another terminal: view status & recent logs
./scripts/status.sh
# Stop and clean up
./scripts/stop.sh
```

````
