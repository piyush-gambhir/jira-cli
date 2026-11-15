#!/usr/bin/env bash
set -euo pipefail

if [[ $# -gt 1 || ( $# -eq 1 && "$1" != "production" ) ]]; then
  echo "usage: bash scripts/deploy-docs.sh [production]" >&2
  exit 2
fi

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DEPLOY_ENV_FILE="$ROOT_DIR/.env.deploy.production"

if [[ -f "$DEPLOY_ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$DEPLOY_ENV_FILE"
  set +a
fi

cd "$ROOT_DIR/web"
pnpm install --frozen-lockfile
pnpm build:cloudflare
pnpm test:search
pnpm deploy:cloudflare
