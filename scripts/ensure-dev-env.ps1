# Ensures .env and JWT keys exist for local / Docker dev. Safe to run repeatedly.
$ErrorActionPreference = "Stop"
$RepoRoot = Split-Path $PSScriptRoot -Parent
Set-Location $RepoRoot

function New-RandomPassword([int]$Length = 32) {
  $chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
  -join (1..$Length | ForEach-Object { $chars[(Get-Random -Maximum $chars.Length)] })
}

if (-not (Test-Path ".env")) {
  $password = New-RandomPassword
  (Get-Content ".env.example" -Raw) -replace 'change-me-before-starting', $password | Set-Content ".env" -NoNewline
  Write-Host "Created .env with a generated POSTGRES_PASSWORD (not committed to git)."
}

if (-not (Test-Path "backend\keys\private.pem")) {
  powershell -ExecutionPolicy Bypass -File .\scripts\generate-jwt-keys.ps1
}
