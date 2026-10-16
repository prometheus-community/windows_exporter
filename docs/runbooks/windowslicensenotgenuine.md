---
title: WindowsLicenseNotGenuine
---

## Meaning

Windows reports a license status other than `genuine` for one hour. The alert
carries the status in its `state` label and needs the `license` collector. The
status comes from the Software Licensing service and changes when an
activation expires, a KMS host is unreachable for more than the grace period,
the hardware changed, or the key was blocked. `offline` means the genuine
validation could not reach Microsoft; on hosts without internet access that
is the permanent state, so turn the alert off there with
`alerts+: { WindowsLicenseNotGenuine: false }` or scope it with the exporter
selector.

## Impact

An unlicensed Windows shows activation notices, may limit personalization and,
on some editions, restarts on a schedule. A Windows Server that leaves its
activation grace period is out of compliance.

## Diagnosis

1. The `state` label of the alert names the status; `windows_license_status`
   shows all states of the host.
2. On the host, read the license details:

   ```powershell
   slmgr.vbs /dlv
   slmgr.vbs /xpr
   Get-CimInstance SoftwareLicensingProduct -Filter "PartialProductKey IS NOT NULL" | Select-Object Name, LicenseStatus, GracePeriodRemaining
   ```

3. For KMS clients, check that the KMS host resolves through the `_vlmcs._tcp`
   DNS record and answers on TCP port 1688. For Active Directory based
   activation, check the activation object in the forest. For MAK or retail
   keys, check whether the activation count is exhausted.

## Mitigation

Restore reachability to the KMS host or the activation object, then run
`slmgr.vbs /ato`. Install the correct key with `slmgr.vbs /ipk <key>` if the
key was wrong, or reactivate after a hardware change. Confirm the status
returns to `genuine`.
