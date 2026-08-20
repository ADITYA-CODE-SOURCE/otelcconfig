# Copyright The otelcconfig Authors
# SPDX-License-Identifier: Apache-2.0

.PHONY: all build test test-race coverage lint vet fmt fmt-check tidy tidy-check generate generate-check bake bake-check check demo-run clean help

MODULE := github.com/ADITYA-CODE-SOURCE/otelcconfig
BIN    := otelcconfig
GO     ?= go
VERSION ?= dev
LDFLAGS := -X main.version=$(VERSION)

all: check

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'

build: ## Build the otelcconfig CLI
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/otelcconfig

test: ## Run all tests
	$(GO) test ./...

test-race: ## Run tests with the race detector
	$(GO) test -race ./...

coverage: ## Write a coverage report to coverage.out
	$(GO) test -coverprofile=coverage.out ./...

vet: ## Run go vet
	$(GO) vet ./...

fmt: ## Format Go sources with gofmt
	@gofmt -w $$(find . -name '*.go' -not -path './.git/*')

fmt-check: ## Fail if Go sources are not formatted
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './.git/*'))" || \
		(echo "Go files need formatting; run 'make fmt'" && \
		 gofmt -l $$(find . -name '*.go' -not -path './.git/*') && exit 1)

lint: ## Run golangci-lint if installed
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint is required; install it from https://golangci-lint.run"; \
		exit 1; \
	fi

tidy: ## Tidy go.mod / go.sum
	$(GO) mod tidy

tidy-check: ## Fail if go.mod or go.sum is not tidy
	$(GO) mod tidy
	@git diff --exit-code -- go.mod go.sum

generate: ## Regenerate committed artifacts from behavior manifests
	$(GO) run ./cmd/otelcconfig generate

generate-check: ## Fail if committed artifacts drift from behavior manifests
	$(GO) run ./cmd/otelcconfig generate --check

bake: ## Freeze resolved examples/demo.yaml configuration into the baked package
	$(GO) run ./cmd/otelcconfig bake --output baked examples/demo.yaml

bake-check: ## Fail if committed baked package drifts from examples/demo.yaml
	$(GO) run ./cmd/otelcconfig bake --output baked --check examples/demo.yaml

demo-run: bake ## Build and run the end-to-end demo binary
	$(GO) run ./cmd/demo

check: fmt-check tidy-check vet test generate-check bake-check build ## Run non-mutating format, module, vet, test, generate drift, bake drift, and build checks
	@echo "✓ check passed"

clean: ## Remove build artifacts
	rm -f $(BIN)
	rm -rf dist/ bin/
