# user_input_delay collector

The user_input_delay collector exposes metrics about user input delays.

|||
-|-
Metric name prefix  | `user_input_delay`
Data source         | Perflib
Enabled by default? | No

## Flags


## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_user_input_delay_session` | Maximum value for queuing delay across all user input waiting to be picked-up by any process in the session during a target time interval | gauge | `session_id`
`windows_user_input_delay_process` | Maximum value for queuing delay across all user input waiting to be picked-up by the process during a target time interval | gauge | `sid`, `pid`, `pname`

### Example metric
Show average delay for all sessions
```
avg(windows_user_input_delay_session)
```

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
