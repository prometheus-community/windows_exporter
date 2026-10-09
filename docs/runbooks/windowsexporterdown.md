---
title: WindowsExporterDown
---

## Meaning

Prometheus has failed to scrape the target for five minutes (`up == 0`).
The alert does not detect hosts removed from service discovery.

## Impact

Monitoring data and resource alerts for the host are unavailable. The host,
exporter service, or the network path may be down.

## Diagnosis

1. Open the target in Prometheus's **Targets** page and read the scrape error.
   Check the target address, TLS settings, authentication, and scrape timeout.
2. From the Prometheus network, check the configured metrics endpoint, normally
   `http://<host>:9182/metrics`. Check DNS resolution and routing.
3. On the affected host, inspect the service and listening socket:

   ```powershell
   Get-Service windows_exporter
   Get-NetTCPConnection -LocalPort 9182 -State Listen
   Test-NetConnection -ComputerName localhost -Port 9182
   ```

4. Review exporter logs and the Windows Application/System event logs for
   startup errors, permission problems, or collector timeouts.

## Mitigation

Restore the network path or correct the scrape configuration. If the exporter
service is stopped, resolve its startup error and start it with
`Start-Service windows_exporter`. If local requests succeed but remote requests
fail, check the host firewall and network ACLs for the configured listener.
Confirm the target becomes healthy and fresh metrics appear before closing the
incident. Update service discovery when a host was intentionally decommissioned.
