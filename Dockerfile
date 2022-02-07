# Note this image doesn't really matter for hostprocess but it is good to build per OS version
# the files in the image are copied to $env:CONTAINER_SANDBOX_MOUNT_POINT on the host
# but the file system is the Host NOT the container
ARG BASE="mcr.microsoft.com/windows/nanoserver:1809"
FROM $BASE

ENV PATH="C:\Windows\system32;C:\Windows;"
COPY output/amd64/windows_exporter.exe /windows_exporter.exe
ENTRYPOINT ["windows_exporter.exe"]
