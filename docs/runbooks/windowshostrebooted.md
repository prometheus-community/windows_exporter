---
title: WindowsHostRebooted
---

## Meaning

The boot time of the host, `windows_system_boot_time_timestamp`, is less than
ten minutes old and differs from the boot time known ten minutes earlier. The
alert fires once for about ten minutes after a reboot and then resolves by
itself. It has the `info` severity and is meant for a dashboard, a ticket or a
chat notification, not for a page.

## Impact

Every service of the host restarted, connections to it were dropped and
caches are cold. A reboot outside a maintenance window can mean a crash, an
unattended Windows Update restart or a power event.

## Diagnosis

1. Compare the boot time with the maintenance calendar and the Windows Update
   schedule. With the `update` collector enabled, `windows_update_pending_info`
   shows updates that were pending before the reboot.
2. Check the System event log: event 1074 names the user or process that
   requested the shutdown, event 6008 marks an unexpected shutdown and event 41
   a reset without a clean shutdown. Check for a memory dump when it was a
   crash.
3. Confirm the host came back with the expected services; the
   `WindowsServiceNotRunning` alert covers automatic services that did not
   start.

## Mitigation

No action is needed for a planned reboot. For an unplanned one, find the cause
in the event log and fix the driver, update setting or power problem. Turn the
alert off with `alerts+: { WindowsHostRebooted: false }` when reboots are not
worth a notification.
