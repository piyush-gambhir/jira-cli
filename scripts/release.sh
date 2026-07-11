#!/usr/bin/env bash
# Build release binaries for one or more platforms.
#
# Usage:
#   ./scripts/release.sh                       # build for the host platform into dist/
#   PLATFORMS="darwin/arm64 linux/amd64" ./scripts/release.sh   # cross-compile matrix
#
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

MODULE="github.com/piyush-gambhir/jira-cli"
BINARY="jira"
VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo dev)"
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
BUILD_TIME="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"

LDFLAGS="-s -w \
  -X ${MODULE}/internal/version.Version=${VERSION} \
  -X ${MODULE}/internal/version.Commit=${COMMIT} \
  -X ${MODULE}/internal/version.BuildTime=${BUILD_TIME}"

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
