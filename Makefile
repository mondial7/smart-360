# Smart 360 Feedback - top-level build & release Makefile.
#
# Usage:
#   make build           # build frontend + single-binary for the host OS/arch
#   make dist            # cross-compile binaries for all supported platforms
#   make release         # build + checksum + tarballs ready for GitHub Releases
#   make clean           # remove build artifacts
#
# The single binary serves both the API and the embedded SPA, which is what
# the Homebrew formula and "from source" install paths consume.

SHELL          := /bin/bash
VERSION        ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT         := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE           := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

BACKEND_DIR    := backend
FRONTEND_DIR   := frontend
WEB_DIST_DIR   := $(BACKEND_DIR)/web/dist
DIST_DIR       := dist
BINARY         := smart360

LDFLAGS        := -s -w \
                  -X 'main.version=$(VERSION)' \
                  -X 'main.commit=$(COMMIT)' \
                  -X 'main.buildDate=$(DATE)'

PLATFORMS      := darwin/arm64 darwin/amd64 linux/amd64 linux/arm64

.PHONY: help build frontend embed backend dist release clean test print-version

help:
	@echo "Smart 360 Feedback - build targets"
	@echo ""
	@echo "  build       Build the single binary for the host platform"
	@echo "  frontend    Build the Vue SPA (vite only, skips vue-tsc)"
	@echo "  dist        Cross-compile binaries for all platforms"
	@echo "  release     Cross-compile + create tarballs and checksums in $(DIST_DIR)/"
	@echo "  test        Run backend test suite"
	@echo "  clean       Remove build artifacts"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  PLATFORMS=$(PLATFORMS)"

print-version:
	@echo $(VERSION)

# Build the SPA into frontend/dist
frontend:
	@echo "==> Building frontend"
	cd $(FRONTEND_DIR) && npm ci --include=dev
	cd $(FRONTEND_DIR) && npm run build

# Sync built SPA into the Go embed folder so `go build` picks it up.
# We keep .gitkeep around so fresh clones (with no built frontend) can still
# compile the //go:embed directive in backend/web/embed.go.
embed: frontend
	@echo "==> Copying SPA into $(WEB_DIST_DIR)"
	rm -rf $(WEB_DIST_DIR)
	mkdir -p $(WEB_DIST_DIR)
	touch $(WEB_DIST_DIR)/.gitkeep
	cp -R $(FRONTEND_DIR)/dist/. $(WEB_DIST_DIR)/

# Build the single binary for the host platform.
build: embed
	@echo "==> Building $(BINARY) for host platform"
	cd $(BACKEND_DIR) && CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o ../$(BINARY) .
	@echo "==> Done: ./$(BINARY) ($(VERSION))"

backend:
	@echo "==> Building backend only (no frontend embed refresh)"
	cd $(BACKEND_DIR) && CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o ../$(BINARY) .

# Cross-compile all platforms into dist/<os>-<arch>/smart360
dist: embed
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%%/*}; arch=$${platform##*/}; \
		out=$(DIST_DIR)/smart360-$(VERSION)-$$os-$$arch; \
		echo "==> Building $$out"; \
		mkdir -p $$out; \
		cd $(BACKEND_DIR) && \
		  CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
		  go build -trimpath -ldflags "$(LDFLAGS)" -o ../$$out/$(BINARY) . && \
		cd - >/dev/null; \
		cp README.md LICENSE .env.example $$out/; \
	done

# Package each cross-compiled folder as a tarball and emit a sha256 sums file.
release: dist
	@echo "==> Creating release archives in $(DIST_DIR)/"
	@cd $(DIST_DIR) && \
	  for d in smart360-$(VERSION)-*; do \
	    [ -d "$$d" ] || continue; \
	    tar -czf "$$d.tar.gz" "$$d"; \
	    rm -rf "$$d"; \
	  done && \
	  shasum -a 256 smart360-$(VERSION)-*.tar.gz > smart360-$(VERSION)-SHA256SUMS.txt
	@echo "==> Release artifacts:"
	@ls -la $(DIST_DIR)/

test:
	cd $(BACKEND_DIR) && go test ./...

clean:
	rm -rf $(BINARY) $(DIST_DIR) $(FRONTEND_DIR)/dist
	find $(WEB_DIST_DIR) -mindepth 1 ! -name '.gitkeep' -delete 2>/dev/null || true
