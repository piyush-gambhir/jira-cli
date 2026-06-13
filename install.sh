#!/usr/bin/env bash
# Install the jira CLI from source.
#
# Usage:
#   ./install.sh            # build and install to $GOBIN (or $GOPATH/bin)
#   PREFIX=/usr/local ./install.sh   # also symlink into $PREFIX/bin
set -euo pipefail

MODULE="github.com/piyush-gambhir/jira-cli"
BINARY="jira"

if ! command -v go >/dev/null 2>&1; then
  echo "error: Go toolchain not found. Install Go 1.22+ from https://go.dev/dl/" >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Load build secrets (e.g. an embedded OAuth app) from .env if present.
if [ -f .env ]; then set -a; . ./.env; set +a; fi

VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo dev)"
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
BUILD_TIME="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
LDFLAGS="-s -w \
  -X ${MODULE}/internal/version.Version=${VERSION} \
  -X ${MODULE}/internal/version.Commit=${COMMIT} \
  -X ${MODULE}/internal/version.BuildTime=${BUILD_TIME} \
  -X ${MODULE}/internal/auth.EmbeddedClientID=${JIRA_OAUTH_CLIENT_ID:-} \
  -X ${MODULE}/internal/auth.EmbeddedClientSecret=${JIRA_OAUTH_CLIENT_SECRET:-}"

echo "Building ${BINARY} ${VERSION} (${COMMIT})..."
go install -ldflags "${LDFLAGS}" .

GOBIN="$(go env GOBIN)"
[ -z "$GOBIN" ] && GOBIN="$(go env GOPATH)/bin"
echo "Installed ${BINARY} to ${GOBIN}/${BINARY}"

if [ -n "${PREFIX:-}" ]; then
  mkdir -p "${PREFIX}/bin"
  ln -sf "${GOBIN}/${BINARY}" "${PREFIX}/bin/${BINARY}"
  echo "Symlinked to ${PREFIX}/bin/${BINARY}"
fi

echo
echo "Run 'jira auth login' to get started, or 'jira --help'."
