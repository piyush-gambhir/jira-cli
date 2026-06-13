# Jira Issues API — Reference

> **Source of truth:** Jira Cloud Platform REST API **v3** (OpenAPI 3.0.1 spec, `swagger-v3.v3.json`).
> All paths, methods, parameters, field names, and JSON examples are taken from that spec and the
> official Atlassian developer docs. Base URL: `https://your-domain.atlassian.net`.

## 0. Cross-cutting concerns (read before implementing any command)

### 0.1 `accountId` everywhere — usernames are gone (GDPR)
The old `name`/`username`/`userkey`/`key` user identifiers were **removed** from Jira Cloud. Use
**`accountId`** (e.g. `5b10ac8d82e05b22cc7d4ef5`) to identify users in: `assignee`/`reporter` on
create/edit, the Assign endpoint, add/remove watcher, `notify` recipients, and comment/worklog
`visibility` (for **groups** prefer the group `identifier` — group names are mutable). Some legacy
spec examples still show `"assignee":{"name":"bob"}` — **do not copy them**; use `accountId`.
Resolve a name/email → accountId via `GET /rest/api/3/user/search?query=<email-or-name>`.

### 0.2 ADF required for rich-text fields (v2 vs v3)
v2 and v3 expose the same operations; the difference is the body format of rich-text fields:

| Field | v2 (`/rest/api/2`) | v3 (`/rest/api/3`) |
|---|---|---|
| `description`, `environment`, comment/worklog body, textarea custom fields | plain text / wiki markup (string) | **Atlassian Document Format (ADF)** — a JSON document |

Minimal ADF doc:
```json
{ "type":"doc", "version":1, "content":[ {"type":"paragraph","content":[{"type":"text","text":"Hello"}]} ] }
```
A CLI taking `--description "plain string"` must wrap it in ADF before sending (or use v2 for wiki text).

### 0.3 `fields` vs `update` — two mutation styles
Create/Edit/Transition accept both in the same body:
- **`fields`** — set/overwrite semantics: `fieldId → value`. e.g. `"summary":"New title"`.
- **`update`** — operation semantics: `fieldId → [ {verb: value} ]` with verbs `set`/`add`/`remove`/`edit`/`copy`.
  Required for incremental list changes (add one label), adding a comment during a transition, etc.

**A field may appear in `fields` OR `update`, never both.**

### 0.4 `X-Atlassian-Token: no-check` for attachments
Uploading attachments **requires** header `X-Atlassian-Token: no-check` (else blocked) and the multipart
form-field name **must be `file`**. See §4.

### 0.5 Success codes
`200` body; `201` created (body); **`204` no body** — common for assign, transition, delete,
add/remove watcher, add/remove vote. Don't JSON-parse a 204.

---

## 1. Issues (core)

| Method & Path | Operation | Purpose |
|---|---|---|
| `POST /rest/api/3/issue` | createIssue | Create an issue or subtask |
| `POST /rest/api/3/issue/bulk` | createIssues | Bulk create (≤ 50) |
| `GET /rest/api/3/issue/{issueIdOrKey}` | getIssue | Get one issue |
| `PUT /rest/api/3/issue/{issueIdOrKey}` | editIssue | Edit fields |
| `DELETE /rest/api/3/issue/{issueIdOrKey}` | deleteIssue | Delete |
| `PUT /rest/api/3/issue/{issueIdOrKey}/assignee` | assignIssue | Assign |
| `GET /rest/api/3/issue/{issueIdOrKey}/transitions` | getTransitions | List transitions |
| `POST /rest/api/3/issue/{issueIdOrKey}/transitions` | doTransition | Perform a transition |
| `GET /rest/api/3/issue/{issueIdOrKey}/editmeta` | getEditIssueMeta | Editable fields |
| `POST /rest/api/3/issue/{issueIdOrKey}/notify` | notify | Email notification |
| `GET /rest/api/3/issue/{issueIdOrKey}/changelog` | getChangeLogs | Paginated history |
| `GET /rest/api/3/issue/createmeta/{projectIdOrKey}/issuetypes` | — | Creatable issue types |
| `GET /rest/api/3/issue/createmeta/{projectIdOrKey}/issuetypes/{issueTypeId}` | — | Creatable fields |

> The old single `GET /rest/api/3/issue/createmeta` is deprecated; use the two paginated forms above.

### 1.1 Create — `POST /rest/api/3/issue`
Required `fields`: `project` (`{id}`|`{key}`), `issuetype` (`{id}`|`{name}`), `summary`.
Query: `updateHistory` (bool, default false).

Field shapes: `description` = ADF doc · `assignee`/`reporter` = `{"id":"<accountId>"}` ·
`priority` = `{"id"}`|`{"name"}` · `labels` = `["a","b"]` · `components`/`fixVersions`/`versions` =
`[{"id"}]` · `duedate` = `"2019-05-11"` · `parent` = `{"key":"PROJ-123"}` (subtasks) ·
`timetracking` = `{"originalEstimate":"10","remainingEstimate":"5"}` · custom = `customfield_NNNNN`.

```json
{ "fields": {
    "project": {"key":"PROJ"}, "issuetype": {"name":"Task"}, "summary": "From CLI",
    "description": {"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"created via CLI"}]}]},
    "assignee": {"id":"5b109f2e9729b51b54dc274d"}, "priority": {"id":"20000"}, "labels": ["bugfix"]
} }
```
Response `201`: `{ "id", "key", "self", "transition": {...} }`.

### 1.2 Bulk create — `POST /rest/api/3/issue/bulk`
Body `{ "issueUpdates": [ <create-body>, ... ] }`, ≤ 50. Response `{ "issues":[...], "errors":[...] }` —
inspect **both** arrays (partial success).

### 1.3 Get — `GET /rest/api/3/issue/{key}`
Query: `fields` (csv; `*all`, `*navigable`, `-field`; **default = ALL fields**), `fieldsByKeys`,
`expand` (`renderedFields,names,schema,transitions,editmeta,changelog`), `properties`, `updateHistory`,
`failFast`. The response `key` is the resolved key (handles moved/renamed issues; no 302).

### 1.4 Edit — `PUT /rest/api/3/issue/{key}`
Uses `fields`/`update` (same rules). **Transitions are ignored here.** Query: `notifyUsers` (default true,
disabling needs admin), `overrideScreenSecurity`, `overrideEditableFlag`, `returnIssue` (true → `200`
with issue, else `204`), `expand`. `update.labels:[{"add":"x"}]` appends; `fields.labels:["x"]` replaces;
`update.components:[{"set":""}]` clears.

### 1.5 Delete — `DELETE /rest/api/3/issue/{key}`
Query `deleteSubtasks` (`true`/`false`, default false). `400` if it has subtasks and the flag is unset.
Response `204`.

### 1.6 Edit meta — `GET /rest/api/3/issue/{key}/editmeta`
Returns `{ "fields": { "<id>": { required, schema, name, key, operations:[...], allowedValues:[...],
hasDefaultValue, defaultValue } } }` — build a valid edit body from this.

### 1.7 Assign — `PUT /rest/api/3/issue/{key}/assignee`
Body `{ "accountId": "<id>" | "-1" (default assignee) | null (unassign) }`. Response `204`.

### 1.8 Transitions — `GET`/`POST /rest/api/3/issue/{key}/transitions`
Jira has **no "set status" endpoint**. (1) `GET` transitions to find the transition `id` whose
`to.name`/`to.id` is your target status; (2) `POST` that id. Transition ids are **per-issue** — always discover.

`GET` query: `expand=transitions.fields` (include screen fields), `transitionId`,
`includeUnavailableTransitions`, `skipRemoteOnlyCondition`, `sortByOpsBarAndStatus`.
`GET` response: `{ "transitions":[ {id,name,to:{id,name,statusCategory},hasScreen,isAvailable,fields:{...}} ] }`.

`POST` body (`transition` required, optional `fields`/`update` to fill the transition screen):
```json
{ "transition": {"id":"5"},
  "fields": {"resolution": {"name":"Fixed"}},
  "update": {"comment": [{"add": {"body": {"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Bug fixed"}]}]}}}]} }
```
Response `204`. CLI pattern for `jira issue transition PROJ-1 Done --comment "…"`: GET → match
`to.name`/transition `name` case-insensitively → POST with id; resolution via `fields.resolution`,
comment via `update.comment[].add.body` (ADF).

### 1.9 Notify — `POST /rest/api/3/issue/{key}/notify`
Body: `subject`, `textBody`, `htmlBody`, `to` (`{reporter,assignee,watchers,voters,users:[{accountId}],
groups:[{name}]}`), `restrict`. Response `204`.

### 1.10 Changelog — `GET /rest/api/3/issue/{key}/changelog`
Paginated (`startAt` 0, `maxResults` 100). Page bean of `{id,author,created,items:[{field,from,fromString,to,toString}]}`.

---

## 2. Issue comments

| Method & Path | Purpose |
|---|---|
| `GET /rest/api/3/issue/{key}/comment` | List (paginated) |
| `POST /rest/api/3/issue/{key}/comment` | Add |
| `GET /rest/api/3/issue/{key}/comment/{id}` | Get one |
| `PUT /rest/api/3/issue/{key}/comment/{id}` | Update |
| `DELETE /rest/api/3/issue/{key}/comment/{id}` | Delete |
| `POST /rest/api/3/comment/list` | Bulk get by IDs (≤ 1000) |

List query: `startAt` (0), `maxResults` (100), `orderBy` (`created`/`-created`), `expand=renderedBody`.
Body (`Comment`): `body` = **ADF doc** (required) · optional `visibility` (`{type:"role"|"group",
value,identifier}`) · optional `properties`. Response includes `id,self,author,updateAuthor,body,
renderedBody,created,updated,visibility`.

```json
{ "body": {"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"hi from CLI"}]}]} }
```

---

## 3. Issue worklogs

| Method & Path | Purpose |
|---|---|
| `GET /rest/api/3/issue/{key}/worklog` | List (maxResults up to 5000; `startedAfter`/`startedBefore` ms) |
| `POST /rest/api/3/issue/{key}/worklog` | Add |
| `GET/PUT/DELETE /rest/api/3/issue/{key}/worklog/{id}` | Get/update/delete |
| `POST /rest/api/3/worklog/list` | Bulk get by IDs (≤ 1000) |

`adjustEstimate` (add/update/delete): `auto` (default) · `leave` · `new` (+`newEstimate` e.g. `2d`) ·
`manual` (+`reduceBy` on add / `increaseBy` on delete). Body (`Worklog`): `timeSpent` (`"3h 30m"`) **or**
`timeSpentSeconds`; `started` (ISO, required on create); `comment` = ADF (optional); `visibility`.

```json
{ "timeSpent":"3h 30m", "started":"2026-06-13T09:00:00.000+0000",
  "comment": {"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"work"}]}]} }
```

---

## 4. Issue attachments

| Method & Path | Purpose |
|---|---|
| `POST /rest/api/3/issue/{key}/attachments` | Upload (multipart) |
| `GET /rest/api/3/attachment/{id}` | Metadata |
| `DELETE /rest/api/3/attachment/{id}` | Delete |
| `GET /rest/api/3/attachment/content/{id}` | Download (binary) |
| `GET /rest/api/3/attachment/thumbnail/{id}` | Thumbnail |
| `GET /rest/api/3/attachment/meta` | Global settings (enabled, uploadLimit) |

**Upload requirements:** header `X-Atlassian-Token: no-check` (else blocked) + multipart field name
exactly `file` (repeat for multiple). Response is an **array** of `{id,self,filename,author,created(ms),
size,mimeType,content(url),thumbnail(url)}`.
```bash
curl -X POST 'https://your-domain.atlassian.net/rest/api/3/issue/TEST-123/attachments' \
  -u 'email:<token>' -H 'X-Atlassian-Token: no-check' --form 'file=@"myfile.txt"'
```
Download: `/attachment/content/{id}` returns a **302** to a temp media URL with `redirect=true`
(default); pass `redirect=false` to get bytes directly. Check `/attachment/meta` (`enabled`,
`uploadLimit` bytes) first. **Go:** `w.CreateFormFile("file", name)` + set the `X-Atlassian-Token` header.

---

## 5. Issue links & link types

| Method & Path | Purpose |
|---|---|
| `POST /rest/api/3/issueLink` | Create a link |
| `GET /rest/api/3/issueLink/{linkId}` | Get a link |
| `DELETE /rest/api/3/issueLink/{linkId}` | Delete a link |
| `GET /rest/api/3/issueLinkType` | List link types |
| `POST/GET/PUT/DELETE /rest/api/3/issueLinkType[/{id}]` | Manage link types (admin) |

Create body (required `type`, `inwardIssue`, `outwardIssue`; optional `comment`):
```json
{ "type":{"name":"Blocks"}, "inwardIssue":{"key":"HSP-1"}, "outwardIssue":{"key":"MKY-1"} }
```
Response `201` **no body**. Link types give direction: `{id,name,inward,outward}` — e.g. Blocks →
outward "Blocks", inward "Blocked by". `DELETE` → `204`.

---

## 6. Issue watchers

| Method & Path | Purpose |
|---|---|
| `GET /rest/api/3/issue/{key}/watchers` | List |
| `POST /rest/api/3/issue/{key}/watchers` | Add |
| `DELETE /rest/api/3/issue/{key}/watchers` | Remove (`?accountId=`) |

**Add body is a bare JSON string** = accountId (`"5b10ac8d82e05b22cc7d4ef5"`); omit body → adds the
caller. **Remove** via `?accountId=` query param. Both → `204`. List → `{self,isWatching,watchCount,
watchers:[{accountId,displayName,active}]}`.

---

## 7. Issue votes

| Method & Path | Purpose |
|---|---|
| `GET /rest/api/3/issue/{key}/votes` | Count + voters |
| `POST /rest/api/3/issue/{key}/votes` | Add caller's vote (`204`) |
| `DELETE /rest/api/3/issue/{key}/votes` | Remove caller's vote (`204`) |

Add/remove always act on the **authenticated user** (no body, can't vote for others; can't vote on your
own reported issue). `GET` → `{self,votes,hasVoted,voters:[...]}` (voters need View Voters & Watchers).

---

## 8. Issue properties

Arbitrary JSON KV storage on an issue. `GET .../properties` (keys), `GET/PUT/DELETE
.../properties/{propertyKey}`. PUT body = raw JSON value (≤ 32768 chars) → `200`/`201`. GET →
`{key,value}`. DELETE → `204`.

---

## 9. Quick endpoint map

```
POST   /rest/api/3/issue                                   createIssue
POST   /rest/api/3/issue/bulk                              createIssues (≤50)
GET    /rest/api/3/issue/{key}            ?fields,fieldsByKeys,expand,properties,updateHistory,failFast
PUT    /rest/api/3/issue/{key}            ?notifyUsers,overrideScreenSecurity,returnIssue,expand
DELETE /rest/api/3/issue/{key}            ?deleteSubtasks
PUT    /rest/api/3/issue/{key}/assignee   body{accountId | "-1" | null}
GET    /rest/api/3/issue/{key}/transitions ?expand=transitions.fields
POST   /rest/api/3/issue/{key}/transitions body{transition.id, fields, update}
GET    /rest/api/3/issue/{key}/editmeta
GET    /rest/api/3/issue/createmeta/{projectIdOrKey}/issuetypes[/{issueTypeId}]
GET/POST/PUT/DELETE  /rest/api/3/issue/{key}/comment[/{id}]
GET/POST/PUT/DELETE  /rest/api/3/issue/{key}/worklog[/{id}]   ?adjustEstimate,newEstimate
POST   /rest/api/3/issue/{key}/attachments   (X-Atlassian-Token: no-check; form field "file")
GET    /rest/api/3/attachment/{id} | /content/{id} | /thumbnail/{id} | /meta
POST   /rest/api/3/issueLink   body{type,inwardIssue,outwardIssue,comment?} -> 201 no body
GET    /rest/api/3/issueLinkType
GET/POST/DELETE  /rest/api/3/issue/{key}/watchers   (add body="accountId"; remove ?accountId)
GET/POST/DELETE  /rest/api/3/issue/{key}/votes      (act on caller; no body)
GET/PUT/DELETE   /rest/api/3/issue/{key}/properties[/{key}]
```

## 10. Implementation gotchas

1. Resolve users → `accountId` before assign/watch/notify (`/user/search`).
2. Wrap rich text in ADF for v3 (description, environment, comment body, worklog comment, textarea CF).
3. `update` vs `fields`: never both for one field; use `update` for additive list ops + transition comments.
4. Transitions = 2-step discover-then-do with per-issue ids; match by `to.name`/transition `name`.
5. Attachments: `X-Atlassian-Token: no-check`, form field `file`, response is an array; check `/meta` first.
6. Add-watcher body is a bare JSON string; remove via `?accountId`.
7. 204s (assign/transition/delete/vote/watch/create-link) have no body — don't parse.
8. Bulk create partially succeeds — inspect both `issues` and `errors`.
9. `deleteSubtasks=true` to delete an issue with subtasks.
10. Get-issue defaults to ALL fields — pass `fields=` to keep payloads small.

### Sources
[Issues](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/) ·
[Comments](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-comments/) ·
[Worklogs](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-worklogs/) ·
[Attachments](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-attachments/) ·
[Links](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-links/) ·
[Watchers](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-watchers/) ·
[Votes](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-votes/) ·
OpenAPI v3 spec `swagger-v3.v3.json`.
