[CmdletBinding()]
Param (
    [Parameter(Mandatory = $true)]
    [String] $PathToExecutable,
    [Parameter(Mandatory = $true)]
    [String] $Version,
    # Numeric version of the package, such as 0.32.0.5 for 0.32.0-beta.5.
    # Defaults to the numeric part of Version.
    [Parameter(Mandatory = $false)]
    [String] $MsiVersion,
    [Parameter(Mandatory = $false)]
    [ValidateSet("amd64", "arm64")]
    [String] $Arch = "amd64"
)
$ErrorActionPreference = "Stop"

# The MSI version is not semver compliant, so just take the numerical parts
if (-not $MsiVersion) {
    $MsiVersion = $Version -replace '^v?([0-9\.]+).*$','$1'
}

# Get absolute path to executable before switching directories
$PathToExecutable = Resolve-Path $PathToExecutable
# Set working dir to this directory and restore the previous one on exit,
# including on errors.
Push-Location $PSScriptRoot
try {
    mkdir -Force Work | Out-Null
    Copy-Item -Force $PathToExecutable Work/windows_exporter.exe

    $msi = "windows_exporter-$Version-$Arch.msi"
    Write-Verbose "Creating $msi"
    $wixArch = @{"amd64" = "x64"; "arm64" = "arm64"}[$Arch]

    # A package left over from an earlier build must not pass for this one.
    if (Test-Path $msi) { Remove-Item -Force $msi }

    # Pass the arguments as an array so that values such as the version are
    # never parsed as PowerShell code.
    $wixArgs = @(
        "build"
        "-sw1149"
        "-arch", $wixArch
        "-o", ".\$msi"
        ".\files.wxs"
        ".\main.wxs"
        "-d", "ProductName=windows_exporter"
        "-d", "Version=$MsiVersion"
        "-ext", "WixToolset.Firewall.wixext"
        "-ext", "WixToolset.UI.wixext"
        "-ext", "WixToolset.Util.wixext"
    )
    & wix @wixArgs
    # $ErrorActionPreference does not apply to native commands.
    if ($LASTEXITCODE -ne 0) {
        throw "wix build failed with exit code $LASTEXITCODE for $msi"
    }

    if (-not (Test-Path -PathType Leaf $msi)) {
        throw "wix build did not create $msi"
    }

    Write-Verbose "Done!"
} finally {
    Pop-Location
}
