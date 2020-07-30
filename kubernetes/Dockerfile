FROM mcr.microsoft.com/dotnet/framework/aspnet:4.8-windowsservercore-ltsc2019

COPY windows_exporter-0.13.0-amd64.exe C:

ENTRYPOINT [ "c:\\windows_exporter-0.13.0-amd64.exe" ]

EXPOSE 9100