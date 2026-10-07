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
