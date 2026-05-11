# Native Deployment Guide (No Docker)

This guide is for an Android box or other Linux-based server where Docker is not available.

## Recommended Topology

- `backend` Go API process
- `worker` Go background process
- `baileys-gateway` Node.js process
- `frontend` Nuxt process
- `MariaDB` external or locally installed
- `Redis` optional for now, required later for queue/realtime scale-out

## Prerequisites

- Go 1.23+
- Node.js 22+
- MariaDB 10.6+ reachable from the box
- Optional: Redis 7+

If the Android box is ARM64, install ARM64 builds of Go and Node.

## Environment

Create a `.env` file for each service from the example files:

- `.env.example`
- `backend/.env` (or export env vars in shell)
- `baileys-gateway/.env`
- `frontend/.env`

For local development without Redis, keep:

```bash
REDIS_OPTIONAL=true
```

## Start Order

1. Prepare MariaDB and create the database:

```sql
CREATE DATABASE IF NOT EXISTS wa_platform CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

2. Run database bootstrap from the backend folder:

```bash
go run ./cmd/dbinit
```

3. Start the API:

```bash
go run ./cmd/api
```

4. Start the worker:

```bash
go run ./cmd/worker
```

5. Start the Baileys gateway:

```bash
cd ../baileys-gateway
npm install
npm run dev
```

6. Start the frontend:

```bash
cd ../frontend
npm install
npm run dev
```

## Production Build

### Backend

```bash
go build -o bin/api ./cmd/api
go build -o bin/worker ./cmd/worker
```

### Baileys Gateway

```bash
npm run build
npm run start
```

### Frontend

```bash
npm run build
npm run start
```

## Notes for Android Box

- If the box has limited RAM, run the frontend in static mode later to reduce Node memory footprint.
- If MariaDB cannot run reliably on the box, use a remote MariaDB instance on the LAN/VPN.
- Redis is currently optional for the starter; the system will start without it.
