##
# Console Colors
##
GREEN  := $(shell printf "\033[0;32m")
YELLOW := $(shell printf "\033[0;33m")
CYAN   := $(shell printf "\033[0;36m")
RESET  := $(shell printf "\033[0m")

# DOCKER_REPO is the official image repository name at docker.io, quay.io.
DOCKER_REPO       ?= prometheuscommunity
DOCKER_IMAGE_NAME ?= windows-exporter

##
# Targets
##
.PHONY: help
help: ## show this help.
	@echo 'Usage:'
	@echo "  ${GREEN}make${RESET} ${YELLOW}<target>${RESET}"
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "  ${GREEN}%-21s${YELLOW}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST) | sort

.PHONY: clean
clean: ## clean build output
	@rm -rf windows_exporter.exe dist/ output/

.PHONY: build
build: export GOOS := windows
build: ## build windows_exporter.exe
	@go build -o windows_exporter.exe ./cmd/windows_exporter

.PHONY: test
test: ## run the tests
	@go test ./...

.PHONY: lint
lint: ## run golangci-lint
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 run ./...

.PHONY: fmt
fmt: ## format the code
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 fmt

# Mandatory target for container description sync action
.PHONY: docker-repo-name
docker-repo-name:
	@echo "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME)"
