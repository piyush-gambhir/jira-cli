// Package client is a thin HTTP client for the Jira REST API. It wraps an
// auth.Authenticator (which stamps credentials and reports the base URL),
// decodes the Jira error envelope, retries on rate limits, and exposes typed
// resource methods (issues, search, projects, users, agile, …).
package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/piyush-gambhir/jira-cli/internal/auth"
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
	return &Client{
		ctx:        context.Background(),
		auth:       a,
		apiVersion: apiVersion,
		verbose:    verbose,
		httpClient: &http.Client{
			Timeout:   60 * time.Second,
			Transport: transport,
			Jar:       jar,
		},
	}
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
