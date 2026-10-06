# A disposable domain guest keeps promotion and reboots off the Actions host.
$ErrorActionPreference = "Stop"
Import-Module Hyper-V
Import-Module Dism
Start-Service vmms
$labDir = Join-Path $env:RUNNER_TEMP "domain-lab"
New-Item -ItemType Directory -Force -Path $labDir | Out-Null
$clock = [Diagnostics.Stopwatch]::StartNew()
$timings = [ordered]@{}
$lastMilestone = 0
$session = $null
$password = "CI-" + [Guid]::NewGuid().ToString('N') + "!a"
Write-Host "::add-mask::$password"
Start-Transcript -Path (Join-Path $labDir "domain-lab.log")
$securePassword = ConvertTo-SecureString $password -AsPlainText -Force
$credential = [PSCredential]::new('CIADDC\Administrator', $securePassword)

function Record-LabTime([string]$Phase) {
    $elapsed = $clock.Elapsed.TotalSeconds
    $timings[$Phase] = [math]::Round($elapsed - $script:lastMilestone, 1)
    $script:lastMilestone = $elapsed
    $timings | ConvertTo-Json | Set-Content (Join-Path $labDir "timings.json")
    Write-Host "$Phase completed in $($timings[$Phase]) seconds"
}

function Connect-LabGuest([PSCredential]$GuestCredential, [datetime]$AfterBoot = [datetime]::MinValue) {
    $deadline = (Get-Date).AddMinutes(5)
    do {
        $candidate = $null
        try {
            $candidate = New-PSSession -VMName CIADDC -Credential $GuestCredential -ErrorAction Stop
            $bootTime = Invoke-Command -Session $candidate -ScriptBlock {
                (Get-CimInstance Win32_OperatingSystem).LastBootUpTime
            }
            if ($bootTime -gt $AfterBoot) { return $candidate }
        } catch {
            Write-Host "Waiting for PowerShell Direct: $($_.Exception.Message)"
        }
        if ($candidate) { Remove-PSSession $candidate -ErrorAction SilentlyContinue }
        Start-Sleep -Seconds 5
    } while ((Get-Date) -lt $deadline)
    throw "Domain guest did not become ready through PowerShell Direct"
}

try {
    # Versioned Microsoft Evaluation Center media; no cloud account is needed.
    $iso = Join-Path $labDir "windows-server-2025.iso"
    $isoUri = "https://software-static.download.prss.microsoft.com/dbazure/998969d5-f34g-4e03-ac9d-1f9786c66749/26100.32230.260111-0550.lt_release_svc_refresh_SERVER_EVAL_x64FRE_en-us.iso"
    curl.exe --fail --location --retry 3 --silent --show-error --output $iso $isoUri
    if ($LASTEXITCODE -ne 0) { throw "Evaluation ISO download failed" }
    $isoHash = (Get-FileHash $iso -Algorithm SHA256).Hash
    Write-Host "Evaluation ISO SHA256: $isoHash"
    $expectedHash = "7b052573ba7894c9924e3e87ba732ccd354d18cb75a883efa9b900ea125bfd51"
    if ($isoHash -ne $expectedHash) {
        throw "Evaluation ISO SHA256 mismatch"
    }
    Record-LabTime "Download ISO"

    $isoMount = Mount-DiskImage -ImagePath $iso -PassThru
    $isoVolume = $isoMount | Get-Volume
    $wim = "$($isoVolume.DriveLetter):\sources\install.wim"
    $images = Get-WindowsImage -ImagePath $wim
    $images | Format-Table ImageIndex, ImageName
    $image = $images | Where-Object { $_.ImageName -match 'Standard' -and $_.ImageName -notmatch 'Desktop' } |
        Select-Object -First 1
    if (-not $image) { throw "Server Core evaluation image was not found" }

    $vhd = Join-Path $labDir "domain-guest.vhdx"
    New-VHD -Path $vhd -Dynamic -SizeBytes 40GB -LogicalSectorSizeBytes 512 -PhysicalSectorSizeBytes 4096 | Out-Null
    $disk = Mount-VHD $vhd -PassThru | Get-Disk | Initialize-Disk -PartitionStyle GPT -PassThru
    $efi = $disk | New-Partition -Size 300MB -GptType '{c12a7328-f81f-11d2-ba4b-00a0c93ec93b}' -AssignDriveLetter |
        Format-Volume -FileSystem FAT32 -NewFileSystemLabel CILabEFI -Confirm:$false
    $disk | New-Partition -Size 16MB -GptType '{e3c9e316-0b5c-4db8-817d-f92df00215ae}' | Out-Null
    $os = $disk | New-Partition -UseMaximumSize -AssignDriveLetter |
        Format-Volume -FileSystem NTFS -NewFileSystemLabel CILabOS -Confirm:$false
    $osRoot = "$($os.DriveLetter):\"
    Expand-WindowsImage -ImagePath $wim -Index $image.ImageIndex -ApplyPath $osRoot `
        -LogPath (Join-Path $labDir "apply-image.log") | Out-Null

    $unattend = Join-Path $labDir "unattend.xml"
    @"
<?xml version="1.0" encoding="utf-8"?>
<unattend xmlns="urn:schemas-microsoft-com:unattend">
  <settings pass="specialize">
    <component name="Microsoft-Windows-Shell-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS">
      <ComputerName>CIADDC</ComputerName><TimeZone>UTC</TimeZone>
    </component>
  </settings>
  <settings pass="oobeSystem">
    <component name="Microsoft-Windows-International-Core" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS">
      <InputLocale>en-US</InputLocale><SystemLocale>en-US</SystemLocale><UILanguage>en-US</UILanguage><UserLocale>en-US</UserLocale>
    </component>
    <component name="Microsoft-Windows-Shell-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS">
      <UserAccounts><AdministratorPassword><Value>$password</Value><PlainText>true</PlainText></AdministratorPassword></UserAccounts>
      <AutoLogon><Enabled>true</Enabled><LogonCount>1</LogonCount><Username>Administrator</Username><Domain>CIADDC</Domain><Password><Value>$password</Value><PlainText>true</PlainText></Password></AutoLogon>
      <OOBE><HideEULAPage>true</HideEULAPage><HideLocalAccountScreen>true</HideLocalAccountScreen><HideOnlineAccountScreens>true</HideOnlineAccountScreens><HideWirelessSetupInOOBE>true</HideWirelessSetupInOOBE><ProtectYourPC>3</ProtectYourPC></OOBE>
    </component>
  </settings>
</unattend>
"@ | Set-Content $unattend -Encoding utf8
    New-Item -ItemType Directory -Force -Path "$osRoot\Windows\Panther" | Out-Null
    Copy-Item $unattend "$osRoot\Windows\Panther\Unattend.xml"
    bcdboot.exe "$osRoot\Windows" /s "$($efi.DriveLetter):" /f UEFI
    if ($LASTEXITCODE -ne 0) { throw "Guest boot files could not be created" }
    $bcd = "$($efi.DriveLetter):\EFI\Microsoft\Boot\BCD"
    foreach ($entry in "device", "osdevice") {
        bcdedit.exe /store $bcd /set '{default}' $entry "partition=$($os.DriveLetter):"
        if ($LASTEXITCODE -ne 0) { throw "Guest boot partition could not be configured" }
    }
    reg.exe load HKLM\CILabSystem "$osRoot\Windows\System32\Config\SYSTEM"
    if ($LASTEXITCODE -ne 0) { throw "Guest registry could not be opened" }
    try {
        reg.exe add HKLM\CILabSystem\ControlSet001\Control\CrashControl /v AutoReboot /t REG_DWORD /d 0 /f
        if ($LASTEXITCODE -ne 0) { throw "Guest crash diagnostics could not be configured" }
    } finally { reg.exe unload HKLM\CILabSystem }
    Dismount-VHD $vhd
    Dismount-DiskImage -ImagePath $iso
    Remove-Item $unattend
    Record-LabTime "Apply Server Core image"

    New-VMSwitch -Name CILabPrivate -SwitchType Private | Out-Null
    New-VM -Name CIADDC -Generation 2 -MemoryStartupBytes 4GB -VHDPath $vhd -SwitchName CILabPrivate | Out-Null
    Set-VMProcessor -VMName CIADDC -Count 2
    Set-VM -Name CIADDC -AutomaticStopAction TurnOff
    Set-VMFirmware -VMName CIADDC -FirstBootDevice (Get-VMHardDiskDrive -VMName CIADDC)
    Start-VM CIADDC
    $session = Connect-LabGuest $credential
    Record-LabTime "Guest first boot"

    $bootBeforePromotion = Invoke-Command -Session $session -ArgumentList $securePassword -ScriptBlock {
        param($RestorePassword)
        $ErrorActionPreference = "Stop"
        $adapter = Get-NetAdapter | Select-Object -First 1
        New-NetIPAddress -InterfaceIndex $adapter.ifIndex -IPAddress 192.0.2.10 -PrefixLength 24 | Out-Null
        Set-DnsClientServerAddress -InterfaceIndex $adapter.ifIndex -ServerAddresses 127.0.0.1
        $roles = Install-WindowsFeature AD-Domain-Services,ADCS-Cert-Authority -IncludeManagementTools
        Write-Host ($roles | Format-List | Out-String)
        if (-not $roles.Success) { throw "Domain roles could not be installed" }
        Install-ADDSForest -DomainName ci.windows-exporter.test -DomainNetbiosName CILAB `
            -InstallDns -SafeModeAdministratorPassword $RestorePassword -NoRebootOnCompletion -Force | Out-Null
        (Get-CimInstance Win32_OperatingSystem).LastBootUpTime
    }
    Invoke-Command -Session $session -ScriptBlock { shutdown.exe /r /t 5 /f }
    Remove-PSSession $session
    $session = $null
    $credential = [PSCredential]::new('CILAB\Administrator', $securePassword)
    $session = Connect-LabGuest $credential $bootBeforePromotion
    Record-LabTime "Domain promotion and guest reboot"

    Invoke-Command -Session $session -ScriptBlock {
        $ErrorActionPreference = "Stop"
        $deadline = (Get-Date).AddMinutes(3)
        do {
            try { Get-ADDomain -ErrorAction Stop | Format-List DNSRoot; break }
            catch { if ((Get-Date) -ge $deadline) { throw }; Start-Sleep -Seconds 5 }
        } while ($true)
        New-ADUser -Name CIFixture -SamAccountName CIFixture
        Get-ADUser CIFixture | Format-Table Name, DistinguishedName
        Install-AdcsCertificationAuthority -CAType EnterpriseRootCA -CACommonName CIEnterpriseCA `
            -CryptoProviderName "RSA#Microsoft Software Key Storage Provider" -KeyLength 2048 `
            -HashAlgorithmName SHA256 -ValidityPeriod Years -ValidityPeriodUnits 1 -Force | Out-Null
        certreq.exe -enroll -machine -q -config "$env:COMPUTERNAME\CIEnterpriseCA" Machine
        if ($LASTEXITCODE -ne 0) { throw "Computer template certificate enrollment failed" }
        winmgmt.exe /resyncperf
    }
    Record-LabTime "Enterprise CA and template enrollment"

    Invoke-Command -Session $session -ScriptBlock { New-Item -ItemType Directory -Force C:\lab\logs | Out-Null }
    Copy-Item domain-tests\* -Destination C:\lab -Recurse -ToSession $session
    "1" | Set-Content domain-test-exit-code.txt
    $testExitCode = 0
    foreach ($collector in "ad", "adcs") {
        $result = Invoke-Command -Session $session -ArgumentList $collector -ScriptBlock {
            param($Name)
            $env:WINDOWS_EXPORTER_TEST_COLLECTORS = "ad,adcs"
            $env:WINDOWS_EXPORTER_TEST_ADCS_TEMPLATE = "Machine"
            $output = & "C:\lab\$Name.test.exe" -test.v=test2json -test.timeout=5m 2>&1
            [PSCustomObject]@{ ExitCode = $LASTEXITCODE; Output = ($output -join "`n") }
        }
        $result.Output | go tool test2json -t -p "github.com/prometheus-community/windows_exporter/internal/collector/$collector" |
            Add-Content test-results.jsonl
        Write-Host $result.Output
        if ($result.ExitCode -ne 0) { $testExitCode = 1 }
    }
    "$testExitCode" | Set-Content domain-test-exit-code.txt
    Record-LabTime "Go race tests"
    if ($testExitCode -ne 0) { throw "Domain collector Go tests failed" }

    Invoke-Command -Session $session -ScriptBlock {
        $ErrorActionPreference = "Stop"
        Set-Location C:\lab
        $env:RUNNER_TEMP = "C:\lab\logs"
        .\tools\end-to-end-test.ps1
    }
    Record-LabTime "Exporter metrics smoke request"
} finally {
    if ($session) {
        Copy-Item C:\lab\logs\* -FromSession $session -Destination $labDir -ErrorAction SilentlyContinue
        Remove-PSSession $session -ErrorAction SilentlyContinue
    }
    $vm = Get-VM -Name CIADDC -ErrorAction SilentlyContinue
    if ($vm -and -not $session) {
        $vm | Format-List Name, State, Status, Uptime
        Get-VMIntegrationService -VMName CIADDC | Format-Table Name, Enabled, PrimaryStatusDescription
        try {
            $namespace = 'root/virtualization/v2'
            $settings = Get-CimInstance -Namespace $namespace -ClassName Msvm_VirtualSystemSettingData |
                Where-Object { $_.ElementName -eq 'CIADDC' -and $_.VirtualSystemType -eq 'Microsoft:Hyper-V:System:Realized' }
            $service = Get-CimInstance -Namespace $namespace -ClassName Msvm_VirtualSystemManagementService
            $thumbnail = Invoke-CimMethod -InputObject $service -MethodName GetVirtualSystemThumbnailImage -Arguments @{
                TargetSystem = $settings; WidthPixels = [uint16]1024; HeightPixels = [uint16]768
            }
            if ($thumbnail.ReturnValue -eq 0) {
                Add-Type -AssemblyName System.Drawing
                $rectangle = [Drawing.Rectangle]::new(0, 0, 1024, 768)
                $bitmap = [Drawing.Bitmap]::new(1024, 768, [Drawing.Imaging.PixelFormat]::Format16bppRgb565)
                try {
                    $bits = $bitmap.LockBits($rectangle, [Drawing.Imaging.ImageLockMode]::WriteOnly, $bitmap.PixelFormat)
                    [byte[]]$pixels = $thumbnail.ImageData
                    if ($pixels.Length -ne [math]::Abs($bits.Stride) * $bits.Height) {
                        throw "Guest thumbnail size does not match its bitmap buffer"
                    }
                    [Runtime.InteropServices.Marshal]::Copy($pixels, 0, $bits.Scan0, $pixels.Length)
                    $bitmap.UnlockBits($bits)
                    $bitmap.Save((Join-Path $labDir "guest-console.png"), [Drawing.Imaging.ImageFormat]::Png)
                } finally { $bitmap.Dispose() }
            }
        } catch { Write-Warning "Guest console capture failed: $($_.Exception.Message)" }
    }
    Get-VM -Name CIADDC -ErrorAction SilentlyContinue |
        Stop-VM -TurnOff -Force -ErrorAction SilentlyContinue
    if ($vm -and -not $session) {
        try {
            $guestDisk = Mount-VHD $vhd -ReadOnly -PassThru | Get-Disk
            $guestVolume = $guestDisk | Get-Partition | Get-Volume | Where-Object FileSystemLabel -eq CILabOS
            $panther = "$($guestVolume.DriveLetter):\Windows\Panther"
            foreach ($log in Get-ChildItem $panther -Filter *.log -Recurse -ErrorAction SilentlyContinue) {
                $safeLog = (Get-Content $log.FullName -Raw).Replace($password, '***')
                $safeLog | Set-Content (Join-Path $labDir "guest-$($log.Directory.Name)-$($log.Name)")
            }
            Dismount-VHD $vhd
        } catch { Write-Warning "Guest setup log collection failed: $($_.Exception.Message)" }
    }
    Stop-Transcript
}
