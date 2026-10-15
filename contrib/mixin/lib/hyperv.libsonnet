// PromQL building blocks shared by the Hyper-V tabs. Selectors are label
// matchers without braces; every function returns a PromQL string.
{
  // Right-hand side of a join that adds vm and path to series of
  // windows_hyperv_virtual_storage_device_*. The virtual_hard_disk metrics carry
  // the counter instance name of each running disk in device. Several disks can
  // share one device (a VHD-format checkpoint uses the name of its base disk), so
  // one is kept per device. Turned into 1, so the join keeps the left values.
  vhdVm(selector):
    '(0 * topk by (instance, device) (1, max by (instance, device, vm, path) (windows_hyperv_virtual_hard_disk_size_bytes{%s, device!=""})) + 1)' % selector,

  // expr, a series of windows_hyperv_virtual_storage_device_*, with the vm and
  // path of its disk. Disks of other VMs drop out when vmSelector selects a VM.
  withVm(expr, vmSelector):
    '%s * on (instance, device) group_left (vm, path) %s' % [expr, self.vhdVm(vmSelector)],

  // Per-second rate of a windows_hyperv_virtual_storage_device_* counter.
  deviceRate(metric, selector):
    'rate(windows_hyperv_virtual_storage_device_%s{%s}[$__rate_interval])' % [metric, selector],

  // numerator per unit of denominator; idle devices drop out instead of
  // dividing by zero. Latency per I/O is
  // latency(rate(io_latency_seconds_total), rate(throughput_total)).
  latency(numerator, denominator): '%s / (%s > 0)' % [numerator, denominator],

  // Adds disk, the file name at the end of path, for legends and tables.
  diskName(expr):
    'label_replace(%s, "disk", "$1", "path", ".*[\\\\\\\\/]([^\\\\\\\\/]+)$")' % expr,

  // Hyper-V names the vNIC of a VM <vm>_<adapter>_<vm_id>--<adapter GUID>. Adds
  // vm_id from the name and vm from windows_hyperv_vm_processors, so vNICs
  // match their VM by GUID and not by a name prefix.
  vnicVm(expr, vmSelector):
    'label_replace(%s, "vm_id", "$1", "adapter", ".*_([0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12})--.*") * on (instance, vm_id) group_left (vm) (0 * max by (instance, vm_id, vm) (windows_hyperv_vm_processors{%s}) + 1)' % [expr, vmSelector],
}
