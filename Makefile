export GOOS=windows
export DOCKER_IMAGE_NAME ?= windows-exporter
DOCKER_REPO:= ghcr.io/prometheus-community docker.io/prometheuscommunity quay.io/prometheuscommunity

VERSION?=$(shell cat VERSION)
DOCKER?=docker

# Image Variables for Hostprocess Container
# Windows image build is heavily influenced by https://github.com/kubernetes/kubernetes/blob/master/cluster/images/etcd/Makefile
OS=ltsc2019
ALL_OS:= ltsc2019 ltsc2022
BASE_IMAGE=mcr.microsoft.com/windows/nanoserver
BASE_HOST_PROCESS_IMAGE=mcr.microsoft.com/oss/kubernetes/windows-host-process-containers-base-image:v1.0.0

.PHONY: build
build: generate windows_exporter.exe

windows_exporter.exe: pkg/**/*.go
	promu build -v

.PHONY: generate
generate:
	go generate ./...

test:
	go test -v ./...

bench:
	go test -v -bench='benchmarkcollector' ./pkg/collector/{cpu,logical_disk,physical_disk,logon,memory,net,process,service,system,tcp,time}

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

crossbuild: generate
	# The prometheus/golang-builder image for promu crossbuild doesn't exist
	# on Windows, so for now, we'll just build twice
	GOARCH=amd64 promu build --prefix=output/amd64
	GOARCH=arm64 promu build --prefix=output/arm64

build-image: crossbuild
	$(DOCKER) build --build-arg=BASE=$(BASE_IMAGE):$(OS) -f Dockerfile -t local/$(DOCKER_IMAGE_NAME):$(VERSION)-$(OS) .

build-hostprocess-image: crossbuild
	$(DOCKER) build --build-arg=BASE=$(BASE_HOST_PROCESS_IMAGE) -f Dockerfile -t local/$(DOCKER_IMAGE_NAME):$(VERSION)-hostprocess .

sub-build-%:
	$(MAKE) OS=$* build-image

build-all: $(addprefix sub-build-,$(ALL_OS)) build-hostprocess-image

push:
	set -x; \
	for repo in ${DOCKER_REPO}; do \
		for osversion in ${ALL_OS}; do \
			$(DOCKER) tag local/$(DOCKER_IMAGE_NAME):$(VERSION)-$${osversion} $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION)-$${osversion}; \
			$(DOCKER) push $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION)-$${osversion}; \
			$(DOCKER) manifest create --amend $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION) $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION)-$${osversion}; \
			full_version=`$(DOCKER) manifest inspect $(BASE_IMAGE):$${osversion} | grep "os.version" | head -n 1 | awk -F\" '{print $$4}'` || true; \
			$(DOCKER) manifest annotate --os windows --arch amd64 --os-version $${full_version} $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION)  $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION)-$${osversion}; \
		done
		$(DOCKER) manifest push --purge $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION); \

		$(DOCKER) tag local/$(DOCKER_IMAGE_NAME):$(VERSION)-hostprocess $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION)-hostprocess; \
		$(DOCKER) push $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(VERSION)-hostprocess; \
	done

push-all: build-all push
