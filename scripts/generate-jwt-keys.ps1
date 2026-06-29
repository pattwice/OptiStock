$KeyDir = if ($args.Count -gt 0) { $args[0] } else { "backend\keys" }
Set-Location (Split-Path $PSScriptRoot -Parent)
go run ./backend/scripts/generatekeys/main.go $KeyDir
