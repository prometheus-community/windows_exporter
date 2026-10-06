param([int]$TestExitCode)

$ErrorActionPreference = "Stop"
$prefix = "github.com/prometheus-community/windows_exporter/internal/collector/"
$required = @($env:WINDOWS_EXPORTER_TEST_COLLECTORS.Split(','))
$results = @{}
$text = [System.Text.StringBuilder]::new()

foreach ($line in Get-Content test-results.jsonl) {
    try {
        $event = $line | ConvertFrom-Json -ErrorAction Stop
    } catch {
        [void]$text.AppendLine($line)
        continue
    }
    if ($event.Output) { [void]$text.Append($event.Output) }
    if (-not $event.Package -or -not $event.Package.StartsWith($prefix) -or $event.Test -ne "TestCollector") {
        continue
    }
    $name = $event.Package.Substring($prefix.Length)
    if (-not $results.ContainsKey($name)) {
        $results[$name] = @{ State = "not completed"; Notes = @() }
    }
    if ($event.Action -in @("pass", "skip", "fail")) {
        $results[$name].State = $event.Action
    }
    if ($event.Output -match 'collector \S+ (?:is not supported|has empty optional instance groups): (.+)') {
        $results[$name].Notes += $Matches[1].Trim()
    }
}
$text.ToString() | Set-Content test-output.txt -NoNewline -Encoding utf8

# Aborted tests must leave required fixtures visible as not run.
foreach ($name in $required) {
    if (-not $results.ContainsKey($name)) {
        $results[$name] = @{ State = "not run"; Notes = @() }
    }
}
$testState = if ($TestExitCode -eq 0) { "passed" } else { "failed" }
$summary = @(
    "### Windows feature tests: $env:WINDOWS_TEST_OS"
    ""
    "Go suite: **$testState**. Required fixtures cannot skip. Empty optional instance groups remain listed below."
    ""
    "| Collector | Required | Live test | Unavailable features / notes |"
    "| --- | --- | --- | --- |"
)
foreach ($name in $results.Keys | Sort-Object) {
    $fixture = if ($name -in $required) { "yes" } else { "no" }
    $state = switch ($results[$name].State) {
        "pass" { "passed" }
        "skip" { "unavailable" }
        "fail" { "failed" }
        default { $results[$name].State }
    }
    $notes = ($results[$name].Notes | Sort-Object -Unique) -join '; '
    $notes = $notes.Replace('|', '\|')
    $summary += "| $name | $fixture | $state | $notes |"
}
$summary | Set-Content feature-summary.md -Encoding utf8
if ($env:GITHUB_STEP_SUMMARY) {
    $summary | Add-Content $env:GITHUB_STEP_SUMMARY
}
