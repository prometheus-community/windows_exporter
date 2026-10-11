# Tests the MSI lifecycle on a disposable Windows host: upgrade from the
# previous release, repair, uninstall and a fresh install. An independent
# windows_exporter.exe from another directory must survive every operation
# of the new package. Run it only on a throwaway machine such as a CI runner:
# it installs and removes the windows_exporter service.

[CmdletBinding()]
Param (
    # MSI built by this pipeline.
    [Parameter(Mandatory = $true)]
    [String] $Msi,
    # MSI used for the upgrade. Its version must be higher than the previous
    # release. Defaults to $Msi.
    [Parameter(Mandatory = $false)]
    [String] $UpgradeMsi = '',
    # Executable packaged in $Msi and $UpgradeMsi.
    [Parameter(Mandatory = $true)]
    [String] $Executable,
    # Release to upgrade from, for example 0.32.0-beta.1. Packages up to 0.31.x
    # stop every windows_exporter.exe with taskkill /T, which also ends
    # unrelated processes whose stale parent process ID matches, like the
    # runner agent of a CI job.
    [Parameter(Mandatory = $true)]
    [String] $PreviousVersion,
    [Parameter(Mandatory = $false)]
    [String] $WorkDir = $(if ($env:RUNNER_TEMP) { Join-Path $env:RUNNER_TEMP 'installer-test' } else { Join-Path $env:TEMP 'installer-test' })
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version 3

if (-not $UpgradeMsi) { $UpgradeMsi = $Msi }

$Msi = (Resolve-Path $Msi).Path
$UpgradeMsi = (Resolve-Path $UpgradeMsi).Path
$Executable = (Resolve-Path $Executable).Path

$serviceName = 'windows_exporter'
$servicePort = 9182
$foreignPort = 19182
$installDir = Join-Path $env:ProgramFiles 'windows_exporter'
$installedExe = Join-Path $installDir 'windows_exporter.exe'
$configFile = Join-Path $installDir 'config.yaml'

New-Item -ItemType Directory -Force -Path $WorkDir | Out-Null

function Write-Step([String] $Message) {
    Write-Host "::group::$Message"
}

# Some runs lose the runner during the repair or uninstall, and then no log
# is uploaded. Print what Windows Installer and Windows did around each
# operation, so passing runs show what could end or starve the runner.
function Get-RunnerProcess {
    return @(Get-Process -Name 'Runner.Listener', 'Runner.Worker' -ErrorAction SilentlyContinue |
            ForEach-Object { "$($_.Name)=$($_.Id)" }) -join ' '
}

function Write-MsiexecDiagnostic([String] $Log, [DateTime] $Since) {
    if (Test-Path $Log) {
        # The verbose log is UTF-16.
        Select-String -Path $Log -Encoding unicode -Pattern 'RESTART MANAGER', 'reboot', 'in use', 'Return value 3', 'KillProcess' |
            ForEach-Object { Write-Host "  msi: $($_.Line.Trim())" }
    }

    $events = Get-WinEvent -ErrorAction SilentlyContinue -FilterHashtable @{
        LogName   = 'System', 'Application'
        StartTime = $Since
    } | Where-Object {
        $_.Level -le 3 -or $_.ProviderName -in 'Microsoft-Windows-RestartManager', 'Service Control Manager', 'MsiInstaller'
    }

    foreach ($entry in (@($events) | Sort-Object TimeCreated)) {
        $message = ($entry.Message -split "`r?`n")[0]
        Write-Host "  event: $($entry.TimeCreated.ToString('HH:mm:ss.fff')) $($entry.LogName)/$($entry.ProviderName)/$($entry.Id): $message"
    }
}

function Invoke-Msiexec([String] $Name, [String[]] $Arguments) {
    $log = Join-Path $WorkDir "$Name.log"
    $allArguments = $Arguments + @('/qn', '/norestart', '/L*v', "`"$log`"")
    Write-Host "msiexec $($allArguments -join ' ')"
    Write-Host "  runner before: $(Get-RunnerProcess)"

    $since = Get-Date
    $process = Start-Process -FilePath 'msiexec.exe' -ArgumentList $allArguments -Wait -PassThru
    Write-Host "  msiexec $Name took $([Math]::Round(((Get-Date) - $since).TotalSeconds, 1)) s, exit code $($process.ExitCode)"
    Write-Host "  runner after: $(Get-RunnerProcess)"
    Write-MsiexecDiagnostic -Log $log -Since $since

    # The package disables the Restart Manager, so no operation may shut down
    # applications that hold its files.
    if (Test-Path $log) {
        $shutdown = Select-String -Path $log -Encoding unicode -Pattern 'RESTART MANAGER: Successfully shut down'
        if ($null -ne $shutdown) {
            throw "msiexec $Name let the Restart Manager shut down applications, see $log"
        }
    }

    # 3010 means success, but a restart is required.
    if ($process.ExitCode -ne 0 -and $process.ExitCode -ne 3010) {
        throw "msiexec $Name failed with exit code $($process.ExitCode), see $log"
    }
}

function Wait-HttpOk([Int32] $Port, [String] $Path, [Int32] $TimeoutSeconds = 60) {
    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    while ($true) {
        try {
            $response = Invoke-WebRequest -Uri "http://127.0.0.1:$Port$Path" -UseBasicParsing -TimeoutSec 30
            if ($response.StatusCode -eq 200) {
                return [String] $response.Content
            }
        } catch {
            if ((Get-Date) -ge $deadline) {
                throw "http://127.0.0.1:$Port$Path did not answer within $TimeoutSeconds seconds: $_"
            }
        }

        Start-Sleep -Seconds 1
    }
}

function Assert-ServiceRunning {
    $deadline = (Get-Date).AddSeconds(60)
    while ($true) {
        $service = Get-Service -Name $serviceName -ErrorAction SilentlyContinue
        if ($null -ne $service -and $service.Status -eq 'Running') {
            break
        }

        if ((Get-Date) -ge $deadline) {
            throw "service $serviceName is not running: $(if ($service) { $service.Status } else { 'not installed' })"
        }

        Start-Sleep -Seconds 1
    }

    $metrics = Wait-HttpOk -Port $servicePort -Path '/metrics'
    if ($metrics -notmatch '(?m)^windows_exporter_build_info\{') {
        throw 'the service scrape has no windows_exporter_build_info metric'
    }

    return $metrics
}

function Assert-ServiceRemoved {
    if ($null -ne (Get-Service -Name $serviceName -ErrorAction SilentlyContinue)) {
        throw "service $serviceName still exists after uninstall"
    }

    if (Test-Path $installedExe) {
        throw "$installedExe still exists after uninstall"
    }
}

function Assert-InstalledExecutable {
    $expected = (Get-FileHash -Algorithm SHA256 $Executable).Hash
    $actual = (Get-FileHash -Algorithm SHA256 $installedExe).Hash
    if ($expected -ne $actual) {
        throw "installed executable $installedExe ($actual) is not the packaged executable $Executable ($expected)"
    }
}

# The independent instance runs from another directory, like a second copy of
# the exporter or a HostProcess container on a Kubernetes node.
$foreignExe = Join-Path $WorkDir 'foreign\windows_exporter.exe'
New-Item -ItemType Directory -Force -Path (Split-Path $foreignExe) | Out-Null
Copy-Item -Force $Executable $foreignExe
$script:foreign = $null

function Start-Foreign {
    $script:foreign = Start-Process -FilePath $foreignExe -PassThru `
        -ArgumentList "--web.listen-address=127.0.0.1:$foreignPort", '--collectors.enabled=os' `
        -RedirectStandardOutput (Join-Path $WorkDir 'foreign.stdout.log') `
        -RedirectStandardError (Join-Path $WorkDir 'foreign.stderr.log')
    Wait-HttpOk -Port $foreignPort -Path '/health' | Out-Null
    Write-Host "independent windows_exporter.exe runs as PID $($script:foreign.Id)"
}

function Test-ForeignAlive {
    return ($null -ne $script:foreign) -and -not $script:foreign.HasExited
}

function Assert-ForeignAlive([String] $Operation) {
    if (-not (Test-ForeignAlive)) {
        throw "$Operation stopped the independent windows_exporter.exe (PID $($script:foreign.Id))"
    }

    Wait-HttpOk -Port $foreignPort -Path '/health' -TimeoutSeconds 10 | Out-Null
}

try {
    Write-Step "Install the previous release $PreviousVersion"
    $previousMsi = Join-Path $WorkDir "windows_exporter-$PreviousVersion-amd64.msi"
    $checksums = Join-Path $WorkDir "sha256sums-$PreviousVersion.txt"
    $releaseUrl = "https://github.com/prometheus-community/windows_exporter/releases/download/v$PreviousVersion"
    Invoke-WebRequest -Uri "$releaseUrl/windows_exporter-$PreviousVersion-amd64.msi" -OutFile $previousMsi -UseBasicParsing
    Invoke-WebRequest -Uri "$releaseUrl/sha256sums.txt" -OutFile $checksums -UseBasicParsing

    $expectedHash = ''
    foreach ($line in Get-Content $checksums) {
        if ($line -match "^([0-9a-f]{64})\s+windows_exporter-$([Regex]::Escape($PreviousVersion))-amd64\.msi$") {
            $expectedHash = $Matches[1]
        }
    }

    if (-not $expectedHash) {
        throw "sha256sums.txt of $PreviousVersion has no entry for the amd64 MSI"
    }

    $actualHash = (Get-FileHash -Algorithm SHA256 $previousMsi).Hash.ToLowerInvariant()
    if ($actualHash -ne $expectedHash) {
        throw "checksum mismatch for $previousMsi`: $actualHash, expected $expectedHash"
    }

    Invoke-Msiexec -Name 'install-previous' -Arguments @('/i', "`"$previousMsi`"")
    Assert-ServiceRunning | Out-Null
    Write-Host '::endgroup::'

    Write-Step 'Configure the previous release'
    # A configuration that differs from the defaults shows whether the upgrade keeps it.
    Set-Content -Path $configFile -Value "collectors:`n  enabled: os,service`n" -Encoding ascii
    Restart-Service -Name $serviceName
    Assert-ServiceRunning | Out-Null
    Write-Host '::endgroup::'

    Write-Step 'Upgrade to this build'
    Start-Foreign
    Invoke-Msiexec -Name 'upgrade' -Arguments @('/i', "`"$UpgradeMsi`"")
    $metrics = Assert-ServiceRunning
    Assert-InstalledExecutable

    if (-not (Test-Path $configFile) -or (Get-Content -Raw $configFile) -notmatch 'enabled: os,service') {
        throw "the upgrade did not keep $configFile"
    }

    if ($metrics -notmatch '(?m)^windows_exporter_collector_success\{collector="os"\} 1' -or
        $metrics -match '(?m)^windows_exporter_collector_success\{collector="cpu"\}') {
        throw 'the upgraded service does not use the kept configuration'
    }

    # The previous package removes itself with its own logic, so it must be
    # 0.32.0-beta.0 or newer to leave the independent instance running.
    Assert-ForeignAlive -Operation 'upgrade'
    Write-Host '::endgroup::'

    Write-Step 'Repair this build'
    Invoke-Msiexec -Name 'repair' -Arguments @('/fa', "`"$UpgradeMsi`"")
    Assert-ServiceRunning | Out-Null
    Assert-InstalledExecutable
    Assert-ForeignAlive -Operation 'repair'
    Write-Host '::endgroup::'

    Write-Step 'Uninstall this build'
    Invoke-Msiexec -Name 'uninstall' -Arguments @('/x', "`"$UpgradeMsi`"")
    Assert-ServiceRemoved
    Assert-ForeignAlive -Operation 'uninstall'
    Write-Host '::endgroup::'

    Write-Step 'Fresh install of this build'
    Invoke-Msiexec -Name 'install' -Arguments @('/i', "`"$Msi`"", "LISTEN_PORT=$servicePort")
    Assert-ServiceRunning | Out-Null
    Assert-InstalledExecutable
    Assert-ForeignAlive -Operation 'install'
    Write-Host '::endgroup::'

    Write-Step 'Uninstall the fresh install'
    Invoke-Msiexec -Name 'uninstall-fresh' -Arguments @('/x', "`"$Msi`"")
    Assert-ServiceRemoved
    Assert-ForeignAlive -Operation 'uninstall of the fresh install'
    Write-Host '::endgroup::'

    Write-Host 'installer test passed'
} finally {
    if (Test-ForeignAlive) {
        Stop-Process -Id $script:foreign.Id -Force
    }
}
