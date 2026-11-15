package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	maxRetries = 4
	maxBackoff = 30 * time.Second
)

type buildFunc func() (*http.Request, error)

func (c *Client) doRetry(build buildFunc, retryable bool) (*http.Response, error) {
	backoff := 2 * time.Second
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := build()
		if err != nil {
			return nil, err
		}
		if req.Header.Get("Accept") == "" {
			req.Header.Set("Accept", "application/json")
		}
		if err := c.auth.Apply(req); err != nil {
			return nil, err
		}
		if c.verbose {
			fmt.Fprintf(os.Stderr, "--> %s %s\n", req.Method, req.URL)
		}
		start := time.Now()
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if c.verbose {
				fmt.Fprintf(os.Stderr, "<-- ERROR: %v (%v)\n", err, time.Since(start))
			}
			if retryable && attempt < maxRetries {
				if err := sleepContext(req.Context(), jitter(backoff)); err != nil {
					return nil, err
				}
				backoff = nextBackoff(backoff)
				continue
			}
			return nil, fmt.Errorf("executing request: %w", err)
		}
		if c.verbose {
			fmt.Fprintf(os.Stderr, "<-- %d %s (%v)\n", resp.StatusCode, resp.Status, time.Since(start))
		}

		if retryable && (resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode >= 500 && resp.StatusCode != http.StatusNotImplemented)) {
			if attempt < maxRetries {
				wait := retryAfter(resp, backoff)
				resp.Body.Close()
				if err := sleepContext(req.Context(), wait); err != nil {
					return nil, err
				}
				backoff = nextBackoff(backoff)
				continue
			}
		}
		return resp, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("request failed after %d retries", maxRetries)
}

func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func retryAfter(resp *http.Response, backoff time.Duration) time.Duration {
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if secs, err := strconv.Atoi(ra); err == nil {
			return time.Duration(secs) * time.Second
		}
	}
	return jitter(backoff)
}

func nextBackoff(b time.Duration) time.Duration {
	b *= 2
	if b > maxBackoff {
		return maxBackoff
	}
	return b
}

func jitter(d time.Duration) time.Duration {
	factor := 0.7 + rand.Float64()*0.6
	return time.Duration(float64(d) * factor)
}

func (c *Client) doJSON(method, path string, query url.Values, body, out any, retryable bool) error {
	return c.doJSONHeader(method, path, query, body, out, retryable, nil)
}

func (c *Client) doJSONHeader(method, path string, query url.Values, body, out any, retryable bool, header http.Header) error {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
	}
	resp, err := c.doRetry(func() (*http.Request, error) {
		var r io.Reader
		if bodyBytes != nil {
			r = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(c.ctx, method, c.fullURL(path, query), r)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		for k, v := range header {
			req.Header[k] = v
		}
		return req, nil
	}, retryable)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := readAllLimited(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return parseAPIError(resp.StatusCode, resp.Status, resp.Request.URL.String(), data)
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

func (c *Client) GetJSON(path string, query url.Values, out any) error {
	return c.doJSON(http.MethodGet, path, query, nil, out, true)
}

func (c *Client) PostJSON(path string, query url.Values, body, out any) error {
	return c.doJSON(http.MethodPost, path, query, body, out, false)
}

func (c *Client) PostJSONRetryable(path string, query url.Values, body, out any) error {
	return c.doJSON(http.MethodPost, path, query, body, out, true)
}

func (c *Client) PutJSON(path string, query url.Values, body, out any) error {
	return c.doJSON(http.MethodPut, path, query, body, out, true)
}

func (c *Client) Delete(path string, query url.Values) error {
	return c.doJSON(http.MethodDelete, path, query, nil, nil, true)
}

func (c *Client) GetBytes(path string, query url.Values, accept string) ([]byte, *http.Response, error) {
	resp, err := c.doRetry(func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, c.fullURL(path, query), nil)
		if err != nil {
			return nil, err
		}
		if accept != "" {
			req.Header.Set("Accept", accept)
		}
		return req, nil
	}, true)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	data, err := readAllLimited(resp.Body)
	if err != nil {
		return nil, resp, fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, resp, parseAPIError(resp.StatusCode, resp.Status, resp.Request.URL.String(), data)
	}
	return data, resp, nil
}
