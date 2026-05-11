# Deployment

## Docker-based local development

Run all services:

```bash
docker compose -f deploy/docker-compose.yml up --build
```

Stop and remove:

```bash
docker compose -f deploy/docker-compose.yml down -v
```

## Native no-Docker deployment

For Android box or other bare-metal installs, follow [deploy/native/README.md](deploy/native/README.md).
