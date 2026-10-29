# Builds the packages compare.ps1 installs: WiX (installer/) and msi-builder
# (examples/windows_exporter/msi.yaml) at 0.90.0, 0.91.0 and 0.92.0, from one
# windows_exporter.exe. Run from the repository root.
[CmdletBinding()]
Param (
    [Parameter(Mandatory = $true)] [String] $MsiBuilder,
    [String] $Dir = 'cmp',
    [String] $Out = 'out'
)
$ErrorActionPreference = 'Stop'
New-Item -ItemType Directory -Force $Dir, $Out | Out-Null
$Dir = (Resolve-Path $Dir).Path

Push-Location cmd/windows_exporter
go run github.com/tc-hib/go-winres@v0.3.3 make --product-version=0.90.0.1000 --file-version=0.90.0.1000 --arch=amd64
if ($LASTEXITCODE) { throw 'go-winres failed' }
Pop-Location
go build -o "$Dir\windows_exporter.exe" ./cmd/windows_exporter
if ($LASTEXITCODE) { throw 'go build windows_exporter failed' }
git checkout -- cmd/windows_exporter 2>$null

Push-Location $MsiBuilder
go build -o "$Dir\msi-builder.exe" ./cmd/msi-builder
if ($LASTEXITCODE) { throw 'go build msi-builder failed' }
Pop-Location

foreach ($v in '0.90.0', '0.91.0', '0.92.0') {
    .\installer\build.ps1 -PathToExecutable "$Dir\windows_exporter.exe" -Version $v -Arch amd64
    Move-Item -Force "installer\windows_exporter-$v-amd64.msi" "$Dir\ref-$v.msi"
    & "$Dir\msi-builder.exe" build --var VERSION=$v --var ARCH=x64 --var "BINARY=$Dir\windows_exporter.exe" --var "ICON=$PWD\installer\icon.ico" -o "$Dir\ours-$v.msi" "$MsiBuilder\examples\windows_exporter\msi.yaml"
    if ($LASTEXITCODE) { throw "msi-builder build $v failed" }
}
& "$Dir\msi-builder.exe" inspect "$Dir\ref-0.90.0.msi" | Set-Content -Encoding utf8 "$Out\ref-tables.txt"
& "$Dir\msi-builder.exe" inspect "$Dir\ours-0.90.0.msi" | Set-Content -Encoding utf8 "$Out\ours-tables.txt"
