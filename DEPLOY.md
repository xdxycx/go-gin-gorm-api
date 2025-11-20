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
