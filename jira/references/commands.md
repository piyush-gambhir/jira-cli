# Jira CLI — Full Command Reference

Complete reference for every command, subcommand, flag, and argument in the `jira` CLI.

## Global Flags

Available on every command:

| Flag | Short | Description |
|---|---|---|
| `--output <format>` | `-o` | Output format: `table` (default), `json`, `yaml` |
| `--profile <name>` | | Configuration profile to use |
| `--site <url>` | `-s` | Jira site URL override |
| `--email <email>` | `-e` | Atlassian account email override (Cloud) |
| `--token <token>` | `-t` | API token / PAT override |
| `--user <name>` | `-u` | Username override (Server/DC basic auth) |
| `--api-version <v>` | | Platform REST version override: `3` (Cloud) or `2` (Server/DC) |
| `--insecure` | `-k` | Skip TLS certificate verification |
| `--no-color` | | Disable color output |
| `--verbose` | `-v` | Verbose HTTP logging to stderr |
| `--read-only` | | Block write operations (agent safety) |
| `--no-input` | | Disable interactive prompts (CI/agents) |
| `--quiet` | `-q` | Suppress informational output |

Configuration precedence: **flags > environment variables > config profile**. Env vars: `JIRA_SITE`,
`JIRA_EMAIL`, `JIRA_TOKEN`, `JIRA_USER`, `JIRA_PASSWORD`, `JIRA_AUTH_TYPE`, `JIRA_CLOUD_ID`,
`JIRA_API_VERSION`, `JIRA_INSECURE`, `JIRA_READ_ONLY`, `JIRA_OAUTH_CLIENT_ID`,
`JIRA_OAUTH_CLIENT_SECRET`, `JIRA_NO_INPUT`, `JIRA_QUIET`.

---

## auth

Manage authentication profiles. See `docs/CREDENTIALS.md` for the full auth reference.

### `jira auth login`
Authenticate and save a profile.

| Flag | Description |
|---|---|
| `--type` | `api_token` (default), `scoped`, `oauth2`, `pat`, `basic` |
| `--name` | Profile name to save under (default `default`) |
| `--password` | Password (basic auth; prompted if omitted) |
| `--client-id`, `--client-secret` | OAuth 2.0 app credentials |
| `--scope-preset` | OAuth scope preset: `read`, `write` (default), `admin`, `all` (prompted if omitted) |
| `--scope` | Add an individual OAuth scope on top of the preset (repeatable) |
| `--scopes` | Raw space-separated scope override (wins over preset/`--scope`) |
| `--port` | OAuth loopback callback port (default 8765) |

OAuth scope presets: `read` (read user/work) · `write` (+ write work) · `admin` (+ manage projects &
configuration) · `all` (+ webhooks & data provider). `offline_access` is always added. The app's
Permissions in the developer console must include the requested scopes.

Site/email/token/user come from the global `--site/--email/--token/--user` flags or are prompted.

```bash
jira auth login
jira auth login --type pat --site https://jira.company.com --token "$PAT"
jira auth login --type scoped --site https://acme.atlassian.net --email me@acme.com --token "$TOK"
jira auth login --type oauth2 --client-id ABC --client-secret XYZ
```

### `jira auth list` (alias `ls`)
List configured profiles (current marked `*`). Columns: CURRENT, NAME, AUTH TYPE, SITE, ACCOUNT.

### `jira auth use <profile>`
Set the active profile.

### `jira auth logout [--name <profile>]`
Remove a profile (and its stored credentials). Defaults to the current profile.

---

## whoami
Show the authenticated user (`GET /myself`). Columns: ACCOUNT ID, NAME, EMAIL, ACTIVE, TIMEZONE.

## status
Show site connectivity + auth status (`serverInfo` + `myself`). Columns: SITE, VERSION, DEPLOYMENT,
AUTHENTICATED, USER. Exits non-zero if not authenticated.

## version
Print the CLI version.

## update
Check GitHub for a newer release and print upgrade instructions.

---

## issue (aliases `issues`, `i`)

### `jira issue list`
List issues using simple filters (builds a JQL query).

| Flag | Short | Description |
|---|---|---|
| `--project` | `-p` | Filter by project key |
| `--status` | | Filter by status name |
| `--assignee` | | Filter by assignee (email, name, `@me`, `id:<accountId>`) |
| `--mine` | | Only issues assigned to me |
| `--type` | | Filter by issue type |
| `--label` | | Filter by label |
| `--limit` | `-n` | Max issues (default 50; ignored with `--all`) |
| `--all` | | Fetch all matching issues (paginate) |
| `--fields` | | Comma-separated fields to request |

### `jira issue search [jql]`
Run a raw JQL query (positional or `--jql`). Cloud uses `/search/jql` (cursor); Server/DC uses classic search.

| Flag | Short | Description |
|---|---|---|
| `--jql` | `-j` | JQL query (or pass positionally) |
| `--limit` | `-n` | Max issues (default 50) |
| `--all` | | Fetch all matching issues |
| `--fields` | | Comma-separated fields |
| `--count` | | Return only the approximate match count |

### `jira issue get <key>`
Get an issue. `--fields` (CSV, default all), `--expand` (e.g. `renderedFields,changelog`).
Table mode prints a detail view; `-o json/yaml` returns the full raw issue.

### `jira issue create`  *(write)*
| Flag | Short | Description |
|---|---|---|
| `--project` | `-p` | Project key (required) |
| `--type` | | Issue type name (required) |
| `--summary` | | Summary (required) |
| `--description` | `-d` | Description text |
| `--markdown` | | Interpret `--description` as lightweight markdown |
| `--priority` | | Priority name |
| `--assignee` | `-a` | Assignee (email, name, `@me`, `id:<accountId>`) |
| `--label` | `-l` | Label (repeatable) |
| `--parent` | | Parent issue key (subtasks) |
| `--field` | | Extra field `name=value` (repeatable) |

### `jira issue edit <key>`  *(write)*
`--summary`, `--description`/`-d`, `--markdown`, `--priority`, `--assignee`/`-a`,
`--add-label` (repeatable), `--remove-label` (repeatable), `--no-notify`.

### `jira issue delete <key>`  *(write)*
`--delete-subtasks`, `--yes` (skip confirmation).

### `jira issue assign <key> --to <user>`  *(write)*
`--to` accepts email, name, `@me`, `id:<accountId>`, `default`, or `none`/`unassign`.

### `jira issue transitions <key>`
List available transitions (ID, NAME, → STATUS).

### `jira issue transition <key> <status-or-transition>` (alias `move`)  *(write)*
Match by target status name or transition name. `--resolution`, `--comment`, `--markdown`.

### `jira issue comment <key> --body <text>`  *(write)*
`--body`/`-b` (required), `--markdown`.

### `jira issue comments <key>`
List comments. `--limit`/`-n`.

### `jira issue worklog <key> --time <dur>`  *(write)*
`--time` (e.g. `"3h 30m"`, required), `--comment`, `--markdown`, `--started` (ISO 8601),
`--adjust` (`auto|leave|new|manual`), `--new-estimate`.

### `jira issue worklogs <key>`
List worklog entries. `--limit`/`-n`.

### `jira issue attach <key> <file>...`  *(write)*
Upload one or more files.

### `jira issue attachments <key>`
List attachment metadata.

### `jira issue download <attachmentId> [--out <path>]`
Download an attachment (defaults to its filename).

### `jira issue link <fromKey> <toKey> --type <Type>`  *(write)*
Create `<fromKey> <type-outward> <toKey>`. `--comment`, `--markdown`.

### `jira issue link-types`
List available link types (ID, NAME, OUTWARD, INWARD).

### `jira issue watch <key>` / `unwatch <key>` / `watchers <key>`  *(watch/unwatch are writes)*
`watch`/`unwatch` take `--user` (defaults to you).

### `jira issue vote <key>` / `unvote <key>` / `votes <key>`  *(vote/unvote are writes)*

---

## project (alias `projects`)

- `jira project list [--query <q>] [--limit/-n N]` — columns KEY, NAME, TYPE, LEAD.
- `jira project get <idOrKey>`
- `jira project create --key <K> --name <N> [--type software|service_desk|business] [--lead <user>] [--template <key>] [-d <desc>] [--field k=v]`  *(write; needs project-admin / manage:jira-project)*
- `jira project update <idOrKey> [--name] [-d <desc>] [--lead <user>] [--field k=v]`  *(write)*
- `jira project delete <idOrKey> [--enable-undo] [--yes]`  *(write)*

## user (alias `users`)

- `jira user search <query> [--limit/-n N]` — resolve a person to an accountId.
- `jira user get <accountId>`

## field (alias `fields`)

- `jira field list [--custom]` — id + name + JQL clause names.

---

## board (alias `boards`) — Agile

- `jira board list [--project/-p <key>] [--type scrum|kanban|simple] [--limit/-n N]`
- `jira board get <boardId>`
- `jira board issues <boardId> [--limit/-n N]`
- `jira board backlog <boardId> [--limit/-n N]`

## sprint (alias `sprints`) — Agile

- `jira sprint list --board/-b <id> [--state future,active,closed] [--limit/-n N]`
- `jira sprint get <sprintId>`
- `jira sprint create --board/-b <id> --name <name> [--start <iso>] [--end <iso>] [--goal <text>]`  *(write)*
- `jira sprint issues <sprintId> [--limit/-n N]`

## epic (alias `epics`) — Agile

- `jira epic get <epicKey>`
- `jira epic issues <epicKey> [--limit/-n N]`

---

## Exit codes

`0` success; `1` on error. With `-o json`, errors are emitted as `{"error": "...", "status_code": N}`
on stderr; otherwise as `Error: <message>`.
