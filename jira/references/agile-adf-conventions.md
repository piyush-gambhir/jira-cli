# Jira Agile API, ADF & API Conventions — Reference

> Compiled from the official Atlassian Cloud developer docs (verified June 2026). Two API surfaces:
> - **Agile (Jira Software) API** — base path **`/rest/agile/1.0`** — boards, sprints, backlog, epics.
> - **Platform (Jira) API** — base path **`/rest/api/3`** (ADF) or **`/rest/api/2`** (plain strings).
>
> Both under `https://your-domain.atlassian.net`. **Boards/sprints are NOT on the platform path.**

## 1. Base URLs & headers
| Concern | Detail |
|---|---|
| Agile base | `.../rest/agile/1.0/...` |
| Platform base | `.../rest/api/3/...` (ADF) or `/rest/api/2/...` (plain/wiki escape hatch) |
| Headers | `Accept: application/json`; writes `Content-Type: application/json`; file/state ops may need `X-Atlassian-Token: no-check` |
| Users | `accountId` only (no usernames) |

---

## 2. Agile API — `/rest/agile/1.0`

### 2.1 Boards
| Method | Path | Purpose |
|---|---|---|
| GET | `/board` | List boards — `startAt,maxResults,type(scrum|kanban|simple),name,projectKeyOrId` |
| POST | `/board` | Create — `name,type,filterId,location` |
| GET | `/board/{boardId}` | Get board |
| DELETE | `/board/{boardId}` | Delete board |
| GET | `/board/{boardId}/backlog` | Backlog issues — `startAt,maxResults,jql,fields,expand` |
| GET | `/board/{boardId}/configuration` | Columns, estimation, ranking field |
| GET | `/board/{boardId}/epic` | Epics — `startAt,maxResults,done` |
| GET | `/board/{boardId}/issue` | All board issues — `startAt,maxResults,jql,fields,expand` |
| GET | `/board/{boardId}/sprint` | **Sprints for board** — `startAt,maxResults,state(future,active,closed)` |
| GET | `/board/{boardId}/sprint/{sprintId}/issue` | Sprint issues scoped to board |
| GET | `/board/{boardId}/project` | Projects mapped to board |
| GET | `/board/{boardId}/version` | Versions — `released` |

Board types: `scrum` (sprints + backlog), `kanban` (continuous; no sprints), `simple`. Calling sprint
endpoints on a kanban board returns none. List responses carry both `total` and `isLast`:
```json
{ "maxResults":2, "startAt":1, "total":5, "isLast":false,
  "values":[ {"id":84,"name":"scrum board","type":"scrum"}, {"id":92,"name":"kanban board","type":"kanban"} ] }
```

### 2.2 Sprints
| Method | Path | Purpose |
|---|---|---|
| POST | `/sprint` | Create — `name`(req), `originBoardId`(req), `startDate`, `endDate`, `goal` |
| GET | `/sprint/{sprintId}` | Get sprint |
| PUT | `/sprint/{sprintId}` | **Full** update (omitted → null) |
| POST | `/sprint/{sprintId}` | **Partial** update (only supplied fields) |
| DELETE | `/sprint/{sprintId}` | Delete |
| GET | `/sprint/{sprintId}/issue` | Issues in sprint — `startAt,maxResults,jql,fields,expand` |
| POST | `/sprint/{sprintId}/issue` | **Move issues into sprint** — `issues`(≤50), `rankBeforeIssue`, `rankCustomFieldId` → 204 |

State machine `future → active → closed`. Start: partial-`POST` with `state:"active"` + valid dates.
Complete: `state:"closed"`. Closed sprint: only `name`/`goal` editable.
```json
// POST /sprint
{ "name":"sprint 1", "originBoardId":5, "startDate":"2015-04-11T15:22:00.000+10:00",
  "endDate":"2015-04-20T01:22:00.000+10:00", "goal":"sprint 1 goal" }
// response 201 adds: "id":37, "state":"future"; closed sprints also have "completeDate"
```

### 2.3 Backlog
`POST /backlog/issue` (move to backlog, removes from sprints) · `POST /backlog/{boardId}/issue`
(+ranking). Body `{ "issues":[...] }` ≤ 50 → `204`.

### 2.4 Epics
`GET /epic/{epicIdOrKey}` · `POST /epic/{epicIdOrKey}` (partial: `name,summary,color,done`) ·
`GET /epic/{epicIdOrKey}/issue` · `POST /epic/{epicIdOrKey}/issue` (move issues to epic) ·
`GET/POST /epic/none/issue` (issues w/o epic; POST removes from epic) · `PUT /epic/{id}/rank`.
```json
{ "id":23, "key":"EX-1", "name":"Epic Name", "summary":"Epic Summary", "color":{"key":"color_4"}, "done":false }
```

---

## 3. Atlassian Document Format (ADF)

In **v3**, all rich-text fields (issue **description**, **comment body**, **environment**, textarea
custom fields, worklog **comment**) MUST be an ADF JSON document — NOT plain strings.

### 3.1 Document structure
```json
{ "version": 1, "type": "doc", "content": [] }
```
- `version`: `1`. `type`: `"doc"`. `content`: array of top-level **block nodes**.
- Block nodes have a `content` array (`paragraph`, `heading`, `bulletList`, `codeBlock`, …).
- Inline nodes live inside blocks (`text`, `hardBreak`).
- Child block nodes only valid inside a parent (`listItem`).
- Top-level allowed: `paragraph,heading,blockquote,bulletList,orderedList,codeBlock,panel,rule,table,...`.

### 3.2 Node reference (CLI subset)
**paragraph** — `{type:"paragraph", content:[inline...]}`
**text** — `{type:"text", text:"...", marks?:[...]}`
**hardBreak** — `{type:"hardBreak"}` (a `<br/>` between text nodes in a paragraph)
**heading** — `{type:"heading", attrs:{level:1-6}, content:[inline]}`
**bulletList** — `{type:"bulletList", content:[listItem...]}`
**orderedList** — `{type:"orderedList", attrs?:{order:N}, content:[listItem...]}`
**listItem** — `{type:"listItem", content:[block... (usually a paragraph)]}`
**codeBlock** — `{type:"codeBlock", attrs?:{language:"go"}, content:[{type:"text", text:"...\n..."}]}` (text has NO marks)
**blockquote** — `{type:"blockquote", content:[paragraph|list|codeBlock...]}`

### 3.3 Marks (on text nodes)
| Mark | Meaning | attrs |
|---|---|---|
| `strong` | bold | — |
| `em` | italic | — |
| `code` | inline mono | — (can't combine with other marks) |
| `strike` | strikethrough | — |
| `underline` | underline | — |
| `link` | hyperlink | `href` (req), `title?` |

```json
{"type":"text","text":"Atlassian","marks":[{"type":"link","attrs":{"href":"http://atlassian.com"}}]}
```

### 3.4 Worked examples
```json
// (a) single paragraph
{"version":1,"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Just one line."}]}]}
// (b) bulleted list
{"version":1,"type":"doc","content":[{"type":"bulletList","content":[
  {"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"one"}]}]},
  {"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"two"}]}]}]}]}
// (c) code block
{"version":1,"type":"doc","content":[{"type":"codeBlock","attrs":{"language":"go"},
  "content":[{"type":"text","text":"func main() {\n\tprintln(\"hi\")\n}"}]}]}
```

### 3.5 `internal/adf` Go helper spec
```go
type Node struct { Type string; Content []Node; Text string; Marks []Mark; Attrs map[string]any }
type Mark struct { Type string; Attrs map[string]any }
type Doc  struct { Version int; Type string; Content []Node } // Version=1, Type="doc"
func FromPlainText(s string) Doc   // \n\n -> paragraph; \n -> hardBreak
func FromMarkdown(s string) Doc    // lightweight md -> ADF
```
**`FromPlainText`:** split on `\n\n+` → one `paragraph` each; within a block split on `\n` joined by
`hardBreak`; empty input → one empty paragraph; wrap in `doc`. **Never emit a bare string for a v3 field.**
**`FromMarkdown` (lightweight):** `#…######` → heading; `- `/`* `/`+ ` → bulletList; `1. ` → orderedList;
fenced ```` ```lang ```` → codeBlock; `> ` → blockquote; inline `` `code` ``/`**bold**`/`*italic*`/
`[label](url)` → marks. **v2 escape hatch:** `/rest/api/2` accepts plain/wiki strings — offer a
`--text`/`--v2` flag for quick one-liners; default to v3 + ADF.

---

## 4. Pagination (two models)

### 4.1 Offset — `startAt`/`maxResults`/`total`/`isLast`
Used by all `/rest/agile/1.0` endpoints and most platform list endpoints.
```json
{ "startAt":0, "maxResults":50, "total":137, "isLast":false, "values":[...] }
```
Loop: advance `startAt` by **`len(values)`** (not requested maxResults — server may clamp); stop on
`isLast` (primary), empty page (safety), or `startAt >= total` (only if `total` present). Some endpoints
name the array `issues` instead of `values`.

### 4.2 Cursor — `nextPageToken`/`isLast`
Used by enhanced JQL search `/rest/api/3/search/jql`.
```json
{ "issues":[...], "nextPageToken":"EgQIBR...", "isLast":false }
```
No `total`. Loop on `nextPageToken` until absent or `isLast`. `maxResults` is a hint.
Abstract both behind one `Paginator`.

---

## 5. Error responses
```json
{ "errorMessages":[ "Issue does not exist or you do not have permission to see it." ],
  "errors":{ "summary":"You must specify a summary of the issue." }, "status":400 }
```
- `errorMessages` — general messages. `errors` — field → message map. Either may be empty.
- 400 validation (don't retry; print each) · 401 bad/expired creds · 403 permission/XSRF (add
  `X-Atlassian-Token: no-check`) · 404 not found / no permission · 429 rate-limited (retry) · 5xx (retry transient).
- **Surfacing:** parse the envelope; if it isn't JSON (HTML gateway page), print status + raw body.

---

## 6. Rate limiting
Several limits run simultaneously: hourly points/cost budget, per-second burst, per-issue write limits.
Headers (read on every response): `Retry-After` (seconds — honor first), `X-RateLimit-Limit/Remaining/Reset`,
`X-RateLimit-NearLimit` (`true` → slow down), `RateLimit-Reason`.
**Behavior:** on 429 sleep `Retry-After` else exponential backoff with jitter (×0.7–1.3, cap ~30s, ~4 tries);
retry transient 5xx similarly; preemptively slow when near-limit; batch writes (≤50). DC limits are admin-configured.
```
delay := 2s; for attempt 0..3 { resp := do(req); if status != 429 && status < 500 { return }
  wait := delay; if Retry-After present { wait = seconds(Retry-After) }; sleep(wait * rand(0.7,1.3)); delay = min(delay*2, 30s) }
```

---

## 7. `serverInfo` & `myself`
- `GET /rest/api/3/serverInfo` — **works unauthenticated**; `{baseUrl,version,versionNumbers,
  deploymentType(Cloud|Server|DataCenter),buildNumber,buildDate,serverTime,serverTitle,serverTimeZone}`.
  Drives `status` (validate base URL + show server info).
- `GET /rest/api/3/myself` — **requires auth** (a 200 confirms creds); `{self,accountId,accountType,
  emailAddress(may be absent),displayName,active,timeZone,locale,avatarUrls,groups?,applicationRoles?}`.
  Drives `whoami` and the auth-check in `status`.

---

## 8. CLI cheat-sheet
- Share one HTTP layer (auth header, retry/backoff §6, error decode §5); expose two base paths:
  `agile = host + /rest/agile/1.0`, `platform = host + /rest/api/3`.
- Boards/sprints → agile path. Issues/search/serverInfo/myself → platform path.
- Rich text → `internal/adf`. Offer `--v2` for raw strings.
- Offset paginator for agile/most-v3; cursor paginator for `/search/jql`.
- Bulk moves capped at 50 — chunk larger inputs.
- `status` = serverInfo (+ myself auth check); `whoami` = myself.

### Sources
[Agile intro](https://developer.atlassian.com/cloud/jira/software/rest/intro/) ·
[Board](https://developer.atlassian.com/cloud/jira/software/rest/api-group-board/) ·
[Sprint](https://developer.atlassian.com/cloud/jira/software/rest/api-group-sprint/) ·
[ADF structure](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/) ·
[v3 intro](https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/) ·
[Rate limiting](https://developer.atlassian.com/cloud/jira/platform/rate-limiting/).
