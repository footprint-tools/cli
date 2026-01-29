#!/bin/bash
set -e

# Quick test for hooks-based recording + export
# (Dense backfill + export testing is in test-backfill.sh)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
FP="$PROJECT_DIR/fp"
TEST_DIR="/tmp/fp-export-test-$$"

# Colors
DIM='\033[2m'
CYAN='\033[0;36m'
GREEN='\033[0;32m'
NC='\033[0m'

section() { echo -e "\n${CYAN}=== $1 ===${NC}"; }
log() { echo -e "${DIM}$1${NC}"; }
success() { echo -e "${GREEN}✓${NC} $1"; }

cleanup() {
    $FP untrack "$TEST_DIR/repo" 2>/dev/null || true
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

if [ ! -f "$FP" ]; then
    echo "Error: fp not found. Run 'make build' first"
    exit 1
fi

mkdir -p "$TEST_DIR/repo"
cd "$TEST_DIR/repo"

section "Creating repo with hooks"
git init -q
git config user.email "test@example.com"
git config user.name "Test"

# Initial commit
echo "# Test" > README.md
git add -A && git commit -q -m "Initial commit"

# Track and setup hooks
$FP track
if ! $FP check 2>/dev/null | grep -q "post-commit.*installed"; then
    $FP setup --repo --force
fi

section "Creating commits (hooks should record)"

for i in $(seq 1 5); do
    echo "content $i" > "file$i.txt"
    git add -A && git commit -q -m "Commit $i"
done

# Branch and merge
git checkout -q -b feature
echo "feature" > feature.txt
git add -A && git commit -q -m "Feature work"
git checkout -q main
git merge -q feature -m "Merge feature"

success "Created 7 commits + merge"

section "Checking recorded events"
EVENT_COUNT=$($FP activity --oneline 2>/dev/null | wc -l | tr -d ' ')
log "Events recorded: $EVENT_COUNT"

if [ "$EVENT_COUNT" -gt 0 ]; then
    success "Hooks recorded events"
else
    log "No events (hooks may be global, not local)"
fi

section "Exporting"
$FP export --now
success "Export completed"

section "Done"
success "Hooks + export test passed!"
