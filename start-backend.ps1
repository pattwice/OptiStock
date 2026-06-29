# Run Go API on host (expects Postgres from .\start-db.ps1).
powershell -ExecutionPolicy Bypass -File .\scripts\ensure-dev-env.ps1
powershell -ExecutionPolicy Bypass -File .\scripts\stop-port.ps1 -Port 8080
Set-Location backend
if (Get-Command air -ErrorAction SilentlyContinue) {
  air -c .air.windows.toml
} else {
  Write-Host "air not found; using go run (install: go install github.com/air-verse/air@latest)"
  go run ./cmd/server
}
