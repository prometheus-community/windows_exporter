---
title: WindowsHyperVWMIUnhealthy
---

## Meaning

The Hyper-V WMI namespace `root/virtualization/v2` on the host has not
answered a minimal query for ten minutes. The alert needs the `hyperv`
collector with its `wmi_health` subcollector. While the namespace is broken,
the virtual machine, replication and virtual hard disk metrics that come from
WMI are missing, so the other Hyper-V alerts cannot fire.

## Impact

Management tools such as Hyper-V Manager, Failover Cluster Manager and System
Center cannot manage the virtual machines on the host. Backups that use the
Hyper-V VSS writer fail. The virtual machines themselves usually keep running.

## Diagnosis

1. Inspect `windows_hyperv_wmi_health` and `windows_exporter_collector_success{collector="hyperv"}`
   for the host.
2. On the host, test the namespace and the management service:

   ```powershell
   Get-CimInstance -Namespace root/virtualization/v2 -ClassName Msvm_ComputerSystem | Select-Object -First 1
   Get-Service vmms, vmcompute, winmgmt
   winmgmt /verifyrepository
   ```

3. Check the System and the Microsoft-Windows-Hyper-V-VMMS logs for errors of
   the Virtual Machine Management service, and the WMI-Activity log for
   provider failures. A stuck `vmms` process or a corrupt WMI repository are
   the usual causes.

## Mitigation

Restart the Virtual Machine Management service with `Restart-Service vmms`;
running virtual machines are not affected. If the repository is inconsistent,
run `winmgmt /salvagerepository`, and register the Hyper-V provider again with
`mofcomp` of the Hyper-V MOF files only when Microsoft support advises it. Plan a
restart of the host if the service cannot be recovered. Confirm the health
metric returns to 1 and the Hyper-V metrics appear again.
