$ErrorActionPreference = 'Stop'
Set-StrictMode -Version 3

if (-not (Test-Path -Path '.\windows_exporter.exe')) {
    Write-Output ".\windows_exporter.exe not found. Consider running \`go build\` first"
}

# cd to location of script
$script_path = $MyInvocation.MyCommand.Path
$working_dir = Split-Path $script_path
Push-Location $working_dir

$temp_dir = Join-Path $env:TEMP $(New-Guid) | ForEach-Object { mkdir $_ }

# Create temporary directory for textfile collector
$textfile_dir = "$($temp_dir)/textfile"
mkdir $textfile_dir | Out-Null
Copy-Item 'e2e-textfile.prom' -Destination "$($textfile_dir)/e2e-textfile.prom"

# Omit dynamic collector information that will change after each run
$skip_re = "^(go_|windows_exporter_build_info|windows_exporter_collector_duration_seconds|windows_exporter_perflib_snapshot_duration_seconds|process_|windows_textfile_mtime_seconds|windows_cpu|windows_cs|windows_logical_disk|windows_net|windows_os|windows_service|windows_system|windows_textfile_mtime_seconds)"

# Start process in background, awaiting HTTP requests.
# Use default collectors, port and address: http://localhost:9182/metrics
$exporter_proc = Start-Process `
    -PassThru `
    -FilePath ..\windows_exporter.exe `
    -ArgumentList "--log.level=debug --collector.textfile.directory=$($textfile_dir)" `
    -WindowStyle Hidden `
    -RedirectStandardOutput "$($temp_dir)/windows_exporter.log" `
    -RedirectStandardError "$($temp_dir)/windows_exporter_error.log"

# Exporter can take some time to start
for ($i=1; $i -le 5; $i++) {
    Start-Sleep 10

    $netstat_output = netstat -anp tcp | Select-String 'listening'
    if ($netstat_output -like '*:9182*') {
            break
    }
    Write-Host "Waiting for exporter to start"
}

$response = Invoke-WebRequest -UseBasicParsing -URI http://127.0.0.1:9182/metrics
# Response output must be split and saved as UTF-8.
$response.content -split "[`r`n]"| Select-String -NotMatch $skip_re | Set-Content -Encoding utf8 "$($temp_dir)/e2e-output.txt"
Stop-Process -Id $exporter_proc.Id
$output_diff = Compare-Object (Get-Content 'e2e-output.txt') (Get-Content "$($temp_dir)/e2e-output.txt")

# Fail if differences in output are detected
if (-not ($null -eq $output_diff)) {
    $output_diff | Format-Table
    exit 1
}
