# SPDX-License-Identifier: MIT
#
# Copyright The Prometheus Authors

# Extracts windows_exporter.exe from an MSI built by build.ps1, without
# installing it. Release builds only sign the MSI, with the executable signed
# inside it, so the standalone executable is taken from the signed package.
#
# The package embeds one cabinet (see Media in main.wxs). Its name comes from
# the Media table, since a signing service may repackage the MSI, and it is
# read from the _Streams table through msi.dll and unpacked with the tar.exe of Windows, which works for
# packages of any architecture, unlike an administrative install. Files in the
# cabinet are stored under their File table key, not their file name.
[CmdletBinding()]
Param (
    [Parameter(Mandatory = $true)]
    [String] $Msi,
    [Parameter(Mandatory = $true)]
    [String] $Destination
)
$ErrorActionPreference = "Stop"

$FileName = "windows_exporter.exe"

# The script may run several times in one PowerShell session.
if (-not ('MsiPackage' -as [type])) {
Add-Type -TypeDefinition @'
using System;
using System.IO;
using System.Runtime.InteropServices;
using System.Text;

public static class MsiPackage
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
    static extern uint MsiRecordReadStream(IntPtr record, uint field, [Out] byte[] buffer, ref uint size);
    [DllImport("msi.dll", CharSet = CharSet.Unicode)]
    static extern uint MsiRecordGetStringW(IntPtr record, uint field, StringBuilder value, ref uint size);
    [DllImport("msi.dll")]
    static extern uint MsiCloseHandle(IntPtr handle);

    const uint ERROR_NO_MORE_ITEMS = 259;

    static void Check(uint result, string call)
    {
        if (result != 0) throw new InvalidOperationException(call + " failed with error " + result);
    }

    // Calls read for each row of a query on the package, opened read-only.
    static void Query(string msi, string query, Action<IntPtr> read)
    {
        IntPtr db = IntPtr.Zero, view = IntPtr.Zero;
        try
        {
            // IntPtr.Zero is MSIDBOPEN_READONLY.
            Check(MsiOpenDatabaseW(msi, IntPtr.Zero, out db), "MsiOpenDatabase");
            Check(MsiDatabaseOpenViewW(db, query, out view), "MsiDatabaseOpenView");
            Check(MsiViewExecute(view, IntPtr.Zero), "MsiViewExecute");
            while (true)
            {
                IntPtr record;
                uint result = MsiViewFetch(view, out record);
                if (result == ERROR_NO_MORE_ITEMS) break;
                Check(result, "MsiViewFetch");
                try { read(record); } finally { MsiCloseHandle(record); }
            }
        }
        finally
        {
            if (view != IntPtr.Zero) MsiCloseHandle(view);
            if (db != IntPtr.Zero) MsiCloseHandle(db);
        }
    }

    static string GetString(IntPtr record, uint field)
    {
        StringBuilder value = new StringBuilder(256);
        uint size = (uint)value.Capacity;
        Check(MsiRecordGetStringW(record, field, value, ref size), "MsiRecordGetString");
        return value.ToString();
    }

    // Returns the File table key of the file with the given long name.
    public static string FileKey(string msi, string fileName)
    {
        string key = null;
        Query(msi, "SELECT `File`, `FileName` FROM `File`", delegate (IntPtr record)
        {
            // FileName is "short|long", or only the name if it is a short name.
            string name = GetString(record, 2);
            int separator = name.IndexOf('|');
            if (separator >= 0) name = name.Substring(separator + 1);
            if (!string.Equals(name, fileName, StringComparison.OrdinalIgnoreCase)) return;
            if (key != null) throw new InvalidOperationException(msi + " contains more than one " + fileName);
            key = GetString(record, 1);
        });
        if (key == null) throw new InvalidOperationException(msi + " does not contain " + fileName);
        return key;
    }

    // Writes an embedded stream, such as a cabinet, to a file.
    public static void ExportStream(string msi, string stream, string output)
    {
        bool found = false;
        Query(msi, "SELECT `Data` FROM `_Streams` WHERE `Name` = '" + stream + "'", delegate (IntPtr record)
        {
            found = true;
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
        });
        if (!found) throw new InvalidOperationException(msi + " has no stream " + stream);
    }

    // Returns the stream name of the only cabinet, which must be embedded.
    public static string EmbeddedCabinet(string msi)
    {
        string cabinet = null;
        Query(msi, "SELECT `Cabinet` FROM `Media`", delegate (IntPtr record)
        {
            string name = GetString(record, 1);
            if (name == "") return;
            if (cabinet != null) throw new InvalidOperationException(msi + " has more than one cabinet");
            if (!name.StartsWith("#")) throw new InvalidOperationException(msi + " uses the external cabinet " + name);
            cabinet = name.Substring(1);
        });
        if (cabinet == null) throw new InvalidOperationException(msi + " has no cabinet");
        return cabinet;
    }
}
'@
}

$Msi = (Resolve-Path $Msi).Path
$Work = Join-Path ([IO.Path]::GetTempPath()) ([Guid]::NewGuid().ToString('N'))
$Files = Join-Path $Work "files"
New-Item -ItemType Directory -Path $Files | Out-Null
try {
    $FileKey = [MsiPackage]::FileKey($Msi, $FileName)
    $Cabinet = [MsiPackage]::EmbeddedCabinet($Msi)
    $CabinetPath = Join-Path $Work $Cabinet
    [MsiPackage]::ExportStream($Msi, $Cabinet, $CabinetPath)

    # expand.exe does not unpack this cabinet. The tar.exe that ships with
    # Windows is libarchive and reads cabinets; the tar of Git for Windows,
    # which can come first in PATH, does not.
    $Tar = Join-Path $env:SystemRoot "System32\tar.exe"
    & $Tar -x -v -f $CabinetPath -C $Files
    # $ErrorActionPreference does not apply to native commands.
    if ($LASTEXITCODE -ne 0) {
        throw "tar.exe failed with exit code $LASTEXITCODE for $Cabinet in $Msi"
    }

    $Extracted = Join-Path $Files $FileKey
    if (-not (Test-Path -PathType Leaf $Extracted)) {
        $Found = (Get-ChildItem -Path $Files -Name) -join ', '
        throw "$Cabinet in $Msi does not contain $FileKey ($FileName), found: $Found"
    }

    $Parent = Split-Path -Parent $Destination
    if ($Parent) {
        New-Item -ItemType Directory -Force -Path $Parent | Out-Null
    }
    Move-Item -Force $Extracted $Destination
} finally {
    Remove-Item -Recurse -Force $Work -ErrorAction SilentlyContinue
}
