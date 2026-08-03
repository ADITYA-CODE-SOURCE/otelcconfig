# Copyright The otelcconfig Authors
# SPDX-License-Identifier: Apache-2.0

.PHONY: all build test lint vet fmt tidy generate check clean help

MODULE := github.com/ADITYA-CODE-SOURCE/otelcconfig
BIN    := otelcconfig
GO     ?= go

all: check

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'

build: ## Build the otelcconfig CLI
	$(GO) build -o $(BIN) ./cmd/otelcconfig

test: ## Run all tests
	$(GO) test ./...

vet: ## Run go vet
	$(GO) vet ./...

fmt: ## Format Go sources with gofmt
	@gofmt -w $$(find . -name '*.go' -not -path './.git/*')
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './.git/*'))"

lint: ## Run golangci-lint if installed
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed; skipping (install: https://golangci-lint.run)"; \
	fi

tidy: ## Tidy go.mod / go.sum
	$(GO) mod tidy

generate: ## Run code generation (Phase 1+)
	@echo "codegen not implemented in Phase 0; see docs/architecture.md"
	@$(GO) generate ./...

check: fmt vet test build ## Run format, vet, test, and build
	@echo "✓ check passed"

clean: ## Remove build artifacts
	rm -f $(BIN)
	rm -rf dist/ bin/
