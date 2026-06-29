# Run Vite dev server on host (proxies /api to localhost:8080).
Set-Location frontend
if (-not (Test-Path "node_modules")) {
  npm install
}
npm run dev
