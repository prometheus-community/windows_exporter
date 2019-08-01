Param(
    [Parameter(Mandatory=$true)]
    $Class,
    [Parameter(Mandatory=$false)]
    $CollectorName = ($Class -replace 'Win32_PerfRawData_Perf',''),
    [Parameter(Mandatory=$false)]
    $ComputerName = "localhost",
    [Parameter(Mandatory=$false)]
    $Credential
)
$ErrorActionPreference = "Stop"

if($Credential -ne $null) {
    $wmiObject = Get-WMIObject -ComputerName $ComputerName -Credential $Credential -Class $Class
}
else {
    $wmiObject = Get-WMIObject -ComputerName $ComputerName -Class $Class
}

$members = $wmiObject `
    | Get-Member -MemberType Properties `
    | Where-Object { $_.Definition -Match '^u?int' -and $_.Name -NotMatch '_' } `
    | Select-Object Name, @{Name="Type";Expression={$_.Definition.Split(" ")[0]}}
$input = @{
    "Class"=$Class;
    "CollectorName"=$CollectorName;
    "Members"=$members
} | ConvertTo-Json
$outFileName = "..\..\collector\$CollectorName.go".ToLower()
$input | .\collector-generator.exe | Out-File -NoClobber -Encoding UTF8 $outFileName
go fmt $outFileName
