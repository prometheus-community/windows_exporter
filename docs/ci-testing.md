# Windows feature tests

The CI Windows job prepares real Windows roles and workloads before running
`go test -json -count=1 -race -timeout=10m ./...`. Collector tests validate the
metrics from those workloads. A final exporter smoke check builds and starts the
binary, requests `/metrics` once, and checks for a nonempty HTTP 200 response.

The job runs on Windows Server 2022; Server 2025 runners take much longer to
provision the same fixtures. It provisions Containers, Hyper-V, SQL Server
Express, IIS, SMTP, MSMQ, NPS, FSRM, SMB, DNS, DHCP, a workgroup failover
cluster, printer and storage fixtures, unlocked and locked BitLocker volumes,
and an authenticated RDP session.

The failover cluster has no administrative access point: the DHCP-only runner
network cannot host a Cluster Name. An internal Hyper-V switch provides a private
client network for its resource fixtures. The `CIResources` group contains an
online Generic Service that depends on an online IP Address, plus an offline IP
Address. The `CIFailedResources` group contains a Generic Service whose binary is
missing, so it stays failed. Native and WMI parity tests require these resources.

`WINDOWS_EXPORTER_TEST_COLLECTORS` lists required collectors. Their tests fail
when setup is missing or collection fails instead of skipping. Fixture assertions
check known instances, including the active RDP user's session and SQL database.
RemoteFX is required when both network and graphics instances are available.

Each Windows job publishes a feature table in its Actions summary and a
`windows-test-results-<OS>` artifact containing `feature-summary.md`, structured
Go test events, readable test output, and setup diagnostics. The table distinguishes
required fixtures, unavailable collectors, and empty optional instance groups.
This records Windows feature availability rather than Go statement coverage.
