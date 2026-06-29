# Stops the process listening on a TCP port (Windows). No-op if port is free.
param([int]$Port = 8080)

$lines = netstat -ano | Select-String ":$Port\s+.*LISTENING"
if (-not $lines) { return }

$pids = $lines | ForEach-Object {
  if ($_ -match '\s+(\d+)\s*$') { [int]$Matches[1] }
} | Sort-Object -Unique

foreach ($procId in $pids) {
  if ($procId -eq 0) { continue }
  Write-Host "Stopping PID $procId on port $Port..."
  Stop-Process -Id $procId -Force -ErrorAction SilentlyContinue
}
