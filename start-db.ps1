# Start PostgreSQL only (local dev: API + frontend run on host).
powershell -ExecutionPolicy Bypass -File .\scripts\ensure-dev-env.ps1
docker compose -f docker-compose.db.yml up -d
Write-Host ""
Write-Host "PostgreSQL is starting on localhost:5432"
Write-Host "Next:"
Write-Host "  Terminal 2: .\start-backend.ps1"
Write-Host "  Terminal 3: .\start-frontend.ps1"
