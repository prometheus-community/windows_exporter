# nsi headers

Read the root [`AGENTS.md`](../../../AGENTS.md) first.

`nsi.dll` is the client of the network store interface (NSI), the store
behind the IP Helper API and `netsh int ipv4|ipv6`. Microsoft does not
document it. Everything here comes from static analysis of the Windows
binaries, from probing a live system and from Wine's `nsi.dll`/`nsiproxy.sys`
reimplementation. Check new calls the same way before relying on them.

## Call path of netsh

Found by static analysis of the x64 binaries of build 26100 (user mode
26100.9818, `netio.sys` 26100.9974) in November 2026, later checked against
the public PDBs from the Microsoft symbol server. No binary was run for it.
Function and object names in backticks below are Microsoft symbol names; the
public PDBs carry no structure types, so structure fields are descriptive.

```text
netsh.exe
  -> netiohlp.dll    command handler, e.g. NhIpHandleShowDynamicPortRange
  -> winnsi.dll      NsiRpcGetAllParameters[Ex]
       RPC context NULL     -> nsi.dll!NsiGetAllParametersEx (local)
       RPC context not NULL -> rpcrt4.dll!Ndr64AsyncClientCall (remote)
  -> nsi.dll         builds the request, DeviceIoControl on \\.\Nsi
  -> nsiproxy.sys    validates the request size, forwards it
  -> netio.sys       NsiGetAllParametersEx, looks up the object descriptor,
                     checks the sizes, calls the provider callback
  -> tcpip.sys       provider of the TCP and UDP modules
```

`netiohlp.dll` keeps the RPC context in a global. It is only set by
`NsiConnectToServer` when netsh targets another machine (`netsh -r`), so a
local netsh call ends in `nsi.dll` like ours. `winnsi.dll` adds no local
logic, so call `nsi.dll` directly: no RPC, no MI, no cgo.

| Operation  | nsi.dll export        | IOCTL      | Request (x64) | netio.sys export        |
|------------|-----------------------|------------|---------------|-------------------------|
| Get one    | `NsiGetParameter`     | `0x120007` | 80 bytes      | `NsiGetParameterEx`     |
| Get all    | `NsiGetAllParameters` | `0x12000F` | 104 bytes     | `NsiGetAllParametersEx` |
| Set all    | `NsiSetAllParameters` | `0x120013` | 72 bytes      | `NsiSetAllParametersEx` |

`nsiproxy.sys` has separate paths for 32-bit callers (`IoIs32bitProcess`).
Use the exported functions; never build the IOCTL requests by hand.

## Object descriptors in netio.sys

Every provider registers one 104-byte (`0x68`) descriptor per object.
`netio.sys` finds it by object index and checks the caller's key and buffer
sizes against it:

| Offset | Field                                    |
|--------|------------------------------------------|
| `0x00` | key size                                 |
| `0x04` | read-write part size (struct type 0)     |
| `0x08` | read-only dynamic part size (type 1)     |
| `0x0C` | read-only static part size (type 2)      |
| `0x20` | single-parameter get callback, or NULL   |
| `0x30` | get-all callback                         |
| `0x38` | set-all callback                         |

So `NSI_STRUCT_TYPE` is 0 = read-write, 1 = dynamic, 2 = static; larger
values are refused. `NsiGetParameterEx` checks `Offset + ParamLen` against the
part size. When the single-parameter callback is NULL, it calls the get-all
callback and copies `ParamLen` bytes from `Offset`.

## NsiGetParameter

The package reads with it. It is the simpler call and, for objects without a
single-parameter callback such as the dynamic port range, reaches the same
provider callback as netsh through the fallback above. Both calls returned the
same bytes on Windows Server 2025 and passed the live test in CI on Windows
Server 2022 and Windows 11 ARM.

```c
DWORD NsiGetParameter(
    DWORD Store, const NPI_MODULEID *ModuleId, DWORD ObjectIndex,
    const void *Key, DWORD KeyLen,
    DWORD StructType,                 // 0 = read-write part
    void *Param, DWORD ParamLen, DWORD Offset);
```

Nine arguments, all passed through `syscall.SyscallN`. It returns a Win32
error code in the low 32 bits; compare `uint32(r0)`, not `r0`. The module ID
and the buffer only need to live for the call: `SyscallN` keeps the Go
pointers alive until it returns, and nsi.dll keeps no reference.

## NsiGetAllParameters

The call local netsh makes. Use it when you need several parts of an object
at once.

```c
DWORD NsiGetAllParameters(
    DWORD Store, const NPI_MODULEID *ModuleId, DWORD ObjectIndex,
    const void *Key, DWORD KeyLen,
    void *Rw, DWORD RwLen,            // read-write part
    void *RoDynamic, DWORD RoDynamicLen,
    void *RoStatic, DWORD RoStaticLen);
```

Eleven arguments, all passed through `syscall.SyscallN`. It returns a Win32
error code in the low 32 bits; compare `uint32(r0)`, not `r0`. Pass `0` and
length `0` for the parts you do not need.

`nsi.dll` turns this into the 104-byte request below (offsets for x64):

| Offset | Field                                         |
|--------|-----------------------------------------------|
| `0x00` | 16 bytes header, zero                         |
| `0x10` | `NPI_MODULEID *`                              |
| `0x18` | object index                                  |
| `0x20` | store                                         |
| `0x24` | unknown, zero                                 |
| `0x28` | key pointer, key length at `0x30`             |
| `0x38` | read-write buffer, length at `0x40`           |
| `0x48` | read-only dynamic buffer, length at `0x50`    |
| `0x58` | read-only static buffer, length at `0x60`     |

## NPI_MODULEID

24 bytes: `USHORT Length` (24), 2 bytes padding, `ULONG Type` (`MIT_GUID` = 1)
at offset 4, then the GUID at offset 8 (a union with a LUID). `npiModuleID`
has a layout test. The module IDs are in Wine's `nsi.h`:

| Module | GUID                                     |
|--------|------------------------------------------|
| TCP    | `{eb004a03-9b1a-11d4-9123-0050047759bc}` |
| UDP    | `{eb004a02-9b1a-11d4-9123-0050047759bc}` |

## Stores

From `NsiGetAllParametersEx` and `NsiSetAllParametersEx` in `netio.sys`:

| Store | Read                                   | Write                         |
|-------|----------------------------------------|-------------------------------|
| 0     | persistent configuration               | persistent only               |
| 1     | active configuration                   | active only                   |
| 2     | invalid parameter                      | active, then persistent       |
| 3     | active provider, else persistent       | invalid parameter             |
| 4     | separate get-all path, purpose unknown | invalid parameter             |

Read store 1: it holds the values in effect. netsh writes `store=persistent`
with store 2, which updates the active value first and then the persistent
one. The two updates are separate, not atomic.

The persistent values live in the registry under
`HKLM\SYSTEM\CurrentControlSet\Control\Nsi\{module GUID}\{object index}`.
Do not read them there: they are the configuration for the next boot, not
necessarily the values in effect.

## Objects

| Module | Object | Key     | Read-write part | tcpip.sys descriptor                         | Content                                  |
|--------|--------|---------|-----------------|----------------------------------------------|------------------------------------------|
| TCP    | 20     | none    | 4 bytes         | `InetPortPoolInformationObject`              | dynamic port range (`show dynamicport`)  |
| UDP    | 4      | none    | 4 bytes         | `InetPortPoolInformationObject`              | dynamic port range                       |
| TCP    | 34     | none    | 44 bytes        | `InetPortPoolExInformationObject`            | extended port pool settings, not decoded |
| TCP    | 29     | 4 bytes | 2 bytes         | `InetPortExclusionGrantingInformationObject` | excluded port ranges, enumerated         |
| UDP    | 5      | 4 bytes | 2 bytes         | `InetPortExclusionGrantingInformationObject` | excluded port ranges, enumerated         |

netsh handles them in `NhIpHandleShowDynamicPortRange`,
`NhIpHandleSetDynamicPortRange` and `NhIpHandleShowExcludedPortRange`.
Object 34 is read by `InetGetAllPortPoolExParameters` and relates to
`InetIsPortInSmartRange`; its layout is unknown, so do not use it yet. Port
reservations (`InetPortReservationInformationObject`) are separate objects
from exclusions. The object indices of objects 29 and 5 come from static
analysis only and have not been probed.

Dynamic port range layout: the first port as a big-endian `uint16` at offset 0,
the number of ports as a little-endian `uint16` at offset 2. The descriptor in
`tcpip.sys` (one template, registered for TCP object 20 and UDP object 4)
has no key, a 4-byte read-write part, no read-only parts and no
single-parameter callback, so the 4 bytes are the whole object. The get-all
callback `InetGetAllPortPoolParameters` copies them from the port pool without
any byte swapping and without an address family. Bytes read with `NsiGetParameter` on
Windows Server 2025 build 26100: `c0 00 00 40` for the default 49152/16384,
`c0 01 ff 3f` after `netsh int ipv4 set dynamicport tcp start=49153
num=16383`. Windows 11 build 26300 read the same. The mixed byte order looks
odd, but the probe and the kernel confirm it; do not "fix" it.

`tcpip.sys` starts with the default `c0 00 00 40`, then loads the persistent
range and keeps it if it is valid. The same check runs when the active range
is changed (`ValidatePortPoolRw`): start port at least 1024, at least 255 ports, and start plus
count at most 65536.

There is one object per protocol, not per address family: setting the IPv4
range with netsh also changed the IPv6 range and every `MSFT_NetTCPSetting`
instance.

## Port pools and compartments

`tcpip.sys` keeps one port pool per protocol and network compartment, not per
address family. `TcpGetPortPoolInCurrentCompartment` and
`UdpGetPortPoolInCurrentCompartment` take the compartment of the calling
thread (from `TcpCompartmentSet` and `UdpCompartmentSet`) and no address
family. The bind handlers of TCP and UDP take a wildcard port from that pool
with `InetAcquireWildcardPortUnderLock`, for IPv4 and IPv6 alike. So IPv4 and
IPv6 connections most likely draw from one pool (strong static evidence, not
measured); the rules for the same port number in both families have not been
traced.

NSI reads go to the compartment of the calling thread. The exporter runs in
the default compartment (1); containers with their own compartment have their
own pools, which the exporter's NSI read does not see. Whether
`GetExtendedTcpTable` lists connections of other compartments has not been
checked.

## Open questions

- Whether a port number held by an IPv4 connection can also be handed to an
  IPv6 connection. The pool is shared; the per-port rules are not traced.
- The layout of TCP object 34 (extended port pool).
- Whether the object index and layout are the same on Windows Server 2016 and
  2019. CI covers Windows Server 2022 and Windows 11 ARM through the live
  test. Callers log a warning and skip the metric when `Load` or a read
  fails; they never substitute a default.
- The layout of the excluded port ranges (objects 29 and 5) has not been
  probed on a live system.
- What stores 3 and 4 are for.

## Adding a call

1. Find the netsh command that shows the value and trace it in
   `netiohlp.dll` to the module, object index, store, key and buffer sizes.
   The provider's descriptor (see above) gives the exact part sizes.
2. Probe it on a live system and change the value with netsh to confirm the
   layout and byte order. Restore the value afterwards.
3. Add a layout test with the observed bytes and a live test with loose
   bounds. Add the proc to the `lazyproc.Once` list of the package, which
   `Load` checks, and call `Load` in the collector's `Build`.
4. In the caller, log a warning and drop only that metric or sub-collector
   when `Load` fails, and return the error without a value when a read fails.
   Do not substitute a default and do not fall back to MI.
