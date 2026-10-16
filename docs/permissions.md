# Running as a least-privilege service account

By default, the MSI installs the `windows_exporter` service as `LocalSystem`.
Every collector works with that account. The `SERVICE_ACCOUNT` property of the
MSI sets another account, but the MSI does not grant that account any rights.

The exporter only reads data and needs no administrator rights for the common
collectors. This page describes how to run the service as an unprivileged
account instead, and which collectors then lose data.

## Choose an account

- A virtual service account, `NT SERVICE\windows_exporter`. Windows creates and
  manages it with the service, and it has no password. It gets the "Log on as
  a service" right through `NT SERVICE\ALL SERVICES`, which Windows includes in
  that right by default. A group policy that sets "Log on as a service" without
  `NT SERVICE\ALL SERVICES` stops virtual accounts from starting. This is the
  simplest choice on a single host.
- A group managed service account (gMSA), for example `CONTOSO\svc-exporter$`,
  when the account needs rights in the domain or the same identity on many
  hosts. Install it on the host with `Install-ADServiceAccount` first. It needs
  the "Log on as a service" user right (`SeServiceLogonRight`).
- A domain or local user. It needs the "Log on as a service" user right
  (`SeServiceLogonRight`) and a password that you have to rotate.

On top of the account itself, add it to the local **Performance Monitor Users**
group. Windows uses this group to grant access to performance counter data
(PDH), which most collectors read. Users who run the exporter like this in
[#594](https://github.com/prometheus-community/windows_exporter/issues/594)
report that the group is enough for the default collectors.
**Performance Log Users** is not needed: it adds the right to create data
collector sets and logs, and the exporter only reads counters.

## Set up the service

Install the MSI with the account:

```powershell
msiexec /i <path-to-msi-file> --% SERVICE_ACCOUNT="NT SERVICE\windows_exporter"
```

For a gMSA, set `SERVICE_ACCOUNT="CONTOSO\svc-exporter$"`. For a user, also set
`SERVICE_PASSWORD`; the MSI keeps it out of its log. A gMSA or a user needs
"Log on as a service" before you install, because the MSI starts the service
right away. The MSI starts it before the account has the other rights below,
so restart the service after granting them.

Then run this in an elevated PowerShell. The service must exist before Windows
can resolve the virtual account `NT SERVICE\windows_exporter`.

```powershell
$service = 'windows_exporter'
$account = "NT SERVICE\$service"

# Only needed if the MSI was installed without SERVICE_ACCOUNT. A virtual
# account has no password.
sc.exe config $service obj= $account

# Grant read access to performance counters. S-1-5-32-558 is the Performance
# Monitor Users group in every display language. SilentlyContinue lets the script
# run again after an MSI upgrade, when the account is already a member.
Add-LocalGroupMember -SID 'S-1-5-32-558' -Member $account -ErrorAction SilentlyContinue

# The default textfile directory and config file below Program Files are readable
# by all users. Grant read access to other locations you use, for example:
# icacls 'D:\metrics' /grant "${account}:(OI)(CI)RX"
# icacls 'D:\windows_exporter\config.yaml' /grant "${account}:R"

Restart-Service $service
```

`Add-LocalGroupMember` and `Remove-LocalGroupMember` come from the
`Microsoft.PowerShell.LocalAccounts` module, which needs 64-bit PowerShell.

Without `SERVICE_ACCOUNT`, use `sc.exe config windows_exporter obj= 'CONTOSO\svc-exporter$'`
for a gMSA, and also add `password= '<password>'` for a user. For both, grant
"Log on as a service" with `secpol.msc` or a group policy: neither the MSI nor
`sc.exe` grants it, only the Services console (`services.msc`) does when you set
the account there.

The account also needs:

- Read access to the config file (`--config.file`), the textfile directories
  (`--collector.textfile.directories`), the files and registry keys that the
  `file` and `registry` collectors read, and the TLS key and certificate of
  `--web.config.file`.
- Write access to the log file and its directory if `--log.file` is a path. The
  default for the service, the Windows event log, needs no extra rights.

Writers of textfile metrics need write access to their directory; the exporter
itself only reads it.

To check the result, look at the account and at the collectors that fail:

```powershell
(Get-CimInstance Win32_Service -Filter "Name='windows_exporter'").StartName
(Invoke-WebRequest -Uri 'http://localhost:9182/metrics' -UseBasicParsing).Content -split "`n" |
    Select-String 'windows_exporter_collector_success'
```

How missing rights show up depends on when the collector reads the data:

- Some collectors query WMI while they start: `dns` (`wmi_stats`), `fsrmquota`
  and `storage_spaces`. Access denied there is not an expected startup error,
  so the exporter exits and the service does not start. The error is in the
  Application event log (source `windows_exporter`). Grant access or disable
  the collector.
- A collector that fails while scraping has
  `windows_exporter_collector_success` 0 and logs the error.
- Partial data loss, such as skipped services, missing scheduled tasks or empty
  process labels, leaves `windows_exporter_collector_success` at 1.

### Reverting and MSI upgrades

To run the service as `LocalSystem` again:

```powershell
sc.exe config windows_exporter obj= LocalSystem
Remove-LocalGroupMember -SID 'S-1-5-32-558' -Member 'NT SERVICE\windows_exporter'
Restart-Service windows_exporter
```

The MSI does not remember `SERVICE_ACCOUNT`, like its other properties. An
upgrade or a repair installs the service again, as `LocalSystem` unless you
pass `SERVICE_ACCOUNT` (and `SERVICE_PASSWORD`) again. Group memberships and
user rights of the account stay, so passing the same properties to every
`msiexec` run is enough.

## Collectors with limits

Collectors that are not listed here read performance counters, the registry
below `HKLM\SOFTWARE`, or other APIs that all users can read, and are expected
to work with the setup above. Only some of them were reported working in the
issue linked above (`cpu`, `logical_disk`, `memory`, `mssql`, `os`, `process`,
`system`); the others are not verified. A counter set whose provider restricts access itself
can still fail.

| Collector | Without administrator rights | What to grant |
|-----------|------------------------------|---------------|
| `container` | The `hcs` sub-collector uses the Host Compute Service, the Host Compute Network service for network statistics and the CRI endpoint (containerd named pipe); memory statistics open the container processes. The `hostprocess` sub-collector also uses the CRI endpoint and opens the job objects of HostProcess containers. Not verified, expected to fail. | Not verified. Use `LocalSystem`. |
| `dhcp` | `server_metrics` reads counters and works. `scope_metrics` uses the DHCP Server API (`dhcpsapi.dll`), which only allows Administrators, DHCP Administrators and DHCP Users. | Add the account to the local **DHCP Users** group (read-only access to the DHCP server). Not verified live. |
| `dns` | `metrics` reads counters and works. `wmi_stats` queries `MicrosoftDNS_Statistic` in `root\MicrosoftDNS`. | WMI namespace security on `root\MicrosoftDNS`, see below. Whether the DNS provider also requires DnsAdmins is not verified. |
| `fsrmquota` | Queries `MSFT_FSRMQuota` in `root\Microsoft\Windows\FSRM`. | WMI namespace security on that namespace. Whether the FSRM provider allows non-administrators is not verified. |
| `hyperv` | The counter sub-collectors work. `host` uses the Host Compute Service; `replica_vm`, `virtual_hard_disk` and `wmi_health` query `root\virtualization\v2`, and `virtual_hard_disk` also opens the virtual disk files. | Membership in **Hyper-V Administrators** gives access to the Hyper-V WMI provider, but it also grants full control over Hyper-V. Read access to the VHD/VHDX files for `virtual_hard_disk`. Not verified; disable these sub-collectors with `--collector.hyperv.enabled` instead. |
| `mscluster` | Reads through the Failover Cluster API (ClusAPI), which requires cluster access. | `Grant-ClusterAccess -User <account> -ReadOnly` on the cluster. Not verified for a virtual account; use a gMSA or domain account, which is the same identity on every node. |
| `process` | All metrics work. The `owner`, `cmdline` and `process_group_id` labels of `windows_process_info` are empty for processes of other users and protected processes (see [`collector.process.md`](collector.process.md)). With `--collector.process.iis`, the application pool comes from the `w3wp` command line, which needs access to the worker process; the fallback queries `root\WebAdministration`. | No grant that opens processes of other users is narrower than administrator rights. WMI namespace security on `root\WebAdministration` for the IIS fallback (not verified). |
| `scheduled_task` | Lists only the tasks whose security descriptor lets the account read them, so tasks are missing. | Not verified. Use `LocalSystem` to see all tasks. |
| `service` | Works for most services. A service whose security descriptor denies the account `SERVICE_QUERY_CONFIG` is skipped with a warning and has no metrics at all, including `windows_service_state`. `windows_service_process` is missing for services whose process the account can't open with `PROCESS_QUERY_LIMITED_INFORMATION`. | Not verified which services are affected. |
| `storage_spaces` | Queries `MSFT_StoragePool` and `MSFT_VirtualDisk` in `root\Microsoft\Windows\Storage`. | WMI namespace security on that namespace. Whether the storage provider allows non-administrators is not verified. |
| `wmi` | Queries the namespaces that you configure. `root\cimv2` allows all authenticated users. | WMI namespace security for other namespaces. |
| `file`, `registry`, `textfile` | Read the paths and registry keys that you configure. | Read access on those files, directories and keys. |
| `update` | Uses the Windows Update Agent API. Not verified. | Not verified. |

The `logical_disk` collector, including `bitlocker_status`, works for
non-elevated accounts (see [`collector.logical_disk.md`](collector.logical_disk.md)).
`cpu_info` and `diskdrive` query `root\cimv2`, which all authenticated users can
read.

### Granting WMI namespace access

By default, all authenticated users have the **Enable Account** permission on
`root\cimv2`. Other namespaces, such as the ones of the DNS, FSRM, storage and
Hyper-V providers, can be limited to administrators. To grant access:

1. Run `wmimgmt.msc`, right-click **WMI Control (Local)** and select
   **Properties**.
2. On the **Security** tab, select the namespace, for example
   `Root\MicrosoftDNS`, and select **Security**.
3. Add the account (`NT SERVICE\windows_exporter`) and allow **Enable Account**.
   The exporter queries locally, so **Remote Enable** is not needed.
4. Restart the service.

A provider can check the caller again after the namespace check and still deny
the query. If the collector keeps failing with access denied, that provider
needs administrator rights or a role-specific group.
