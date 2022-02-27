$ErrorActionPreference = 'Stop'
Set-StrictMode -Version 3

if (-not (Test-Path -Path '.\windows_exporter.exe')) {
    Write-Output ".\windows_exporter.exe not found. Consider running \`go build\` first"
}

# Powershell pipes & Get-Content command rather unhelpfully add a carriage return at the end of the string, so
# passing the string as bytes is a messy but necessary workaround for processes that are sensitive to
# line endings, like promtool.
function Start-RawProcess {
    param(
        # String to pass to $CommandName via STDIN
        [Parameter(Mandatory=$true)][String]$InputVar,
        # Command to run
        [Parameter(Mandatory=$true)][String]$CommandName,
        # Arguments provided to $CommandName
        [Parameter(Mandatory=$false)][String[]]$CommandArgs
    )
    # Buffer & initial size of MemoryStream
    $BufferSize = 4096

    # Convert text to bytes and write to MemoryStream
    [byte[]]$InputBytes = [Text.Encoding]::UTF8.GetBytes($InputVar)
    $MemStream = New-Object -TypeName System.IO.MemoryStream -ArgumentList $BufferSize
    $MemStream.Write($InputBytes, 0, $InputBytes.Length)
    [Void]$MemStream.Seek(0, 'Begin')

    # Setup stdin\stdout redirection for our process
    if ($CommandArgs) {
        $StartInfo = New-Object -TypeName System.Diagnostics.ProcessStartInfo -Property @{
            FileName = $CommandName
            UseShellExecute = $false
            RedirectStandardInput = $true
            RedirectStandardError = $true
            Arguments = $CommandArgs
        }
    } else {
        $StartInfo = New-Object -TypeName System.Diagnostics.ProcessStartInfo -Property @{
            FileName = $CommandName
            UseShellExecute = $false
            RedirectStandardInput = $true
            RedirectStandardError = $true
        }
    }

    # Create new process
    $Process = New-Object -TypeName System.Diagnostics.Process

    # Assign previously created StartInfo properties
    $Process.StartInfo = $StartInfo
    # Start process
    [void]$Process.Start()

    # Pipe data
    $Buffer = New-Object -TypeName byte[] -ArgumentList $BufferSize
    $StdinStream = $Process.StandardInput.BaseStream

    try {
        do {
            $ReadCount = $MemStream.Read($Buffer, 0, $Buffer.Length)
            $StdinStream.Write($Buffer, 0, $ReadCount)
            $StdinStream.Flush()
        }
        while($ReadCount -gt 0)
    }
    catch
    {
       throw 'Error streaming buffer to STDIN'
    } finally {
        # Close streams
        $StdinStream.Close()
        $MemStream.Close()
    }
    $Process.WaitForExit()
    if ($Process.ExitCode -ne 0) {
        Write-Host $Process.StandardError.ReadToEnd()
    }

    return $Process.ExitCode
}

# cd to location of script
$script_path = $MyInvocation.MyCommand.Path
$working_dir = Split-Path $script_path
Push-Location $working_dir

$temp_dir = Join-Path $env:TEMP $(New-Guid) | ForEach-Object { mkdir $_ }

# Start process in background, awaiting HTTP requests.
# Listen on 9183/TCP, preventing conflicts with 9182/TCP used by end-to-end-test.ps1
# Not an issue when run individually, but will cause failures when run concurrently in CI.
$exporter_proc = Start-Process `
    -PassThru `
    -FilePath ..\windows_exporter.exe `
    -ArgumentList '--telemetry.addr="127.0.0.1:9183" --log.level=debug' `
    -WindowStyle Hidden `
    -RedirectStandardOutput "$($temp_dir)/windows_exporter.log" `
    -RedirectStandardError "$($temp_dir)/windows_exporter_error.log"

# Exporter can take some time to start
for ($i=1; $i -le 5; $i++) {
    Start-Sleep 10

    $netstat_output = netstat -anp tcp | Select-String 'listening'
    if ($netstat_output -like '*:9183*') {
            break
    }
    Write-Host "Waiting for exporter to start"
}

# Omit metrics from client_golang library; we're not responsible for these
$skip_re = "^[#]?\s*(HELP|TYPE)?\s*go_"

# Need to remove carriage returns, as promtool expects LF line endings
$output = ((Invoke-WebRequest -UseBasicParsing -URI http://127.0.0.1:9183/metrics).Content) -Split "`r?`n" | Select-String -NotMatch $skip_re | Join-String -Separator "`n"
# Join the split lines back to a single String (with LF line endings!)
$output = $output -Join "`n"
Stop-Process -Id $exporter_proc.Id
$ExitCode = Start-RawProcess -InputVar $output -CommandName promtool.exe -CommandArgs @("check metrics")
if ($ExitCode -ne 0) {
    Write-Host "Promtool command returned exit code $($ExitCode). See output for details."
    EXIT 1
}
