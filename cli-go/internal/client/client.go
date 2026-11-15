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

type Client struct {
	ctx        context.Context
	auth       auth.Authenticator
	apiVersion string
	httpClient *http.Client
	verbose    bool
}

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

func siteOrigin(base string) string {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return ""
	}
	return origin(u)
}

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

func (c *Client) WithContext(ctx context.Context) *Client {
	if ctx == nil {
		ctx = context.Background()
	}
	c.ctx = ctx
	return c
}

func (c *Client) APIVer() string { return c.apiVersion }

func (c *Client) IsServer() bool { return c.apiVersion == "2" }

func (c *Client) BaseURL() string { return c.auth.BaseURL() }

func (c *Client) AuthDescription() string { return c.auth.Describe() }

func (c *Client) api(p string, args ...any) string {
	return fmt.Sprintf("/rest/api/%s/%s", c.apiVersion, fmt.Sprintf(p, args...))
}

func (c *Client) agile(p string, args ...any) string {
	return "/rest/agile/1.0/" + fmt.Sprintf(p, args...)
}

func (c *Client) fullURL(path string, query url.Values) string {
	u := strings.TrimRight(c.auth.BaseURL(), "/") + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	return u
}
