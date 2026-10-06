$ErrorActionPreference = 'Stop'
Set-StrictMode -Version 3

$binary = Join-Path $PSScriptRoot '..\windows_exporter.exe'
$logDir = if ($env:RUNNER_TEMP) { $env:RUNNER_TEMP } else { $env:TEMP }
$stdout = Join-Path $logDir 'windows_exporter.stdout.log'
$stderr = Join-Path $logDir 'windows_exporter.stderr.log'
$exporter = Start-Process -FilePath $binary -PassThru `
    -ArgumentList '--web.listen-address=127.0.0.1:9182' `
    -RedirectStandardOutput $stdout -RedirectStandardError $stderr

try {
    $deadline = (Get-Date).AddSeconds(60)
    do {
        if ($exporter.HasExited) {
            throw "windows_exporter exited with code $($exporter.ExitCode)"
        }

        $client = [System.Net.Sockets.TcpClient]::new()
        try {
            $client.Connect('127.0.0.1', 9182)
            break
        } catch {
            if ((Get-Date) -ge $deadline) { throw 'windows_exporter did not start within 60 seconds' }
            Start-Sleep -Milliseconds 200
        } finally {
            $client.Dispose()
        }
    } while ($true)

    # Metric validation belongs in Go tests. This is one binary smoke request.
    $response = Invoke-WebRequest -UseBasicParsing -Uri 'http://127.0.0.1:9182/metrics' -TimeoutSec 30
    if ($response.StatusCode -ne 200 -or [string]::IsNullOrWhiteSpace($response.Content)) {
        throw 'windows_exporter returned an unsuccessful or empty metrics response'
    }
    Write-Host "windows_exporter returned HTTP $($response.StatusCode)"
} finally {
    if (-not $exporter.HasExited) { Stop-Process -Id $exporter.Id }
    Get-Content $stdout -ErrorAction SilentlyContinue
    Get-Content $stderr -ErrorAction SilentlyContinue
}
