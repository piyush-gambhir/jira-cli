BINARY := jira
MODULE := github.com/piyush-gambhir/jira-cli/cli-go
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -s -w \
	-X $(MODULE)/internal/version.Version=$(VERSION) \
	-X $(MODULE)/internal/version.Commit=$(COMMIT) \
	-X $(MODULE)/internal/version.BuildTime=$(BUILD_TIME)

.PHONY: build install test lint fmt vet clean tidy

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
