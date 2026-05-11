# WhatsApp Platform (Go + Nuxt + Baileys + MariaDB)

Production-oriented starter for a WhatsApp bot platform with:

- Multi-device lifecycle management
- Professional monitoring dashboard
- Message logs and delivery states
- Webhook management baseline

## Monorepo Layout

- `backend/` Go API + worker
- `baileys-gateway/` Node.js Baileys adapter
- `frontend/` Nuxt 3 dashboard UI
- `deploy/` Docker Compose and infra helpers

## Quick Start

### Option A: Docker (local development)

1. Copy env files:
   - Root: `.env.example` to `.env`
   - Frontend: `frontend/.env.example` to `frontend/.env`
   - Gateway: `baileys-gateway/.env.example` to `baileys-gateway/.env`
2. Start services:

```bash
docker compose -f deploy/docker-compose.yml up --build
```

3. Open:

- Frontend: `http://localhost:3000`
- API health: `http://localhost:8080/healthz`

### Option B: Native install (Android box without Docker)

See [deploy/native/README.md](deploy/native/README.md) for the no-Docker deployment flow.

## Current Scope

This initial implementation includes:

- API skeleton, auth login endpoint, health/readiness checks
- Worker process skeleton
- Dashboard pages (login, dashboard, devices, logs)
- Baileys gateway skeleton with event bridge endpoint
- MariaDB migrations baseline

Next iterations will implement full WhatsApp session orchestration, queue pipeline, webhook retries, and full realtime metrics stream.
