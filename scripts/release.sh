#!/usr/bin/env bash
# Build release binaries with the embedded OAuth app baked in from .env.
#
# Usage:
#   ./scripts/release.sh                       # build for the host platform into dist/
#   PLATFORMS="darwin/arm64 linux/amd64" ./scripts/release.sh   # cross-compile matrix
#
# Reads JIRA_OAUTH_CLIENT_ID / JIRA_OAUTH_CLIENT_SECRET from .env (or the
# environment). If empty, the binaries are built WITHOUT an embedded OAuth app
# (users then bring their own API token or OAuth app).
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

# Load secrets from .env (robust shell sourcing — handles values Make can't).
if [ -f .env ]; then
  set -a
  . ./.env
  set +a
fi
: "${JIRA_OAUTH_CLIENT_ID:=}"
: "${JIRA_OAUTH_CLIENT_SECRET:=}"

MODULE="github.com/piyush-gambhir/jira-cli"
BINARY="jira"
VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo dev)"
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
BUILD_TIME="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"

LDFLAGS="-s -w \
  -X ${MODULE}/internal/version.Version=${VERSION} \
  -X ${MODULE}/internal/version.Commit=${COMMIT} \
  -X ${MODULE}/internal/version.BuildTime=${BUILD_TIME} \
  -X ${MODULE}/internal/auth.EmbeddedClientID=${JIRA_OAUTH_CLIENT_ID} \
  -X ${MODULE}/internal/auth.EmbeddedClientSecret=${JIRA_OAUTH_CLIENT_SECRET}"

if [ -n "${JIRA_OAUTH_CLIENT_ID}" ] && [ -n "${JIRA_OAUTH_CLIENT_SECRET}" ]; then
  echo "Embedding built-in OAuth app (client_id=${JIRA_OAUTH_CLIENT_ID:0:6}…)"
else
  echo "No OAuth app embedded (set JIRA_OAUTH_CLIENT_ID/SECRET in .env to enable browser login out of the box)"
fi

PLATFORMS="${PLATFORMS:-$(go env GOOS)/$(go env GOARCH)}"
mkdir -p dist
for p in $PLATFORMS; do
  os="${p%/*}"
  arch="${p#*/}"
  out="dist/${BINARY}-${os}-${arch}"
  [ "$os" = "windows" ] && out="${out}.exe"
  echo "  building $out (${VERSION})"
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o "$out" .
done

echo "Done. Artifacts in ./dist/"
