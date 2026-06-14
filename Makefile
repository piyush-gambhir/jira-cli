BINARY := jira
MODULE := github.com/piyush-gambhir/jira-cli
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Optional: bake a built-in OAuth app into the binary so end users can do
# `jira auth login --type oauth2` without registering their own app. Put the
# values in a local .env (gitignored) or pass them in the environment; empty by default.
-include .env
JIRA_OAUTH_CLIENT_ID ?=
JIRA_OAUTH_CLIENT_SECRET ?=

LDFLAGS := -s -w \
	-X $(MODULE)/internal/version.Version=$(VERSION) \
	-X $(MODULE)/internal/version.Commit=$(COMMIT) \
	-X $(MODULE)/internal/version.BuildTime=$(BUILD_TIME) \
	-X $(MODULE)/internal/auth.EmbeddedClientID=$(JIRA_OAUTH_CLIENT_ID) \
	-X $(MODULE)/internal/auth.EmbeddedClientSecret=$(JIRA_OAUTH_CLIENT_SECRET)

.PHONY: build install test lint fmt vet clean tidy

# NOTE: build/install recipes are silenced (@) so the embedded-OAuth ldflags
# (which may contain a client secret from .env) are never echoed to the terminal
# or CI logs. Embedding for distribution should go through scripts/release.sh or
# CI secrets (auto-masked), not a chatty local build.
build:
	@echo "building $(BINARY) $(VERSION)"
	@go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

install:
	@echo "installing $(BINARY) $(VERSION)"
	@go install -ldflags "$(LDFLAGS)" .

test:
	go test -v -race ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

clean:
	rm -rf bin/
	go clean

tidy:
	go mod tidy

.PHONY: llms-check
llms-check: ## verify docs/llms.txt documents every command
	@bash scripts/check-llms.sh
