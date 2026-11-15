package auth

import (
	"fmt"
	"strings"
)

const (
	ScopeReadUser           = "read:jira-user"
	ScopeReadWork           = "read:jira-work"
	ScopeWriteWork          = "write:jira-work"
	ScopeManageProject      = "manage:jira-project"
	ScopeManageConfig       = "manage:jira-configuration"
	ScopeManageWebhook      = "manage:jira-webhook"
	ScopeManageDataProvider = "manage:jira-data-provider"
	ScopeOffline            = "offline_access"
)

var ScopePresets = map[string][]string{
	"read":  {ScopeReadUser, ScopeReadWork},
	"write": {ScopeReadUser, ScopeReadWork, ScopeWriteWork},
	"admin": {ScopeReadUser, ScopeReadWork, ScopeWriteWork, ScopeManageProject, ScopeManageConfig},
	"all":   {ScopeReadUser, ScopeReadWork, ScopeWriteWork, ScopeManageProject, ScopeManageConfig, ScopeManageWebhook, ScopeManageDataProvider},
}

var ScopePresetOrder = []string{"read", "write", "admin", "all"}

var ScopePresetDescriptions = map[string]string{
	"read":  "Read only (issues, projects, users)",
	"write": "Read + write (create/edit/transition issues, comments) — default",
	"admin": "Read + write + manage projects & configuration",
	"all":   "Everything (all classic Jira scopes incl. webhooks & data provider)",
}

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
