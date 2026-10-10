# Runs the test binaries built by build-go-tests.ps1 like `go test -json ./...`:
# each binary runs in its package directory, packages run in parallel, and the
# test2json events are written to test-results.jsonl in package order.
param(
    [Parameter(Mandatory)]
    [string]$BinaryDirectory,
    [string]$Timeout = '10m',
    [int]$ThrottleLimit = [Environment]::ProcessorCount,
    [string]$ResultsFile = 'test-results.jsonl'
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version 3

$BinaryDirectory = (Resolve-Path $BinaryDirectory).Path
$manifest = Get-Content (Join-Path $BinaryDirectory 'manifest.json') -Raw | ConvertFrom-Json
$logDirectory = Join-Path $BinaryDirectory 'results'
New-Item -ItemType Directory -Force -Path $logDirectory | Out-Null

$runs = $manifest | ForEach-Object -ThrottleLimit $ThrottleLimit -Parallel {
    $package = $_
    $timeout = $using:Timeout
    $binary = Join-Path $using:BinaryDirectory $package.Binary
    $test2json = Join-Path $using:BinaryDirectory 'test2json.exe'
    $log = Join-Path $using:logDirectory "$($package.Binary).jsonl"

    # Native commands start in the runspace location, so tests see their
    # package directory as the working directory, as with `go test`.
    Set-Location $package.Dir
    $events = & $test2json -t -p $package.ImportPath $binary `
        '-test.v=test2json' "-test.timeout=$timeout" 2>&1 |
        ForEach-Object { $_.ToString() }
    $exitCode = $LASTEXITCODE
    $events | Set-Content -Encoding utf8 $log

    # Print each package as one block so parallel output does not interleave.
    $text = [System.Text.StringBuilder]::new()
    foreach ($line in $events) {
        try {
            $event = $line | ConvertFrom-Json -ErrorAction Stop
            if ($event.PSObject.Properties['Output']) { [void]$text.Append($event.Output) }
        } catch {
            [void]$text.AppendLine($line)
        }
    }

    [pscustomobject]@{
        ImportPath = $package.ImportPath
        Binary     = $package.Binary
        ExitCode   = $exitCode
        Text       = $text.ToString()
    }
} | ForEach-Object {
    [Console]::Write($_.Text)
    $_
}

$manifest | ForEach-Object { Get-Content (Join-Path $logDirectory "$($_.Binary).jsonl") } |
    Set-Content -Encoding utf8 $ResultsFile

$failed = @($runs | Where-Object ExitCode -ne 0)
if ($failed.Count -gt 0) {
    Write-Host "Failed packages:"
    $failed | ForEach-Object { Write-Host "  $($_.ImportPath) (exit code $($_.ExitCode))" }
    exit 1
}

exit 0
