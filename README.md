# OptiStock

Production & inventory management for contract packaging and secondary manufacturing.

## Stack

- **Backend:** Go + Fiber + PostgreSQL
- **Frontend:** React + TypeScript + Ant Design + Vite
- **Docs:** `doc/srs/`, `doc/architecture/`, `doc/plan/`

## Quick start (Docker)

```powershell
# Creates .env with a generated DB password on first run (never committed)
.\start.ps1
```

Or manually:

```powershell
# 1. Create local secrets file and JWT keys (first time only)
Copy-Item .env.example .env
# Edit .env — set POSTGRES_PASSWORD to a strong value
powershell -ExecutionPolicy Bypass -File .\scripts\generate-jwt-keys.ps1

# 2. Start all services
docker compose -f docker-compose.dev.yml up --build
```

- Frontend: http://localhost:5173
- API: http://localhost:8080/api/v1/health
- App login (seeded dev user): `admin@optistock.local` / `changeme`

> **Secrets:** `.env` and `backend/keys/` are gitignored. Only `.env.example` is tracked, with placeholders — never real passwords.

## Local development (without Docker)

### Backend

```powershell
cd backend
go run ./scripts/generatekeys/main.go keys
Copy-Item ..\.env.example .env
# Edit .env — set POSTGRES_PASSWORD (and POSTGRES_HOST=localhost)
go run ./cmd/server
```

### Frontend

```powershell
cd frontend
npm install
npm run dev
```

## API (Phase 0)

| Method | Path | Auth |
| :--- | :--- | :--- |
| GET | `/api/v1/health` | Public |
| POST | `/api/v1/auth/login` | Public |
| POST | `/api/v1/auth/refresh` | Cookie |
| POST | `/api/v1/auth/logout` | Cookie |
| GET | `/api/v1/me` | Bearer JWT |

## Project status

- [x] Phase 0 — Foundation
- [ ] Phase 1 — Item & LOT Management
