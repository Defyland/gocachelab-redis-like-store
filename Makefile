GO_TOOLCHAIN ?= go1.25.11
GO ?= env GOTOOLCHAIN=$(GO_TOOLCHAIN) go
GOFMT ?= $(shell env GOTOOLCHAIN=$(GO_TOOLCHAIN) go env GOROOT 2>/dev/null)/bin/gofmt
APP ?= gocachelab
PKGS := ./...
GOVULNCHECK_VERSION ?= v1.3.0
GOVULNCHECK ?= $(GO) run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

.PHONY: fmt fmt-check vet test coverage race bench bench-compile build build-check security validate-repository docker-build compose-config review

fmt:
	$(GOFMT) -w cmd internal

fmt-check:
	test -z "$$($(GOFMT) -l cmd internal)"

vet:
	$(GO) vet $(PKGS)

test:
	$(GO) test $(PKGS)

coverage:
	$(GO) test -coverprofile=coverage.out $(PKGS)
	$(GO) tool cover -func=coverage.out

race:
	$(GO) test -race $(PKGS)

bench:
	$(GO) test -bench=. -benchmem $(PKGS)

bench-compile:
	$(GO) test -run '^$$' -bench=. -benchtime=1x $(PKGS)

build:
	mkdir -p bin
	$(GO) build -o bin/$(APP) ./cmd/gocachelab

build-check:
	@tmpdir="$$(mktemp -d)"; \
	trap 'rm -rf "$$tmpdir"' EXIT; \
	$(GO) build -o "$$tmpdir/$(APP)" ./cmd/gocachelab

security:
	$(GOVULNCHECK) ./...

validate-repository:
	./scripts/validate_repository.sh

docker-build:
	docker build -t gocachelab-redis-like-store:local .

compose-config:
	docker compose config

review:
	$(MAKE) validate-repository
	$(MAKE) fmt-check
	$(MAKE) vet
	$(MAKE) test
	$(MAKE) race
	$(MAKE) bench-compile
	$(MAKE) build-check
	$(MAKE) security
