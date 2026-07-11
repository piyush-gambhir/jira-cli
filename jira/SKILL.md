---
name: jira
description: "Expert guide for using the jira CLI to manage Jira (Cloud and Server/Data Center) from the terminal. Use this skill whenever the user mentions Jira issues, tickets, JQL, sprints, boards, epics, backlog, issue transitions, assigning/creating/closing tickets, comments, worklogs or time logging, attachments, issue links, watchers, votes, Jira projects/users/fields, or automating Jira from the command line. Also trigger when the user wants to search issues with JQL, create/edit/transition an issue, log work, move issues across sprints, or authenticate to Jira via API token, OAuth 2.0, or a personal access token."
---

# Jira CLI -- Agent Skill Guide

Complete guide for coding agents to manage Jira via the `jira` CLI. Covers issues, JQL search,
transitions, comments, worklogs, attachments, links, watchers, votes, projects, users, fields, and
Agile boards/sprints/epics — on both Jira Cloud and Server/Data Center.

## 1. Prerequisites & Setup

### Install

```bash
# go install
go install github.com/piyush-gambhir/jira-cli@latest
# from source
git clone https://github.com/piyush-gambhir/jira-cli.git && cd jira-cli && make install
# install script
curl -sSL https://raw.githubusercontent.com/piyush-gambhir/jira-cli/main/install.sh | bash
```

The binary is `jira`.

### Authenticate

Easiest (and best for CI/agents) — set environment variables, no `auth login` needed:

```bash
export JIRA_SITE=https://acme.atlassian.net   # Cloud site, or https://jira.company.com for Server/DC
export JIRA_EMAIL=me@acme.com                  # Atlassian account email (Cloud)
export JIRA_TOKEN=xxxx                          # API token from id.atlassian.com
```

Or save a profile interactively (also supports scoped token, OAuth 2.0, PAT, basic):

```bash
jira auth login --client-id "$ID" --client-secret "$SECRET" # browser OAuth with your app
jira auth login --type api_token                 # paste site / email / API token (no app needed)
jira auth login --type pat --site https://jira.company.com --token "$PAT"   # Server/DC
```

Config priority: **CLI flags > environment variables > profile config** (`~/.config/jira-cli/config.yaml`).

### Verify

```bash
jira whoami -o json     # current user (includes accountId)
jira status -o json     # site reachability + auth check + serverInfo
```

## 2. Core Principles for Agents

- **ALWAYS use `-o json` for parsing.** Every list/get supports `-o json` / `-o yaml`; default is human
  `table`. Informational lines go to **stderr**, data to **stdout**, so `jira ... -o json | jq` is safe.
- **Users are accountIds (GDPR — no usernames).** Pass an email, display name, `@me`, or `id:<accountId>`
  to `--assignee` / `--to` / `--user`; the CLI resolves it. To look one up: `jira user search "<name>" -o json`.
- **Rich text is ADF.** `--description` / `--body` / `--comment` are sent as Atlassian Document Format.
  Add `--markdown` to interpret the text as lightweight markdown (headings, lists, bold/italic/code/links).
- **Transitions are discover-then-do.** There is no "set status". List first, then move:
  `jira issue transitions ABC-1` → `jira issue transition ABC-1 "In Progress"` (matches status OR
  transition name, case-insensitively).
- **JQL must be bounded** (contain a real restriction). Get a count with `jira issue search "<jql>" --count`.
  Cloud paginates with a cursor; use `--all` to fetch everything, `--limit N` otherwise.
- **Safety:** `--read-only` (or `read_only` in the profile) blocks every write command; `jira issue delete`
  and `jira project delete` need `--yes`. Use `--no-input` in non-interactive contexts.
- **Multiple sites:** `--profile <name>`, managed with `jira auth list|use|logout`.

## 3. Common Workflows

### Find issues

```bash
jira issue list --mine                                   # my open work
jira issue list -p ABC --status "In Progress" -o json
jira issue search "project = ABC AND statusCategory != Done ORDER BY updated DESC" -o json
jira issue search "assignee = currentUser() AND resolution IS EMPTY" --all -o json
jira issue search "project = ABC AND issuetype = Bug" --count
```

### Inspect an issue

```bash
jira issue get ABC-123 -o json
jira issue get ABC-123 --fields summary,status,assignee --expand changelog -o json
```

### Create / edit / assign / delete

```bash
jira issue create -p ABC --type Task --summary "Set up CI" -d "Steps:\n- one\n- two" --markdown -a @me -l ci
jira issue create -p ABC --type Bug --summary "X" --field customfield_10010=5     # custom field
jira issue edit ABC-123 --summary "New title" --add-label triaged --remove-label blocker
jira issue assign ABC-123 --to jane@acme.com         # or @me / id:<accountId> / default / none
jira issue delete ABC-123 --yes
```

### Transition (change status)

```bash
jira issue transitions ABC-123                       # discover available transitions
jira issue transition ABC-123 "In Progress"
jira issue transition ABC-123 Done --resolution Fixed --comment "Shipped"
```

### Comments, worklogs, attachments, links

```bash
jira issue comment ABC-123 --body "Looks good" [--markdown]
jira issue comments ABC-123 -o json
jira issue worklog ABC-123 --time "2h 30m" --comment "review" --started "2026-06-13T09:00:00.000+0000"
jira issue attach ABC-123 ./error.log ./screenshot.png
jira issue attachments ABC-123 -o json   ;   jira issue download <attachmentId> --out ./file
jira issue link ABC-1 ABC-2 --type Blocks   ;   jira issue link-types
```

### Watch / vote

```bash
jira issue watch ABC-123 ; jira issue unwatch ABC-123 ; jira issue watchers ABC-123 -o json
jira issue vote  ABC-123 ; jira issue unvote  ABC-123 ; jira issue votes   ABC-123 -o json
```

### Projects, users, fields

```bash
jira project list --query mobile -o json     ;   jira project get ABC -o json
jira project create --key MOB --name "Mobile" --lead @me --template <projectTemplateKey>   # needs admin
jira user search "jane" -o json              # resolve a person to an accountId
jira field list --custom -o json             # custom field ids + their JQL clause names
```

### Agile (boards, sprints, epics)

```bash
jira board list -p ABC -o json   ;   jira board issues 42 -o json   ;   jira board backlog 42 -o json
jira sprint list --board 42 --state active -o json
jira sprint get 100 -o json      ;   jira sprint issues 100 -o json
jira sprint create --board 42 --name "Sprint 5" --goal "Ship X"
jira epic get ABC-10 -o json     ;   jira epic issues ABC-10 -o json
```

## 4. Command Reference (compact)

| Group | Key commands |
|---|---|
| **Top-level** | `auth (login/list/use/logout)`, `whoami`, `status`, `version`, `update` |
| **issue** | `list`, `search`, `get`, `create`, `edit`, `delete`, `assign`, `transitions`, `transition` |
| **issue (more)** | `comment(s)`, `worklog(s)`, `attach`, `attachments`, `download`, `link`, `link-types`, `watch`, `unwatch`, `watchers`, `vote`, `unvote`, `votes` |
| **project** | `list`, `get`, `create`, `update`, `delete` |
| **user / field** | `user search`, `user get`, `field list` |
| **board** | `list`, `get`, `issues`, `backlog` |
| **sprint / epic** | `sprint list/get/create/issues`, `epic get/issues` |

### Global flags

| Flag | Description |
|---|---|
| `-o, --output` | `table` (default), `json`, `yaml` |
| `--profile` | Profile to use |
| `-s, --site` | Site URL override |
| `-e, --email` | Cloud account email override |
| `-t, --token` | API token / PAT override |
| `-u, --user` | Username override (Server/DC basic) |
| `--api-version` | `3` (Cloud) or `2` (Server/DC) |
| `-k, --insecure` | Skip TLS verification |
| `--read-only` | Block writes (agent safety) |
| `--no-input` | No prompts (CI/agents) |
| `-v, --verbose` / `-q, --quiet` | Verbose HTTP / suppress info |

### Notable per-command flags

| Command | Flags |
|---|---|
| `issue list` | `-p/--project`, `--status`, `--assignee`, `--mine`, `--type`, `--label`, `-n/--limit`, `--all`, `--fields` |
| `issue search` | `-j/--jql`, `-n/--limit`, `--all`, `--fields`, `--count` |
| `issue create` | `-p/--project`, `--type`, `--summary`, `-d/--description`, `--markdown`, `--priority`, `-a/--assignee`, `-l/--label`, `--parent`, `--field k=v` |
| `issue transition` | `--resolution`, `--comment`, `--markdown` |
| `issue worklog` | `--time`, `--comment`, `--started`, `--adjust`, `--new-estimate` |
| `auth login` | `--type`, `--name`, `--scope-preset read\|write\|admin\|all`, `--scope`, `--client-id/secret` |

See [references/commands.md](references/commands.md) for the complete reference.

## 5. Authentication methods

| `--type` | Deployment | Credential |
|---|---|---|
| `oauth2` (default) | Cloud | OAuth 2.0 (3LO) browser login (refresh tokens; requires your app credentials) |
| `api_token` | Cloud | email + API token (no app needed) |
| `scoped` | Cloud | email + scoped token (api.atlassian.com gateway) |
| `pat` | Server/DC | bearer personal access token |
| `basic` | Server/DC | username + password |

Full reference (flows, scopes, gateway/cloudId, refresh): [../docs/CREDENTIALS.md](../docs/CREDENTIALS.md).

## 6. Troubleshooting

| Symptom | Cause / fix |
|---|---|
| **"no Jira site configured"** | Run `jira auth login` or set `JIRA_SITE`/`JIRA_EMAIL`/`JIRA_TOKEN`. |
| **401 Unauthorized** | Bad/expired API token. Re-create at id.atlassian.com; check `JIRA_EMAIL` matches the token owner. |
| **403 Forbidden** | The user lacks permission for that action/project (or, for OAuth, the scope isn't granted). |
| **400 on search** | JQL must be **bounded** — add a restriction (project/assignee/date), not just `ORDER BY`. |
| **"no transition matching X"** | Run `jira issue transitions KEY` to see valid targets; match the status or transition name. |
| **Assignee/user not found** | Resolve via `jira user search "<email-or-name>"`; pass the accountId or `@me`. |
| **Description/comment looks wrong** | v3 needs ADF — the CLI converts text automatically; use `--markdown` for formatting. |
| **429 Too Many Requests** | The CLI retries with backoff automatically; if persistent, slow down or reduce `--all` sweeps. |
| **scoped token fails on site URL** | A scoped token must use the gateway; `jira auth login --type scoped` resolves the cloudId for you. |
| **Wrong site/profile** | Pass `--profile <name>` or `--site`; inspect `~/.config/jira-cli/config.yaml`. |

## 7. References

- Full command reference: [references/commands.md](references/commands.md)
- API references: [references/issues-api.md](references/issues-api.md),
  [references/search-jql-projects-users.md](references/search-jql-projects-users.md),
  [references/agile-adf-conventions.md](references/agile-adf-conventions.md)
- Authentication: [../docs/CREDENTIALS.md](../docs/CREDENTIALS.md)
- Agent-optimized guide: [../CLAUDE.md](../CLAUDE.md) · CLI README: [../README.md](../README.md)
- Config file: `~/.config/jira-cli/config.yaml` · API token: https://id.atlassian.com/manage-profile/security/api-tokens
