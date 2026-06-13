# Jira CLI — Agent Guide

A command-line interface for Jira (Cloud and Server/Data Center) over the REST API.
Built for humans and coding agents. This file is the canonical reference.

## Quick Reference

- **Binary:** `jira`
- **Config file:** `~/.config/jira-cli/config.yaml` (XDG-aware; mode 0600)
- **Env vars:** `JIRA_SITE`, `JIRA_EMAIL`, `JIRA_TOKEN`, `JIRA_USER`, `JIRA_PASSWORD`,
  `JIRA_AUTH_TYPE`, `JIRA_CLOUD_ID`, `JIRA_API_VERSION`, `JIRA_INSECURE`, `JIRA_READ_ONLY`,
  `JIRA_OAUTH_CLIENT_ID`, `JIRA_OAUTH_CLIENT_SECRET`, `JIRA_NO_INPUT`, `JIRA_QUIET`
- **Config priority:** CLI flags > environment variables > config profile
- **Machine output:** add `-o json` (or `-o yaml`) to any list/get command
- **Agent safety:** `--read-only` (or `read_only` in the profile) blocks all write commands;
  `--no-input` disables prompts

## Authentication

Five methods, all via `jira auth login --type <method>` (see `docs/CREDENTIALS.md` for the
full reference). The auth type is stored per profile.

| `--type` | Deployment | Credential | Notes |
|---|---|---|---|
| `oauth2` (default) | Cloud | OAuth 2.0 (3LO) | Browser loopback flow; refresh tokens auto-rotate; app baked into released binaries |
| `api_token` | Cloud | email + API token | Token from id.atlassian.com; no app needed |
| `scoped` | Cloud | email + scoped token | Resolves cloudId; routes via api.atlassian.com gateway |
| `pat` | Server/DC | bearer PAT | Jira 8.14+ |
| `basic` | Server/DC | username + password | Legacy; watch CAPTCHA lockout |

```bash
# Browser OAuth — the default (released binaries have the app baked in)
jira auth login
# API token instead (no app needed) — prompts for site/email/token
jira auth login --type api_token
# Non-interactive (CI/agents) — uses API token
jira auth login --type api_token --site https://acme.atlassian.net --email me@acme.com --token "$JIRA_TOKEN"
# Server/DC with a PAT
jira auth login --type pat --site https://jira.company.com --token "$PAT"
# OAuth 2.0 3LO with your own app (bring-your-own-app)
jira auth login --type oauth2 --client-id "$ID" --client-secret "$SECRET"

jira auth list            # show profiles (current marked *)
jira auth use staging     # switch active profile
jira auth logout --name staging
```

For OAuth, choose how much access to grant with `--scope-preset read|write|admin|all` (or an
interactive picker if omitted), add individual scopes with repeatable `--scope`, or override raw with
`--scopes`. `admin`/`all` enable project & configuration management. The app's Permissions in the
developer console must include whatever you request. Details: `docs/CREDENTIALS.md`.

For agents/CI, the simplest path is environment variables (no `auth login` needed):

```bash
export JIRA_SITE=https://acme.atlassian.net JIRA_EMAIL=me@acme.com JIRA_TOKEN=xxxx
jira whoami -o json
```

## Output Formats

`-o table` (default, human), `-o json` (parse with jq), `-o yaml`. **Agents: always use `-o json`.**
Informational lines (e.g. "Created ABC-1", the JQL used by `issue list`) go to **stderr**, so
stdout stays clean for parsing.

## Common Workflows

### Identity & connectivity
```bash
jira whoami -o json          # current user (also prints accountId)
jira status -o json          # site reachability + auth check + serverInfo
```

### Find issues (JQL)
```bash
# Convenience filters (builds JQL for you)
jira issue list --mine
jira issue list -p ABC --status "In Progress" -o json
jira issue list -p ABC --type Bug --label regression

# Raw JQL (Cloud uses /search/jql cursor paging; --all to fetch everything)
jira issue search "project = ABC AND statusCategory != Done ORDER BY updated DESC" -o json
jira issue search --jql "assignee = currentUser() AND resolution IS EMPTY" --all
jira issue search "project = ABC" --count        # approximate count only
```

### Inspect / create / edit
```bash
jira issue get ABC-123 -o json
jira issue get ABC-123 --fields summary,status,assignee --expand changelog

jira issue create -p ABC --type Task --summary "Title" -d "Plain description"
jira issue create -p ABC --type Bug --summary "X" -d "**bold** details" --markdown -a me@acme.com -l urgent
jira issue create -p ABC --type Task --summary "X" --field customfield_10010=5

jira issue edit ABC-123 --summary "New title" --add-label triaged --remove-label blocker
jira issue assign ABC-123 --to me@acme.com      # or @me, id:<accountId>, default, none
jira issue delete ABC-123 --yes
```

### Transition (change status)
```bash
jira issue transitions ABC-123                  # list available transitions
jira issue transition ABC-123 "In Progress"     # match by target status or transition name
jira issue transition ABC-123 Done --resolution Fixed --comment "Shipped"
```

### Comments, worklogs, attachments, links, watch, vote
```bash
jira issue comment ABC-123 --body "Looks good" [--markdown]
jira issue comments ABC-123 -o json
jira issue worklog ABC-123 --time "2h 30m" --comment "review" --started "2026-06-13T09:00:00.000+0000"
jira issue attach ABC-123 ./error.log ./screenshot.png
jira issue attachments ABC-123
jira issue download <attachmentId> --out ./file
jira issue link ABC-1 ABC-2 --type Blocks       # "ABC-1 Blocks ABC-2"
jira issue link-types
jira issue watch ABC-123 ; jira issue unwatch ABC-123 ; jira issue watchers ABC-123
jira issue vote  ABC-123 ; jira issue unvote  ABC-123 ; jira issue votes   ABC-123
```

### Projects, users, fields
```bash
jira project list --query mobile -o json
jira project get ABC
# Manage projects (needs project-admin permission; OAuth needs the admin/all scope preset)
jira project create --key MOB --name "Mobile" --lead @me --template <projectTemplateKey>
jira project update ABC --name "New name" --lead me@acme.com
jira project delete ABC --enable-undo --yes
jira user search "jane" -o json      # resolve a person to an accountId
jira user get <accountId>
jira field list --custom             # field ids + JQL clause names
```

### Agile (boards, sprints, epics)
```bash
jira board list -p ABC -o json
jira board issues 42 ; jira board backlog 42
jira sprint list --board 42 --state active
jira sprint get 100 ; jira sprint issues 100
jira sprint create --board 42 --name "Sprint 5" --goal "Ship X"
jira epic get ABC-10 ; jira epic issues ABC-10
```

## Notes & gotchas (Jira-specific)

- **Users are accountIds (GDPR).** Commands accept email / display name / `@me` / `id:<accountId>`
  and resolve to an accountId via user search before use.
- **Rich text is ADF.** Descriptions/comments/worklog comments are sent as Atlassian Document Format.
  `--markdown` interprets the text as lightweight markdown (headings, lists, bold/italic/code/links).
- **Transitions are discover-then-do.** There's no "set status"; the CLI lists transitions and matches
  your argument against the transition name or its target status.
- **Search pagination differs by deployment.** Cloud uses the cursor-based `/search/jql`; Server/DC
  uses the classic `/search`. The CLI handles both. JQL must be bounded.
- **Rate limits (429)** are retried automatically (honor `Retry-After`, else exponential backoff).
- **API version** defaults to v3 on Cloud, v2 on Server/DC; override with `--api-version`.
- **Agile commands over OAuth:** `board`/`sprint`/`epic` use the Jira Software (Agile) API, which over
  OAuth needs granular `jira-software` scopes — classic scopes give `401 "scope does not match"`. They
  work with API-token/PAT auth. See `docs/CREDENTIALS.md`.

## Discovering commands

Every command has `-h/--help`:
```bash
jira --help
jira issue --help
jira issue create --help
```

Full written reference: `jira/references/commands.md`. API references:
`jira/references/{issues-api,search-jql-projects-users,agile-adf-conventions}.md`.
Auth reference: `docs/CREDENTIALS.md`.

## Maintainer note

When you change the CLI (commands, flags, auth, output), update **this file**, `README.md`,
`jira/SKILL.md`, and `jira/references/commands.md` in the same change. See `CONTRIBUTING.md`.
