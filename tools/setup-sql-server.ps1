# Install a real Database Engine instance; LocalDB does not expose server counters.
$ErrorActionPreference = "Stop"

$mediaDir = Join-Path $env:RUNNER_TEMP "sql-server"
New-Item -ItemType Directory -Force -Path $mediaDir | Out-Null
$installer = Join-Path $mediaDir "SQL2025-SSEI-Expr.exe"
Invoke-WebRequest `
    -Uri "https://download.microsoft.com/download/ffd82b4c-9955-47c0-8efe-6290f7795cf6/SQL2025-SSEI-Expr.exe" `
    -OutFile $installer
$expectedHash = "fa7e1fabc9a2e9c9cdab0d1512bcb30d2949133147db057ac530f510e5270680"
if ((Get-FileHash $installer -Algorithm SHA256).Hash -ne $expectedHash) {
    throw "SQL Server bootstrapper SHA256 mismatch"
}

$download = Start-Process $installer -Wait -PassThru `
    -ArgumentList "/ACTION=Download", "/MEDIATYPE=Core", "/QUIET", "/MEDIAPATH=$mediaDir"
if ($download.ExitCode -ne 0) {
    throw "SQL Server media download failed: $($download.ExitCode)"
}

$setupDir = Join-Path $mediaDir "setup"
$extract = Start-Process (Join-Path $mediaDir "SQLEXPR_x64_ENU.exe") -Wait -PassThru `
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
    '/SQLSYSADMINACCOUNTS="BUILTIN\Administrators"'
    "/UPDATEENABLED=False"
    "/TCPENABLED=0"
    "/NPENABLED=0"
)
if ($setup.ExitCode -ne 0) {
    Get-Content "C:\Program Files\Microsoft SQL Server\170\Setup Bootstrap\Log\Summary.txt" `
        -ErrorAction SilentlyContinue
    throw "SQL Server setup failed or requires a restart: $($setup.ExitCode)"
}

$connection = New-Object System.Data.SqlClient.SqlConnection
$connection.ConnectionString = "Data Source=lpc:.\CISQL;Integrated Security=True;Connect Timeout=30"
try {
    $connection.Open()
    $command = $connection.CreateCommand()
    $command.CommandText = @'
CREATE DATABASE CIWindowsExporter;
USE CIWindowsExporter;
CREATE TABLE dbo.Fixture (ID int NOT NULL, Data char(8000) NOT NULL);
INSERT INTO dbo.Fixture SELECT TOP (1000) ROW_NUMBER() OVER (ORDER BY (SELECT NULL)), 'fixture' FROM sys.all_objects;
CHECKPOINT;
'@
    $command.ExecuteNonQuery() | Out-Null
} finally {
    $connection.Dispose()
}
Get-Service 'MSSQL$CISQL' | Format-Table Name, Status
