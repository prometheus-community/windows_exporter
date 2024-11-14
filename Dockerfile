# mcr.microsoft.com/oss/kubernetes/windows-host-process-containers-base-image:v1.0.0
# Using this image as a base for HostProcess containers has a few advantages over using other base images for Windows containers including:
# - Smaller image size
# - OS compatibility (works on any Windows version that supports containers)

# This image MUST be built with docker buildx build (buildx) command on a Linux system.
# Ref: https://github.com/microsoft/windows-host-process-containers-base-image

ARG BASE="mcr.microsoft.com/oss/kubernetes/windows-host-process-containers-base-image:v1.0.0"
FROM $BASE

COPY windows_exporter*-amd64.exe /windows_exporter.exe
ENTRYPOINT ["windows_exporter.exe"]
