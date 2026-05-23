package leaksnow

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// Downloads returns your IntelX download history. Cost: 0.
func (s *IntelXService) Downloads(ctx context.Context) (json.RawMessage, error) {
	return s.client.doJSON(ctx, http.MethodGet, "/api/v1/intelx/downloads", nil)
}

// Download requests an IntelX file (server-cached). Cost: 5.
func (s *IntelXService) Download(ctx context.Context, req IntelXDownloadRequest) (json.RawMessage, error) {
	return s.client.doJSON(ctx, http.MethodPost, "/api/v1/intelx/download", req)
}

// DeleteDownload hides an entry in the IntelX history. Cost: 0.
func (s *IntelXService) DeleteDownload(ctx context.Context, id string) (json.RawMessage, error) {
	return s.client.doJSON(ctx, http.MethodDelete, "/api/v1/intelx/downloads/"+url.PathEscape(id), nil)
}

// GetFile downloads the cached binary for a download id (no re-billing). Cost: 0.
func (s *IntelXService) GetFile(ctx context.Context, id string) (*BinaryFile, error) {
	resp, err := s.client.doRequest(ctx, http.MethodGet, "/api/v1/intelx/downloads/"+url.PathEscape(id)+"/file", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &TransportError{Op: "GET intelx file", Err: err}
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	return &BinaryFile{
		Data:        data,
		ContentType: ct,
		Filename:    parseFilename(resp.Header.Get("Content-Disposition")),
	}, nil
}

var (
	reFilenameExt   = regexp.MustCompile(`(?i)filename\*\s*=\s*(?:[\w-]+)?'[^']*'([^;]+)`)
	reFilenameBasic = regexp.MustCompile(`(?i)filename\s*=\s*"?([^";]+)"?`)
)

// parseFilename extracts a filename from a Content-Disposition header.
// The RFC 5987 extended form (filename*) takes precedence and is percent-decoded.
func parseFilename(disposition string) string {
	if disposition == "" {
		return ""
	}
	if m := reFilenameExt.FindStringSubmatch(disposition); m != nil {
		raw := strings.TrimSpace(m[1])
		if dec, err := url.PathUnescape(raw); err == nil {
			return dec
		}
		return raw
	}
	if m := reFilenameBasic.FindStringSubmatch(disposition); m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}
