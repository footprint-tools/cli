VERSION := $(shell git describe --tags --dirty --always 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/footprint-tools/footprint-cli/internal/app.Version=$(VERSION)"

.PHONY: all build test lint fmt clean install wipe integration simulate-activity

# Default target
all: build

# Build binary (runs tests first)
build: test
	go build $(LDFLAGS) -o fp ./cmd/fp

# Build without tests (for quick iteration)
build-fast:
	go build $(LDFLAGS) -o fp ./cmd/fp

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
	rm -f fp
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
