# Copy env and generate JWT keys on first run
if (-not (Test-Path ".env")) {
  Copy-Item ".env.example" ".env"
}

if (-not (Test-Path "backend\keys\private.pem")) {
  powershell -ExecutionPolicy Bypass -File .\scripts\generate-jwt-keys.ps1
}

docker compose -f docker-compose.dev.yml up --build
