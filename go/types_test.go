package leaksnow

import (
	"encoding/json"
	"testing"
)

func TestEnumValues(t *testing.T) {
	if ScopeLeak != "leak" || ScopeService != "service" || ScopeShodan != "shodan" {
		t.Fatal("scope constants wrong")
	}
	if SeverityAll != "all" || SeverityCritical != "critical" {
		t.Fatal("severity constants wrong")
	}
	if ULPTypeDomain != "domain" || ULPTypeEmail != "email" {
		t.Fatal("ulp type constants wrong")
	}
	if BucketLeaksPrivateGeneral != "leaks.private.general" || BucketDarknetTor != "darknet.tor" {
		t.Fatal("bucket constants wrong")
	}
}

func TestSearchRequestJSONOmitsEmpty(t *testing.T) {
	b, err := json.Marshal(SearchRequest{Query: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"query":"x"}` {
		t.Fatalf("marshal = %s, want {\"query\":\"x\"}", b)
	}
	b, _ = json.Marshal(SearchRequest{Query: "x", Scope: ScopeLeak, Severity: SeverityAll, Page: 0})
	if string(b) != `{"query":"x","scope":"leak","severity":"all"}` {
		t.Fatalf("marshal = %s", b)
	}
}

func TestULPAndIntelXRequestJSON(t *testing.T) {
	b, _ := json.Marshal(ULPSearchRequest{Type: ULPTypeDomain, Value: "example.com"})
	if string(b) != `{"type":"domain","value":"example.com"}` {
		t.Fatalf("ulp marshal = %s", b)
	}
	b, _ = json.Marshal(IntelXDownloadRequest{SystemID: "abc", Bucket: BucketLeaksLogs})
	if string(b) != `{"systemId":"abc","bucket":"leaks.logs"}` {
		t.Fatalf("intelx marshal = %s", b)
	}
}
