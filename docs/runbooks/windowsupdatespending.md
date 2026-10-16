---
title: WindowsUpdatesPending
---

## Meaning

The host has updates waiting for installation for longer than
`updatePendingDuration`, seven days by default. The alert is an info alert, its
value is the number of pending updates, and it needs the `update` collector,
which queries the Windows Update agent in the background at the interval of
its scrape setting. Updates waiting for uninstallation do not count. Prometheus counts the
duration and restores it after a restart shorter than its
`--rules.alert.for-outage-tolerance`, one hour by default; a longer outage or
a new server starts the wait again.

## Impact

Security updates that wait longer than the patch window leave known
vulnerabilities open. Other updates carry fixes for reliability problems. The
host may also be waiting for a restart to finish an installation.

## Diagnosis

1. List the pending updates with their category and severity:

   ```promql
   windows_update_pending_info{job="windows_exporter", instance="<instance>", deployment_action="installation"}
   ```

   `windows_update_pending_published_timestamp` shows how old each update is.
2. On the host, check the update client and the last installation attempts:

   ```powershell
   Get-WindowsUpdateLog
   Get-WinEvent -LogName Microsoft-Windows-WindowsUpdateClient/Operational -MaxEvents 50
   Get-CimInstance -Namespace root/cimv2 -ClassName Win32_QuickFixEngineering | Sort-Object InstalledOn -Descending | Select-Object -First 5
   ```

3. Check whether the host is in a maintenance window, a deferral period or a
   WSUS or Configuration Manager deployment that has not approved the updates
   yet, and whether a pending restart blocks the installation.

## Mitigation

Install the updates in the next maintenance window, or approve them in the
update management tool. Restart the host if an installation waits for it.
Raise `updatePendingDuration` to the patch window of the environment if seven
days is shorter than the policy, or turn the alert off with
`alerts+: { WindowsUpdatesPending: false }` where another system tracks patch
compliance.
