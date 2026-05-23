package leaksnow

// Service namespaces grouping the API endpoints. Methods live in shodan.go,
// ulp.go and intelx.go.
type ShodanService struct{ client *Client }
type ULPService struct{ client *Client }
type IntelXService struct{ client *Client }
