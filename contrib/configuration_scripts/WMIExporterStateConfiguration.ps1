cls
$serviceName = "wmi_exporter"
$MSMQCollector = "msmq"
$IISCollector = "iis"
$MSSQLCollector = "mssql"
$MetricsPath = "C:\Metrics"
$ImagePathlocation = "HKLM:\SYSTEM\CurrentControlSet\Services\wmi_exporter\"
$RegString = "ImagePath"
$ImagePathstr1 = '"C:\Program Files\wmi_exporter\wmi_exporter.exe" --log.format logger:eventlog?name=wmi_exporter --telemetry.addr :9182 --collector.textfile.directory C:\Metrics --collectors.enabled '
$global:CoreCollectorsEnabled = "cpu,cs,logical_disk,net,os,process,service,textfile,system" -split ","
$CollectorServiceQuery = @"
--collector.service.services-where "Name='wmi_exporter' or Name='TermService' or Name='BITS' or Name='NTD' or Name='ADWS' or Name='DFSR' or Name='DNS' or Name='ADWS' or Name='wuauserv' or Name='VMTools' or Name='W3SVC'"
"@
$collectorprocessqry = @"
--collector.process.processes-where "Name LIKE '%w3wp%' or Name='wmi_exporter' or Name='java' or Name='Coalmine.Prometheus.Service' or Name LIKE or Name LIKE '%sqlservr%' or Name='amazon-ssm-agent' or Name='mqsvc'"
"@


Function IsIISInstalled {

return (Get-WindowsFeature -Name Web-Server).Installed

}

Function IsMSMQInstalled {

return (Get-WindowsFeature -Name MSMQ-Server).Installed

}

Function IsMSSQLInstalled {

return (Test-Path “HKLM:\Software\Microsoft\Microsoft SQL Server\Instance Names\SQL”)

}

function BuildArgsRegValue() {
    if (IsIISInstalled) {$CoreCollectorsEnabled = $CoreCollectorsEnabled + $IISCollector}
	if (IsMSMQInstalled) {$CoreCollectorsEnabled = $CoreCollectorsEnabled + $MSMQCollector}
    if (IsMSSQLInstalled) {$CoreCollectorsEnabled = $CoreCollectorsEnabled + $MSSQLCollector}
return $CoreCollectorsEnabled -join ","

}
$allcollectors = BuildArgsRegValue
Function RegKeyCreate ($allcollectors){

$ImagePathValue = "$ImagePathstr1 $allcollectors $CollectorServiceQuery $collectorprocessqry"
write-host "$ImagePathValue"

$value1 = (Test-Path $ImagePathlocation)
If ($value1 -eq $true) 
{

Set-ItemProperty -Path $ImagePathlocation -Name "$Regstring" "$ImagePathValue" -Force

}

else {
Write-Host "WMI Exporter Registry key does not exist!"
}

 }

 Function CreateMetricsDir () {
 
 If(!(test-path $MetricsPath)){
      write-host "Creating $MetricsPath"
      New-Item -ItemType Directory -Force -Path $MetricsPath | Out-Null
 }
    }
Function RestartService() {

If (Get-Service $serviceName -ErrorAction SilentlyContinue) {

    If ((Get-Service $serviceName).Status -eq 'Running') {

        Restart-Service $serviceName
        Write-Host "Restarting $serviceName"

    } Else {

        Write-Host "$serviceName found, but it is not running, lets attempt to restart it"
        Start-Service $serviceName
    }

} Else {

    Write-Host "$serviceName service installed"

}


}

BuildArgsRegValue
RegKeyCreate $allcollectors
CreateMetricsDir
RestartService