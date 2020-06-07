# netframework_clrloading collector

The netframework_clrloading collector exposes metrics about the dotnet loader.

|||
-|-
Metric name prefix  | `netframework_clrloading`
Classes             | `Win32_PerfRawData_NETFramework_NETCLRLoading`
Enabled by default? | No

## Flags

None

## Metrics

Name | Description | Type | Labels
-----|-------------|------|-------
`windows_netframework_clrloading_loader_heap_size_bytes` | Displays the current size, in bytes, of the memory committed by the class loader across all application domains. Committed memory is the physical space reserved in the disk paging file. | gauge | `process`
`windows_netframework_clrloading_appdomains_loaded_current` | Displays the current number of application domains loaded in this application. | gauge | `process`
`windows_netframework_clrloading_assemblies_loaded_current` | Displays the current number of assemblies loaded across all application domains in the currently running application. If the assembly is loaded as domain-neutral from multiple application domains, this counter is incremented only once. | gauge | `process`
`windows_netframework_clrloading_classes_loaded_current` | Displays the current number of classes loaded in all assemblies. | gauge | `process`
`windows_netframework_clrloading_appdomains_loaded_total` | Displays the peak number of application domains loaded since the application started. | counter | `process`
`windows_netframework_clrloading_appdomains_unloaded_total` | Displays the total number of application domains unloaded since the application started. If an application domain is loaded and unloaded multiple times, this counter increments each time the application domain is unloaded. | counter | `process`
`windows_netframework_clrloading_assemblies_loaded_total` | Displays the total number of assemblies loaded since the application started. If the assembly is loaded as domain-neutral from multiple application domains, this counter is incremented only once. | counter | `process`
`windows_netframework_clrloading_classes_loaded_total` | Displays the cumulative number of classes loaded in all assemblies since the application started. | counter | `process`
`windows_netframework_clrloading_class_load_failures_total` | Displays the peak number of classes that have failed to load since the application started. | counter | `process`

### Example metric
_This collector does not yet have explained examples, we would appreciate your help adding them!_

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
