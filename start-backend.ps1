# Run Go API on host (expects Postgres from .\start-db.ps1).
powershell -ExecutionPolicy Bypass -File .\scripts\ensure-dev-env.ps1
Set-Location backend
if (Get-Command air -ErrorAction SilentlyContinue) {
  air -c .air.toml
} else {
  Write-Host "air not found; using go run (install: go install github.com/air-verse/air@latest)"
  go run ./cmd/server
}
