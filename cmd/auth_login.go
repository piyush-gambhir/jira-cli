package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/jira-cli/internal/auth"
	"github.com/piyush-gambhir/jira-cli/internal/client"
	"github.com/piyush-gambhir/jira-cli/internal/config"
)

func newAuthLoginCmd() *cobra.Command {
	var (
		authType     string
		profileName  string
		password     string
		clientID     string
		clientSecret string
		scopes       string
		scopePreset  string
		extraScopes  []string
		port         int
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate to a Jira site and save a profile",
		Long: `Authenticate to a Jira site. The auth method is chosen with --type
(default: oauth2 — browser login):

  oauth2        (default) Cloud: OAuth 2.0 (3LO) browser login. Pass your app's
                --client-id/--client-secret (callback http://localhost:<port>/callback).
  api_token     Cloud: prompts for site, email, API token (no app needed).
                Create a token at https://id.atlassian.com/manage-profile/security/api-tokens
  scoped_token  Cloud: like api_token but for a scoped token; resolves the cloudId
                and routes via the api.atlassian.com gateway.
  pat           Server/DC: prompts for site (host) and a personal access token.
  basic         Server/DC: prompts for site (host), username and password.

Inputs may be supplied via flags (--site, --email, --token, --user) for
non-interactive use; missing required values are prompted unless --no-input.

Examples:
  jira auth login                         # browser OAuth (default)
  jira auth login --type api_token        # paste site / email / API token
  jira auth login --type pat --site https://jira.company.com --token "$PAT"
  jira auth login --type oauth2 --client-id ABC --client-secret XYZ   # bring-your-own app`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			at := normalizeAuthType(authType)
			if at == "" {
				return fmt.Errorf("unknown --type %q (use oauth2, api_token, scoped, pat, or basic)", authType)
			}
			if profileName == "" {
				profileName = "default"
			}

			profile := config.Profile{AuthType: at, Insecure: insecureFlag}
			var err error
			switch at {
			case config.AuthAPIToken, config.AuthScopedToken:
				if profile.Site, err = need(siteFlag, "Jira site (https://your-domain.atlassian.net): "); err != nil {
					return err
				}
				if profile.Email, err = need(emailFlag, "Atlassian account email: "); err != nil {
					return err
				}
				if profile.Token, err = needSecret(tokenFlag, "API token: "); err != nil {
					return err
				}
				profile.Site = strings.TrimRight(profile.Site, "/")
				if at == config.AuthScopedToken {
					info("Resolving cloudId for %s ...", profile.Site)
					if profile.CloudID, err = auth.ResolveCloudID(profile.Site, insecureFlag); err != nil {
						return err
					}
				}
			case config.AuthPAT:
				if profile.Site, err = need(siteFlag, "Jira site (https://jira.host): "); err != nil {
					return err
				}
				if profile.Token, err = needSecret(tokenFlag, "Personal access token: "); err != nil {
					return err
				}
				profile.Site = strings.TrimRight(profile.Site, "/")
			case config.AuthBasic:
				if profile.Site, err = need(siteFlag, "Jira site (https://jira.host): "); err != nil {
					return err
				}
				if profile.Username, err = need(userFlag, "Username: "); err != nil {
					return err
				}
				if profile.Password, err = needSecret(password, "Password: "); err != nil {
					return err
				}
				profile.Site = strings.TrimRight(profile.Site, "/")
			case config.AuthOAuth2:
				if clientID, err = need(clientID, "OAuth client id: "); err != nil {
					return err
				}
				if clientSecret, err = needSecret(clientSecret, "OAuth client secret: "); err != nil {
					return err
				}
				if scopes == "" { // --scopes (raw) wins; otherwise resolve from preset/picker
					if scopes, err = resolveLoginScopes(scopePreset, extraScopes); err != nil {
						return err
					}
				}
				if !strings.Contains(scopes, auth.ScopeOffline) {
					scopes = strings.TrimSpace(scopes + " " + auth.ScopeOffline)
				}
				info("Requesting scopes: %s", scopes)
				info("(Your app's Permissions in the developer console must include these scopes.)")
				info("Starting OAuth 2.0 (3LO) flow on http://localhost:%d/callback ...", oauthPort(port))
				res, err := auth.OAuthLogin(clientID, clientSecret, scopes, port, insecureFlag)
				if err != nil {
					return err
				}
				// Resolve which site this token can access.
				resources, err := auth.AccessibleResources(res.AccessToken, insecureFlag)
				if err != nil {
					return err
				}
				if len(resources) == 0 {
					return fmt.Errorf("the authorized account has no accessible Jira sites")
				}
				chosen := resources[0]
				if len(resources) > 1 {
					chosen, err = pickResource(resources)
					if err != nil {
						return err
					}
				}
				profile.ClientID = clientID
				profile.ClientSecret = clientSecret
				profile.Scopes = scopes
				profile.AccessToken = res.AccessToken
				profile.RefreshToken = res.RefreshToken
				profile.TokenExpiry = nowUnix() + int64(res.ExpiresIn)
				profile.CloudID = chosen.ID
				profile.Site = strings.TrimRight(chosen.URL, "/")
			}

			// Test the connection.
			info("Testing connection ...")
			me, err := testProfile(profile, profileName)
			if err != nil {
				return fmt.Errorf("connection test failed: %w", err)
			}

			if err := config.Update(func(c *config.Config) error {
				config.SetProfile(c, profileName, profile)
				if c.CurrentProfile == "" {
					c.CurrentProfile = profileName
				}
				return nil
			}); err != nil {
				return err
			}
			info("Authenticated as %s. Profile %q saved to %s", me.DisplayName, profileName, config.ConfigPath())
			return nil
		},
	}

	cmd.Flags().StringVar(&authType, "type", "oauth2", "Auth method: oauth2 (default, browser), api_token, scoped, pat, basic")
	cmd.Flags().StringVar(&profileName, "name", "default", "Profile name to save under")
	cmd.Flags().StringVar(&password, "password", "", "Password (basic auth; prompted if omitted)")
	cmd.Flags().StringVar(&clientID, "client-id", "", "OAuth 2.0 client id")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "OAuth 2.0 client secret")
	cmd.Flags().StringVar(&scopes, "scopes", "", "OAuth 2.0 scopes, space-separated (raw override; wins over --scope-preset/--scope)")
	cmd.Flags().StringVar(&scopePreset, "scope-preset", "", "OAuth scope preset: read, write, admin, all (prompted if omitted and interactive)")
	cmd.Flags().StringArrayVar(&extraScopes, "scope", nil, "Add an individual OAuth scope on top of the preset (repeatable)")
	cmd.Flags().IntVar(&port, "port", auth.DefaultCallbackPort, "Local callback port for the OAuth flow")
	return cmd
}

// resolveLoginScopes determines the OAuth scope string from a preset, defaulting
// without a prompt for non-interactive builds and showing the interactive picker
// otherwise. Extra granular scopes are appended.
func resolveLoginScopes(preset string, extra []string) (string, error) {
	if preset == "" {
		switch {
		case noInputFlag:
			preset = "write"
		default:
			p, err := promptScopePreset()
			if err != nil {
				return "", err
			}
			preset = p
		}
	}
	if preset == "custom" {
		raw, err := prompt("Enter space-separated scopes: ")
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(raw) == "" {
			return auth.DefaultScopes, nil
		}
		return raw, nil
	}
	return auth.ResolveScopes(preset, extra)
}

// promptScopePreset shows the scope presets and reads the user's choice.
func promptScopePreset() (string, error) {
	fmt.Fprintln(stderr(), "Select the access to grant (OAuth scopes to request):")
	for i, name := range auth.ScopePresetOrder {
		fmt.Fprintf(stderr(), "  [%d] %-6s %s\n", i+1, name, auth.ScopePresetDescriptions[name])
	}
	fmt.Fprintf(stderr(), "  [%d] custom — enter your own space-separated scopes\n", len(auth.ScopePresetOrder)+1)
	choice, err := prompt("Choice [2]: ")
	if err != nil {
		return "", err
	}
	choice = strings.ToLower(strings.TrimSpace(choice))
	if choice == "" {
		return "write", nil
	}
	var n int
	if _, e := fmt.Sscanf(choice, "%d", &n); e == nil {
		if n >= 1 && n <= len(auth.ScopePresetOrder) {
			return auth.ScopePresetOrder[n-1], nil
		}
		if n == len(auth.ScopePresetOrder)+1 {
			return "custom", nil
		}
	}
	if _, ok := auth.ScopePresets[choice]; ok {
		return choice, nil
	}
	if choice == "custom" {
		return "custom", nil
	}
	return "", fmt.Errorf("invalid choice %q", choice)
}

// normalizeAuthType maps friendly aliases to the stored auth_type value.
func normalizeAuthType(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "", "api_token", "apitoken", "token", "api":
		return config.AuthAPIToken
	case "scoped", "scoped_token":
		return config.AuthScopedToken
	case "oauth2", "oauth", "3lo":
		return config.AuthOAuth2
	case "pat":
		return config.AuthPAT
	case "basic", "password":
		return config.AuthBasic
	}
	return ""
}

// need returns the flag value if set, else prompts for it; errors if still empty.
func need(flagVal, label string) (string, error) {
	if flagVal != "" {
		return flagVal, nil
	}
	v, err := prompt(label)
	if err != nil {
		return "", err
	}
	if v == "" {
		return "", fmt.Errorf("%s is required", strings.TrimRight(strings.TrimRight(label, " "), ":"))
	}
	return v, nil
}

func needSecret(flagVal, label string) (string, error) {
	if flagVal != "" {
		return flagVal, nil
	}
	v, err := promptSecret(label)
	if err != nil {
		return "", err
	}
	if v == "" {
		return "", fmt.Errorf("%s is required", strings.TrimRight(strings.TrimRight(label, " "), ":"))
	}
	return v, nil
}

// testProfile builds an authenticator + client for the profile and calls myself.
func testProfile(profile config.Profile, name string) (*client.User, error) {
	persist := func(updated config.Profile) error { return config.PersistProfile(name, updated) }
	authr, err := auth.New(profile, name, persist)
	if err != nil {
		return nil, err
	}
	c := client.NewClient(authr, profile.EffectiveAPIVersion(), profile.Insecure, verboseFlag)
	return c.Myself()
}

func pickResource(resources []auth.AccessibleResource) (auth.AccessibleResource, error) {
	fmt.Fprintln(stderr(), "Multiple Jira sites are accessible:")
	for i, r := range resources {
		fmt.Fprintf(stderr(), "  [%d] %s (%s)\n", i+1, r.Name, r.URL)
	}
	choice, err := prompt("Choose a site [1]: ")
	if err != nil {
		return auth.AccessibleResource{}, err
	}
	if choice == "" {
		return resources[0], nil
	}
	idx := 0
	if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx < 1 || idx > len(resources) {
		return auth.AccessibleResource{}, fmt.Errorf("invalid choice %q", choice)
	}
	return resources[idx-1], nil
}

func oauthPort(p int) int {
	if p == 0 {
		return auth.DefaultCallbackPort
	}
	return p
}
