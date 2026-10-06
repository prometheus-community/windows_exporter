# Windows feature tests

The CI Windows job prepares real Windows roles and workloads before running
`go test -json -count=1 -race -timeout=10m ./...`. Collector tests validate the
metrics from those workloads. A final exporter smoke check builds and starts the
binary, requests `/metrics` once, and checks for a nonempty HTTP 200 response.

The job runs on Windows Server 2022 and 2025. It provisions Containers, Hyper-V,
SQL Server Express, IIS, MSMQ, NPS, FSRM, SMB, DNS, DHCP, a workgroup failover
cluster, printer and storage fixtures, and an authenticated RDP session. SMTP is
also tested on Server 2022; the role was removed from Server 2025.

`WINDOWS_EXPORTER_TEST_COLLECTORS` lists required collectors. Their tests fail
when setup is missing or collection fails instead of skipping. Fixture assertions
check known instances, including the active RDP user's session and SQL database.
RemoteFX is required when both network and graphics instances are available.

Each Windows job publishes a feature table in its Actions summary and a
`windows-test-results-<OS>` artifact containing `feature-summary.md`, structured
Go test events, readable test output, and setup diagnostics. The table distinguishes
required fixtures, unavailable collectors, and empty optional instance groups.
This records Windows feature availability rather than Go statement coverage.

## Optional domain lab

Dispatch the CI workflow with `domain_lab=true` to run only the domain pilot:

```sh
gh workflow run ci.yml --repo prometheus-community/windows_exporter \
  --ref master -f domain_lab=true
```

For an unmerged branch, use its repository and branch instead. Actions must be
enabled in that repository.

The pilot applies versioned Microsoft Windows Server 2025 evaluation media to a
disposable Hyper-V guest, promotes it to a domain controller, installs an
enterprise root CA, and enrolls a computer certificate. Promotion and reboot
happen inside the guest. PowerShell Direct controls the guest through an isolated
private switch, so it needs no external VM or cloud account.

The host compiles the AD and ADCS Go tests with the race detector and runs them
inside the guest with both collectors required. It then performs the same single
exporter smoke request. The `domain-lab-results` artifact includes a feature table,
test output, provisioning diagnostics, and `timings.json` with each setup phase's
duration.

The pilot downloads an approximately 8 GB ISO and boots a 4 GB RAM guest. It is
manual so its download, image deployment, and reboot costs do not slow every PR.
The measured phase durations inform whether to reuse this approach for routine
CI or a scheduled lab. Multi-server scenarios such as DFS replication need a
second guest and remain a follow-up.
