# Jira Authentication — Reference

> Authoritative reference for every authentication method supported by the Jira REST API, across
> **Jira Cloud** and **Jira Server / Data Center**. Compiled from the live Atlassian developer and
> support documentation (verified June 2026). This drives the `jira-cli` `internal/auth` package.

## TL;DR — what `jira-cli` implements and defaults to

| Deployment | `auth_type` | Header sent | Base URL |
|---|---|---|---|
| **Jira Cloud** (default) | `api_token` | `Authorization: Basic base64(email:api_token)` | `https://<your-domain>.atlassian.net` |
| **Jira Cloud** (scoped token) | `scoped_token` | `Authorization: Basic base64(email:api_token)` | `https://api.atlassian.com/ex/jira/{cloudId}` |
| **Jira Cloud** (login w/ Atlassian account) | `oauth2` | `Authorization: Bearer <access_token>` | `https://api.atlassian.com/ex/jira/{cloudId}` |
| **Jira Server / Data Center** | `pat` | `Authorization: Bearer <PAT>` | `https://<host>` |
| **Jira Server / Data Center** (legacy) | `basic` | `Authorization: Basic base64(user:password)` | `https://<host>` |

- For a CLI, **API token basic auth (Cloud)** and **PAT (DC)** are the right defaults: no browser,
  no callback server, one secret to store.
- **OAuth 2.0 3LO** gives the "log in with your Atlassian account" experience (loopback + browser).
- Cookie/session, username+password on Cloud, and OAuth 1.0a are **deprecated or legacy**.

## Method comparison matrix

| # | Method | Deployment | Header | Status | CLI-friendly? |
|---|---|---|---|---|---|
| 1 | Basic auth + API token (unscoped) | Cloud | `Basic base64(email:token)` | Supported (recommended for scripts) | Yes |
| 1b | Basic auth + API token (scoped) | Cloud | `Basic base64(email:token)` | Supported (newer, least-privilege) | Yes (needs cloudId) |
| 2 | OAuth 2.0 (3LO) auth-code | Cloud | `Bearer <access_token>` | Supported (recommended for apps) | Yes (loopback flow) |
| 3 | Personal Access Token (PAT) | Server / DC | `Bearer <PAT>` | Supported (recommended for DC) | Yes |
| 4 | Basic auth + username/password | Server / DC | `Basic base64(user:pass)` | DC: supported · Cloud: removed | Yes (DC only) |
| 5 | Cookie / session auth | Both (legacy) | Cookie `JSESSIONID=…` | Deprecated | No (avoid) |
| 6 | OAuth 1.0a | Server / DC | Signed `OAuth …` (RSA-SHA1) | Legacy | Painful |
| 7 | Connect JWT / Forge | Cloud (apps only) | `JWT <token>` / managed | App-only | No |
| 8 | Anonymous / unauthenticated | Both | none | Public data only | N/A |

---

## 1. Basic auth with API token (Jira Cloud) — DEFAULT

**Applies to:** Jira Cloud only. Account-password basic auth is **deprecated on Cloud** — use an
API token in place of the password.

| Field | Value |
|---|---|
| Header | `Authorization: Basic <base64(email:api_token)>` |
| Base64 input | `email@example.com:api_token` (UTF-8, no trailing newline) |
| Base URL | `https://<your-domain>.atlassian.net` then `/rest/api/3/...` |

**Create the token** at `https://id.atlassian.com/manage-profile/security/api-tokens`
("Create API token" → classic/unscoped; "Create API token with scopes" → §1b).

**Token expiry:** since 15 Dec 2024 new tokens require an expiry (1–365 days, default 1 year).
Tokens created before that date expire between 14 Mar and 12 May 2026 — so a long-lived CLI must
handle token-expired (401) and prompt the user to mint a new one. Shown once; individually revocable.

```bash
curl -u 'fred@example.com:freds_api_token' \
  -H 'Accept: application/json' \
  https://your-domain.atlassian.net/rest/api/3/myself
```

**Go:** `req.SetBasicAuth(email, apiToken)` does exactly `base64(email:token)` — correct and cleanest.
Never log the token; store it 0600.

## 1b. Basic auth with **scoped** API token (Jira Cloud)

Same `Basic base64(email:token)` mechanism, but the token carries explicit scopes (least-privilege).
**Critical difference:** a scoped token will **not** work against `your-domain.atlassian.net` — it must
target the gateway `https://api.atlassian.com/ex/jira/{cloudId}` (the same base URL OAuth uses),
still with **Basic** auth (OAuth not required).

**Resolve cloudId (no OAuth):**
```bash
curl https://your-domain.atlassian.net/_edge/tenant_info   # → {"cloudId":"<uuid>", ...}
```

```bash
curl --url 'https://api.atlassian.com/ex/jira/<cloud_id>/rest/api/3/myself' \
  --user '<email>:<scoped_api_token>' -H 'Accept: application/json'
```

---

## 2. OAuth 2.0 (3LO) — authorization code grant (Jira Cloud)

**Applies to:** Jira Cloud. Use when users should log in with their Atlassian account and consent to
scopes rather than pasting a token.

| Phase | Header |
|---|---|
| API calls | `Authorization: Bearer <access_token>` |
| Token endpoints | `Content-Type: application/json` (credentials in JSON body) |

**Base URL for API calls:** `https://api.atlassian.com/ex/jira/{cloudId}/rest/api/3/...`

### One-time app registration (developer console)
developer.atlassian.com → Developer console → create app → **Authorization** → configure
**OAuth 2.0 (3LO)** → set Callback URL (for a CLI, a loopback like `http://localhost:PORT/callback`)
→ get **Client ID** / **Client Secret** → add Jira scopes under **Permissions**.

### Step 1 — Authorization request (browser)
```
https://auth.atlassian.com/authorize?
  audience=api.atlassian.com&
  client_id=YOUR_CLIENT_ID&
  scope=read%3Ajira-work%20write%3Ajira-work%20offline_access&
  redirect_uri=http://localhost:PORT/callback&
  state=RANDOM_CSRF_VALUE&
  response_type=code&
  prompt=consent
```
| Param | Required | Notes |
|---|---|---|
| `audience` | yes | `api.atlassian.com` |
| `client_id` | yes | from Settings |
| `scope` | yes | space-separated; add `offline_access` to get a refresh token |
| `redirect_uri` | yes | must exactly match a registered Callback URL |
| `state` | yes | opaque, unguessable; verify on return (CSRF) |
| `response_type` | yes | `code` |
| `prompt` | yes | `consent` |

After consent → redirect to `redirect_uri?code=AUTH_CODE&state=...`.

### Step 2 — Exchange code for tokens
```bash
curl -X POST 'https://auth.atlassian.com/oauth/token' \
  -H 'Content-Type: application/json' \
  -d '{ "grant_type":"authorization_code", "client_id":"...", "client_secret":"...",
        "code":"AUTH_CODE", "redirect_uri":"http://localhost:PORT/callback" }'
```
```json
{ "access_token":"...", "expires_in":3600, "scope":"...", "refresh_token":"...", "token_type":"Bearer" }
```
- Access token lifetime: **60 minutes** (not configurable).
- `refresh_token` present only if `offline_access` was requested.

### Step 3 — Resolve cloudId (accessible-resources)
```bash
curl https://api.atlassian.com/oauth/token/accessible-resources \
  -H 'Authorization: Bearer ACCESS_TOKEN' -H 'Accept: application/json'
```
```json
[ { "id":"<cloudId>", "name":"Site", "url":"https://your-domain.atlassian.net", "scopes":[...] } ]
```
Use `id` as `{cloudId}`. If multiple sites → prompt the user to pick.

### Step 4 — Call the API
`GET https://api.atlassian.com/ex/jira/<cloudId>/rest/api/3/myself` with the bearer token.

### Step 5 — Refresh (rotating refresh tokens)
```bash
curl -X POST 'https://auth.atlassian.com/oauth/token' \
  -H 'Content-Type: application/json' \
  -d '{ "grant_type":"refresh_token", "client_id":"...", "client_secret":"...", "refresh_token":"..." }'
```
**Rules (critical for a long-lived CLI):**
- Refresh tokens **rotate** — each use returns a NEW refresh token; persist it, discard the old.
- **Inactivity expiry: 90 days**, reset on each rotation. Idle > 90 days → re-authorize.
- **Reuse interval ~10 min**: replaying the same refresh token inside the window is tolerated; outside
  it fails with `invalid_grant`. Serialize refreshes (mutex / file lock) to avoid burning the token.
- Refreshing does not invalidate already-issued access tokens (they live out their 60 min).

### Scopes
Prefer **classic** scopes; use granular only when needed; keep an app under ~50 scopes.
Classic: `read:jira-user`, `read:jira-work`, `write:jira-work`, `manage:jira-project`,
`manage:jira-configuration`, `manage:jira-webhook`, `manage:jira-data-provider`.
`offline_access` → required to receive a refresh token.

**Choosing scopes in `jira-cli`.** `jira auth login --type oauth2` lets you pick how much access to
grant. If you don't pass a scope flag (and you're interactive), it shows a picker; otherwise use:

| Preset | Scopes requested |
|---|---|
| `read`  | `read:jira-user read:jira-work` |
| `write` (default) | `+ write:jira-work` |
| `admin` | `+ manage:jira-project manage:jira-configuration` (manage projects & config) |
| `all`   | `+ manage:jira-webhook manage:jira-data-provider` (everything) |

```bash
jira auth login --type oauth2 --client-id … --client-secret … --scope-preset admin
jira auth login --type oauth2 … --scope-preset write --scope manage:jira-webhook   # preset + extra
jira auth login --type oauth2 … --scopes "read:jira-work write:jira-work offline_access"  # raw override
```
`offline_access` is always added automatically. **Important:** the OAuth request can only ask for
scopes your app actually has — add the matching scopes under **Permissions** in the developer console,
or authorization will fail. Broader scopes (e.g. `admin`) enable `jira project create/update/delete`.

**Agile API + OAuth caveat.** The `board` / `sprint` / `epic` commands call the Jira **Software (Agile)**
API (`/rest/agile/1.0`), which over OAuth requires **granular `jira-software` scopes** (e.g.
`read:board-scope:jira-software`, `read:sprint:jira-software`). The classic scopes above return
`401 "scope does not match"` for those endpoints. They work out of the box with **API-token / PAT**
auth (which use your full account permissions, not scopes). To use the Agile commands over OAuth, add
the granular `jira-software` scopes under your app's **Permissions** and re-login.

### Distributing OAuth builds safely

Release binaries must not embed OAuth client secrets: users can extract any value shipped in a public
binary. End users either paste an API token (easiest) or bring their own OAuth app via
`--client-id/--client-secret`. A zero-configuration browser-login experience requires a server-side
authorization broker that retains the confidential client secret and applies appropriate abuse,
rate-limit, and redirect controls.

### CLI loopback flow (what `jira auth login --type oauth2` does)
1. Start `http.Server` on `127.0.0.1:PORT` with a `/callback` handler.
2. Generate random `state`; open the authorize URL in the browser.
3. Callback validates `state`, captures `code`, shows "you can close this tab".
4. Exchange code → tokens; call `accessible-resources`; store `{access,refresh,expires_at,cloudId}`.
5. Before each API call, if `now >= expires_at - skew`, refresh and atomically persist the rotated token.
6. **No device-code flow exists for Jira** — on headless boxes fall back to API token (§1).

---

## 3. Personal Access Tokens (PAT) — Jira Server / Data Center — DEFAULT for DC

**Applies to:** Jira Server / Data Center (Jira 8.14+, JSM 4.15+). Token-management UI is a DC feature.

| Field | Value |
|---|---|
| Header | `Authorization: Bearer <PAT>` (literal `Bearer`, space, token — **not** base64, no email) |
| Base URL | `https://<host>/rest/api/2/...` (DC uses v2; `/latest` also works) |

**Create:** avatar → Profile → **Personal access tokens** → Create token. Optional expiry (days),
up to 10 per user, shown once, inherits the creating user's permissions.

```bash
curl -H "Authorization: Bearer <yourToken>" https://jira.example.com/rest/api/2/myself
```

**Go:** `req.Header.Set("Authorization", "Bearer "+pat)`.

---

## 4. Basic auth with username + password — Server / Data Center

**Applies to:** Server / DC. **Removed on Cloud.** Prefer PAT (§3); use this only for pre-8.14 instances.

| Field | Value |
|---|---|
| Header | `Authorization: Basic <base64(username:password)>` |
| Base URL | `https://<host>/rest/api/2/...` |

**CAPTCHA lockout:** after several failed logins Jira triggers CAPTCHA and REST basic auth stops working
(response header `X-Seraph-LoginReason: AUTHENTICATION_DENIED`). Surface this clearly — **do not auto-retry**
(retries make it worse). Send the `Authorization` header preemptively (Go's `SetBasicAuth` does this).

---

## 5. Cookie-based / session auth (legacy — avoid)

**Deprecated.** Documented for old Server instances only.
```bash
curl -X POST -H 'Content-Type: application/json' \
  --data '{"username":"u","password":"p"}' \
  http://jira.example.com/rest/auth/1/session     # → Set-Cookie: JSESSIONID=...
```
Then send `Cookie: JSESSIONID=...`. `DELETE /rest/auth/1/session` to log out. Interacts badly with
XSRF checks; subject to CAPTCHA lockout. Use API token (Cloud) / PAT (DC) instead.

## 6. OAuth 1.0a (Server / DC — legacy)

Server/DC only, predates PATs. Uses an **Application Link** with an RSA key pair; **RSA-SHA1** signed
requests; long-lived (~5y) access token. Three-leg dance:
`POST /plugins/servlet/oauth/request-token` → `GET /plugins/servlet/oauth/authorize` →
`POST /plugins/servlet/oauth/access-token`. Each call carries `Authorization: OAuth ...`. Requires an
admin to register your public key first. Prefer PAT unless app-link semantics are specifically required.

## 7. App-only (not for a CLI) + anonymous

- **Connect JWT** (`Authorization: JWT <token>`, signed with the app's shared secret + `qsh` claim) and
  **Forge** (managed `asUser()`/`asApp()`) are app frameworks, not user CLI flows — skip.
- **Anonymous:** send no `Authorization` header; reads whatever is public. Most useful endpoints require auth.

---

## Rate limiting by auth method

| Aspect | Detail |
|---|---|
| Status | `429 Too Many Requests` |
| By auth | **API-token** traffic uses legacy **burst** limits (not the points quota). **OAuth/Connect/Forge** use the **points-based + burst** system. |
| Headers | `Retry-After` (seconds), `X-RateLimit-Limit/Remaining/Reset`, `X-RateLimit-NearLimit` (`true` < 20% left), `RateLimit-Reason`. |
| DC | Rate limiting configured per instance by the admin; also returns `429` + `Retry-After`. |

**Client behavior:** on 429 honor `Retry-After`, else exponential backoff with jitter (cap ~30s, ~4 tries).
Never tight-loop (also avoids CAPTCHA on DC). Treat `X-RateLimit-NearLimit: true` as a slow-down signal.

## Practical design summary

1. Detect deployment: `*.atlassian.net` → Cloud; else Server/DC.
2. Cloud default → API token basic auth (§1). Scoped token → resolve cloudId via `/_edge/tenant_info`
   and use the `api.atlassian.com/ex/jira/{cloudId}` base URL (§1b).
3. Cloud "login" UX → OAuth 2.0 3LO (§2), loopback + browser; persist rotated refresh token carefully.
4. DC default → PAT (§3). Fall back to username/password (§4) only for old instances (watch CAPTCHA).
5. Avoid cookie/session (§5) and OAuth 1.0a (§6) unless integrating with legacy Server.
6. Secrets: 0600 file (or OS keychain); never log tokens. Central handling for 401/403/429.

### Sources
- Security overview — https://developer.atlassian.com/cloud/jira/platform/security-overview/
- Basic auth — https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/
- OAuth 2.0 (3LO) — https://developer.atlassian.com/cloud/jira/platform/oauth-2-3lo-apps/
- Scopes — https://developer.atlassian.com/cloud/jira/platform/scopes-for-oauth-2-3LO-and-forge-apps/
- Cookie auth — https://developer.atlassian.com/cloud/jira/platform/jira-rest-api-cookie-based-authentication/
- Rate limiting — https://developer.atlassian.com/cloud/jira/platform/rate-limiting/
- Manage API tokens — https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/
- Find your Cloud ID — https://support.atlassian.com/jira/kb/retrieve-my-atlassian-sites-cloud-id/
- PAT (Server/DC) — https://confluence.atlassian.com/enterprise/using-personal-access-tokens-1026032365.html
- OAuth 1.0a (Server) — https://developer.atlassian.com/server/jira/platform/oauth/
