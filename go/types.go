package leaksnow

// Scope selects the search backend for Search.
type Scope string

const (
	ScopeLeak    Scope = "leak"
	ScopeService Scope = "service"
	ScopeShodan  Scope = "shodan"
)

// Severity filters search results by severity.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
	SeverityAll      Severity = "all"
)

// ULPType selects the field searched by ULPService.Search.
type ULPType string

const (
	ULPTypeDomain   ULPType = "domain"
	ULPTypeEmail    ULPType = "email"
	ULPTypeUsername ULPType = "username"
	ULPTypePassword ULPType = "password"
)

// IntelXBucket selects the Intelligence X data bucket.
type IntelXBucket string

const (
	BucketLeaksPrivateGeneral IntelXBucket = "leaks.private.general"
	BucketLeaksLogs           IntelXBucket = "leaks.logs"
	BucketWhois               IntelXBucket = "whois"
	BucketDNS                 IntelXBucket = "dns"
	BucketDarknetTor          IntelXBucket = "darknet.tor"
)

// SearchRequest is the body for Search.
type SearchRequest struct {
	Query    string   `json:"query"`
	Scope    Scope    `json:"scope,omitempty"`
	Plugin   string   `json:"plugin,omitempty"`
	Severity Severity `json:"severity,omitempty"`
	Page     int      `json:"page,omitempty"`
}

// ShodanCustomScanRequest is the body for ShodanService.CustomScan.
type ShodanCustomScanRequest struct {
	Target string `json:"target"`
}

// ULPSearchRequest is the body for ULPService.Search.
type ULPSearchRequest struct {
	Type  ULPType `json:"type"`
	Value string  `json:"value"`
}

// ULPDownloadRequest is the body for ULPService.Download.
type ULPDownloadRequest struct {
	Token string `json:"token"`
}

// IntelXDownloadRequest is the body for IntelXService.Download.
type IntelXDownloadRequest struct {
	SystemID string       `json:"systemId"`
	Bucket   IntelXBucket `json:"bucket"`
}

// BinaryFile is the result of IntelXService.GetFile.
type BinaryFile struct {
	Data        []byte
	ContentType string
	Filename    string
}
