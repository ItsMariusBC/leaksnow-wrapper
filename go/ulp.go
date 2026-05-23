package leaksnow

import (
	"context"
	"encoding/json"
	"net/http"
)

// Search runs a ULP search. The response includes a downloadToken (~15 min). Cost: 1.
func (s *ULPService) Search(ctx context.Context, req ULPSearchRequest) (json.RawMessage, error) {
	return s.client.doJSON(ctx, http.MethodPost, "/api/v1/ulp/search", req)
}

// Download exports the JSON rows for a previously returned token. Cost: 10.
func (s *ULPService) Download(ctx context.Context, req ULPDownloadRequest) (json.RawMessage, error) {
	return s.client.doJSON(ctx, http.MethodPost, "/api/v1/ulp/download", req)
}
