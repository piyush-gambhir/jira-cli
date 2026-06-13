# Jira Search, JQL, Projects, Users & Metadata — Reference

> Jira Cloud Platform REST API **v3** — Issue Search (new JQL endpoints), JQL syntax, Projects, Users,
> Fields & Metadata, and static enums. Base URL: `https://<your-site>.atlassian.net`. Verified against
> the current Atlassian OpenAPI v3 spec (openapi 3.0.1) and official support docs, June 2026.

## 0. TL;DR for CLI builders
- **Search changed.** The classic `GET/POST /rest/api/3/search` is **deprecated and being removed**
  (announced 2024-10-31; deprecated 2025-05-01; traffic blocked by 2025-10-31). Use
  **`GET`/`POST /rest/api/3/search/jql`**. It paginates with an opaque **`nextPageToken`** (no
  `startAt`/`total`) and by default returns **only issue IDs** — pass `fields` to get more.
- **Count separately:** `total` is gone; use **`POST /rest/api/3/search/approximate-count`**.
- **No usernames** — users are `accountId` (GDPR).
- **Resolve names → IDs for JQL.** Status/priority/etc. names are human-facing; list them via the enum endpoints (§6).
- **Creating issues:** discover required fields via `GET /rest/api/3/issue/createmeta/{projectIdOrKey}/issuetypes[/{issueTypeId}]`.

---

## 1. Issue Search (JQL) — the NEW endpoints

### 1.1 `GET /rest/api/3/search/jql`
| Param | Default | Notes |
|---|---|---|
| `jql` | – | **Bounded** JQL (must contain a restriction; `ORDER BY` alone → 400). ORDER BY ≤ 7 fields. |
| `nextPageToken` | – | Cursor; omit for first page. Expires after 7 days. |
| `maxResults` | 50 | Max **5000** (only when requesting id/key only). |
| `fields` | **`id`** | csv/repeatable: `*all`, `*navigable`, names, `-field`. |
| `expand` | – | `renderedFields,names,schema,transitions,changelog,...` |
| `properties` | – | up to 5 keys |
| `fieldsByKeys` | false | reference fields by key |
| `reconcileIssues` | – | up to 50 issue IDs to force strong consistency |

```bash
curl -G 'https://your-site.atlassian.net/rest/api/3/search/jql' -u 'email:token' \
  --data-urlencode 'jql=project = ABC AND statusCategory != Done ORDER BY updated DESC' \
  --data-urlencode 'fields=summary,status,assignee,updated' --data-urlencode 'maxResults=100'
```

### 1.2 `POST /rest/api/3/search/jql`
Same semantics; use when JQL is large. Body: `jql`, `nextPageToken`, `maxResults`, `fields` (array,
default `["id"]`), `expand` (comma-string), `properties`, `fieldsByKeys`, `reconcileIssues`.

### 1.3 `POST /rest/api/3/search/approximate-count`
Body `{ "jql": "<bounded JQL>" }` → `{ "count": <int> }`.

### 1.4 Response shape (`SearchAndReconcileResults`)
`{ issues:[IssueBean], nextPageToken, isLast, names?, schema?, warnings? }` — **no `total`/`startAt`/
`maxResults`**. Drive pagination off `nextPageToken`/`isLast`; counts via approximate-count.

### 1.5 The `fields` parameter
`id` (default for search) · `*all` · `*navigable` · `summary,status` · `-description` (exclude) ·
`customfield_10010`. Multiple `fields` params merge.

### 1.6 CLI pagination loop (cursor)
```
token := ""
for {
  body := {jql, fields:[...], maxResults:100}; if token != "" { body.nextPageToken = token }
  resp := POST /search/jql
  emit(resp.issues)
  if resp.isLast || resp.nextPageToken == "" { break }
  token = resp.nextPageToken
}
```

---

## 2. JQL syntax reference
A query is `field operator value/function` clauses joined by keywords, optionally `ORDER BY` last.

### 2.1 Keywords
`AND` · `OR` · `NOT` · `EMPTY`/`NULL` (with `IS`/`=`) · `ORDER BY ... ASC|DESC`.
Precedence: `NOT` > `AND` > `OR`; use parentheses. ORDER BY ≤ 7 fields in `/search/jql`.

### 2.2 Operators
`=` `!=` `>` `>=` `<` `<=` · `IN` / `NOT IN` · `~` (contains, text) / `!~` · `IS` / `IS NOT` (EMPTY/NULL) ·
`WAS` / `WAS IN` / `WAS NOT` (history) · `CHANGED`. History predicates: `AFTER`,`BEFORE`,`ON`,
`DURING(s,e)`,`BY`,`FROM`,`TO`.

### 2.3 Functions
**Dates:** `now()`, `currentLogin()`, `lastLogin()`, `startOfDay(["±n(y|M|w|d|h|m)"])`, `endOfDay`,
`startOfWeek/endOfWeek`, `startOfMonth/endOfMonth`, `startOfYear/endOfYear`. e.g. `created > startOfDay("-3d")`.
**Users/groups:** `currentUser()` (=/≠), `membersOf("group")` / `membersOf("id:<id>")` (IN/NOT IN).
**Sprints:** `openSprints()`, `closedSprints()`, `futureSprints()`.
**Versions:** `latestReleasedVersion(proj)`, `earliestUnreleasedVersion(proj)`, `releasedVersions([proj])`, `unreleasedVersions([proj])`.
**Other:** `linkedIssues(key[,type])`, `votedIssues()`, `watchedIssues()`, `updatedBy(user[,from[,to]])`, `parentEpic(key)`.

### 2.4 Key fields (with types)
`assignee`/`reporter`/`creator`/`watcher`/`voter` (USER) · `project` (key/name/id) · `issuetype`/`type` ·
`status` (name or id; `WAS*` on name) · `statusCategory` (`To Do`/`In Progress`/`Done`) · `priority` ·
`resolution` (`Unresolved` / `IS EMPTY`; absent in team-managed → use `statusCategory`) · `labels` ·
`component` · `fixVersion`/`affectedVersion` · `sprint` · `parent` (Epic Link retired Feb 2024 → use `parent`) ·
`created`/`updated`/`due`/`resolved`/`lastViewed` (DATE) · `summary`/`description`/`comment`/`environment`/`text` (TEXT, `~`) ·
`issuekey`/`key`/`id` · custom fields by `"Field Name"` or `cf[ID]`.

**Quoting:** quote values with spaces/reserved chars; quote relative dates (`"-7d"`). Date formats
`yyyy/MM/dd[ HH:mm]`, `yyyy-MM-dd[ HH:mm]`, or relative `w/d/h/m`.

### 2.5 Ready-to-use examples
```text
project = ABC AND statusCategory != Done AND assignee = currentUser() ORDER BY updated DESC
project = ABC AND issuetype = Bug AND created >= startOfWeek() AND resolution IS EMPTY ORDER BY priority DESC
statusCategory = "In Progress" AND updated <= "-2w" ORDER BY updated ASC
project = ABC AND sprint IN openSprints() AND priority IN (Highest, High) AND statusCategory != Done
project = ABC AND status WAS "Blocked"
```

---

## 3. Projects

| Method & Path | Purpose | Key params |
|---|---|---|
| `GET /rest/api/3/project/search` | Paginated list/search (preferred) | `startAt`,`maxResults`(≤100),`orderBy`,`query`,`id[]`,`keys[]`,`typeKey`,`categoryId`,`action`,`status[]`,`expand` |
| `GET /rest/api/3/project` | All projects — **DEPRECATED** → use `/search` | `expand`,`recent`,`properties[]` |
| `GET /rest/api/3/project/{projectIdOrKey}` | Get one | `expand`(`description,issueTypes,lead,projectKeys,...`),`properties[]` |
| `POST /rest/api/3/project` | Create | body `CreateProjectDetails` |
| `PUT /rest/api/3/project/{idOrKey}` | Update | body `UpdateProjectDetails` |
| `DELETE /rest/api/3/project/{idOrKey}` | Delete | `enableUndo` |
| `GET /rest/api/3/project/{idOrKey}/statuses` | Issue types + their statuses | – |

`Project`: `id,key,name,projectTypeKey,simplified,style,lead,components[],issueTypes[],versions[],archived,isPrivate,roles{},avatarUrls{},self`.
**`PageBean*`** responses use `startAt/maxResults/total/isLast` (+`nextPage` URL) — only issue-search moved to cursors.

---

## 4. Users (accountId-based)

> Identify users by **`accountId`**. `query` matches displayName + emailAddress. User-directory
> endpoints need the **Browse users and groups** global permission.

| Method & Path | Purpose | Key params |
|---|---|---|
| `GET /rest/api/3/myself` | Current user | `expand=groups,applicationRoles` |
| `GET /rest/api/3/user` | One user by accountId | `accountId`,`expand` |
| `GET /rest/api/3/user/search` | Find users | `query` or `accountId` (one req), `startAt`,`maxResults` |
| `GET /rest/api/3/user/assignable/search` | Assignable to issue/project | `query`/`accountId`, **`project` or `issueKey`** (req), `startAt`,`maxResults` |
| `GET /rest/api/3/user/bulk` | Bulk by accountId | `accountId` (repeatable), `maxResults`(10) |
| `GET /rest/api/3/users/search` | All users (admin) | `startAt`,`maxResults`(≤1000) |

`User`: `accountId,accountType,active,displayName,emailAddress(nullable),timeZone,locale,avatarUrls{},self`.
**CLI tip:** accept email/display-name fragment, hit `/user/search`, resolve to `accountId` before use.

---

## 5. Fields & metadata

### 5.1 Fields
`GET /rest/api/3/field` (all) · `GET /rest/api/3/field/search` (paginated custom-field search:
`type[],id[],query,orderBy,projectIds[]`) · `POST /rest/api/3/field` (create custom field).
`FieldDetails`: `id` (e.g. `customfield_10010`), `key,name,custom,navigable,orderable,searchable,
clauseNames[]` (JQL clause names), `schema{type,...},scope`.

### 5.2 Create-issue metadata (field discovery)
| Method & Path | Purpose |
|---|---|
| `GET /rest/api/3/issue/createmeta/{projectIdOrKey}/issuetypes` | Creatable issue types for a project |
| `GET /rest/api/3/issue/createmeta/{projectIdOrKey}/issuetypes/{issueTypeId}` | Create-screen fields |
| `GET /rest/api/3/issue/createmeta` | Old all-in-one — **DEPRECATED** |

`FieldMetadata`: `required,name,key,schema{type,items,system|custom,customId},operations[],
hasDefaultValue,defaultValue,allowedValues[],autoCompleteUrl`.
**`issue create` workflow:** pick project + issue type → call issuetype-field createmeta → prompt for each
`required:true` field (offer `allowedValues`) → build `fields{}` (scalars/options as `{id}`/`{name}`,
users as `{accountId}`, labels as `string[]`, rich text as ADF) → `POST /rest/api/3/issue`.

---

## 6. Static enums

> In JQL use the **name** (quote if spaced) or numeric **ID**. List valid values to validate user input.

### 6.1 Issue types
`GET /rest/api/3/issuetype` (all) · `GET /rest/api/3/issuetype/project?projectId=&level=` ·
`GET /rest/api/3/issuetype/{id}`. `IssueTypeDetails`: `id,name,description,subtask,hierarchyLevel,iconUrl,scope`.

### 6.2 Priorities
`GET /rest/api/3/priority` (deprecated) · `/priority/search` · `/priority/{id}`.
`Priority`: `id,name,description,statusColor,iconUrl,isDefault`.

### 6.3 Statuses
`GET /rest/api/3/status` (all) · `/status/{idOrName}` · `GET /rest/api/3/statuses/search`
(`projectId,searchString,statusCategory(TODO|IN_PROGRESS|DONE),startAt,maxResults(200)`) ·
`GET /rest/api/3/project/{idOrKey}/statuses` (grouped by issue type).
`StatusDetails`: `id,name,description,iconUrl,statusCategory{id,key(new|indeterminate|done),name,colorName},scope`.

### 6.4 Resolutions
`GET /rest/api/3/resolution` (deprecated) · `/resolution/search` · `/resolution/{id}`.
`Resolution`: `id,name,description`. Unresolved = `resolution = Unresolved` / `resolution IS EMPTY`.

---

## 7. CLI cheat-sheet
- **Two pagination models.** Issue search → cursor (`nextPageToken`, no total). Everything `PageBean*`/
  `PageOf*` → offset (`startAt/maxResults/total/isLast`). Don't mix.
- **`fields` default is `id`** on `/search/jql` — always pass what you want.
- **Bounded JQL required** — add a real restriction or you get 400.
- **Counts:** `POST /search/approximate-count`.
- **Prefer IDs** for versions/components/statuses/priorities/resolutions/issue types in JQL & bodies.
- **Users are accountId-only** — resolve via `/user/search`.
- **ADF for rich text** in v3.
- **Spec source of truth:** `https://developer.atlassian.com/cloud/jira/platform/swagger-v3.v3.json`.

### Sources
[Issue Search](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/) ·
[JQL fields](https://support.atlassian.com/jira-software-cloud/docs/jql-fields/) ·
[JQL operators](https://support.atlassian.com/jira-software-cloud/docs/jql-operators/) ·
[JQL functions](https://support.atlassian.com/jira-software-cloud/docs/jql-functions/) ·
[Projects](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-projects/) ·
[Users](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-users/) · OpenAPI v3 spec.
