# Compares the WiX-built windows_exporter MSI with the msi-builder one on a
# disposable Windows runner: installs both with the same properties and
# records the service, firewall, registry, ARP and file state each leaves.
# Throwaway experiment for the msi-builder real-world check; runner only.

[CmdletBinding()]
Param (
    [Parameter(Mandatory = $true)] [String] $Dir,
    [Parameter(Mandatory = $true)] [String] $Out
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version 3
New-Item -ItemType Directory -Force -Path $Out | Out-Null

$installDir = Join-Path $env:ProgramFiles 'windows_exporter'
$summary = [System.Collections.Generic.List[String]]::new()

function Msi([String] $Kind, [String] $Version) {
    return Join-Path $Dir "$Kind-$Version.msi"
}

function Invoke-Msi([String] $Name, [String[]] $Arguments) {
    $log = Join-Path $Out "$Name.log"
    $all = $Arguments + @('/qn', '/norestart', '/L*v', "`"$log`"")
    $since = Get-Date
    $p = Start-Process -FilePath msiexec.exe -ArgumentList $all -Wait -PassThru
    $line = "{0,-40} exit {1,5}  {2,5:N1}s  msiexec {3}" -f $Name, $p.ExitCode, ((Get-Date) - $since).TotalSeconds, ($Arguments -join ' ')
    Write-Host $line
    $summary.Add($line)
    return $p.ExitCode
}

function Get-ArpKey {
    $root = 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall'
    return @(Get-ChildItem $root | Where-Object { (Get-ItemProperty $_.PSPath).PSObject.Properties['DisplayName'] -and (Get-ItemProperty $_.PSPath).DisplayName -eq 'windows_exporter' })
}

function Remove-Installed([String] $Name) {
    foreach ($key in Get-ArpKey) {
        Invoke-Msi -Name "$Name-cleanup" -Arguments @('/x', $key.PSChildName) | Out-Null
    }
    Get-Process -Name windows_exporter -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
    if (Test-Path $installDir) { Remove-Item -Recurse -Force $installDir -ErrorAction SilentlyContinue }
    & netsh.exe advfirewall firewall delete rule name=windows_exporter | Out-Null
}

function Format-Props($Object, [String[]] $Skip = @()) {
    if ($null -eq $Object) { return @('  <none>') }
    return @($Object.PSObject.Properties | Where-Object { $_.Name -notlike 'PS*' -and $_.Name -notin $Skip } |
            Sort-Object Name | ForEach-Object { "  $($_.Name) = $(@($_.Value) -join ';')" })
}

function Save-State([String] $Name) {
    Start-Sleep -Seconds 3
    $lines = [System.Collections.Generic.List[String]]::new()

    $lines.Add('[service]')
    $svc = Get-CimInstance Win32_Service -Filter "Name='windows_exporter'"
    if ($svc) {
        foreach ($p in 'DisplayName', 'Description', 'PathName', 'StartMode', 'DelayedAutoStart', 'StartName', 'State', 'ErrorControl', 'ServiceType') {
            $lines.Add("  $p = $($svc.$p)")
        }
        foreach ($q in 'qc', 'qfailure', 'qfailureflag', 'qsidtype', 'qtriggerinfo', 'qprivs') {
            $lines.Add("  sc ${q}:")
            & sc.exe $q windows_exporter | Where-Object { $_.Trim() } | ForEach-Object { $lines.Add("    $($_.Trim())") }
        }
        try {
            $r = Invoke-WebRequest -UseBasicParsing -TimeoutSec 20 -Uri 'http://127.0.0.1:9182/health'
            $lines.Add("  health 9182 = $($r.StatusCode)")
        } catch { $lines.Add('  health 9182 = unreachable') }
    } else { $lines.Add('  <none>') }

    $lines.Add('[eventlog source]')
    $ev = Get-ItemProperty 'HKLM:\SYSTEM\CurrentControlSet\Services\EventLog\Application\windows_exporter' -ErrorAction SilentlyContinue
    $lines.AddRange([String[]] (Format-Props $ev))

    $lines.Add('[firewall]')
    $rules = @(Get-NetFirewallRule -ErrorAction SilentlyContinue | Where-Object { $_.DisplayName -like '*windows_exporter*' })
    if (-not $rules) { $lines.Add('  <none>') }
    foreach ($rule in $rules) {
        $lines.Add("  rule:")
        foreach ($p in 'DisplayName', 'Description', 'Direction', 'Action', 'Enabled', 'Profile', 'Group', 'EdgeTraversalPolicy') { $lines.Add("    $p = $($rule.$p)") }
        $pf = $rule | Get-NetFirewallPortFilter
        $af = $rule | Get-NetFirewallAddressFilter
        $ap = $rule | Get-NetFirewallApplicationFilter
        $lines.Add("    Protocol = $($pf.Protocol)  LocalPort = $($pf.LocalPort -join ',')  RemotePort = $($pf.RemotePort -join ',')")
        $lines.Add("    RemoteAddress = $($af.RemoteAddress -join ',')  LocalAddress = $($af.LocalAddress -join ',')")
        $lines.Add("    Program = $($ap.Program)")
    }

    $lines.Add('[arp]')
    foreach ($key in Get-ArpKey) {
        $lines.AddRange([String[]] (Format-Props (Get-ItemProperty $key.PSPath) -Skip @('InstallDate', 'InstallSource', 'ModifyPath', 'UninstallString', 'LocalPackage')))
    }

    $lines.Add('[files]')
    if (Test-Path $installDir) {
        Get-ChildItem -Recurse -Force $installDir | Sort-Object FullName | ForEach-Object {
            $rel = $_.FullName.Substring($installDir.Length)
            if ($_.PSIsContainer) { $lines.Add("  $rel\") } else { $lines.Add("  $rel ($($_.Length) bytes)") }
        }
        $acl = (Get-Acl $installDir).Sddl
        $lines.Add("  sddl = $acl")
    } else { $lines.Add('  <no install dir>') }

    $lines | Set-Content -Encoding utf8 (Join-Path $Out "$Name.state.txt")
}

function Compare-States([String] $Scenario) {
    $refFile = Join-Path $Out "ref-$Scenario.state.txt"
    $oursFile = Join-Path $Out "ours-$Scenario.state.txt"
    $diff = & git.exe diff --no-index --no-color -U0 $refFile $oursFile
    if ($diff) {
        $summary.Add("DIFF $Scenario")
        $diff | Select-Object -Skip 4 | ForEach-Object { $summary.Add("    $_") }
    } else { $summary.Add("SAME $Scenario") }
}

try {
    foreach ($kind in 'ref', 'ours') {
        Remove-Installed "$kind-pre"

        # Defaults.
        Invoke-Msi "$kind-default-install" @('/i', (Msi $kind '0.90.0')) | Out-Null
        Save-State "$kind-default"
        Invoke-Msi "$kind-default-uninstall" @('/x', (Msi $kind '0.90.0')) | Out-Null
        Save-State "$kind-default-removed"
        Remove-Installed "$kind-default"

        # Every property, firewall feature on.
        $props = @('ADDLOCAL=ALL', 'LISTEN_PORT=9182', 'ENABLED_COLLECTORS=os,cpu,service', 'METRICS_PATH=/metrics',
            'TEXTFILE_DIRS=C:\tf', 'REMOTE_ADDR=10.0.0.0/8', 'EXTRA_FLAGS=--log.level=debug')
        Invoke-Msi "$kind-custom-install" (@('/i', (Msi $kind '0.90.0')) + $props) | Out-Null
        Save-State "$kind-custom"
        Invoke-Msi "$kind-custom-repair" @('/fa', (Msi $kind '0.90.0')) | Out-Null
        Save-State "$kind-custom-repair"
        # Rerunning the package without properties reinstalls everything.
        Invoke-Msi "$kind-custom-rerun" @('/i', (Msi $kind '0.90.0')) | Out-Null
        Save-State "$kind-custom-rerun"
        Invoke-Msi "$kind-custom-uninstall" @('/x', (Msi $kind '0.90.0')) | Out-Null
        Save-State "$kind-custom-removed"
        Remove-Installed "$kind-custom"

        # Listen address and port, no config file.
        Invoke-Msi "$kind-noconfig-install" @('/i', (Msi $kind '0.90.0'), 'CONFIG_FILE=', 'LISTEN_ADDR=127.0.0.1', 'LISTEN_PORT=9183', 'ADDLOCAL=ALL') | Out-Null
        Save-State "$kind-noconfig"
        Remove-Installed "$kind-noconfig"

        # A flag the exporter rejects: the service cannot start.
        Invoke-Msi "$kind-badflag-install" @('/i', (Msi $kind '0.90.0'), 'EXTRA_FLAGS=--no-such-flag') | Out-Null
        Save-State "$kind-badflag"
        Remove-Installed "$kind-badflag"

        # --config.file in EXTRA_FLAGS is refused.
        Invoke-Msi "$kind-checkextra-install" @('/i', (Msi $kind '0.90.0'), 'EXTRA_FLAGS=--config.file=C:\x.yaml') | Out-Null
        Save-State "$kind-checkextra"
        Remove-Installed "$kind-checkextra"

        # Upgrade keeps the feature selection and config.yaml; downgrade is refused.
        Invoke-Msi "$kind-v90-install" @('/i', (Msi $kind '0.90.0'), 'ADDLOCAL=ALL') | Out-Null
        Set-Content -Path (Join-Path $installDir 'config.yaml') -Value "collectors:`n  enabled: os`n" -Encoding ascii
        Invoke-Msi "$kind-v91-upgrade" @('/i', (Msi $kind '0.91.0')) | Out-Null
        Save-State "$kind-upgraded"
        Invoke-Msi "$kind-v90-downgrade" @('/i', (Msi $kind '0.90.0')) | Out-Null
        Save-State "$kind-after-downgrade"
        Remove-Installed "$kind-upgrade"
    }

    # Switching tools: WiX release -> msi-builder -> WiX again.
    Invoke-Msi 'cross-ref90-install' @('/i', (Msi 'ref' '0.90.0'), 'ADDLOCAL=ALL', 'ENABLED_COLLECTORS=os') | Out-Null
    Set-Content -Path (Join-Path $installDir 'config.yaml') -Value "collectors:`n  enabled: os`n" -Encoding ascii
    Save-State 'cross-1-ref90'
    Invoke-Msi 'cross-ours91-upgrade' @('/i', (Msi 'ours' '0.91.0')) | Out-Null
    Save-State 'cross-2-ours91'
    Invoke-Msi 'cross-ref92-upgrade' @('/i', (Msi 'ref' '0.92.0')) | Out-Null
    Save-State 'cross-3-ref92'
    Invoke-Msi 'cross-ref92-uninstall' @('/x', (Msi 'ref' '0.92.0')) | Out-Null
    Save-State 'cross-4-removed'
    Remove-Installed 'cross'

    foreach ($s in 'default', 'default-removed', 'custom', 'custom-repair', 'custom-rerun', 'custom-removed', 'noconfig', 'badflag', 'checkextra', 'upgraded', 'after-downgrade') {
        Compare-States $s
    }
} finally {
    $summary | Set-Content -Encoding utf8 (Join-Path $Out 'SUMMARY.txt')
    $summary | ForEach-Object { Write-Host $_ }
}
