# OptiStock

Production & inventory management for contract packaging and secondary manufacturing.

## Stack

- **Backend:** Go + Fiber + PostgreSQL
- **Frontend:** React + TypeScript + Ant Design + Vite
- **Docs:** `doc/srs/`, `doc/architecture/`, `doc/plan/`

## Quick start — local dev (recommended)

Postgres in Docker; API and frontend on your machine (faster reload/debug).

**Prerequisites:** Go 1.26+, Node 20+, Docker Desktop

```powershell
# Terminal 1 — database
.\start-db.ps1

# Terminal 2 — API (reads .env from repo root)
.\start-backend.ps1

# Terminal 3 — frontend
.\start-frontend.ps1
```

- Frontend: http://localhost:5173
- API: http://localhost:8080/api/v1/health
- Login: `admin@optistock.local` / `changeme`

First run creates `.env` (random DB password) and `backend/keys/` if missing.

> **Secrets:** `.env` and `backend/keys/` are gitignored. Only `.env.example` is tracked, with placeholders.

## Full stack in Docker (optional)

Use for onboarding or prod-like checks:

```powershell
.\start.ps1
```

Same URLs as above. API runs with `air` inside the container.

## Manual setup (without scripts)

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\ensure-dev-env.ps1
docker compose -f docker-compose.db.yml up -d

cd backend
go run ./cmd/server

# separate terminal
cd frontend
npm install
npm run dev
```

`.env.example` sets `POSTGRES_HOST=localhost` for host API. Docker full-stack overrides to `postgres` in `docker-compose.dev.yml`.

## API

| Method | Path | Auth |
| :--- | :--- | :--- |
| GET | `/api/v1/health` | Public |
| POST | `/api/v1/auth/login` | Public |
| POST | `/api/v1/auth/refresh` | Cookie |
| POST | `/api/v1/auth/logout` | Cookie |
| GET | `/api/v1/me` | Bearer JWT |
| GET/POST/PATCH | `/api/v1/items` | Bearer JWT |
| GET/POST | `/api/v1/bom` | Bearer JWT |
| GET/POST | `/api/v1/lots` | Bearer JWT |
| GET | `/api/v1/ledger` | Bearer JWT |
| POST | `/api/v1/receiving/po`, `/adjustment` | Bearer JWT |

## Project status

- [x] Phase 0 — Foundation
- [ ] Phase 1 — Item & LOT Management (backend APIs done; frontend pages pending)
