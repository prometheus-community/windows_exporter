# container collector

The container collector exposes metrics about containers running on a Hyper-V system

|||
-|-
Metric name prefix  | `container`
Data source         | [HCS](https://learn.microsoft.com/en-us/virtualization/api/hcs/overview), [Kubernetes CRI](https://kubernetes.io/docs/concepts/architecture/cri/)
Enabled by default? | No

## Flags

### `--collector.container.enabled`

Comma-separated list of collectors to use. Defaults to all, if not specified. Available collectors: `hcs`, `hostprocess`.

- `hcs` collects containers managed by the Host Compute Service, for example Docker or process-isolated containerd containers.
- `hostprocess` collects Kubernetes HostProcess containers, which run in Win32 job objects instead of HCS.

### `--collector.container.cri-endpoint`

Kubernetes Container Runtime Interface (CRI) endpoint. Defaults to `npipe:////./pipe/containerd-containerd`, the endpoint of containerd.

The collector reads the running containers and pod sandboxes from this endpoint to add the `namespace`, `pod` and `container` labels, to skip pause containers and to find HostProcess containers.
If the endpoint is not available, for example on hosts without Kubernetes, the HCS containers are exported without Kubernetes labels and HostProcess containers aren't collected.

### `--collector.container.containerd-state-dir`

Deprecated and ignored. Kubernetes metadata is read from `--collector.container.cri-endpoint`.

## Metrics

| Name                                                       | Description                            | Type    | Labels                                                     |
|------------------------------------------------------------|----------------------------------------|---------|------------------------------------------------------------|
| `windows_container_available`                              | Available                              | gauge   | `container_id`,`namespace`,`pod`,`container`,`hostprocess` |
| `windows_container_count`                                  | Number of running HCS containers       | gauge   | None                                                       |
| `windows_container_cpu_usage_seconds_kernelmode`           | Runtime in Kernel mode in Seconds      | counter | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_cpu_usage_seconds_usermode`             | Runtime in User mode in Seconds        | counter | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_cpu_usage_seconds_total`                | Total Runtime in Seconds               | counter | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_memory_usage_commit_bytes`              | Memory Usage Commit Bytes              | gauge   | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_memory_usage_commit_peak_bytes`         | Memory Usage Commit Peak Bytes         | gauge   | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_memory_usage_private_working_set_bytes` | Memory Usage Private Working Set Bytes | gauge   | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_network_receive_bytes_total`            | Bytes Received on Interface            | counter | `container_id`,`namespace`,`pod`,`container`,`interface`   |
| `windows_container_network_receive_packets_total`          | Packets Received on Interface          | counter | `container_id`,`namespace`,`pod`,`container`,`interface`   |
| `windows_container_network_receive_packets_dropped_total`  | Dropped Incoming Packets on Interface  | counter | `container_id`,`namespace`,`pod`,`container`,`interface`   |
| `windows_container_network_transmit_bytes_total`           | Bytes Sent on Interface                | counter | `container_id`,`namespace`,`pod`,`container`,`interface`   |
| `windows_container_network_transmit_packets_total`         | Packets Sent on Interface              | counter | `container_id`,`namespace`,`pod`,`container`,`interface`   |
| `windows_container_network_transmit_packets_dropped_total` | Dropped Outgoing Packets on Interface  | counter | `container_id`,`namespace`,`pod`,`container`,`interface`   |
| `windows_container_storage_read_count_normalized_total`    | Read Count Normalized                  | counter | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_storage_read_size_bytes_total`          | Read Size Bytes                        | counter | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_storage_write_count_normalized_total`   | Write Count Normalized                 | counter | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_storage_write_size_bytes_total`         | Write Size Bytes                       | counter | `container_id`,`namespace`,`pod`,`container`               |

### Example metric
_windows_container_network_receive_bytes_total{container_id="docker://1bd30e8b8ac28cbd76a9b697b4d7bb9d760267b0733d1bc55c60024e98d1e43e",interface="822179E7-002C-4280-ABBA-28BCFE401826"} 9.3305343e+07_

This metric means that total _9.3305343e+07_ bytes received on interface _822179E7-002C-4280-ABBA-28BCFE401826_ for container _docker://1bd30e8b8ac28cbd76a9b697b4d7bb9d760267b0733d1bc55c60024e98d1e43e_

## Useful queries
Attach labels namespace/pod/container for Windows container metrics.
```
# kube_pod_container_info(a metric of kube-state-metrics) has labels namespace/pod/container/container_id for a container, while Windows container metrics only have container_id.
# Attaching labels namespace/pod/container for Windows container metrics, is useful to query for Windows pods.
windows_container_network_receive_bytes_total * on(container_id) group_left(namespace, pod, container) kube_pod_container_info{container_id!=""}
```
## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
