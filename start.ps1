# Full stack in Docker (postgres + api + frontend). For local dev, prefer .\start-db.ps1 + host API/FE.
powershell -ExecutionPolicy Bypass -File .\scripts\ensure-dev-env.ps1
docker compose -f docker-compose.dev.yml up --build
