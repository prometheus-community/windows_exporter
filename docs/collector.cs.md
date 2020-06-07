# cs collector

The cs collector exposes metrics detailing the hardware of the computer system

|||
-|-
Metric name prefix  | `cs`
Classes             | [`Win32_ComputerSystem`](https://msdn.microsoft.com/en-us/library/aa394102)
Enabled by default? | Yes

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_cs_logical_processors` | Number of installed logical processors | gauge | None
`windows_cs_physical_memory_bytes` | Total installed physical memory | gauge | None
`windows_cs_hostname` | Labeled system hostname information | gauge | `hostname`, `domain`, `fqdn`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
