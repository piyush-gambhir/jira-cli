package auth

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type tokenTransport func(*http.Request) (*http.Response, error)

func (f tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestTokenRequestRejectsHTTPFailure(t *testing.T) {
	for _, body := range []string{`{"access_token":"must-not-be-used"}`, "temporarily unavailable"} {
		client := &http.Client{Transport: tokenTransport(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
		})}
		token, err := tokenRequest(client, map[string]string{"grant_type": "refresh_token"})
		if err == nil || !strings.Contains(err.Error(), "HTTP 503") || token != nil {
			t.Fatalf("tokenRequest = %v, %v; want HTTP 503 failure", token, err)
		}
	}
}

func TestTokenRequestAcceptsSuccessfulResponse(t *testing.T) {
	client := &http.Client{Transport: tokenTransport(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"access_token":"valid","expires_in":3600}`)), Header: make(http.Header)}, nil
	})}
	token, err := tokenRequest(client, map[string]string{"grant_type": "refresh_token"})
	if err != nil || token.AccessToken != "valid" {
		t.Fatalf("tokenRequest = %v, %v", token, err)
	}
}
