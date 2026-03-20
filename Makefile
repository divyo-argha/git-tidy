# ─────────────────────────────────────────────────────────────────────────────
#  git-tidy — Makefile
#  Usage: make <target>
#  Run `make help` to list all targets.
# ─────────────────────────────────────────────────────────────────────────────

BINARY      := git-tidy
MODULE      := github.com/yourusername/git-tidy
BUILD_DIR   := ./bin
INSTALL_DIR ?= /usr/local/bin

# Build metadata — overridden by goreleaser / CI.
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE     ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags "\
  -X main.version=$(VERSION) \
  -X main.commit=$(COMMIT) \
  -X main.date=$(DATE) \
  -s -w"

# Cross-compile targets for `make release`.
PLATFORMS := \
  linux/amd64 \
  linux/arm64 \
  darwin/amd64 \
  darwin/arm64

.DEFAULT_GOAL := help

# ── Primary targets ───────────────────────────────────────────────────────────

.PHONY: build
## build: Compile the binary for the current platform → ./bin/git-tidy
build:
	@mkdir -p $(BUILD_DIR)
	@echo "  building $(BINARY) $(VERSION) ($(GOOS)/$(GOARCH))…"
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) .
	@echo "  → $(BUILD_DIR)/$(BINARY)"

.PHONY: install
## install: Build and install to INSTALL_DIR (default: /usr/local/bin)
install: build
	@echo "  installing $(BINARY) → $(INSTALL_DIR)/$(BINARY)"
	@install -d $(INSTALL_DIR)
	@install -m 0755 $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "  done — run: git tidy --help"

.PHONY: uninstall
## uninstall: Remove the installed binary from INSTALL_DIR
uninstall:
	@if [ -f "$(INSTALL_DIR)/$(BINARY)" ]; then \
		rm -f $(INSTALL_DIR)/$(BINARY); \
		echo "  removed $(INSTALL_DIR)/$(BINARY)"; \
	else \
		echo "  $(INSTALL_DIR)/$(BINARY) not found — nothing to remove"; \
	fi

.PHONY: reinstall
## reinstall: Uninstall then install (useful after a pull)
reinstall: uninstall install

# ── Development ───────────────────────────────────────────────────────────────

.PHONY: run
## run: Build and run with default args (shows help)
run: build
	$(BUILD_DIR)/$(BINARY) --help

.PHONY: tidy
## tidy: Run go mod tidy to clean up dependencies
tidy:
	go mod tidy

.PHONY: test
## test: Run all unit tests
test:
	go test -v -race ./...

.PHONY: test-cover
## test-cover: Run tests with coverage report
test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "  coverage report → coverage.html"

.PHONY: lint
## lint: Run golangci-lint (install: brew install golangci-lint)
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "  golangci-lint not found — install: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	}
	golangci-lint run ./...

.PHONY: vet
## vet: Run go vet
vet:
	go vet ./...

.PHONY: fmt
## fmt: Format all Go source files
fmt:
	gofmt -w .
	@echo "  formatted all .go files"

.PHONY: check
## check: Run vet + fmt check (no lint dependency) — good for pre-commit
check: vet
	@test -z "$$(gofmt -l .)" || { \
		echo "  the following files need formatting:"; \
		gofmt -l .; \
		exit 1; \
	}
	@echo "  all checks passed"

# ── Release ───────────────────────────────────────────────────────────────────

.PHONY: release
## release: Cross-compile for all platforms → ./bin/git-tidy-<os>-<arch>
release:
	@mkdir -p $(BUILD_DIR)
	@echo "  cross-compiling $(VERSION) for $(words $(PLATFORMS)) platforms…"
	@$(foreach platform,$(PLATFORMS), \
		$(eval OS   := $(word 1,$(subst /, ,$(platform)))) \
		$(eval ARCH := $(word 2,$(subst /, ,$(platform)))) \
		$(eval OUT  := $(BUILD_DIR)/$(BINARY)-$(OS)-$(ARCH)) \
		echo "  → $(OUT)"; \
		GOOS=$(OS) GOARCH=$(ARCH) go build $(LDFLAGS) -o $(OUT) . || exit 1; \
	)
	@echo "  release binaries in $(BUILD_DIR)/"

.PHONY: snapshot
## snapshot: Build a quick local snapshot (like goreleaser --snapshot)
snapshot:
	@which goreleaser >/dev/null 2>&1 && \
		goreleaser build --snapshot --clean || \
		$(MAKE) release

.PHONY: checksums
## checksums: Generate SHA256 checksums for release binaries
checksums:
	@cd $(BUILD_DIR) && sha256sum $(BINARY)-* > checksums.txt && \
		echo "  checksums → $(BUILD_DIR)/checksums.txt"

# ── Housekeeping ──────────────────────────────────────────────────────────────

.PHONY: clean
## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html
	@echo "  cleaned"

.PHONY: version
## version: Print the version that would be embedded in a build
version:
	@echo $(VERSION)

.PHONY: help
## help: List all available targets with descriptions
help:
	@echo ""
	@echo "  git-tidy build system"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) \
		| sed 's/## //' \
		| awk -F': ' '{ printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2 }'
	@echo ""
	@echo "  Variables (override with make <target> VAR=value):"
	@echo "    INSTALL_DIR  default: /usr/local/bin"
	@echo "    VERSION      default: git describe --tags"
	@echo ""
