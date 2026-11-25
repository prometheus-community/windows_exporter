# nsi headers

Read the root [`AGENTS.md`](../../../AGENTS.md) first.

`nsi.dll` is the client of the network store interface (NSI), the store
behind the IP Helper API and `netsh int ipv4|ipv6`. Microsoft does not
document it. Everything here comes from static analysis of the Windows
binaries, from probing a live system and from Wine's `nsi.dll`/`nsiproxy.sys`
reimplementation. Check new calls the same way before relying on them.

## Call path of netsh

Found by static analysis of build 26100.9818 (x64) in November 2026. No
binary was run for it.

```text
netsh.exe
  -> netiohlp.dll    command handler, e.g. show dynamicport (RVA 0x7DC0)
  -> winnsi.dll      NsiRpcGetAllParameters[Ex]
       RPC context NULL     -> nsi.dll!NsiGetAllParametersEx (local)
       RPC context not NULL -> rpcrt4.dll!Ndr64AsyncClientCall (remote)
  -> nsi.dll         builds the request, DeviceIoControl on \\.\Nsi
  -> nsiproxy.sys    validates the request size, forwards it
  -> netio.sys       NsiGetAllParametersEx, dispatches to the provider
  -> tcpip.sys       provider of the TCP and UDP modules (not traced)
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

## NsiGetAllParameters

Use it for reads. It is the call local netsh makes, and it needs no
`NSI_STRUCT_TYPE`, whose numeric values are not proven by the analysis.

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

| Store | Read                         | Write                              |
|-------|------------------------------|------------------------------------|
| 0     | persistent configuration     | netsh maps `store=persistent` to 2 |
| 1     | active configuration         | active configuration               |
| 2     | refused by the remote path   | persistent                         |

Read store 1: it holds the values in effect. Whether a write to store 2 also
changes the active values is not known.

## Objects

| Module | Object | Key  | Read-write part | Content                                 |
|--------|--------|------|-----------------|-----------------------------------------|
| TCP    | 20     | none | 4 bytes         | dynamic port range (`show dynamicport`) |
| UDP    | 4      | none | 4 bytes         | dynamic port range                      |

Dynamic port range layout: the first port as a big-endian `uint16` at offset 0,
the number of ports as a little-endian `uint16` at offset 2. Bytes read on
Windows Server 2025 build 26100: `c0 00 00 40` for the default 49152/16384,
`c0 01 ff 3f` after `netsh int ipv4 set dynamicport tcp start=49153
num=16383`. Windows 11 build 26300 read the same. The mixed byte order looks
odd, but the probe confirmed it; do not "fix" it.

There is one object per protocol, not per address family: setting the IPv4
range with netsh also changed the IPv6 range and every `MSFT_NetTCPSetting`
instance. Whether IPv4 and IPv6 also share one pool of ports in `tcpip.sys`
is not known.

## Open questions

- Whether the 4 bytes are the whole read-write part of the object, or only
  its start. A larger buffer has not been tried.
- Whether IPv4 and IPv6 share one ephemeral pool (needs `tcpip.sys`).
- Whether the object index and layout are the same on Windows Server 2016 and
  2019. CI covers Windows Server 2022 and Windows 11 ARM through the live
  test. Callers must keep a fallback (the tcp collector falls back to
  `MSFT_NetTCPSetting`).
- The excluded port ranges (`show excludedportrange`) have not been traced.
- `netio.sys` would answer the struct type values and the provider dispatch.

## Adding a call

1. Find the netsh command that shows the value and trace it in
   `netiohlp.dll` to the module, object index, store, key and buffer sizes.
2. Probe it on a live system and change the value with netsh to confirm the
   layout and byte order. Restore the value afterwards.
3. Add a layout test with the observed bytes and a live test with loose
   bounds, and add the function to `lazyproc` via `Load`.
4. Keep a fallback in the caller and drop only that sub-collector when both
   sources fail.
