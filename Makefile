# enim-core — common tasks
# Run `make` or `make help` to see everything.

SHELL  := /usr/bin/env bash
MODULE := $(shell awk '/^module /{print $$2; exit}' go.mod)
LATEST := $(shell git tag --list 'v[0-9]*' --sort=-v:refname 2>/dev/null | head -n1)

.DEFAULT_GOAL := help

## ---------- Development ----------

.PHONY: test
test: ## Run all tests
	go test ./...

.PHONY: test-race
test-race: ## Run tests with the race detector
	go test -race -count=1 ./...

.PHONY: cover
cover: ## Run tests and open the coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -n1
	go tool cover -html=coverage.out

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: fmt
fmt: ## Format all Go files
	gofmt -s -w .

.PHONY: tidy
tidy: ## Tidy go.mod / go.sum
	go mod tidy

.PHONY: lint
lint: ## Run golangci-lint (if installed)
	@command -v golangci-lint >/dev/null 2>&1 \
		&& golangci-lint run ./... \
		|| echo "golangci-lint not installed: brew install golangci-lint"

.PHONY: check
check: tidy fmt vet test ## Tidy, format, vet and test (run before committing)

## ---------- Release ----------

.PHONY: version
version: ## Show the latest released version
	@echo "$(MODULE) $(if $(LATEST),$(LATEST),(no releases yet))"

.PHONY: release
release: release-patch ## Alias for release-patch

.PHONY: release-patch
release-patch: ## Release a patch version  (v0.1.3 -> v0.1.4)
	@./scripts/release.sh patch

.PHONY: release-minor
release-minor: ## Release a minor version  (v0.1.4 -> v0.2.0)
	@./scripts/release.sh minor

.PHONY: release-major
release-major: ## Release a major version  (v0.2.0 -> v1.0.0)
	@./scripts/release.sh major

.PHONY: release-version
release-version: ## Release an explicit version: make release-version V=v0.5.0
	@test -n "$(V)" || (echo "usage: make release-version V=v0.5.0"; exit 1)
	@./scripts/release.sh $(V)

## ---------- Misc ----------

.PHONY: clean
clean: ## Remove generated files
	rm -f coverage.out

.PHONY: help
help: ## Show this help
	@echo "Usage: make <target>"
	@awk 'BEGIN {FS = ":.*## "} \
		/^## -+/ { gsub(/^## -+ | -+$$/, ""); printf "\n\033[1m%s\033[0m\n", $$0 } \
		/^[a-zA-Z_-]+:.*## / { printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)