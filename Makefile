VERSION := $(shell git describe --tags --dirty --always 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/footprint-tools/cli/internal/app.Version=$(VERSION)"
LDFLAGS_RELEASE := -ldflags "-s -w -X github.com/footprint-tools/cli/internal/app.Version=$(VERSION)"

DEV_HOME := $(HOME)/.footprint-dev
DEV_LDFLAGS := -ldflags "-X github.com/footprint-tools/cli/internal/paths.DefaultHome=$(DEV_HOME) -X github.com/footprint-tools/cli/internal/app.Version=$(VERSION)"

.PHONY: all build test lint fmt clean install wipe integration simulate-activity changelog release fpdev fpdev-fast snapshot

# Default target
all: build

# Build binary with debug symbols (runs tests first)
build: test
	go build $(LDFLAGS) -o fp ./cmd/fp

# Build without tests (for quick iteration)
build-fast:
	go build $(LDFLAGS) -o fp ./cmd/fp

# Build optimized release binary (smaller, no debug symbols)
release: test
	go build $(LDFLAGS_RELEASE) -o fp ./cmd/fp
	@echo "Built release binary: $$(ls -lh fp | awk '{print $$5}')"

# Build dev binary with isolated data directory
fpdev: test
	go build $(DEV_LDFLAGS) -o fpdev ./cmd/fp

# Build dev binary without tests (quick iteration)
fpdev-fast:
	go build $(DEV_LDFLAGS) -o fpdev ./cmd/fp

# Copy production DB and config to dev environment
snapshot:
	@mkdir -p "$(DEV_HOME)"
	@PROD_DB="$(HOME)/Library/Application Support/footprint/store.db"; \
	if [ -f "$$PROD_DB" ]; then \
		sqlite3 "$$PROD_DB" ".backup '$(DEV_HOME)/store.db'"; \
		echo "DB copiada a $(DEV_HOME)/store.db"; \
	else \
		echo "No se encontró DB de producción en $$PROD_DB"; \
	fi
	@if [ -f "$(HOME)/.fprc" ]; then \
		cp "$(HOME)/.fprc" "$(DEV_HOME)/.fprc"; \
		echo "Config copiada a $(DEV_HOME)/.fprc"; \
	fi
	@echo "Snapshot listo. Usa ./fpdev para probar."

# Run unit tests
test:
	go test ./...

# Run linter
lint:
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Clean build artifacts
clean:
	rm -f fp fpdev
	go clean

# Install to GOPATH/bin
install: test
	go install $(LDFLAGS) ./cmd/fp

# Run integration tests (slow, requires built binary)
integration: build
	./scripts/test-hooks.sh
	./scripts/test-export-flow.sh
	./scripts/test-backfill.sh

# Wipe all local data (database, exports, config), uninstall hooks, and remove fpdev binary
wipe:
	@# Uninstall hooks from all tracked repos before deleting the database
	@DB="$${XDG_CONFIG_HOME:-$$HOME/.config}/Footprint/store.db"; \
	if [ -f "$$DB" ]; then \
		echo "Removing hooks from tracked repositories..."; \
		sqlite3 "$$DB" "SELECT repo_path FROM tracked_repos" 2>/dev/null | while read -r repo; do \
			if [ -d "$$repo/.git/hooks" ]; then \
				for hook in post-commit post-merge post-checkout post-rewrite pre-push; do \
					if [ -f "$$repo/.git/hooks/$$hook" ] && grep -q "fp record" "$$repo/.git/hooks/$$hook" 2>/dev/null; then \
						rm -f "$$repo/.git/hooks/$$hook"; \
						echo "  Removed $$hook from $$repo"; \
					fi; \
				done; \
			fi; \
		done; \
	fi
	rm -rf "$(HOME)/Library/Application Support/Footprint"
	rm -rf "$(HOME)/Library/Application Support/footprint"
	rm -rf "$${XDG_CONFIG_HOME:-$$HOME/.config}/Footprint"
	rm -rf "$${XDG_DATA_HOME:-$$HOME/.local/share}/footprint"
	rm -f ~/.fprc
	rm -f ./fp
	@echo "Wiped hooks, database, exports, config, and fpdev binary"

# Simulate continuous git activity for testing watch -i (Ctrl+C to stop)
# Creates temporary repos, tracks them, and generates commits/merges/checkouts
# Usage: make simulate-activity [REPOS=5]
simulate-activity: build
	./scripts/event-generator.sh $(or $(REPOS),3)

# Generate changelog from git history using git-cliff
# Requires: brew install git-cliff
changelog:
	@git-cliff --output CHANGELOG.md
	@# Add comparison links for versions that are ancestors of HEAD
	@echo "" >> CHANGELOG.md
	@TAGS=$$(git tag -l 'v*' --sort=-version:refname --merged HEAD); \
	PREV="HEAD"; \
	for TAG in $$TAGS; do \
		VER=$${TAG#v}; \
		if [ "$$PREV" = "HEAD" ]; then \
			echo "[Unreleased]: https://github.com/footprint-tools/cli/compare/$$TAG...HEAD" >> CHANGELOG.md; \
		else \
			PREV_VER=$${PREV#v}; \
			echo "[$$PREV_VER]: https://github.com/footprint-tools/cli/compare/$$TAG...$$PREV" >> CHANGELOG.md; \
		fi; \
		PREV=$$TAG; \
	done; \
	LAST_VER=$${PREV#v}; \
	echo "[$$LAST_VER]: https://github.com/footprint-tools/cli/releases/tag/$$PREV" >> CHANGELOG.md
	@echo "Generated CHANGELOG.md from git history"
