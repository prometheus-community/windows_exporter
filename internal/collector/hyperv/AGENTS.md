# hyperv agent notes

- `virtual_hard_disk` lists attached disks through WMI (`Msvm_ComputerSystem`,
  `Msvm_StorageAllocationSettingData` with `ResourceType = 31`, filtered by
  `ResourceSubType` because ISO and floppy images share the type) and reads
  sizes with `virtdisk.dll`. The VM GUID in `InstanceID`
  (`Microsoft:<owner GUID>\...`) drops settings that belong to checkpoints.
  Don't filter WMI by `Caption`; it is localized.
- It refreshes in the background (`refresher`), so a hanging disk read can't
  block `Collect`. `stop` waits for the WMI and PDH reads, which use resources
  released after `Close`, but not for disk reads.
- `Hyper-V Virtual Storage Device` instance names are the disk path with `\`
  replaced by `-` and `(…)` by `[…]` (`\\?\` becomes `--?-`, seen on `.vmgs`
  guest state files). Only running VMs have instances. A VHDX checkpoint
  (`.avhdx`) has its own instance; a VHD checkpoint (`.avhd`) is named after
  the base disk of its chain. The `device` label is matched against the real
  instance names, following the parent chain, instead of being derived only
  from the path.
