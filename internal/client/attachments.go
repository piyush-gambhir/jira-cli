package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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

// DownloadAttachment fetches an attachment's bytes (following the redirect to
// the media store) along with its filename.
func (c *Client) DownloadAttachment(id string) ([]byte, string, error) {
	meta, err := c.GetAttachmentMeta(id)
	if err != nil {
		return nil, "", err
	}
	data, _, err := c.GetBytes(c.api("attachment/content/%s", id), nil, "*/*")
	if err != nil {
		return nil, "", err
	}
	return data, meta.Filename, nil
}

// DeleteAttachment removes an attachment.
func (c *Client) DeleteAttachment(id string) error {
	return c.Delete(c.api("attachment/%s", id), nil)
}
