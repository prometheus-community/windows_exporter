# Installer

This page adds details on the MSI installer to the
[installation section](../README.md#installation) of the readme.

## Upgrades and repairs

The installer remembers the parameters of each install, `APPLICATIONFOLDER` included and
`SERVICE_PASSWORD` excepted, in the registry key `HKLM\SOFTWARE\prometheus-community\windows_exporter`.
An upgrade, a repair or another run of the
installer, for example by `winget upgrade`, keeps the earlier settings for every parameter that is
not on its command line. A parameter on the command line replaces the remembered value, and
`RESET_SETTINGS=1` returns every parameter that is not on the command line to its default, except
`APPLICATIONFOLDER`: only an explicit `APPLICATIONFOLDER` moves the installation. The uninstall removes
the remembered settings.

Installers released before this feature (0.32.2-alpha.0 and earlier) do not remember their
parameters: pass them once more on the first upgrade from such a version.

Remembered `ENABLED_COLLECTORS`, `EXTRA_FLAGS` and `TEXTFILE_DIRS` carry over into the new version.
If a release removes or renames a flag or a collector you use, pass the new values, or
`RESET_SETTINGS=1`, with the upgrade, or the service does not start.

An empty parameter on the command line counts as not set, so it keeps its remembered value. To
remove one setting, pass `RESET_SETTINGS=1` together with the parameters to keep.

A service that runs as a user account needs `SERVICE_PASSWORD` again on every upgrade, repair or
other run of the installer, since each of them configures the service again. Without it, the run
stops with an error before it changes anything.

Settings that were changed outside the installer, for example by editing the service's command line,
are not remembered. Use the config file for settings that the installer has no parameter for.
