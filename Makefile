export GOOS=windows

.PHONY: build
build: windows_exporter.exe
windows_exporter.exe: **/*.go
	promu build -v

test:
	go test -v ./...

bench:
	go test -v -bench='benchmark(cpu|logicaldisk|logon|memory|net|process|service|system|tcp|time)collector' ./...

lint:
	golangci-lint -c .golangci.yaml run

.PHONY: e2e-test
e2e-test: windows_exporter.exe
	pwsh -NonInteractive -ExecutionPolicy Bypass -File .\tools\end-to-end-test.ps1

fmt:
	gofmt -l -w -s .

crossbuild:
	# The prometheus/golang-builder image for promu crossbuild doesn't exist
	# on Windows, so for now, we'll just build twice
	GOARCH=amd64 promu build --prefix=output/amd64
	GOARCH=386   promu build --prefix=output/386
