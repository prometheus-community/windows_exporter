// Checks the PromQL the Hyper-V helpers build. Run with jsonnet tests/hyperv.test.jsonnet.
local h = import '../lib/hyperv.libsonnet';
local config = (import '../config.libsonnet')._config;

local eq(name, got, want) =
  assert got == want : name + ':\n got  ' + got + '\n want ' + want;
  true;

std.all([
  eq(
    'vhdVm',
    h.vhdVm('job=~"$job", vm="$vm"'),
    '(0 * topk by (instance, device) (1, max by (instance, device, vm, path) (windows_hyperv_virtual_hard_disk_size_bytes{job=~"$job", vm="$vm", device!=""})) + 1)'
  ),
  eq(
    'withVm',
    h.withVm('rate(x[1m])', 'vm="a"'),
    'rate(x[1m]) * on (instance, device) group_left (vm, path) (0 * topk by (instance, device) (1, max by (instance, device, vm, path) (windows_hyperv_virtual_hard_disk_size_bytes{vm="a", device!=""})) + 1)'
  ),
  eq(
    'deviceRate',
    h.deviceRate('throughput_total', 'job=~"$job"'),
    'rate(windows_hyperv_virtual_storage_device_throughput_total{job=~"$job"}[$__rate_interval])'
  ),
  eq('latency', h.latency('a', 'b'), 'a / (b > 0)'),
  eq(
    'diskName',
    h.diskName('x'),
    'label_replace(x, "disk", "$1", "path", ".*[\\\\\\\\/]([^\\\\\\\\/]+)$")'
  ),
  eq(
    'vnicVm',
    h.vnicVm('rate(y[1m])', 'vm="a"'),
    'label_replace(rate(y[1m]), "vm_id", "$1", "adapter", ".*_([0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12})--.*") * on (instance, vm_id) group_left (vm) (0 * max by (instance, vm_id, vm) (windows_hyperv_vm_processor_count{vm="a"}) + 1)'
  ),
  eq('threshold', std.toString(config.hypervDiskLatencyThresholdSeconds), '0.01'),
])
