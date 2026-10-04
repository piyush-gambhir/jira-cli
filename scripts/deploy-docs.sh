#!/usr/bin/env bash
# Build the docs site in web/ and deploy it as the Cloudflare Worker that serves
# projects.piyushgambhir.com/jira-cli. Run from an up-to-date main after
# `wrangler login`, or put CLOUDFLARE_API_TOKEN in .env.deploy.production.
set -euo pipefail

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
