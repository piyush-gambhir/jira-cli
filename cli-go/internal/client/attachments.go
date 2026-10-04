package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// AddAttachment uploads one or more files to an issue. It sends the mandatory
// X-Atlassian-Token: no-check header and uses the required form field name "file".
func (c *Client) AddAttachment(key string, files []string) ([]Attachment, error) {
	type fileData struct {
		name string
		data []byte
	}
	var loaded []fileData
	for _, p := range files {
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", p, err)
		}
		loaded = append(loaded, fileData{name: filepath.Base(p), data: data})
	}

	resp, err := c.doRetry(func() (*http.Request, error) {
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		for _, f := range loaded {
			part, err := w.CreateFormFile("file", f.name)
			if err != nil {
				return nil, err
			}
			if _, err := part.Write(f.data); err != nil {
				return nil, err
			}
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(c.ctx, http.MethodPost, c.fullURL(c.api("issue/%s/attachments", key), nil), &buf)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", w.FormDataContentType())
		req.Header.Set("X-Atlassian-Token", "no-check")
		return req, nil
	}, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := readAllLimited(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, resp.Status, resp.Request.URL.String(), data)
	}
	var out []Attachment
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("decoding attachments: %w", err)
	}
	return out, nil
}

// ListAttachments returns the attachment metadata on an issue.
func (c *Client) ListAttachments(key string) ([]Attachment, error) {
	var out struct {
		Fields struct {
			Attachment []Attachment `json:"attachment"`
		} `json:"fields"`
	}
	if err := c.GetJSON(c.api("issue/%s", key), issueQuery("attachment", ""), &out); err != nil {
		return nil, err
	}
	return out.Fields.Attachment, nil
}

// GetAttachmentMeta returns metadata for a single attachment.
func (c *Client) GetAttachmentMeta(id string) (*Attachment, error) {
	var out Attachment
	if err := c.GetJSON(c.api("attachment/%s", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DownloadAttachment fetches an attachment's bytes along with its filename. On
// Cloud it uses attachment/content/{id} (following the redirect to the media
// store); Server/DC has no such endpoint, so it downloads the metadata's
// content URL instead.
func (c *Client) DownloadAttachment(id string) ([]byte, string, error) {
	meta, err := c.GetAttachmentMeta(id)
	if err != nil {
		return nil, "", err
	}
	path := c.api("attachment/content/%s", id)
	if c.IsServer() {
		if path, err = c.sitePath(meta.Content); err != nil {
			return nil, "", fmt.Errorf("downloading attachment %s: %w", id, err)
		}
	}
	data, _, err := c.GetBytes(path, nil, "*/*")
	if err != nil {
		return nil, "", err
	}
	return data, meta.Filename, nil
}

// sitePath turns an absolute URL Jira returned (such as an attachment's content
// URL) into a path under the configured site. The URL must be on the site's
// exact origin (scheme, host, and port) and under its base path, with no ".."
// segments, so the request and its credentials can only reach the configured
// site's own resources.
func (c *Client) sitePath(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Path == "" {
		return "", fmt.Errorf("unusable URL %q", raw)
	}
	base, err := url.Parse(strings.TrimRight(c.auth.BaseURL(), "/"))
	if err != nil {
		return "", err
	}
	if (u.Scheme != "" || u.Host != "") && origin(u) != origin(base) {
		return "", fmt.Errorf("URL %q is not on the configured site %s (set --site to Jira's base URL)", raw, c.auth.BaseURL())
	}
	for _, seg := range strings.FieldsFunc(u.Path, func(r rune) bool { return r == '/' || r == '\\' }) {
		if seg == ".." {
			return "", fmt.Errorf("URL %q contains a \"..\" path segment", raw)
		}
	}
	p := u.EscapedPath()
	if !strings.HasPrefix(p, base.EscapedPath()+"/") {
		return "", fmt.Errorf("URL %q is outside the configured site %s", raw, c.auth.BaseURL())
	}
	return strings.TrimPrefix(p, base.EscapedPath()), nil
}

// DeleteAttachment removes an attachment.
func (c *Client) DeleteAttachment(id string) error {
	return c.Delete(c.api("attachment/%s", id), nil)
}
