---
name: jira-cli
description: "Expert guide for the jira CLI — authenticate to Jira Cloud or Server/Data Center (API token, scoped token, OAuth 2.0 3LO, PAT, or username/password) and manage issues, JQL search, transitions, comments, worklogs, attachments, links, watchers, votes, projects, users, fields, boards, sprints, and epics from the terminal. Use this skill when the user mentions Jira from the command line, jira-cli, JQL, issue transitions, creating/assigning tickets, or automating Jira with a CLI or coding agent. Trigger for agents that need exact commands, flags, auth env vars, or JSON output for scripts."
---

# Jira CLI (`jira`) — agent skill

## Maintainer note

**When you change the CLI** (commands, flags, auth, config, output): update **`CLAUDE.md`** and the
affected **`README.md` / `docs/` / `jira/references/`** in the same change. If the change affects
what this skill claims (scope, install, auth methods, major workflows), update the YAML
**`description`** above or the sections below. See **`CONTRIBUTING.md`**.

## Canonical guide

The full command reference, workflows, and examples live in **`CLAUDE.md`**. Read that file for
complete coverage; the full per-flag reference is in **`jira/references/commands.md`** and the
authentication deep-dive in **`docs/CREDENTIALS.md`**.

## Quick orientation

| Item | Value |
|------|--------|
| Binary | `jira` |
| Config | `~/.config/jira-cli/config.yaml` |
| Machine-readable output | `-o json` (prefer for agents) |
| Non-interactive | set env vars (`JIRA_SITE`, `JIRA_EMAIL`, `JIRA_TOKEN`, …) and/or `--no-input` |
| Safety | `--read-only` blocks all write commands |

## Auth in one line (agents/CI)

```bash
export JIRA_SITE=https://acme.atlassian.net JIRA_EMAIL=me@acme.com JIRA_TOKEN=xxxx
jira whoami -o json
```

Other methods: `jira auth login --type {api_token|scoped|oauth2|pat|basic}`.

## Most-used commands

```bash
jira issue search "project = ABC AND statusCategory != Done" -o json
jira issue get ABC-123 -o json
jira issue create -p ABC --type Task --summary "Title" -d "Body" [--markdown]
jira issue assign ABC-123 --to @me
jira issue transitions ABC-123                # discover, then:
jira issue transition ABC-123 "In Progress"
jira issue comment ABC-123 --body "..."
jira project list -o json
jira user search "jane" -o json               # email/name -> accountId
```

## Discovering commands

Cobra adds `-h`/`--help` on the root and every subcommand:

```bash
jira --help
jira issue --help
jira issue create --help
```

## Install (reference)

```bash
go install github.com/piyush-gambhir/jira-cli@latest
# or: clone repo && make install — see README.md
```

## Gotchas for agents

- Users are **accountIds** — pass email/name/`@me` and the CLI resolves them.
- Rich text is **ADF**; use `--markdown` for formatting in `-d/--body/--comment`.
- Transitions are **discover-then-do** (match by status or transition name).
- JQL must be **bounded**; counts via `jira issue search "<jql>" --count`.
- stdout is data, stderr is informational — safe to pipe `-o json` into `jq`.
