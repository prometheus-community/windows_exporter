---
title: WindowsServiceCrashLooping
---

## Meaning

The service `service` has run as more processes in the last 15 minutes than
the configured number of restarts allows. The default is two restarts. Every
service process has its own `windows_service_process` series, identified by
`process_id`, so the series seen within 15 minutes count the restarts.

## Impact

The service crashes or is stopped and started repeatedly. It is unavailable
during every restart, and the recovery options of the Service Control Manager
give up after a configurable number of failures, after which
`WindowsServiceNotRunning` follows.

## Diagnosis

1. List the processes the service ran as:

   ```promql
   count_over_time(windows_service_process{job="windows_exporter", service="<service>"}[15m])
   ```

   The value of each series is the start time of the process.
2. Check the Application log for crashes of the service executable (events
   1000 and 1026, .NET Runtime errors) and the System log for Service Control
   Manager events 7031 and 7034, which also show the recovery action.
3. Check whether a deployment, configuration push, scheduled task or another
   automation restarts the service on purpose.
4. Check for resources the service fails to acquire at start, such as a port
   that is in use, a certificate, a database or a license.

## Mitigation

Fix the crash or the start failure, roll back the deployment that caused it, or
stop the automation that restarts the service. Keep the service stopped until
the cause is known if every restart has side effects. Confirm the service keeps
one process and stays running. Raise the threshold for services that restart by
design, for example worker services that recycle themselves.
