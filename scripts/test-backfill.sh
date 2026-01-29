#!/bin/bash
set -e

# Comprehensive backfill + export test
# Creates dense git history, backfills it, then exports

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
FP="$PROJECT_DIR/fp"
TEST_DIR="/tmp/fp-backfill-test-$$"

# Platform-specific export directory
if [ "$(uname)" = "Darwin" ]; then
    EXPORT_DIR="$HOME/Library/Application Support/footprint/export"
else
    EXPORT_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/footprint/export"
fi

# Colors
DIM='\033[2m'
CYAN='\033[0;36m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

section() { echo -e "\n${CYAN}=== $1 ===${NC}"; }
log() { echo -e "${DIM}$1${NC}"; }
success() { echo -e "${GREEN}✓${NC} $1"; }
fail() { echo -e "${RED}✗${NC} $1"; exit 1; }

# Cleanup
cleanup() {
    log "Cleaning up..."
    for repo in webapp api-server shared-lib devops; do
        $FP untrack "$TEST_DIR/$repo" 2>/dev/null || true
    done
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

# Check fp exists
if [ ! -f "$FP" ]; then
    echo "Error: fp binary not found. Run 'make build' first"
    exit 1
fi

mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

log "Test directory: $TEST_DIR"

# Clear previous export data
rm -rf "$EXPORT_DIR/repos"

# Helper to create commits with dates
commit_at() {
    local date="$1"
    local msg="$2"
    GIT_AUTHOR_DATE="$date" GIT_COMMITTER_DATE="$date" git commit -q -m "$msg"
}

###############################################################################
section "REPO 1: webapp - Frontend application (50+ commits)"
###############################################################################

mkdir -p webapp && cd webapp
git init -q
git config user.email "frontend@example.com"
git config user.name "Frontend Dev"

# Create directories
mkdir -p src/components src/pages

# Initial setup
echo "# WebApp" > README.md
echo "node_modules/" > .gitignore
echo '{"name": "webapp"}' > package.json
git add -A && commit_at "2024-01-01T09:00:00" "Initial commit"

# Week 1: Project setup
for day in $(seq 2 7); do
    echo "// Day $day setup" >> src/index.js
    git add -A && commit_at "2024-01-0${day}T10:00:00" "chore: day $day setup work"
done

# Feature: Authentication (branch + merge)
git checkout -q -b feature/auth
echo "export const login = () => {}" > src/auth.js
git add -A && commit_at "2024-01-08T09:00:00" "feat(auth): add login stub"
echo "export const logout = () => {}" >> src/auth.js
git add -A && commit_at "2024-01-08T11:00:00" "feat(auth): add logout"
echo "export const refresh = () => {}" >> src/auth.js
git add -A && commit_at "2024-01-08T14:00:00" "feat(auth): add token refresh"
echo "// JWT validation" >> src/auth.js
git add -A && commit_at "2024-01-08T16:00:00" "feat(auth): add JWT validation"
git checkout -q main
git merge -q feature/auth -m "Merge feature/auth"

# Week 2: Components
for i in $(seq 1 10); do
    echo "export const Component$i = () => <div>$i</div>" > "src/components/Component$i.jsx"
    git add -A && commit_at "2024-01-$((9 + i / 3))T$((9 + i % 8)):00:00" "feat: add Component$i"
done

# Feature: Dashboard (branch + rebase + merge)
git checkout -q -b feature/dashboard
echo "export const Dashboard = () => {}" > src/pages/Dashboard.jsx
git add -A && commit_at "2024-01-15T09:00:00" "feat(dashboard): scaffold"
echo "// widgets" >> src/pages/Dashboard.jsx
git add -A && commit_at "2024-01-15T11:00:00" "feat(dashboard): add widgets"
echo "// charts" >> src/pages/Dashboard.jsx
git add -A && commit_at "2024-01-15T14:00:00" "feat(dashboard): add charts"

# Meanwhile, commits on main
git checkout -q main
echo "// hotfix" >> src/index.js
git add -A && commit_at "2024-01-15T12:00:00" "fix: critical hotfix"
echo "// another fix" >> src/index.js
git add -A && commit_at "2024-01-15T13:00:00" "fix: follow-up fix"

# Rebase dashboard onto main
git checkout -q feature/dashboard
git rebase -q main

# More dashboard work
echo "// filters" >> src/pages/Dashboard.jsx
git add -A && commit_at "2024-01-16T09:00:00" "feat(dashboard): add filters"
echo "// export" >> src/pages/Dashboard.jsx
git add -A && commit_at "2024-01-16T11:00:00" "feat(dashboard): add export"

git checkout -q main
git merge -q feature/dashboard -m "Merge feature/dashboard"

# Week 3: Styling and polish
for i in $(seq 1 8); do
    echo ".style$i { color: blue; }" >> src/styles.css
    git add -A && commit_at "2024-01-$((17 + i / 4))T$((9 + i)):00:00" "style: add style$i"
done

# Bug fixes
git checkout -q -b bugfix/memory-leak
echo "// fix leak" >> src/utils.js
git add -A && commit_at "2024-01-20T10:00:00" "fix: memory leak in utils"
git checkout -q main
git merge -q bugfix/memory-leak -m "Merge bugfix/memory-leak"

# Final commits
for i in $(seq 1 5); do
    echo "// polish $i" >> src/index.js
    git add -A && commit_at "2024-01-2${i}T15:00:00" "chore: polish $i"
done

WEBAPP_COMMITS=$(git rev-list --count HEAD)
success "webapp: $WEBAPP_COMMITS commits created"
cd "$TEST_DIR"

###############################################################################
section "REPO 2: api-server - Backend API (60+ commits)"
###############################################################################

mkdir -p api-server && cd api-server
git init -q
git config user.email "backend@example.com"
git config user.name "Backend Dev"

# Create directories
mkdir -p db handlers middleware cache security

# Initial setup
echo "package main" > main.go
echo "go.mod" > .gitignore
git add -A && commit_at "2024-01-01T08:00:00" "Initial commit"

# Database layer
git checkout -q -b feature/database
for i in $(seq 1 6); do
    echo "func dbFunc$i() {}" >> db/database.go
    git add -A && commit_at "2024-01-0$((2 + i / 2))T$((8 + i)):00:00" "feat(db): add dbFunc$i"
done
git checkout -q main
git merge -q feature/database -m "Merge feature/database"

# API endpoints - users
git checkout -q -b feature/users-api
for endpoint in list get create update delete; do
    echo "func ${endpoint}User() {}" >> handlers/users.go
    git add -A && commit_at "2024-01-08T$((9 + ${#endpoint})):00:00" "feat(users): add $endpoint endpoint"
done
git checkout -q main
git merge -q feature/users-api -m "Merge feature/users-api"

# API endpoints - products (with amends)
git checkout -q -b feature/products-api
echo "func listProducts() {}" > handlers/products.go
git add -A && commit_at "2024-01-10T09:00:00" "feat(products): add list"
echo "func getProduct() {}" >> handlers/products.go
git add -A && commit_at "2024-01-10T10:00:00" "feat(products): add get"
echo "// with filtering" >> handlers/products.go
git add -A && git commit -q --amend --no-edit  # amend
echo "func createProduct() {}" >> handlers/products.go
git add -A && commit_at "2024-01-10T11:00:00" "feat(products): add create"
echo "func updateProduct() {}" >> handlers/products.go
git add -A && commit_at "2024-01-10T12:00:00" "feat(products): add update"
echo "func deleteProduct() {}" >> handlers/products.go
git add -A && commit_at "2024-01-10T13:00:00" "feat(products): add delete"
git checkout -q main
git merge -q feature/products-api -m "Merge feature/products-api"

# Middleware
git checkout -q -b feature/middleware
for mw in auth logging cors ratelimit; do
    echo "func ${mw}Middleware() {}" >> middleware/${mw}.go
    git add -A && commit_at "2024-01-12T$((9 + ${#mw} % 6)):00:00" "feat(mw): add $mw middleware"
done
git checkout -q main
git merge -q feature/middleware -m "Merge feature/middleware"

# Testing
git checkout -q -b feature/tests
for i in $(seq 1 10); do
    echo "func Test$i(t *testing.T) {}" >> handlers/handlers_test.go
    git add -A && commit_at "2024-01-$((14 + i / 4))T$((9 + i % 8)):00:00" "test: add Test$i"
done
git checkout -q main
git merge -q feature/tests -m "Merge feature/tests"

# Performance optimizations (parallel branch work)
git checkout -q -b perf/caching
echo "var cache = map[string]interface{}{}" > cache/cache.go
git add -A && commit_at "2024-01-18T09:00:00" "perf: add caching layer"
echo "func getFromCache() {}" >> cache/cache.go
git add -A && commit_at "2024-01-18T10:00:00" "perf: add cache get"
echo "func setToCache() {}" >> cache/cache.go
git add -A && commit_at "2024-01-18T11:00:00" "perf: add cache set"

git checkout -q main
# Work on main while perf branch exists
echo "// config update" >> config.go
git add -A && commit_at "2024-01-18T10:30:00" "chore: update config"

git merge -q perf/caching -m "Merge perf/caching"

# Security patches (hotfix)
git checkout -q -b hotfix/security
echo "func sanitize() {}" > security/sanitize.go
git add -A && commit_at "2024-01-19T08:00:00" "fix(security): add input sanitization"
echo "func validateToken() {}" >> security/auth.go
git add -A && commit_at "2024-01-19T09:00:00" "fix(security): add token validation"
echo "func rateLimit() {}" >> security/ratelimit.go
git add -A && commit_at "2024-01-19T10:00:00" "fix(security): add rate limiting"
git checkout -q main
git merge -q hotfix/security -m "Merge hotfix/security"

# Final cleanup commits
for i in $(seq 1 8); do
    echo "// cleanup $i" >> main.go
    git add -A && commit_at "2024-01-2$i T14:00:00" "chore: cleanup $i"
done

API_COMMITS=$(git rev-list --count HEAD)
success "api-server: $API_COMMITS commits created"
cd "$TEST_DIR"

###############################################################################
section "REPO 3: shared-lib - Shared library (40+ commits)"
###############################################################################

mkdir -p shared-lib && cd shared-lib
git init -q
git config user.email "platform@example.com"
git config user.name "Platform Dev"

# Create directories
mkdir -p src

echo "# Shared Library" > README.md
git add -A && commit_at "2024-01-01T07:00:00" "Initial commit"

# Utilities
for util in string array date number object; do
    echo "export const ${util}Utils = {}" > "src/${util}.js"
    git add -A && commit_at "2024-01-0$((2 + ${#util} % 5))T09:00:00" "feat: add $util utilities"
done

# Validators
git checkout -q -b feature/validators
for v in email phone url uuid; do
    echo "export const validate_$v = () => {}" >> src/validators.js
    git add -A && commit_at "2024-01-08T$((9 + ${#v})):00:00" "feat(validate): add $v validator"
done
git checkout -q main
git merge -q feature/validators -m "Merge feature/validators"

# Formatters
git checkout -q -b feature/formatters
for f in currency date number percent; do
    echo "export const format_$f = () => {}" >> src/formatters.js
    git add -A && commit_at "2024-01-12T$((9 + ${#f})):00:00" "feat(format): add $f formatter"
done
git checkout -q main
git merge -q feature/formatters -m "Merge feature/formatters"

# Types/interfaces
git checkout -q -b feature/types
for i in $(seq 1 8); do
    echo "export interface Type$i {}" >> src/types.ts
    git add -A && commit_at "2024-01-$((15 + i / 3))T$((9 + i)):00:00" "feat(types): add Type$i"
done
git checkout -q main
git merge -q feature/types -m "Merge feature/types"

# Constants
for i in $(seq 1 6); do
    echo "export const CONST_$i = '$i'" >> src/constants.js
    git add -A && commit_at "2024-01-$((18 + i / 3))T10:00:00" "feat: add CONST_$i"
done

# Bug fixes across multiple files
git checkout -q -b bugfix/edge-cases
echo "// edge case 1" >> src/string.js
git add -A && commit_at "2024-01-20T09:00:00" "fix: string edge case"
echo "// edge case 2" >> src/array.js
git add -A && commit_at "2024-01-20T10:00:00" "fix: array edge case"
echo "// edge case 3" >> src/date.js
git add -A && commit_at "2024-01-20T11:00:00" "fix: date edge case"
git checkout -q main
git merge -q bugfix/edge-cases -m "Merge bugfix/edge-cases"

SHARED_COMMITS=$(git rev-list --count HEAD)
success "shared-lib: $SHARED_COMMITS commits created"
cd "$TEST_DIR"

###############################################################################
section "REPO 4: devops - Infrastructure (30+ commits)"
###############################################################################

mkdir -p devops && cd devops
git init -q
git config user.email "devops@example.com"
git config user.name "DevOps Engineer"

# Create directories
mkdir -p docker k8s .github/workflows terraform monitoring scripts

echo "# DevOps" > README.md
git add -A && commit_at "2024-01-01T06:00:00" "Initial commit"

# Dockerfiles
for svc in webapp api worker; do
    echo "FROM node:18" > "docker/Dockerfile.$svc"
    git add -A && commit_at "2024-01-05T$((9 + ${#svc})):00:00" "infra: add $svc Dockerfile"
done

# Kubernetes configs
git checkout -q -b feature/k8s
for resource in deployment service ingress configmap secret; do
    echo "apiVersion: v1" > "k8s/${resource}.yaml"
    git add -A && commit_at "2024-01-10T$((9 + ${#resource} % 6)):00:00" "k8s: add $resource"
done
git checkout -q main
git merge -q feature/k8s -m "Merge feature/k8s"

# CI/CD pipelines
git checkout -q -b feature/cicd
echo "name: CI" > .github/workflows/ci.yml
git add -A && commit_at "2024-01-12T09:00:00" "ci: add CI pipeline"
echo "name: CD" > .github/workflows/cd.yml
git add -A && commit_at "2024-01-12T10:00:00" "ci: add CD pipeline"
echo "name: Tests" > .github/workflows/tests.yml
git add -A && commit_at "2024-01-12T11:00:00" "ci: add test pipeline"
git checkout -q main
git merge -q feature/cicd -m "Merge feature/cicd"

# Terraform
git checkout -q -b feature/terraform
for tf in main variables outputs providers; do
    echo "# Terraform $tf" > "terraform/${tf}.tf"
    git add -A && commit_at "2024-01-15T$((9 + ${#tf} % 4)):00:00" "infra(tf): add $tf"
done
git checkout -q main
git merge -q feature/terraform -m "Merge feature/terraform"

# Monitoring
for mon in prometheus grafana alerts; do
    echo "# $mon config" > "monitoring/${mon}.yaml"
    git add -A && commit_at "2024-01-18T$((9 + ${#mon} % 4)):00:00" "monitoring: add $mon"
done

# Scripts
for script in deploy rollback backup restore; do
    echo "#!/bin/bash" > "scripts/${script}.sh"
    git add -A && commit_at "2024-01-20T$((9 + ${#script} % 4)):00:00" "scripts: add $script"
done

DEVOPS_COMMITS=$(git rev-list --count HEAD)
success "devops: $DEVOPS_COMMITS commits created"
cd "$TEST_DIR"

###############################################################################
section "TRACKING REPOS"
###############################################################################

for repo in webapp api-server shared-lib devops; do
    $FP track "$TEST_DIR/$repo"
done
success "All 4 repos tracked"

###############################################################################
section "BACKFILL: Testing flags"
###############################################################################

cd "$TEST_DIR/webapp"

log "Testing --dry-run..."
DRY_RUN=$($FP backfill --dry-run 2>&1)
DRY_COUNT=$(echo "$DRY_RUN" | grep -c "^  " || true)
if [ "$DRY_COUNT" -gt 40 ]; then
    success "Dry run found $DRY_COUNT commits"
else
    fail "Expected 40+ commits, got $DRY_COUNT"
fi

log "Testing --limit..."
LIMIT_OUTPUT=$($FP backfill --dry-run --limit=10 2>&1)
if echo "$LIMIT_OUTPUT" | grep -q "Found 10 commits"; then
    success "Limit flag works"
else
    fail "Limit flag broken"
fi

log "Testing --since..."
SINCE_OUTPUT=$($FP backfill --dry-run --since=2024-01-20 2>&1)
SINCE_COUNT=$(echo "$SINCE_OUTPUT" | grep -c "^  " || true)
if [ "$SINCE_COUNT" -gt 0 ] && [ "$SINCE_COUNT" -lt 20 ]; then
    success "Since flag works (found $SINCE_COUNT recent commits)"
else
    fail "Since flag broken"
fi

cd "$TEST_DIR"

###############################################################################
section "BACKFILL: Running for all repos"
###############################################################################

TOTAL_EXPECTED=0
for repo in webapp api-server shared-lib devops; do
    cd "$TEST_DIR/$repo"
    REPO_COMMITS=$(git rev-list --count HEAD)
    TOTAL_EXPECTED=$((TOTAL_EXPECTED + REPO_COMMITS))
    log "Backfilling $repo ($REPO_COMMITS commits)..."
    $FP backfill --background
    cd "$TEST_DIR"
done

sleep 2  # Let background processes finish

log "Verifying backfill..."
BACKFILL_COUNT=$($FP activity --source=backfill --oneline 2>/dev/null | grep -c "BACKFILL" || true)
success "Backfilled $BACKFILL_COUNT events (expected ~$TOTAL_EXPECTED)"

###############################################################################
section "BACKFILL: Testing idempotency"
###############################################################################

log "Running backfill again..."
for repo in webapp api-server shared-lib devops; do
    cd "$TEST_DIR/$repo"
    $FP backfill --background
    cd "$TEST_DIR"
done
sleep 2

BACKFILL_COUNT_2=$($FP activity --source=backfill --oneline 2>/dev/null | grep -c "BACKFILL" || true)
if [ "$BACKFILL_COUNT_2" -eq "$BACKFILL_COUNT" ]; then
    success "Idempotent: still $BACKFILL_COUNT_2 events (no duplicates)"
else
    fail "Duplicates created: was $BACKFILL_COUNT, now $BACKFILL_COUNT_2"
fi

###############################################################################
section "EXPORT: Exporting backfilled data"
###############################################################################

log "Pending events before export:"
$FP activity --oneline 2>/dev/null | head -5
echo "  ... ($BACKFILL_COUNT total)"

log "Running export..."
$FP export --now

log "Export directory:"
find "$EXPORT_DIR" -type f -name "*.csv" 2>/dev/null | while read -r f; do
    lines=$(($(wc -l < "$f") - 1))
    echo "  $f ($lines rows)"
done

success "Export completed"

###############################################################################
section "VERIFICATION"
###############################################################################

log "Checking CSV contents..."
TOTAL_ROWS=0
for csv in "$EXPORT_DIR"/*.csv "$EXPORT_DIR"/commits-*.csv; do
    if [ -f "$csv" ]; then
        rows=$(($(wc -l < "$csv") - 1))
        TOTAL_ROWS=$((TOTAL_ROWS + rows))
        echo "  $(basename "$csv"): $rows commits"
    fi
done

echo ""
echo "Total commits exported: $TOTAL_ROWS"
echo "Total events backfilled: $BACKFILL_COUNT"

if [ "$TOTAL_ROWS" -gt 100 ]; then
    success "Export contains substantial data ($TOTAL_ROWS rows)"
else
    fail "Export seems incomplete ($TOTAL_ROWS rows)"
fi

###############################################################################
section "SUMMARY"
###############################################################################

echo ""
echo "Repos: 4 (webapp, api-server, shared-lib, devops)"
echo "Total commits: $TOTAL_EXPECTED"
echo "Backfilled events: $BACKFILL_COUNT"
echo "Exported rows: $TOTAL_ROWS"
echo ""
success "All backfill + export tests passed!"
