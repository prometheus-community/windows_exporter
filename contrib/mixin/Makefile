JSONNET ?= jsonnet
JSONNETFMT ?= jsonnetfmt
PROMTOOL ?= promtool
PINT ?= pint

# renovate: go=github.com/google/go-jsonnet
JSONNET_VERSION := v0.22.0
# renovate: go=github.com/cloudflare/pint
PINT_VERSION := v0.89.2
SOURCES := $(wildcard *.libsonnet *.jsonnet lib/*.libsonnet rules/*.libsonnet alerts/*.libsonnet tests/*.libsonnet)
FMT_FLAGS := -n 2 --max-blank-lines 2 --string-style s --comment-style s

.DEFAULT_GOAL := generate
.DELETE_ON_ERROR:
.PHONY: generate tools fmt lint test clean

generate: windows_rules.yaml windows_alerts.yaml

tools:
	go install github.com/google/go-jsonnet/cmd/jsonnet@$(JSONNET_VERSION)
	go install github.com/google/go-jsonnet/cmd/jsonnetfmt@$(JSONNET_VERSION)
	go install github.com/cloudflare/pint/cmd/pint@$(PINT_VERSION)

windows_rules.yaml: $(SOURCES)
	$(JSONNET) -S rules.jsonnet > $@

windows_alerts.yaml: $(SOURCES)
	$(JSONNET) -S alerts.jsonnet > $@

tests/custom_rules.yaml: $(SOURCES)
	$(JSONNET) -S -e 'std.manifestYamlDoc((import "tests/config.libsonnet").prometheusRules)' > $@

tests/custom_alerts.yaml: $(SOURCES)
	$(JSONNET) -S -e 'std.manifestYamlDoc((import "tests/config.libsonnet").prometheusAlerts)' > $@

fmt:
	$(JSONNETFMT) $(FMT_FLAGS) -i $(SOURCES)

lint: generate tests/custom_rules.yaml tests/custom_alerts.yaml
	@set -e; for file in $(SOURCES); do \
		$(JSONNETFMT) $(FMT_FLAGS) "$$file" | diff -u "$$file" -; \
	done
	$(PROMTOOL) check rules windows_rules.yaml windows_alerts.yaml tests/custom_rules.yaml tests/custom_alerts.yaml
	$(PINT) --offline lint windows_rules.yaml windows_alerts.yaml
	$(PINT) --offline lint tests/custom_rules.yaml tests/custom_alerts.yaml

test: lint
	$(PROMTOOL) test rules tests/rules.test.yaml tests/alerts.test.yaml tests/custom.test.yaml

clean:
	rm -f windows_rules.yaml windows_alerts.yaml tests/custom_rules.yaml tests/custom_alerts.yaml
