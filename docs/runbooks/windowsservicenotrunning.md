---
title: WindowsServiceNotRunning
---

## Meaning

The service `service` has the start mode `auto` but has not been in the
`running` state for 15 minutes. Trigger-start services, which Windows stops by
design when their trigger condition ends, are left out through the
`trigger_start` label of `windows_service_info`. Delayed-start services are
included; the 15 minute duration covers their later start after a reboot.

## Impact

A service that should always run is stopped, so the function it provides is
unavailable, for example printing, time synchronization, a site or a
database instance. Windows restarts a crashed service only when its recovery
options say so.

## Diagnosis

1. Inspect the current state of the service and when it changed:

   ```promql
   windows_service_state{job="windows_exporter", service="<service>"} == 1
   ```

2. Check the System event log for Service Control Manager events 7000, 7009,
   7031 and 7034, and the Application log for crashes of the service process.
3. Check `windows_service_info` for the account in `run_as`; an expired
   password or a removed right to log on as a service stops the service at the
   next start. `windows_service_process` shows whether a process is running.
4. Check whether the service was stopped on purpose for maintenance or
   disabled by policy, and whether it depends on another service that is also
   stopped.

## Mitigation

Start the service after fixing the cause: the account, a missing dependency, a
configuration error or a crash. Set recovery options so Windows restarts the
service itself. Change the start mode to `manual` or `disabled` for services
that should not run, or exclude them with the `serviceSelector` when the mixin
should not watch them. Confirm the service stays running.
