export GOOS=windows
export DOCKER_IMAGE_NAME ?= windows-exporter
export DOCKER_REPO ?= ghcr.io/prometheus-community

VERSION?=$(shell cat VERSION)
DOCKER?=docker

# Image Variables for Hostprocess Container
# Windows image build is heavily influenced by https://github.com/kubernetes/kubernetes/blob/master/cluster/images/etcd/Makefile
OS=1809
ALL_OS:= 1809 ltsc2022
BASE_IMAGE=mcr.microsoft.com/windows/nanoserver

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

.PHONY: promtool
promtool: windows_exporter.exe
	pwsh -NonInteractive -ExecutionPolicy Bypass -File .\tools\promtool.ps1

fmt:
	gofmt -l -w -s .

crossbuild:
	# The prometheus/golang-builder image for promu crossbuild doesn't exist
	# on Windows, so for now, we'll just build twice
	GOARCH=amd64 promu build --prefix=output/amd64
	GOARCH=386   promu build --prefix=output/386

build-image: crossbuild
	$(DOCKER) build --build-arg=BASE=$(BASE_IMAGE):$(OS) -f Dockerfile -t $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION)-$(OS) .

sub-build-%:
	$(MAKE) OS=$* build-image

build-all: $(addprefix sub-build-,$(ALL_OS))

push:
	set -x; \
	for osversion in ${ALL_OS}; do \
		$(DOCKER) push $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION)-$${osversion}; \
		$(DOCKER) manifest create --amend $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION) $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION)-$${osversion}; \
		full_version=`$(DOCKER) manifest inspect $(BASE_IMAGE):$${osversion} | grep "os.version" | head -n 1 | awk -F\" '{print $$4}'` || true; \
		$(DOCKER) manifest annotate --os windows --arch amd64 --os-version $${full_version} $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION)  $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION)-$${osversion}; \
	done
	$(DOCKER) manifest push --purge $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION)

push-all: build-all push
