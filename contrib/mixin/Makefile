JSONNET ?= jsonnet
JSONNETFMT ?= jsonnetfmt
JB ?= jb
PROMTOOL ?= promtool
PINT ?= pint
PYTHON ?= python3

# renovate: go=github.com/google/go-jsonnet
JSONNET_VERSION := v0.22.0
# renovate: go=github.com/jsonnet-bundler/jsonnet-bundler
JB_VERSION := v0.6.0
# renovate: go=github.com/cloudflare/pint
PINT_VERSION := v0.89.2
SOURCES := $(wildcard *.libsonnet *.jsonnet lib/*.libsonnet rules/*.libsonnet alerts/*.libsonnet dashboards/*.libsonnet tests/*.libsonnet tests/*.jsonnet)
FMT_FLAGS := -n 2 --max-blank-lines 2 --string-style s --comment-style s
DASHBOARD := ../../dashboard/windows-exporter-dashboard.json

.DEFAULT_GOAL := generate
.DELETE_ON_ERROR:
.PHONY: generate rules dashboard dashboards check-dashboards tools jsonnet-tools fmt lint test check-runbooks clean

generate: rules dashboard

rules: windows_rules.yaml windows_alerts.yaml

tools: jsonnet-tools
	go install github.com/cloudflare/pint/cmd/pint@$(PINT_VERSION)

jsonnet-tools:
	go install github.com/google/go-jsonnet/cmd/jsonnet@$(JSONNET_VERSION)
	go install github.com/google/go-jsonnet/cmd/jsonnetfmt@$(JSONNET_VERSION)
	go install github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@$(JB_VERSION)

# Only the dashboard needs the Jsonnet libraries (grafonnet).
vendor: jsonnetfile.json jsonnetfile.lock.json
	$(JB) install
	@touch $@

windows_rules.yaml: $(SOURCES)
	$(JSONNET) -S rules.jsonnet > $@

windows_alerts.yaml: $(SOURCES)
	$(JSONNET) -S alerts.jsonnet > $@

tests/custom_rules.yaml: $(SOURCES)
	$(JSONNET) -S -e 'std.manifestYamlDoc((import "tests/config.libsonnet").prometheusRules)' > $@

tests/custom_alerts.yaml: $(SOURCES)
	$(JSONNET) -S -e 'std.manifestYamlDoc((import "tests/config.libsonnet").prometheusAlerts)' > $@

dashboard: vendor
	$(JSONNET) -J vendor -S dashboard.jsonnet > $(DASHBOARD).tmp
	mv $(DASHBOARD).tmp $(DASHBOARD)

dashboards: dashboard

check-dashboards: dashboard
	git diff --exit-code -- $(DASHBOARD)

fmt:
	$(JSONNETFMT) $(FMT_FLAGS) -i $(SOURCES)

lint: rules tests/custom_rules.yaml tests/custom_alerts.yaml vendor
	@set -e; for file in $(SOURCES); do \
		$(JSONNETFMT) $(FMT_FLAGS) "$$file" | diff -u "$$file" -; \
	done
	@$(JSONNET) -J vendor -S dashboard.jsonnet | diff -u $(DASHBOARD) - \
		|| { echo "$(DASHBOARD) is out of date; run make dashboard"; exit 1; }
	$(PROMTOOL) check rules windows_rules.yaml windows_alerts.yaml tests/custom_rules.yaml tests/custom_alerts.yaml
	$(PINT) --offline lint windows_rules.yaml windows_alerts.yaml
	$(PINT) --offline lint tests/custom_rules.yaml tests/custom_alerts.yaml

check-runbooks:
	$(JSONNET) -e '(import "mixin.libsonnet") { _config+:: { enableActiveDirectory: true, enableTime: true } }.prometheusAlerts' | $(PYTHON) tests/check_runbooks.py ../../docs/runbooks

test: lint check-runbooks
	$(JSONNET) tests/runbooks.test.jsonnet
	$(PROMTOOL) test rules tests/*.test.yaml

clean:
	rm -f windows_rules.yaml windows_alerts.yaml tests/custom_rules.yaml tests/custom_alerts.yaml
