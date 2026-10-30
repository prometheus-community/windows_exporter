# Privacy policy

windows_exporter does not collect telemetry, send usage reports or check for
updates. This program will not transfer any information to other networked
systems unless specifically requested by the user or the person installing or
operating it.

windows_exporter only serves the metrics it collects on the local machine over
HTTP, by default on port 9182, to clients that request them, such as a
Prometheus server you run. You control who can reach that endpoint with the
listen address, the firewall and the [web configuration file](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md)
for TLS and authentication.

When enabled with `--collector.update.online`, the `update` collector asks the
Windows Update Agent to search for updates online. Windows then contacts the
update service configured on the machine, such as Microsoft Update or WSUS.
This option is off by default.
