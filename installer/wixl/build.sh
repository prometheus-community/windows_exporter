#!/usr/bin/env bash
# Probe: build the windows_exporter MSI on Linux with wixl.
# Usage: build.sh <windows_exporter.exe> <msi-version> <amd64|arm64> <out.msi>
set -euo pipefail

exe=$(realpath "$1") version=$2 arch=$3 out=$(realpath -m "$4")
WIXL=${WIXL:-wixl}

cd "$(dirname "$0")"
mkdir -p Work
cp -f "$exe" Work/windows_exporter.exe

wixl_arch=$([[ $arch == amd64 ]] && echo x64 || echo arm64)
rm -f "$out"
"$WIXL" --ext ui ${WIXL_EXTDIR:+--extdir "$WIXL_EXTDIR"} -a "$wixl_arch" -D Version="$version" -o "$out" windows_exporter.wxs 2>Work/wixl.log || { cat Work/wixl.log >&2; exit 1; }
# wixl only warns about attributes it does not support, which then silently go missing.
if [[ -s Work/wixl.log ]]; then
	cat Work/wixl.log >&2
	exit 1
fi
