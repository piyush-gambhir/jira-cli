package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Both servers listen on 127.0.0.1, so the off-site one differs only by port:
// Go's own redirect rule and cookie scoping both ignore the port.
func TestRedirectSendsCredentialsOnlyToTheSiteOrigin(t *testing.T) {
	var offAuth, offCookie string
	off := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offAuth, offCookie = r.Header.Get("Authorization"), r.Header.Get("Cookie")
		_, _ = io.WriteString(w, "media bytes")
	}))
	defer off.Close()

	var siteAuth, siteCookie string
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rest/api/3/attachment/1":
			http.SetCookie(w, &http.Cookie{Name: "JSESSIONID", Value: "session", Path: "/"})
			_, _ = io.WriteString(w, `{"id":"1","filename":"a.txt"}`)
		case "/rest/api/3/attachment/content/1":
			http.Redirect(w, r, off.URL+"/file/1", http.StatusFound)
		case "/to-self":
			http.Redirect(w, r, "/landing", http.StatusFound)
		case "/landing":
			siteAuth, siteCookie = r.Header.Get("Authorization"), r.Header.Get("Cookie")
			_, _ = io.WriteString(w, "ok")
		default:
			http.NotFound(w, r)
		}
	}))
	defer site.Close()
	if !strings.HasPrefix(off.URL, "http://127.0.0.1:") || !strings.HasPrefix(site.URL, "http://127.0.0.1:") {
		t.Fatalf("test servers are not on the same host: %s, %s", site.URL, off.URL)
	}

	c := testClient(t, site.URL)
	data, _, err := c.DownloadAttachment("1")
	if err != nil || string(data) != "media bytes" {
		t.Fatalf("DownloadAttachment = %q, %v", data, err)
	}
	if offAuth != "" || offCookie != "" {
		t.Fatalf("redirect to another port carried credentials: Authorization=%q Cookie=%q", offAuth, offCookie)
	}

	if _, _, err := c.GetBytes("/to-self", nil, ""); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(siteAuth, "Basic ") || !strings.Contains(siteCookie, "JSESSIONID=session") {
		t.Fatalf("same-origin redirect lost credentials: Authorization=%q Cookie=%q", siteAuth, siteCookie)
	}
}
