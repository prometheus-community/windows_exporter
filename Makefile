GOOS    ?= windows
# VERSION is the version the binaries are staged with, like in CI: the content of the
# VERSION file if it exists, else git describe without the leading v.
VERSION ?= $(patsubst v%,%,$(subst +,_,$(if $(wildcard VERSION),$(shell cat VERSION),$(shell git describe --tags --always))))
DOCKER  ?= docker

# DOCKER_REPO is the list of repositories the image is tagged and pushed with.
# The official image repository name at docker.io, quay.io is prometheuscommunity.
DOCKER_REPO       ?= prometheuscommunity
DOCKER_IMAGE_NAME ?= windows-exporter

# ALL_DOCKER_REPOS is the list of repositories to push the image to. ghcr.io requires that org name be the same as the image repo name.
ALL_DOCKER_REPOS  ?= docker.io/$(DOCKER_REPO) ghcr.io/prometheus-community quay.io/$(DOCKER_REPO)

COLLECTOR_BENCHMARKS ?= cpu logical_disk physical_disk memory net printer process service system tcp time

.PHONY: build
build: windows_exporter.exe

windows_exporter.exe: generate
	CGO_ENABLED=0 go build -trimpath -tags=trimpath -o $@ ./cmd/windows_exporter

.PHONY: generate
generate:
	go generate ./...

test:
	go test -v ./...

bench:
	go test -v -run='^$$' -bench='Collector' $(addprefix ./internal/collector/,$(COLLECTOR_BENCHMARKS))

lint:
	golangci-lint -c .golangci.yaml run

.PHONY: e2e-test
e2e-test: windows_exporter.exe
	powershell -NonInteractive -ExecutionPolicy Bypass -File ./tools/end-to-end-test.ps1

fmt:
	gofmt -l -w -s .

# crossbuild stages the binaries like the CI build job:
# output/windows_exporter-$(VERSION)-<arch>.exe
.PHONY: crossbuild
crossbuild: generate
	mkdir -p output
	rm -f output/windows_exporter-*.exe
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -tags=trimpath -o output/windows_exporter-$(VERSION)-amd64.exe ./cmd/windows_exporter
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -trimpath -tags=trimpath -o output/windows_exporter-$(VERSION)-arm64.exe ./cmd/windows_exporter

# package builds the MSI installers into output/. It runs on Windows and needs WiX.
.PHONY: package
package: crossbuild
	powershell -NonInteractive -ExecutionPolicy Bypass -File ./installer/build.ps1 -PathToExecutable ./output/windows_exporter-$(VERSION)-amd64.exe -Version $(VERSION) -Arch amd64
	powershell -NonInteractive -ExecutionPolicy Bypass -File ./installer/build.ps1 -PathToExecutable ./output/windows_exporter-$(VERSION)-arm64.exe -Version $(VERSION) -Arch arm64
	mv installer/*.msi output/

# The container image is a HostProcess image for windows/amd64. Like the CI docker job,
# it is built with docker buildx from the staged amd64 binary, so it can be built on Linux.
# build-image builds it with the current buildx builder. The default builder of a Linux
# Docker engine can only load it with the containerd image store. push builds and pushes it.
DOCKER_BUILD = $(DOCKER) buildx build --platform windows/amd64 -f Dockerfile $(foreach repo,$(DOCKER_REPO),-t $(repo)/$(DOCKER_IMAGE_NAME):$(VERSION))

.PHONY: build-image
build-image: crossbuild
	$(DOCKER_BUILD) output

.PHONY: push
push: crossbuild
	$(DOCKER_BUILD) --push output

.PHONY: push-all
push-all:
	$(MAKE) DOCKER_REPO="$(ALL_DOCKER_REPOS)" push

# Mandatory target for container description sync action
.PHONY: docker-repo-name
docker-repo-name:
	@echo "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME)"
