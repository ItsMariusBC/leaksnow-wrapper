package leaksnow

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// CustomScans lists your custom Shodan scans. Cost: 0.
func (s *ShodanService) CustomScans(ctx context.Context) (json.RawMessage, error) {
	return s.client.doJSON(ctx, http.MethodGet, "/api/v1/shodan/custom-scans", nil)
}

// CustomScan launches a Shodan scan on the target. Cost: 1.
func (s *ShodanService) CustomScan(ctx context.Context, req ShodanCustomScanRequest) (json.RawMessage, error) {
	return s.client.doJSON(ctx, http.MethodPost, "/api/v1/shodan/custom-scan", req)
}

// GetScan returns the status of a scan by id. Cost: 0.
func (s *ShodanService) GetScan(ctx context.Context, id string) (json.RawMessage, error) {
	return s.client.doJSON(ctx, http.MethodGet, "/api/v1/shodan/custom-scans/"+url.PathEscape(id), nil)
}

// DeleteScan removes a scan entry by id. Cost: 0.
func (s *ShodanService) DeleteScan(ctx context.Context, id string) (json.RawMessage, error) {
	return s.client.doJSON(ctx, http.MethodDelete, "/api/v1/shodan/custom-scans/"+url.PathEscape(id), nil)
}

// Host returns the aggregated Shodan view for an IPv4. Cost: 0.
func (s *ShodanService) Host(ctx context.Context, ip string) (json.RawMessage, error) {
	return s.client.doJSON(ctx, http.MethodGet, "/api/v1/shodan/host/"+url.PathEscape(ip), nil)
}
