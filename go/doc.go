// Package leaksnow is a Go client for the leaks.now /api/v1 OSINT API
// (Leak, Service, Shodan, ULP, IntelX).
//
// Construct a client with NewClient and call methods with a context:
//
//	c := leaksnow.NewClient(os.Getenv("LEAKSNOW_API_KEY"))
//	res, err := c.Search(ctx, leaksnow.SearchRequest{Query: "host:example.com"})
//
// Authentication uses an "Authorization: Bearer ms_..." header. Each successful
// call consumes the credits documented in docs/openapi.yaml at the repo root.
package leaksnow
