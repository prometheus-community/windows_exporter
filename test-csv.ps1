# Quick test script for CSV collector

Write-Host "Starting windows_exporter with CSV collector..." -ForegroundColor Green

go run ./cmd/windows_exporter `
    --collectors.enabled=csv `
    --log.level=debug `
    --web.listen-address=:9999

# Usage: .\test-csv.ps1
# Then visit http://localhost:9999/metrics in your browser
