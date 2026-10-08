# mcr.microsoft.com/oss/kubernetes/windows-host-process-containers-base-image:v1.0.0
# Using this image as a base for HostProcess containers has a few advantages over using other base images for Windows containers including:
# - Smaller image size
# - OS compatibility (works on any Windows version that supports containers)

# This image MUST be built with docker buildx build (buildx) command on a Linux system.
# Ref: https://github.com/microsoft/windows-host-process-containers-base-image

# The build context must contain the amd64 binary, named windows_exporter-<version>-amd64.exe
# like the CI build stages it. `make build-image` uses output/ as the build context.

ARG BASE="mcr.microsoft.com/oss/kubernetes/windows-host-process-containers-base-image:v1.0.0"
FROM $BASE

COPY windows_exporter*-amd64.exe /windows_exporter.exe
ENTRYPOINT ["windows_exporter.exe"]
