# jira-cli

A fast, scriptable command-line interface for **Jira** — Cloud and Server/Data Center — over the
REST API. Manage issues, JQL search, transitions, comments, worklogs, attachments, links, boards
and sprints from your terminal, with first-class JSON/YAML output for scripts and coding agents.

> This is an independent, open wrapper over the Jira REST API. It is **not** affiliated with
> Atlassian and is unrelated to Atlassian's official `acli`.

## Features

- **Every auth method**: Cloud API token (default), scoped API token, OAuth 2.0 (3LO) with browser
  login + auto-refresh, Server/DC personal access token, and username/password. See
  [docs/CREDENTIALS.md](docs/CREDENTIALS.md).
- **Issues**: search (JQL), list, get, create, edit, delete, assign, transition.
- **Around issues**: comments, worklogs, attachments (upload/download), links, watchers, votes.
- **Projects, users, fields**, and **Agile** boards / sprints / epics.
- **ADF-aware**: descriptions/comments are sent as Atlassian Document Format; `--markdown` supported.
- **Agent-friendly**: `-o json|yaml`, `--read-only` safety mode, `--no-input`, env-var config,
  clean stderr/stdout separation, automatic 429 backoff.
- Multiple named **profiles** with `flags > env > config` precedence.

## Install

```bash
# From source (Go 1.22+)
git clone https://github.com/piyush-gambhir/jira-cli
cd jira-cli && make install        # or: ./install.sh

# Or directly
go install github.com/piyush-gambhir/jira-cli@latest
```

The binary is `jira`.

## Quick start

```bash
# 1. Authenticate (Cloud, API token — create one at id.atlassian.com)
jira auth login
#    or non-interactively / in CI:
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

## Authentication methods

| `--type`  | Deployment   | Credential            |
|-----------|--------------|-----------------------|
| `api_token` (default) | Cloud | email + API token |
| `scoped`  | Cloud        | email + scoped token (api.atlassian.com gateway) |
| `oauth2`  | Cloud        | OAuth 2.0 3LO (browser; refresh tokens) |
| `pat`     | Server/DC    | bearer personal access token |
| `basic`   | Server/DC    | username + password |

```bash
jira auth login --type pat --site https://jira.company.com --token "$PAT"
jira auth login --type oauth2 --client-id "$ID" --client-secret "$SECRET"
jira auth list
jira auth use staging
```

### For end users (after publishing)

The CLI ships with **no embedded credentials** — every user authenticates to **their own** Jira
account. The simplest path needs nothing from the maintainer: install, then `jira auth login` and
paste your own API token (create one at id.atlassian.com in ~30s). OAuth is optional and "bring your
own app" by default.

Maintainers who want **one-command browser login** for end users (no per-user app registration) can
publish a build with a built-in OAuth app baked in:

```bash
JIRA_OAUTH_CLIENT_ID=… JIRA_OAUTH_CLIENT_SECRET=… make build
```

Each user still logs into their own account in the browser; only the app identity and rate-limit
quota are shared. The secret lives only in the released binary, not in source. See
[docs/CREDENTIALS.md](docs/CREDENTIALS.md#distributing-a-branded-build-built-in-oauth-app) for the tradeoffs.

## Output

`-o table` (default), `-o json`, `-o yaml`. Informational messages go to stderr; data goes to stdout.

## Documentation

- [CLAUDE.md](CLAUDE.md) — complete command guide and workflows
- [jira/references/commands.md](jira/references/commands.md) — full command/flag reference
- [docs/CREDENTIALS.md](docs/CREDENTIALS.md) — all authentication methods, in depth
- API references: [issues](jira/references/issues-api.md),
  [search/JQL/projects/users](jira/references/search-jql-projects-users.md),
  [agile/ADF/conventions](jira/references/agile-adf-conventions.md)

## Configuration

Profiles live in `~/.config/jira-cli/config.yaml` (mode 0600). Override per-invocation with flags
(`--site`, `--email`, `--token`, `--profile`, …) or environment variables (`JIRA_SITE`, `JIRA_EMAIL`,
`JIRA_TOKEN`, …). Precedence: **flags > env > config**.

## License

See [LICENSE](LICENSE).
