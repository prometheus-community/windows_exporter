# Background CI steps attach virtual disks at the same time, which makes
# Mount-VHD fail with 0x800703E3. Serialize attaching and formatting them.
function Invoke-WithVhdLock([scriptblock]$ScriptBlock) {
    $mutex = [System.Threading.Mutex]::new($false, "Global\windows-exporter-ci-vhd")
    try {
        try {
            [void]$mutex.WaitOne()
        } catch [System.Threading.AbandonedMutexException] {
            # The previous owner exited without releasing; the lock is ours now.
        }

        try {
            & $ScriptBlock
        } finally {
            $mutex.ReleaseMutex()
        }
    } finally {
        $mutex.Dispose()
    }
}

# New-Partition can return before the storage provider exposes the new volume,
# which makes Format-Volume fail with CmdletizationQuery_NotFound. Retry until
# the volume appears.
function Format-NewDisk($Disk, [string]$Label) {
    $partition = $Disk | Initialize-Disk -PartitionStyle GPT -PassThru |
        New-Partition -UseMaximumSize -AssignDriveLetter
    $deadline = (Get-Date).AddSeconds(30)
    while ($true) {
        try {
            return $partition | Format-Volume -FileSystem NTFS -NewFileSystemLabel $Label -Confirm:$false
        } catch {
            if ($_.FullyQualifiedErrorId -notlike "CmdletizationQuery_NotFound*" -or (Get-Date) -ge $deadline) {
                throw
            }
            Start-Sleep -Seconds 1
        }
    }
}
