module github.com/prometheus-community/windows_exporter

go 1.13

require (
	github.com/Microsoft/go-winio v0.4.14
	github.com/Microsoft/hcsshim v0.8.6
	github.com/StackExchange/wmi v0.0.0-20180725035823-b12b22c5341f
	github.com/dimchansky/utfbom v1.1.0
	github.com/elastic/go-sysinfo v1.4.0
	github.com/elastic/go-windows v1.0.1 // indirect
	github.com/go-ole/go-ole v1.2.1 // indirect
	github.com/leoluk/perflib_exporter v0.1.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v0.9.2
	github.com/prometheus/client_model v0.0.0-20180712105110-5c3871d89910
	github.com/prometheus/common v0.2.0
	github.com/prometheus/procfs v0.2.0 // indirect
	golang.org/x/sys v0.0.0-20201029080932-201ba4db2418
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	howett.net/plist v0.0.0-20201026045517-117a925f2150 // indirect
)
