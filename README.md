# OptiStock

Production & inventory management for contract packaging and secondary manufacturing.

## Stack

- **Backend:** Go + Fiber + PostgreSQL
- **Frontend:** React + TypeScript + Ant Design + Vite
- **Docs:** `doc/srs/`, `doc/architecture/`, `doc/plan/`

## Quick start (Docker)

```powershell
# 1. Copy env and generate JWT keys (first time only)
Copy-Item .env.example .env
powershell -ExecutionPolicy Bypass -File .\scripts\generate-jwt-keys.ps1

# 2. Start all services
docker compose -f docker-compose.dev.yml up --build
```

- Frontend: http://localhost:5173
- API: http://localhost:8080/api/v1/health
- Default login: `admin@optistock.local` / `changeme`

## Local development (without Docker)

### Backend

```powershell
cd backend
go run ./scripts/generatekeys/main.go keys
Copy-Item ..\.env.example .env
# Start PostgreSQL locally and set DATABASE_URL in .env
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
