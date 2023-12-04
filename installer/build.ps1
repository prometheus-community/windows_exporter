[CmdletBinding()]
Param (
    [Parameter(Mandatory = $true)]
    [String] $PathToExecutable,
    [Parameter(Mandatory = $true)]
    [String] $Version,
    [Parameter(Mandatory = $false)]
    [ValidateSet("amd64", "arm64", "386")]
    [String] $Arch = "amd64"
)
$ErrorActionPreference = "Stop"

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
$wixArch = @{"amd64" = "x64"; "arm64" = "arm64"; "386" = "x86"}[$Arch]
$wixOpts = "-ext WixFirewallExtension -ext WixUtilExtension"
Invoke-Expression "wix build -arch $wixArch -o .\windows_exporter-$($Version)-$($Arch).msi .\windows_exporter.wxs -d Version=$($Version) -ext WixToolset.Firewall.wixext -ext WixToolset.Util.wixext"

Write-Verbose "Done!"
Pop-Location
