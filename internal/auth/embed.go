package auth

// EmbeddedClientID and EmbeddedClientSecret let a distributor bake a built-in
// OAuth 2.0 (3LO) app into the binary, so end users can run
// `jira auth login --type oauth2` and sign in to THEIR OWN Jira account in the
// browser without registering their own app.
//
// These are empty in source and in a plain `go build`. Inject them at build time:
//
//	JIRA_OAUTH_CLIENT_ID=... JIRA_OAUTH_CLIENT_SECRET=... make build
//
// which sets them via -ldflags -X. Keeping them out of source means the secret
// lives only in the released artifact, not the repository. Note the tradeoffs:
// all users share this app's rate-limit quota and app identity, and rotating the
// secret invalidates older binaries. See docs/CREDENTIALS.md.
var (
	EmbeddedClientID     string
	EmbeddedClientSecret string
)

// HasEmbeddedOAuthApp reports whether this build ships a built-in OAuth app.
func HasEmbeddedOAuthApp() bool {
	return EmbeddedClientID != "" && EmbeddedClientSecret != ""
}
