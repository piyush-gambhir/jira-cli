# jira-cli — Architecture & Build Plan

A first-party-style command-line interface for Jira, built to match the conventions of the
sibling CLIs in this folder (`jenkins-cli`, `cubeapm-cli`, `es-cli`, `grafana`, `nginxpm-cli`):
Go + [Cobra](https://github.com/spf13/cobra), profile-based config, `flags > env > config`
precedence, `-o table|json|yaml` output, read-only safety mode, background update check.

This is **not** Atlassian's official `acli` — it is our own thin, scriptable wrapper over the
Jira REST API, designed for humans and coding agents alike.

## Identity

| Item | Value |
|------|-------|
| Module | `github.com/piyush-gambhir/jira-cli` |
| Binary | `jira` |
| Config | `~/.config/jira-cli/config.yaml` (XDG-aware) |
| Token store | secrets live in the profile in `config.yaml` (0600); OAuth tokens cached alongside |
| Go | 1.22+ |
| Deps | `spf13/cobra`, `olekukonko/tablewriter`, `gopkg.in/yaml.v3`, `pkg/browser` (OAuth) |

## Authentication (all methods — see `docs/CREDENTIALS.md`)

The CLI supports every realistic Jira auth path. A profile carries an `auth_type`:

| `auth_type` | Deployment | Header | Inputs | Base URL |
|-------------|-----------|--------|--------|----------|
| `api_token` (default, Cloud) | Cloud | `Basic base64(email:token)` | site, email, API token | `https://<site>.atlassian.net` |
| `scoped_token` | Cloud | `Basic base64(email:token)` | site, email, scoped token | `https://api.atlassian.com/ex/jira/{cloudId}` |
| `oauth2` (3LO) | Cloud | `Bearer <access_token>` | client id/secret, scopes (browser loopback) | `https://api.atlassian.com/ex/jira/{cloudId}` |
| `pat` | Server/DC | `Bearer <PAT>` | host, PAT | `https://<host>` |
| `basic` | Server/DC | `Basic base64(user:pass)` | host, username, password | `https://<host>` |

`internal/auth` exposes an `Authenticator` interface (`Apply(*http.Request)` + `BaseURL()`),
with one implementation per method. OAuth handles the loopback-redirect browser flow,
`accessible-resources` cloudId resolution, and rotating-refresh-token persistence.
cloudId for scoped tokens is resolved via `/_edge/tenant_info`.

## Layout

```
main.go                     thin entrypoint; maps APIError -> exit code
cmd/                        one file per command (noun.go + noun_verb.go), Cobra
internal/config/            Config/Profile types, Load/Save, ResolveAuth (flags>env>config)
internal/auth/              Authenticator interface + api_token/scoped/oauth2/pat/basic
internal/client/            HTTP core, request, errors (Jira envelope), pagination, ratelimit,
                            + resource files: issues, search, transitions, comments, worklogs,
                              attachments, links, watchers, votes, projects, users, fields,
                              boards, sprints, epics, serverinfo
internal/adf/               Atlassian Document Format: FromPlainText / FromMarkdown -> doc JSON
internal/output/            table/json/yaml formatters, WriteError, TableDef
internal/version/           ldflags-injected version/commit/buildtime
internal/update/            background GitHub release check (24h cache)
jira/references/            verified API reference docs (issues, search/jql, agile/adf)
docs/CREDENTIALS.md         exhaustive auth guide (all methods)
CLAUDE.md / README.md / SKILL.md / CONTRIBUTING.md
```

## Command surface

```
jira auth login [--type api_token|scoped|oauth2|pat|basic]   # interactive, per method
jira auth list | use <profile> | logout [--profile p]
jira whoami                       GET /rest/api/3/myself
jira status                       serverInfo (+ myself auth check)

jira issue list [--project --status --assignee --mine --jql]   # convenience -> /search/jql
jira issue search --jql "..."     cursor pagination (nextPageToken)
jira issue get KEY [--fields --expand]
jira issue create --project --type --summary --description ... # ADF; createmeta-aware
jira issue edit KEY [--summary --description --add-label ...]
jira issue delete KEY [--delete-subtasks]
jira issue assign KEY --to <email|accountId|-1|@me|null>
jira issue transition KEY [<status>] [--comment --resolution]  # discover-then-do
jira issue transitions KEY                                     # list available
jira issue comment KEY --body "..."     |  jira issue comments KEY
jira issue worklog add KEY --time 3h [--comment --started]
jira issue attach KEY <file>...  |  jira issue attachments KEY  |  download
jira issue link KEY --type Blocks --to KEY2
jira issue watch/unwatch/watchers KEY
jira issue vote/unvote/votes KEY

jira project list|get      jira user search|get      jira field list
jira board list|get        jira sprint list|get|create|issues
jira epic get|issues
jira version | update
```

## Cross-cutting decisions

- **ADF**: all rich-text writes (description, comment, worklog) flow through `internal/adf`.
  A `--text`/`--v2` escape hatch can target `/rest/api/2` with a raw string when desired.
- **Pagination**: two models. `internal/client` provides `PageThrough` (offset: startAt/
  maxResults/isLast, advance by len) and `SearchAll` (cursor: nextPageToken). Search uses cursor;
  everything else uses offset.
- **Users**: usernames are gone (GDPR). Commands accept email / display-name / `@me` and resolve
  to `accountId` via `/user/search` before use.
- **Errors**: parse the `{errorMessages, errors{field:msg}}` envelope into a readable message.
- **Rate limits**: honor `Retry-After` on 429, else exponential backoff with jitter (cap ~30s, ~4 tries).
- **Read-only mode**: commands tagged `mutates=true` are blocked when a profile/flag sets read-only
  (agent safety), exactly like jenkins-cli.

## Build order

1. References + this plan (done first; they drive the code).
2. Scaffold: go.mod, Makefile, install.sh, main.go, version.
3. config + auth (all methods).
4. client (core + resources) + adf.
5. output.
6. cmd: root, auth/login, whoami, status, version, update.
7. cmd: issue (+ comment/worklog/attach/link/watch/vote) and search.
8. cmd: project, user, field, board, sprint, epic.
9. build + vet + gofmt + `--help` smoke test.
10. CLAUDE.md / README.md / SKILL.md / CONTRIBUTING.md / commands.md.
11. Live verification against a test Jira instance.
