module github.com/prometheus-community/windows_exporter

go 1.23

require (
	github.com/Microsoft/hcsshim v0.12.6
	github.com/alecthomas/kingpin/v2 v2.4.0
	github.com/dimchansky/utfbom v1.1.1
	github.com/go-ole/go-ole v1.3.0
	github.com/google/uuid v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.20.3
	github.com/prometheus/client_model v0.6.1
	github.com/prometheus/common v0.59.1
	github.com/prometheus/exporter-toolkit v0.13.0
	github.com/stretchr/testify v1.9.0
	github.com/yusufpapurcu/wmi v1.2.4
	golang.org/x/sys v0.25.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/alecthomas/units v0.0.0-20240626203959-61d1e3462e30 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/containerd/cgroups/v3 v3.0.3 // indirect
	github.com/containerd/errdefs v0.1.0 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/mdlayher/socket v0.5.1 // indirect
	github.com/mdlayher/vsock v1.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/oauth2 v0.23.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/grpc v1.66.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

// https://github.com/prometheus/common/pull/694
replace github.com/prometheus/common v0.59.1 => github.com/jkroepke/prometheus-common v0.0.0-20240907211841-5f9af24b97ad
