# Builds the race-enabled test binaries ahead of time, so CI can link them in
# the background while Windows fixtures are provisioned. run-go-tests.ps1 runs
# the binaries listed in manifest.json.
param(
    [Parameter(Mandatory)]
    [string]$OutputDirectory
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version 3

New-Item -ItemType Directory -Force -Path $OutputDirectory | Out-Null
$OutputDirectory = (Resolve-Path $OutputDirectory).Path

$listing = go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{"\t"}}{{.Dir}}{{end}}' ./...
if ($LASTEXITCODE -ne 0) { throw "go list failed with exit code $LASTEXITCODE" }

$packages = @($listing | Where-Object { $_ } | ForEach-Object {
    $importPath, $dir = $_ -split "`t", 2
    [pscustomobject]@{
        ImportPath = $importPath
        Dir        = $dir
        Base       = $importPath.Split('/')[-1]
        Binary     = $null
    }
})

# `go test -c -o dir/` names binaries after the last import path element, so
# packages sharing that name (internal/collector, pkg/collector) build separately.
$groups = $packages | Group-Object Base
$shared = @($groups | Where-Object Count -eq 1 | ForEach-Object { $_.Group[0] })
$clashing = @($groups | Where-Object Count -gt 1 | ForEach-Object { $_.Group })

foreach ($package in $shared) {
    $package.Binary = "$($package.Base).test.exe"
}

go test -race -c -o "$OutputDirectory/" @($shared.ImportPath)
if ($LASTEXITCODE -ne 0) { throw "go test -c failed with exit code $LASTEXITCODE" }

foreach ($package in $clashing) {
    $package.Binary = ($package.ImportPath -replace '[/.]', '_') + '.test.exe'
    go test -race -c -o (Join-Path $OutputDirectory $package.Binary) $package.ImportPath
    if ($LASTEXITCODE -ne 0) { throw "go test -c $($package.ImportPath) failed with exit code $LASTEXITCODE" }
}

# Go does not ship test2json prebuilt; `go tool test2json` would build it on
# first use in every parallel runner.
go build -o (Join-Path $OutputDirectory 'test2json.exe') cmd/test2json
if ($LASTEXITCODE -ne 0) { throw "go build cmd/test2json failed with exit code $LASTEXITCODE" }

$packages | Select-Object ImportPath, Dir, Binary |
    ConvertTo-Json -AsArray |
    Set-Content -Encoding utf8 (Join-Path $OutputDirectory 'manifest.json')

Write-Host "Built $($packages.Count) test binaries in $OutputDirectory"
