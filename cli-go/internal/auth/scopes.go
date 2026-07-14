package auth

import (
	"fmt"
	"strings"
)

// Classic Jira OAuth 2.0 (3LO) scopes. See
// https://developer.atlassian.com/cloud/jira/platform/scopes-for-oauth-2-3LO-and-forge-apps/
const (
	ScopeReadUser           = "read:jira-user"            // read user info
	ScopeReadWork           = "read:jira-work"            // read issues, projects, worklogs
	ScopeWriteWork          = "write:jira-work"           // create/edit/transition issues, comments
	ScopeManageProject      = "manage:jira-project"       // project administration
	ScopeManageConfig       = "manage:jira-configuration" // global configuration
	ScopeManageWebhook      = "manage:jira-webhook"       // manage webhooks
	ScopeManageDataProvider = "manage:jira-data-provider" // dev-info / data provider
	ScopeOffline            = "offline_access"            // required for a refresh token
)

// ScopePresets maps friendly preset names to the scopes they request.
// offline_access is always added by ResolveScopes (so refresh works).
var ScopePresets = map[string][]string{
	"read":  {ScopeReadUser, ScopeReadWork},
	"write": {ScopeReadUser, ScopeReadWork, ScopeWriteWork},
	"admin": {ScopeReadUser, ScopeReadWork, ScopeWriteWork, ScopeManageProject, ScopeManageConfig},
	"all":   {ScopeReadUser, ScopeReadWork, ScopeWriteWork, ScopeManageProject, ScopeManageConfig, ScopeManageWebhook, ScopeManageDataProvider},
}

// ScopePresetOrder is the display order for the presets.
var ScopePresetOrder = []string{"read", "write", "admin", "all"}

// ScopePresetDescriptions describes each preset for the interactive picker.
var ScopePresetDescriptions = map[string]string{
	"read":  "Read only (issues, projects, users)",
	"write": "Read + write (create/edit/transition issues, comments) — default",
	"admin": "Read + write + manage projects & configuration",
	"all":   "Everything (all classic Jira scopes incl. webhooks & data provider)",
}

// ResolveScopes builds the space-separated scope string for a preset plus any
// extra granular scopes, always including offline_access.
func ResolveScopes(preset string, extra []string) (string, error) {
	base, ok := ScopePresets[preset]
	if !ok {
		return "", fmt.Errorf("unknown scope preset %q (use one of: %s)", preset, strings.Join(ScopePresetOrder, ", "))
	}
	seen := map[string]bool{}
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	for _, s := range base {
		add(s)
	}
	for _, s := range extra {
		add(s)
	}
	add(ScopeOffline)
	return strings.Join(out, " "), nil
}
