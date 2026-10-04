#!/usr/bin/env bash
#
# release.sh — bump version, tag, and push a Go module to GitHub.
#
# Usage:
#   ./scripts/release.sh            # patch: v0.1.3 -> v0.1.4 (default)
#   ./scripts/release.sh minor      # minor: v0.1.4 -> v0.2.0
#   ./scripts/release.sh major      # major: v0.2.0 -> v1.0.0
#   ./scripts/release.sh v0.3.0     # explicit version
#
# Options (env vars):
#   YES=1        skip the confirmation prompt (CI)
#   SKIP_TESTS=1 skip `go test ./...`
#   NO_GH=1      don't create a GitHub Release even if `gh` is installed
#   BRANCH=main  branch releases must come from (default: main)
#
# Works with the bash 3.2 that ships with macOS.

set -euo pipefail

BRANCH="${BRANCH:-main}"
REMOTE="${REMOTE:-origin}"
BUMP="${1:-patch}"

# ---------- helpers ----------
if [ -t 1 ]; then
  RED=$'\033[31m'; GREEN=$'\033[32m'; YELLOW=$'\033[33m'; BOLD=$'\033[1m'; RESET=$'\033[0m'
else
  RED=""; GREEN=""; YELLOW=""; BOLD=""; RESET=""
fi
info() { printf '%s==>%s %s\n' "$BOLD" "$RESET" "$*"; }
ok()   { printf '%s✓%s %s\n' "$GREEN" "$RESET" "$*"; }
warn() { printf '%s!%s %s\n' "$YELLOW" "$RESET" "$*"; }
die()  { printf '%s✗ %s%s\n' "$RED" "$*" "$RESET" >&2; exit 1; }

# ---------- preflight ----------
cd "$(git rev-parse --show-toplevel 2>/dev/null)" || die "not inside a git repository"
[ -f go.mod ] || die "go.mod not found in repo root"

MODULE="$(awk '/^module /{print $2; exit}' go.mod)"
[ -n "$MODULE" ] || die "could not read module path from go.mod"

current_branch="$(git rev-parse --abbrev-ref HEAD)"
[ "$current_branch" = "$BRANCH" ] || die "you are on '$current_branch'; releases must be made from '$BRANCH'"

[ -z "$(git status --porcelain)" ] || die "working tree has uncommitted changes; commit or stash them first"

info "Fetching $REMOTE..."
git fetch --quiet --tags "$REMOTE" "$BRANCH"

local_sha="$(git rev-parse HEAD)"
remote_sha="$(git rev-parse "$REMOTE/$BRANCH" 2>/dev/null || echo "")"
if [ -n "$remote_sha" ] && [ "$local_sha" != "$remote_sha" ]; then
  if git merge-base --is-ancestor "$local_sha" "$remote_sha"; then
    die "local $BRANCH is behind $REMOTE/$BRANCH; run 'git pull' first"
  elif ! git merge-base --is-ancestor "$remote_sha" "$local_sha"; then
    die "local $BRANCH and $REMOTE/$BRANCH have diverged; sync them first"
  fi
  # local is ahead: those commits get pushed below
fi

# ---------- compute next version ----------
LATEST="$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -n1 || true)"
[ -n "$LATEST" ] || LATEST="v0.0.0"

ver="${LATEST#v}"
MAJOR="${ver%%.*}"; rest="${ver#*.}"
MINOR="${rest%%.*}"; PATCH="${rest#*.}"

case "$BUMP" in
  patch) NEXT="v${MAJOR}.${MINOR}.$((PATCH + 1))" ;;
  minor) NEXT="v${MAJOR}.$((MINOR + 1)).0" ;;
  major) NEXT="v$((MAJOR + 1)).0.0" ;;
  v[0-9]*)
    echo "$BUMP" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$' || die "invalid version '$BUMP' (want vX.Y.Z)"
    NEXT="$BUMP" ;;
  -h|--help) sed -n '2,18p' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
  *) die "unknown argument '$BUMP' (use patch, minor, major, or vX.Y.Z)" ;;
esac

git rev-parse -q --verify "refs/tags/$NEXT" >/dev/null && die "tag $NEXT already exists"

# Go requires the module path to end in /vN for major versions >= 2.
next_major="${NEXT#v}"; next_major="${next_major%%.*}"
if [ "$next_major" -ge 2 ]; then
  case "$MODULE" in
    */v"$next_major") ;;
    *) die "$NEXT is a v$next_major release, so go.mod must be 'module $MODULE/v$next_major' (currently '$MODULE')" ;;
  esac
fi

# ---------- checks ----------
info "Checking go.mod/go.sum are tidy..."
go mod tidy
if [ -n "$(git status --porcelain go.mod go.sum 2>/dev/null)" ]; then
  git --no-pager diff --stat go.mod go.sum || true
  git checkout -- go.mod 2>/dev/null || true
  git checkout -- go.sum 2>/dev/null || true
  die "'go mod tidy' changed go.mod/go.sum; run it, commit, and try again"
fi
ok "go.mod is tidy"

info "Running go vet..."
go vet ./...
ok "vet passed"

if [ "${SKIP_TESTS:-0}" != "1" ]; then
  info "Running tests..."
  go test ./...
  ok "tests passed"
else
  warn "skipping tests (SKIP_TESTS=1)"
fi

# ---------- changelog ----------
if [ "$LATEST" = "v0.0.0" ] && ! git rev-parse -q --verify "refs/tags/v0.0.0" >/dev/null; then
  RANGE="HEAD"
else
  RANGE="$LATEST..HEAD"
fi
CHANGES="$(git log --no-merges --pretty='- %s (%h)' "$RANGE")"
[ -n "$CHANGES" ] || die "no new commits since $LATEST; nothing to release"

echo
printf '%sModule:%s  %s\n' "$BOLD" "$RESET" "$MODULE"
printf '%sVersion:%s %s -> %s%s%s\n' "$BOLD" "$RESET" "$LATEST" "$GREEN" "$NEXT" "$RESET"
printf '%sChanges:%s\n%s\n\n' "$BOLD" "$RESET" "$CHANGES"

if [ "${YES:-0}" != "1" ]; then
  printf 'Release %s? [y/N] ' "$NEXT"
  read -r answer
  case "$answer" in y|Y|yes|YES) ;; *) die "aborted" ;; esac
fi

# ---------- tag & push ----------
info "Pushing $BRANCH..."
git push --quiet "$REMOTE" "$BRANCH"

info "Creating tag $NEXT..."
git tag -a "$NEXT" -m "Release $NEXT

$CHANGES"
git push --quiet "$REMOTE" "$NEXT"
ok "pushed tag $NEXT"

# ---------- GitHub release (optional) ----------
if [ "${NO_GH:-0}" != "1" ] && command -v gh >/dev/null 2>&1; then
  if gh release create "$NEXT" --title "$NEXT" --notes "$CHANGES" --verify-tag >/dev/null 2>&1; then
    ok "created GitHub Release $NEXT"
  else
    warn "tag pushed, but creating the GitHub Release failed (run 'gh auth login'?)"
  fi
fi

echo
ok "Released ${BOLD}${MODULE}@${NEXT}${RESET}"
echo "  Update other repos with:"
echo "    go get ${MODULE}@${NEXT}"