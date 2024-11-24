# msmq collector

The msmq collector exposes metrics about the queues on a MSMQ server

|                     |                      |
|---------------------|----------------------|
| Metric name prefix  | `msmq`               |
| Spource             | Performance Counters |
| Enabled by default? | No                   |

## Flags

## Metrics

| Name                                     | Description                     | Type  | Labels |
|------------------------------------------|---------------------------------|-------|--------|
| `windows_msmq_bytes_in_journal_queue`    | Size of queue journal in bytes  | gauge | `name` |
| `windows_msmq_bytes_in_queue`            | Size of queue in bytes          | gauge | `name` |
| `windows_msmq_messages_in_journal_queue` | Count messages in queue journal | gauge | `name` |
| `windows_msmq_messages_in_queue`         | Count messages in queue         | gauge | `name` |

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
