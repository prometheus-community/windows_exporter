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
  Kubernetes containers with Hyper-V isolation are collected from the CRI endpoint, see [Hyper-V isolated containers](#hyper-v-isolated-containers).
- `hostprocess` collects Kubernetes HostProcess containers, which run in Win32 job objects instead of HCS.

### `--collector.container.cri-endpoint`

Kubernetes Container Runtime Interface (CRI) endpoint. Defaults to `npipe:////./pipe/containerd-containerd`, the endpoint of containerd.

The collector reads the running containers and pod sandboxes from this endpoint to add the `namespace`, `pod` and `container` labels, to skip pause containers and to find HostProcess containers.
It also reads the container stats for the writable layer usage and the metrics of Hyper-V isolated containers.
If the endpoint is not available, for example on hosts without Kubernetes, the HCS containers are exported without Kubernetes labels and HostProcess and Hyper-V isolated containers aren't collected.

## Metrics

| Name                                                       | Description                                         | Type    | Labels                                                     |
|------------------------------------------------------------|-----------------------------------------------------|---------|------------------------------------------------------------|
| `windows_container_available`                              | Available                                           | gauge   | `container_id`,`namespace`,`pod`,`container`,`hostprocess` |
| `windows_container_count`                                  | Number of running HCS containers                    | gauge   | None                                                       |
| `windows_container_cpu_usage_seconds_kernelmode`           | Runtime in Kernel mode in Seconds                   | counter | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_cpu_usage_seconds_usermode`             | Runtime in User mode in Seconds                     | counter | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_cpu_usage_seconds_total`                | Total Runtime in Seconds                            | counter | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_memory_page_faults_total`               | Total number of page faults (HostProcess only)      | counter | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_memory_usage_commit_bytes`              | Memory Usage Commit Bytes                           | gauge   | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_memory_usage_commit_peak_bytes`         | Memory Usage Commit Peak Bytes                      | gauge   | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_memory_usage_private_working_set_bytes` | Memory Usage Private Working Set Bytes              | gauge   | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_network_receive_bytes_total`            | Bytes Received on Interface                         | counter | `container_id`,`namespace`,`pod`,`container`,`interface`   |
| `windows_container_network_receive_packets_total`          | Packets Received on Interface                       | counter | `container_id`,`namespace`,`pod`,`container`,`interface`   |
| `windows_container_network_receive_packets_dropped_total`  | Dropped Incoming Packets on Interface               | counter | `container_id`,`namespace`,`pod`,`container`,`interface`   |
| `windows_container_network_transmit_bytes_total`           | Bytes Sent on Interface                             | counter | `container_id`,`namespace`,`pod`,`container`,`interface`   |
| `windows_container_network_transmit_packets_total`         | Packets Sent on Interface                           | counter | `container_id`,`namespace`,`pod`,`container`,`interface`   |
| `windows_container_network_transmit_packets_dropped_total` | Dropped Outgoing Packets on Interface               | counter | `container_id`,`namespace`,`pod`,`container`,`interface`   |
| `windows_container_processes`                              | Number of processes running in the container        | gauge   | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_start_time_seconds`                     | Start time of the container since Unix epoch        | gauge   | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_storage_read_count_normalized_total`    | Read Count Normalized                               | counter | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_storage_read_size_bytes_total`          | Read Size Bytes                                     | counter | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_storage_write_count_normalized_total`   | Write Count Normalized                              | counter | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_storage_write_size_bytes_total`         | Write Size Bytes                                    | counter | `container_id`,`namespace`,`pod`,`container`               |
| `windows_container_storage_writable_layer_usage_bytes`     | Used bytes of the writable layer                    | gauge   | `container_id`,`namespace`,`pod`,`container`               |

`windows_container_start_time_seconds` is the start time reported by HCS. HostProcess and Hyper-V isolated containers have no start time in HCS on the host, they report their creation time from the CRI endpoint instead.

`windows_container_memory_page_faults_total` is only available for HostProcess containers, because HCS doesn't report page faults. The job object counts page faults in 32 bits, so the counter wraps around after 2^32 page faults.

`windows_container_storage_writable_layer_usage_bytes` is read from the CRI endpoint and only exported for Kubernetes containers.
containerd measures it periodically, so it may lag behind and is missing until the first measurement.

`windows_container_count` only counts HCS containers. HostProcess and Hyper-V isolated containers aren't counted.

### Hyper-V isolated containers

Kubernetes containers with Hyper-V isolation run inside a utility VM. HCS on the host only sees the utility VM, not the containers inside it.
The collector reads their metrics from the container stats of the CRI endpoint instead, if the `hcs` collector is enabled.
containerd fills them from the HCS statistics that the shim queries inside the utility VM, but the CRI API only carries a subset of them:

| Name                                                       | Source                                 |
|------------------------------------------------------------|----------------------------------------|
| `windows_container_available`                              | always `1`, with `hostprocess="false"` |
| `windows_container_cpu_usage_seconds_total`                | `cpu.usage_core_nano_seconds`          |
| `windows_container_memory_usage_private_working_set_bytes` | `memory.working_set_bytes`             |
| `windows_container_memory_usage_commit_bytes`              | `memory.usage_bytes`                   |
| `windows_container_start_time_seconds`                     | creation time of the container         |
| `windows_container_storage_writable_layer_usage_bytes`     | `writable_layer.used_bytes`            |

The user and kernel mode CPU time, the commit peak, page faults, the process count, storage I/O and network metrics aren't available for Hyper-V isolated containers.
The collector identifies them as running Kubernetes containers that are neither HCS containers nor HostProcess containers.

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
