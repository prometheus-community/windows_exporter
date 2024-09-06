[CmdletBinding()]
Param (
    [Parameter(Mandatory = $true)]
    [String] $PathToExecutable,
    [Parameter(Mandatory = $true)]
    [String] $Version,
    [Parameter(Mandatory = $false)]
    [ValidateSet("amd64", "arm64")]
    [String] $Arch = "amd64"
)
$ErrorActionPreference = "Stop"

# The MSI version is not semver compliant, so just take the numerical parts
$MsiVersion = $Version -replace '^v?([0-9\.]+).*$','$1'

# Get absolute path to executable before switching directories
$PathToExecutable = Resolve-Path $PathToExecutable
# Set working dir to this directory, reset previous on exit
Push-Location $PSScriptRoot
Trap {
    # Reset working dir on error
    Pop-Location
}

mkdir -Force Work | Out-Null
Copy-Item -Force $PathToExecutable Work/windows_exporter.exe

Write-Verbose "Creating windows_exporter-${Version}-${Arch}.msi"
$wixArch = @{"amd64" = "x64"; "arm64" = "arm64"}[$Arch]

Invoke-Expression "wix build -arch $wixArch -o .\windows_exporter-$($Version)-$($Arch).msi .\files.wxs .\main.wxs -d ProductName=windows_exporter -d Version=$($MsiVersion) -ext WixToolset.Firewall.wixext -ext WixToolset.UI.wixext -ext WixToolset.Util.wixext"

Write-Verbose "Done!"
Pop-Location
