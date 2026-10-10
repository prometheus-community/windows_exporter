# Runs the test binaries built by build-go-tests.ps1 like `go test -v ./...`:
# each binary runs in its package directory and packages run in parallel.
param(
    [Parameter(Mandatory)]
    [string]$BinaryDirectory,
    [string]$Timeout = '10m',
    [int]$ThrottleLimit = [Environment]::ProcessorCount
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version 3

$BinaryDirectory = (Resolve-Path $BinaryDirectory).Path
$manifest = Get-Content (Join-Path $BinaryDirectory 'manifest.json') -Raw | ConvertFrom-Json

$runs = $manifest | ForEach-Object -ThrottleLimit $ThrottleLimit -Parallel {
    $package = $_
    $timeout = $using:Timeout
    $binary = Join-Path $using:BinaryDirectory $package.Binary

    # Native commands start in the runspace location, so tests see their
    # package directory as the working directory, as with `go test`.
    Set-Location $package.Dir
    $started = Get-Date
    $output = & $binary '-test.v' "-test.timeout=$timeout" 2>&1 |
        ForEach-Object { $_.ToString() }
    $exitCode = $LASTEXITCODE
    $elapsed = ((Get-Date) - $started).TotalSeconds

    $status = if ($exitCode -eq 0) { 'ok  ' } else { 'FAIL' }
    [pscustomobject]@{
        ImportPath = $package.ImportPath
        ExitCode   = $exitCode
        # Print each package as one block so parallel output does not interleave.
        Text       = (@($output) + ("{0}`t{1}`t{2:N3}s" -f $status, $package.ImportPath, $elapsed)) -join "`n"
    }
} | ForEach-Object {
    Write-Host $_.Text
    $_
}

$failed = @($runs | Where-Object ExitCode -ne 0)
if ($failed.Count -gt 0) {
    Write-Host "Failed packages:"
    $failed | ForEach-Object { Write-Host "  $($_.ImportPath) (exit code $($_.ExitCode))" }
    exit 1
}

exit 0
