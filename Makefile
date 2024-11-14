GOOS    ?= windows
VERSION ?= $(shell cat VERSION)
DOCKER  ?= docker

# DOCKER_REPO is the official image repository name at docker.io, quay.io.
DOCKER_REPO       ?= prometheuscommunity
DOCKER_IMAGE_NAME ?= windows-exporter

# ALL_DOCKER_REPOS is the list of repositories to push the image to. ghcr.io requires that org name be the same as the image repo name.
ALL_DOCKER_REPOS  ?= docker.io/$(DOCKER_REPO) ghcr.io/prometheus-community # quay.io/$(DOCKER_REPO)

# Image Variables for host process Container
# Windows image build is heavily influenced by https://github.com/kubernetes/kubernetes/blob/master/cluster/images/etcd/Makefile
OS                ?= ltsc2019
ALL_OS            ?= ltsc2019 ltsc2022
BASE_IMAGE        ?= mcr.microsoft.com/windows/nanoserver

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
	go test -v -bench='benchmarkcollector' ./internal/collectors/{cpu,logical_disk,physical_disk,logon,memory,net,printer,process,service,system,tcp,time}

lint:
	golangci-lint -c .golangci.yaml run

.PHONY: e2e-test
e2e-test: windows_exporter.exe
	powershell -NonInteractive -ExecutionPolicy Bypass -File .\tools\end-to-end-test.ps1

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

.PHONY: package
package: crossbuild
	powershell -NonInteractive -ExecutionPolicy Bypass -File .\installer\build.ps1 -PathToExecutable .\output\amd64\windows_exporter.exe -Version $(shell git describe --tags --abbrev=0)

build-image: crossbuild
	$(DOCKER) build --build-arg=BASE=$(BASE_IMAGE):$(OS) -f Dockerfile -t local/$(DOCKER_IMAGE_NAME):$(VERSION)-$(OS) .

build-hostprocess:
	$(DOCKER) buildx build --build-arg=BASE=mcr.microsoft.com/oss/kubernetes/windows-host-process-containers-base-image:v1.0.0 -f Dockerfile -t local/$(DOCKER_IMAGE_NAME):$(VERSION)-hostprocess .

sub-build-%:
	$(MAKE) OS=$* build-image

build-all: crossbuild
	@for docker_repo in ${DOCKER_REPO}; do \
		echo $(DOCKER) buildx build -f Dockerfile -t $${docker_repo}/$(DOCKER_IMAGE_NAME):$(VERSION) .; \
	done

push:
	@for docker_repo in ${DOCKER_REPO}; do \
		echo $(DOCKER) buildx build --push -f Dockerfile -t $${docker_repo}/$(DOCKER_IMAGE_NAME):$(VERSION) .; \
	done

.PHONY: push-all
push-all: build-all
	$(MAKE) DOCKER_REPO="$(ALL_DOCKER_REPOS)" push

# Mandatory target for container description sync action
.PHONY: docker-repo-name
docker-repo-name:
	@echo "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME)"
