# Jira CLI

A fast, scriptable command-line interface for **Jira** — Cloud and Server/Data Center — over the
REST API. Manage issues, JQL search, transitions, comments, worklogs, attachments, links, boards
and sprints from your terminal.

Designed for both human operators and coding agents. All list/get commands support `-o json` and
`-o yaml` for machine-readable output.

[![Go Version](https://img.shields.io/github/go-mod/go-version/piyush-gambhir/jira-cli)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/piyush-gambhir/jira-cli)](https://github.com/piyush-gambhir/jira-cli/releases)
[![License](https://img.shields.io/github/license/piyush-gambhir/jira-cli)](LICENSE)
[![CI](https://github.com/piyush-gambhir/jira-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/piyush-gambhir/jira-cli/actions/workflows/ci.yml)

> Independent open-source wrapper over the Jira REST API. Not affiliated with Atlassian and unrelated
> to Atlassian's official `acli`.

## Features

- **Every auth method** — Cloud API token (default), scoped API token, OAuth 2.0 (3LO) browser login
  with auto-refresh, Server/DC personal access token, and username/password. See [docs/CREDENTIALS.md](docs/CREDENTIALS.md).
- **Issues** — JQL search, list, get, create, edit, delete, assign, transition.
- **Around issues** — comments, worklogs, attachments (upload/download), links, watchers, votes.
- **Projects, users, fields** and **Agile** boards / sprints / epics.
- **ADF-aware** — descriptions and comments are sent as Atlassian Document Format; `--markdown` supported.
- **Agent-friendly** — `-o json|yaml`, `--read-only` safety mode, duplicate-safe issue creation,
  `--no-input`, env-var config, clean stdout/stderr separation, automatic rate-limit (429) backoff.
- **Cross-platform** — macOS, Linux, Windows (amd64 and arm64). Multiple named profiles.

## Installation

### Install script (recommended — no Go required)

Installs the latest prebuilt binary for your OS/arch:

```bash
curl -sSfL https://raw.githubusercontent.com/piyush-gambhir/jira-cli/main/install.sh | sh
```

Pin a version or change the install directory:

```bash
curl -sSfL https://raw.githubusercontent.com/piyush-gambhir/jira-cli/main/install.sh | VERSION=0.1.1 sh
curl -sSfL https://raw.githubusercontent.com/piyush-gambhir/jira-cli/main/install.sh | INSTALL_DIR=~/.local/bin sh
```

### Download a release manually

Prebuilt binaries for every platform are attached to each
[GitHub Release](https://github.com/piyush-gambhir/jira-cli/releases). Download the archive for your
OS/arch, extract, and put `jira` on your `PATH`:

```bash
# example (macOS arm64) — adjust the version/asset name
curl -sSfL https://github.com/piyush-gambhir/jira-cli/releases/latest/download/jira-cli_darwin_arm64.tar.gz | tar xz
sudo mv jira /usr/local/bin/
```

Assets are named `jira-cli_<os>_<arch>.tar.gz` (lowercase), where `<os>` is `darwin` or `linux` and
`<arch>` is `amd64` or `arm64` (Windows ships as `jira-cli_windows_amd64.zip`).

### Alternative (build from source, requires Go)

Requires the Go 1.22+ toolchain. The CLI lives in the [`cli-go/`](cli-go) directory of this
monorepo (the marketing site and docs live in [`web/`](web)).

```bash
git clone https://github.com/piyush-gambhir/jira-cli.git
cd jira-cli/cli-go
make install          # builds and installs `jira` to $GOBIN / $GOPATH/bin
```

The binary is `jira`.

## Quick start

```bash
# 1. Authenticate — API token is the simplest setup
jira auth login --type api_token
#    prefer an API token (no app needed)? — prompts for site, email, token
jira auth login --type api_token
#    or non-interactively / in CI (uses API token under the hood):
export JIRA_SITE=https://acme.atlassian.net JIRA_EMAIL=me@acme.com JIRA_TOKEN=xxxx

# 2. Confirm
jira whoami
jira status

# 3. Work
jira issue search "assignee = currentUser() AND statusCategory != Done" -o json
jira issue create -p ABC --type Task --summary "Set up CI" -d "Details here"
jira issue transition ABC-123 "In Progress"
jira issue comment ABC-123 --body "On it"
```

`issue create` is safe to rerun: an exact type/summary/parent match created by
the current user in the same project within 10 minutes is reused. Pass
`--allow-duplicate` when another identical issue is intentional.

## Authentication methods

| `--type`  | Deployment   | Credential            |
|-----------|--------------|-----------------------|
| `oauth2` (default) | Cloud | OAuth 2.0 3LO browser login (requires your app client ID and secret) |
| `api_token` | Cloud      | email + API token (no app needed) |
| `scoped`  | Cloud        | email + scoped token (api.atlassian.com gateway) |
| `pat`     | Server/DC    | bearer personal access token |
| `basic`   | Server/DC    | username + password |

```bash
jira auth login --type pat --site https://jira.company.com --token "$PAT"
jira auth login --type oauth2 --client-id "$ID" --client-secret "$SECRET" --scope-preset admin
jira auth list   ;   jira auth use staging   ;   jira auth logout --name staging
```

OAuth scopes are selectable: `--scope-preset read|write|admin|all` (or an interactive picker), plus
granular `--scope`. See [docs/CREDENTIALS.md](docs/CREDENTIALS.md).

### For end users

Everyone authenticates to **their own** Jira account. For OAuth, register a 3LO app and pass its
client ID and secret; these credentials are stored only in your local protected config. For the
simplest setup, run `jira auth login --type api_token` and paste a token from id.atlassian.com.

## Output

`-o table` (default), `-o json`, `-o yaml`. Informational messages go to stderr; data goes to stdout,
so `jira ... -o json | jq` is always safe.

## Documentation

- [CLAUDE.md](CLAUDE.md) — complete command guide and workflows
- [jira/references/commands.md](jira/references/commands.md) — full command/flag reference
- [docs/CREDENTIALS.md](docs/CREDENTIALS.md) — all authentication methods, in depth
- API references: [issues](jira/references/issues-api.md),
  [search/JQL/projects/users](jira/references/search-jql-projects-users.md),
  [agile/ADF/conventions](jira/references/agile-adf-conventions.md)

## Releasing (maintainers)

Releases are built by [GoReleaser](https://goreleaser.com) via GitHub Actions on any `v*` tag — see
[`.github/workflows/release.yml`](.github/workflows/release.yml) and [`.goreleaser.yaml`](.goreleaser.yaml).

```bash
git tag -s v0.1.0 -m "v0.1.0"     # signed tag
git push origin v0.1.0            # -> CI builds cross-platform binaries + checksums and publishes a Release
```

Release binaries never embed OAuth client secrets. Users bring their own 3LO app credentials or use
an API token; a future zero-configuration OAuth experience should use a server-side authorization
broker so no reusable client secret is distributed in a public binary.

## Development

```bash
make build     # -> bin/jira
make test      # go test -race ./...
make vet
make fmt       # gofmt -s -w .
```

Commits and tags are **signed and required** — the repository enforces a "require signed commits" ruleset
on GitHub, and local config (`commit.gpgsign`, `tag.gpgsign`) signs by default. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

See [LICENSE](LICENSE).
