# Windows feature tests

The CI Windows job prepares real Windows roles and workloads, then runs the Go
tests with the race detector. While the fixtures are provisioned, a background
step builds every test binary and the exporter with
[`tools/build-go-tests.ps1`](../tools/build-go-tests.ps1); linking the race test
binaries would otherwise take minutes after provisioning.
[`tools/run-go-tests.ps1`](../tools/run-go-tests.ps1) then runs the binaries in
parallel through `test2json`, each in its package directory, producing
the same events as `go test -json ./...`. Collector tests validate the metrics
from those workloads. A final exporter smoke check starts the binary, requests
`/metrics` once, and checks for a nonempty HTTP 200 response.

The job runs on Windows Server 2022; Server 2025 runners take much longer to
provision the same fixtures. It provisions Containers, Hyper-V, SQL Server
Express, IIS, SMTP, MSMQ, NPS, FSRM, SMB, DNS, DHCP, a workgroup failover
cluster, printer and storage fixtures, unlocked and locked BitLocker volumes,
and an authenticated RDP session.

`WINDOWS_EXPORTER_TEST_COLLECTORS` lists required collectors. Their tests fail
when setup is missing or collection fails instead of skipping. Fixture assertions
check known instances, including the active RDP user's session and SQL database.
RemoteFX is required when both network and graphics instances are available.

Each Windows job uploads a `windows-test-results-<OS>` artifact containing the
Go test events (`test-results.jsonl`) and setup diagnostics. Unavailable
optional collectors show up as skipped `TestCollector` tests with the reason in
their output.
