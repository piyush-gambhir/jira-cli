// Package client is a thin HTTP client for the Jira REST API. It wraps an
// auth.Authenticator (which stamps credentials and reports the base URL),
// decodes the Jira error envelope, retries on rate limits, and exposes typed
// resource methods (issues, search, projects, users, agile, …).
package client

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/auth"
)

// Client talks to one Jira site using one authenticator.
type Client struct {
	ctx        context.Context
	auth       auth.Authenticator
	apiVersion string // platform REST version: "3" (Cloud) or "2" (Server/DC)
	httpClient *http.Client
	verbose    bool
}

// NewClient builds a client for the given authenticator and platform API version.
func NewClient(a auth.Authenticator, apiVersion string, insecure, verbose bool) *Client {
	if apiVersion == "" {
		apiVersion = "3"
	}
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: insecure},
	}
	jar, _ := cookiejar.New(nil)
	site := siteOrigin(a.BaseURL())
	return &Client{
		ctx:        context.Background(),
		auth:       a,
		apiVersion: apiVersion,
		verbose:    verbose,
		httpClient: &http.Client{
			Timeout:       60 * time.Second,
			Transport:     transport,
			Jar:           siteJar{jar, site},
			CheckRedirect: stripAuthOffSite(site),
		},
	}
}

// siteOrigin returns the scheme://host:port origin of the configured site.
func siteOrigin(base string) string {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return ""
	}
	return origin(u)
}

// origin returns u's scheme://host:port, with the default port made explicit
// so https://h and https://h:443 compare equal.
func origin(u *url.URL) string {
	scheme := strings.ToLower(u.Scheme)
	port := u.Port()
	if port == "" {
		switch scheme {
		case "https":
			port = "443"
		case "http":
			port = "80"
		}
	}
	return scheme + "://" + net.JoinHostPort(strings.ToLower(u.Hostname()), port)
}

// stripAuthOffSite is the redirect policy. Go keeps the Authorization header
// on redirects to another port or a subdomain of the same host, so it is
// removed whenever the target is not exactly the configured site's origin.
func stripAuthOffSite(site string) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		if origin(req.URL) != site {
			req.Header.Del("Authorization")
			req.Header.Del("Cookie")
		}
		return nil
	}
}

// siteJar keeps session cookies to the configured site's exact origin (cookie
// scope otherwise ignores the port), so a redirect elsewhere cannot carry them.
type siteJar struct {
	jar  http.CookieJar
	site string
}

func (j siteJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	if origin(u) == j.site {
		j.jar.SetCookies(u, cookies)
	}
}

func (j siteJar) Cookies(u *url.URL) []*http.Cookie {
	if origin(u) != j.site {
		return nil
	}
	return j.jar.Cookies(u)
}

// WithContext sets the default context for subsequent requests and returns c.
func (c *Client) WithContext(ctx context.Context) *Client {
	if ctx == nil {
		ctx = context.Background()
	}
	c.ctx = ctx
	return c
}

// APIVer returns the platform REST API version in use ("3" or "2").
func (c *Client) APIVer() string { return c.apiVersion }

// IsServer reports whether the client targets Jira Server/Data Center. The
// platform API version is the deployment signal: v2 is the Server/DC default,
// v3 is Cloud (see config.Profile.EffectiveAPIVersion).
func (c *Client) IsServer() bool { return c.apiVersion == "2" }

// BaseURL returns the site base URL the client targets.
func (c *Client) BaseURL() string { return c.auth.BaseURL() }

// AuthDescription returns a human-readable description of the active auth method.
func (c *Client) AuthDescription() string { return c.auth.Describe() }

// api builds a platform REST path: /rest/api/{ver}/<p>.
func (c *Client) api(p string, args ...any) string {
	return fmt.Sprintf("/rest/api/%s/%s", c.apiVersion, fmt.Sprintf(p, args...))
}

// agile builds an Agile REST path: /rest/agile/1.0/<p>.
func (c *Client) agile(p string, args ...any) string {
	return "/rest/agile/1.0/" + fmt.Sprintf(p, args...)
}

// fullURL joins the auth base URL, a path, and query parameters.
func (c *Client) fullURL(path string, query url.Values) string {
	u := strings.TrimRight(c.auth.BaseURL(), "/") + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	return u
}
