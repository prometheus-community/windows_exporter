# SPDX-License-Identifier: MIT
#
# Copyright The Prometheus Authors

# Extracts windows_exporter.exe from an MSI built by build.ps1, without
# installing it. Release builds only sign the MSI, with the executable signed
# inside it, so the standalone executable is taken from the signed package.
#
# The package embeds one cabinet (see Media in main.wxs). It is read from the
# _Streams table through msi.dll and unpacked with expand.exe, which works for
# packages of any architecture, unlike an administrative install.
[CmdletBinding()]
Param (
    [Parameter(Mandatory = $true)]
    [String] $Msi,
    [Parameter(Mandatory = $true)]
    [String] $Destination
)
$ErrorActionPreference = "Stop"

$Cabinet = "windows_exporter.cab"
# Files in a WiX cabinet are named after their File Id in files.wxs.
$FileId = "windows_exporter.exe"

Add-Type -TypeDefinition @'
using System;
using System.IO;
using System.Runtime.InteropServices;

public static class MsiStream
{
    [DllImport("msi.dll", CharSet = CharSet.Unicode)]
    static extern uint MsiOpenDatabaseW(string path, IntPtr persist, out IntPtr db);
    [DllImport("msi.dll", CharSet = CharSet.Unicode)]
    static extern uint MsiDatabaseOpenViewW(IntPtr db, string query, out IntPtr view);
    [DllImport("msi.dll")]
    static extern uint MsiViewExecute(IntPtr view, IntPtr record);
    [DllImport("msi.dll")]
    static extern uint MsiViewFetch(IntPtr view, out IntPtr record);
    [DllImport("msi.dll")]
    static extern uint MsiRecordReadStream(IntPtr record, uint field, byte[] buffer, ref uint size);
    [DllImport("msi.dll")]
    static extern uint MsiCloseHandle(IntPtr handle);

    static void Check(uint result, string call)
    {
        if (result != 0) throw new InvalidOperationException(call + " failed with error " + result);
    }

    public static void Export(string msi, string stream, string output)
    {
        IntPtr db = IntPtr.Zero, view = IntPtr.Zero, record = IntPtr.Zero;
        try
        {
            // MSIDBOPEN_READONLY
            Check(MsiOpenDatabaseW(msi, IntPtr.Zero, out db), "MsiOpenDatabase");
            Check(MsiDatabaseOpenViewW(db, "SELECT `Data` FROM `_Streams` WHERE `Name` = '" + stream + "'", out view), "MsiDatabaseOpenView");
            Check(MsiViewExecute(view, IntPtr.Zero), "MsiViewExecute");
            // ERROR_NO_MORE_ITEMS means the stream does not exist.
            Check(MsiViewFetch(view, out record), "MsiViewFetch for stream " + stream);

            using (FileStream file = File.Create(output))
            {
                byte[] buffer = new byte[1 << 16];
                while (true)
                {
                    uint size = (uint)buffer.Length;
                    Check(MsiRecordReadStream(record, 1, buffer, ref size), "MsiRecordReadStream");
                    if (size == 0) break;
                    file.Write(buffer, 0, (int)size);
                }
            }
        }
        finally
        {
            if (record != IntPtr.Zero) MsiCloseHandle(record);
            if (view != IntPtr.Zero) MsiCloseHandle(view);
            if (db != IntPtr.Zero) MsiCloseHandle(db);
        }
    }
}
'@

$Msi = (Resolve-Path $Msi).Path
$Work = Join-Path ([IO.Path]::GetTempPath()) ([Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $Work | Out-Null
try {
    $CabinetPath = Join-Path $Work $Cabinet
    [MsiStream]::Export($Msi, $Cabinet, $CabinetPath)

    & expand.exe $CabinetPath "-F:$FileId" $Work | Out-Null
    # $ErrorActionPreference does not apply to native commands.
    if ($LASTEXITCODE -ne 0) {
        throw "expand.exe failed with exit code $LASTEXITCODE for $Msi"
    }

    $Extracted = Join-Path $Work $FileId
    if (-not (Test-Path -PathType Leaf $Extracted)) {
        throw "$Msi does not contain $FileId in $Cabinet"
    }

    $Parent = Split-Path -Parent $Destination
    if ($Parent) {
        New-Item -ItemType Directory -Force -Path $Parent | Out-Null
    }
    Move-Item -Force $Extracted $Destination
} finally {
    Remove-Item -Recurse -Force $Work -ErrorAction SilentlyContinue
}
