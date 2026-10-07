# Install a real Database Engine instance; LocalDB does not expose server counters.
$ErrorActionPreference = "Stop"

$mediaDir = Join-Path $env:RUNNER_TEMP "sql-server"
New-Item -ItemType Directory -Force -Path $mediaDir | Out-Null
$installer = Join-Path $mediaDir "SQL2025-SSEI-Expr.exe"
# Microsoft download endpoints occasionally reset hosted runner connections.
curl.exe --fail --location --retry 5 --retry-all-errors --silent --show-error --output $installer `
    "https://download.microsoft.com/download/ffd82b4c-9955-47c0-8efe-6290f7795cf6/SQL2025-SSEI-Expr.exe"
if ($LASTEXITCODE -ne 0) { throw "SQL Server bootstrapper download failed" }
$expectedHash = "fa7e1fabc9a2e9c9cdab0d1512bcb30d2949133147db057ac530f510e5270680"
if ((Get-FileHash $installer -Algorithm SHA256).Hash -ne $expectedHash) {
    throw "SQL Server bootstrapper SHA256 mismatch"
}

$media = Join-Path $mediaDir "SQLEXPR_x64_ENU.exe"
foreach ($attempt in 1..3) {
    $download = Start-Process $installer -Wait -PassThru `
        -ArgumentList "/ACTION=Download", "/MEDIATYPE=Core", "/QUIET", "/MEDIAPATH=$mediaDir"
    if ($download.ExitCode -eq 0 -and (Test-Path $media)) { break }
    if ($attempt -eq 3) { throw "SQL Server media download failed: $($download.ExitCode)" }
    Write-Warning "SQL Server media download attempt $attempt failed: $($download.ExitCode)"
    Start-Sleep -Seconds 10
}

# Azure runners can expose sectors larger than SQL Server's supported 4 KB.
# Keep every instance file on a disk with a known, compatible sector size.
Import-Module Hyper-V
$diskPath = Join-Path $env:RUNNER_TEMP "sql-server.vhdx"
New-VHD -Path $diskPath -Dynamic -SizeBytes 8GB `
    -LogicalSectorSizeBytes 512 -PhysicalSectorSizeBytes 4096 | Out-Null
$disk = Mount-VHD -Path $diskPath -PassThru | Get-Disk
if ($disk.PhysicalSectorSize -gt 4096) {
    throw "SQL Server fixture disk exposes unsupported physical sectors"
}
$volume = $disk | Initialize-Disk -PartitionStyle GPT -PassThru |
    New-Partition -UseMaximumSize -AssignDriveLetter |
    Format-Volume -FileSystem NTFS -NewFileSystemLabel CISQL -Confirm:$false
$instanceDir = "$($volume.DriveLetter):\SQLServer"
fsutil.exe fsinfo sectorinfo "$($volume.DriveLetter):"

$setupDir = Join-Path $mediaDir "setup"
$extract = Start-Process $media -Wait -PassThru `
    -ArgumentList "/Q", "/X:$setupDir"
if ($extract.ExitCode -ne 0) {
    throw "SQL Server media extraction failed: $($extract.ExitCode)"
}

$setup = Start-Process (Join-Path $setupDir "setup.exe") -Wait -PassThru -ArgumentList @(
    "/Q"
    "/IACCEPTSQLSERVERLICENSETERMS"
    "/ACTION=Install"
    "/FEATURES=SQLEngine"
    "/INSTANCENAME=CISQL"
    "/INSTANCEDIR=$instanceDir"
    "/INSTALLSQLDATADIR=$instanceDir"
    '/SQLSYSADMINACCOUNTS="BUILTIN\Administrators"'
    "/UPDATEENABLED=False"
    "/TCPENABLED=0"
    "/NPENABLED=0"
)
$summaryLog = Join-Path $env:RUNNER_TEMP "sql-server-summary.txt"
Copy-Item "C:\Program Files\Microsoft SQL Server\170\Setup Bootstrap\Log\Summary.txt" `
    $summaryLog -ErrorAction SilentlyContinue
if ($setup.ExitCode -ne 0) {
    Get-Content $summaryLog -ErrorAction SilentlyContinue
    throw "SQL Server setup failed or requires a restart: $($setup.ExitCode)"
}

$connection = New-Object System.Data.SqlClient.SqlConnection
$connection.ConnectionString = "Data Source=lpc:.\CISQL;Integrated Security=True;Connect Timeout=30"
try {
    $connection.Open()
    $command = $connection.CreateCommand()
    $command.CommandText = "CREATE DATABASE CIWindowsExporter"
    $command.ExecuteNonQuery() | Out-Null
    # Express can close idle databases and remove their performance counter instances.
    $command.CommandText = "ALTER DATABASE CIWindowsExporter SET AUTO_CLOSE OFF"
    $command.ExecuteNonQuery() | Out-Null
    $connection.ChangeDatabase("CIWindowsExporter")
    $command.CommandText = @'
CREATE TABLE dbo.Fixture (ID int NOT NULL, Data char(8000) NOT NULL);
INSERT INTO dbo.Fixture SELECT TOP (1000) ROW_NUMBER() OVER (ORDER BY (SELECT NULL)), 'fixture' FROM sys.all_objects;
CHECKPOINT;
'@
    $command.ExecuteNonQuery() | Out-Null
} finally {
    $connection.Dispose()
}
Get-Service 'MSSQL$CISQL' | Format-Table Name, Status
