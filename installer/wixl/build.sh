#!/usr/bin/env bash
# Probe: build the windows_exporter MSI on Linux with wixl, then add what wixl cannot express.
# Usage: build.sh <windows_exporter.exe> <msi-version> <amd64|arm64> <out.msi>
set -euo pipefail

exe=$(realpath "$1") version=$2 arch=$3 out=$(realpath -m "$4")
WIXL=${WIXL:-wixl}
MSIBUILD=${MSIBUILD:-msibuild}

cd "$(dirname "$0")"
mkdir -p Work
cp -f "$exe" Work/windows_exporter.exe

wixl_arch=$([[ $arch == amd64 ]] && echo x64 || echo arm64)
rm -f "$out"
"$WIXL" --ext ui ${WIXL_EXTDIR:+--extdir "$WIXL_EXTDIR"} -a "$wixl_arch" -D Version="$version" -o "$out" windows_exporter.wxs

# wixl ignores Property/@Secure, ServiceInstall/@Vital and has no ServiceConfig element.
secure="ENABLED_COLLECTORS;EXTRA_FLAGS;CONFIG_FILE;LISTEN_ADDR;LISTEN_PORT;METRICS_PATH;REMOTE_ADDR;TEXTFILE_DIRS;SERVICE_ACCOUNT;SERVICE_PASSWORD"
"$MSIBUILD" "$out" -q "UPDATE \`Property\` SET \`Value\` = '$secure;WIX_DOWNGRADE_DETECTED;WIX_SAME_VERSION_UPGRADE_DETECTED;WIX_UPGRADE_DETECTED' WHERE \`Property\` = 'SecureCustomProperties'"
# msidbServiceInstallErrorControlVital (0x8000) | SERVICE_ERROR_NORMAL (1)
"$MSIBUILD" "$out" -q "UPDATE \`ServiceInstall\` SET \`ErrorControl\` = 32769 WHERE \`ServiceInstall\` = 'InstallExporterService'"

# Native Windows Installer 5 service configuration, applied by the MsiConfigureServices action:
#   ConfigType 3 = delayed auto start, 4 = failure actions also on non-crash failures (sc failureflag 1)
#   Event 5 = install + reinstall
#   The failure actions themselves are set by a custom action, see windows_exporter.wxs.
"$MSIBUILD" "$out" -q "CREATE TABLE \`MsiServiceConfig\` (\`MsiServiceConfig\` CHAR(72) NOT NULL, \`Name\` CHAR(255) NOT NULL, \`Event\` SHORT NOT NULL, \`ConfigType\` LONG NOT NULL, \`Argument\` CHAR(0), \`Component_\` CHAR(72) NOT NULL PRIMARY KEY \`MsiServiceConfig\`)"
"$MSIBUILD" "$out" -q "INSERT INTO \`MsiServiceConfig\` (\`MsiServiceConfig\`, \`Name\`, \`Event\`, \`ConfigType\`, \`Argument\`, \`Component_\`) VALUES ('DelayedAutoStart', 'windows_exporter', 5, 3, '1', 'windows_exporter.exe')"
"$MSIBUILD" "$out" -q "INSERT INTO \`MsiServiceConfig\` (\`MsiServiceConfig\`, \`Name\`, \`Event\`, \`ConfigType\`, \`Argument\`, \`Component_\`) VALUES ('FailureActionsFlag', 'windows_exporter', 5, 4, '1', 'windows_exporter.exe')"
"$MSIBUILD" "$out" -q "INSERT INTO \`InstallExecuteSequence\` (\`Action\`, \`Condition\`, \`Sequence\`) VALUES ('MsiConfigureServices', 'VersionNT>=600', 5850)"
