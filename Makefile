# Smart 360 Feedback — build & dev Makefile.
#
# The app is a single Go binary with templates and static assets embedded via
# //go:embed. There is no separate frontend build step.

SHELL    := /bin/bash
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE     := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BINARY   := smart360
DIST_DIR := dist
LDFLAGS  := -s -w \
            -X 'main.version=$(VERSION)' \
            -X 'main.commit=$(COMMIT)' \
            -X 'main.buildDate=$(DATE)'
PLATFORMS := darwin/arm64 darwin/amd64 linux/amd64 linux/arm64

.PHONY: help build run test test-short vet fmt lint dist release clean docker-up docker-down

help:
	@echo "Targets:"
	@echo "  build       Build the single binary for the host OS/arch"
	@echo "  run         Run the server (needs a running Postgres; see docker-up)"
	@echo "  test        Full test pyramid (gateway tests need Docker)"
	@echo "  test-short  Unit + in-memory handler tests only (no Docker)"
	@echo "  vet fmt     go vet / gofmt"
	@echo "  dist        Cross-compile bare release binaries into $(DIST_DIR)/"
	@echo "  release     Cross-compile + tar.gz + SHA256SUMS (VERSION=vX.Y.Z)"
	@echo "  backup      Dump the database via scripts/backup.sh (needs DATABASE_URL)"
	@echo "  docker-up   Start the Postgres dependency via docker compose"
	@echo "  clean       Remove build artifacts"

build:
	@echo "==> Building $(BINARY) ($(VERSION))"
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY) ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

test-short:
	go test -short ./...

vet:
	go vet ./...

fmt:
	gofmt -w cmd internal web

lint: vet
	@test -z "$$(gofmt -l cmd internal web)" || (echo "gofmt needed:"; gofmt -l cmd internal web; exit 1)

dist:
	@echo "==> Cross-compiling $(VERSION)"
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		out=$(DIST_DIR)/$(BINARY)-$${os}-$${arch}; \
		echo "  $$out"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -ldflags="$(LDFLAGS)" -o $$out ./cmd/server; \
	done

# release produces one tar.gz per platform plus a SHA256SUMS file, laid out the
# way the Homebrew formula and the Release workflow expect:
#   dist/smart360-<version>-<os>-<arch>.tar.gz   (contains smart360 + .env.example)
#   dist/smart360-<version>-SHA256SUMS.txt
release: clean
	@echo "==> Building release $(VERSION)"
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		name=$(BINARY)-$(VERSION)-$${os}-$${arch}; \
		stage=$(DIST_DIR)/$${name}; \
		mkdir -p $$stage; \
		echo "  $$name.tar.gz"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -ldflags="$(LDFLAGS)" -o $$stage/$(BINARY) ./cmd/server; \
		cp .env.example $$stage/.env.example; \
		tar -czf $(DIST_DIR)/$${name}.tar.gz -C $(DIST_DIR) $${name}; \
		rm -rf $$stage; \
	done
	@cd $(DIST_DIR) && shasum -a 256 *.tar.gz > $(BINARY)-$(VERSION)-SHA256SUMS.txt
	@echo "==> Wrote $(DIST_DIR)/$(BINARY)-$(VERSION)-SHA256SUMS.txt"

backup:
	./scripts/backup.sh

docker-up:
	docker compose up -d postgres

docker-down:
	docker compose down

clean:
	rm -rf $(BINARY) $(DIST_DIR)
