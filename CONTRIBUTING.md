# Contributing

## Layout

```
main.go                 entrypoint (maps APIError -> exit code)
cmd/                    one file per command group (Cobra); noun.go + behavior files
internal/auth/          Authenticator interface + api_token/scoped/oauth2/pat/basic
internal/config/        Profile/Config types, Load/Save, ResolveAuth (flags>env>config)
internal/client/        HTTP core (errors, retry/backoff) + resource methods
internal/adf/           Atlassian Document Format build + extract
internal/output/        table/json/yaml formatters
internal/{version,update}/
jira/references/        verified API reference docs
docs/CREDENTIALS.md     authentication reference
```

## Build & test

The Go CLI lives in `cli-go/`; run the toolchain from there:

```bash
cd cli-go
make build      # -> bin/jira
make test       # go test -race ./...
make vet
make fmt        # gofmt -s -w .
make install
```

Keep `go vet ./...` and `gofmt -l .` clean before committing.

## Conventions

- **Commands**: each command is a `newXxxCmd() *cobra.Command`. Group parents add their children.
  Tag write operations with `Annotations: mutates` so `--read-only` blocks them.
- **Shorthands**: `-o -s -e -t -u -k -v -q` are reserved by global persistent flags. Don't redefine
  them on subcommands (cobra will panic). Free letters used: `-p -d -a -l -n -b -j`.
- **Output**: data to stdout via `render(...)`; informational lines to stderr via `info(...)`.
  Every list/get must support `-o json|yaml` (pass a `TableDef` for table mode).
- **Users**: accept email/name/`@me`/`id:<accountId>` and resolve via `client.ResolveUser`.
- **Rich text**: build with `internal/adf` (`FromPlainText` / `FromMarkdown`).
- **New endpoints**: add a typed method in `internal/client/`, then the command in `cmd/`.

## Documentation and agent materials

When you change commands, flags, auth, config, or output, update **`CLAUDE.md`**, **`README.md`**,
**`jira/SKILL.md`**, and **`jira/references/commands.md`** in the same change. If scope/auth/major
workflows change, update the `jira/SKILL.md` front-matter `description` too.

## Commits

Conventional, imperative commit subjects (e.g. `feat(issue): add bulk create`).

**Signed commits are mandatory.** The repository enforces a GitHub ruleset that rejects any unsigned
push, so every commit (and tag) must carry a valid signature. Configure signing once:

```bash
# SSH signing (simplest; no GPG needed). Use a key you've added to GitHub as a *signing* key.
git config --global gpg.format ssh
git config --global user.signingkey ~/.ssh/<your_key>.pub
git config --global commit.gpgsign true
git config --global tag.gpgsign true
```

Then commit normally — git signs automatically. Don't use `--no-gpg-sign`; the push will be rejected.
Tag releases with `git tag -s vX.Y.Z -m "vX.Y.Z"`.
