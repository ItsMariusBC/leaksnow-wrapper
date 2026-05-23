export const SCOPES = ["leak", "service", "shodan"] as const;
export type Scope = (typeof SCOPES)[number];

export const SEVERITIES = ["critical", "high", "medium", "low", "info", "all"] as const;
export type Severity = (typeof SEVERITIES)[number];

export const ULP_TYPES = ["domain", "email", "username", "password"] as const;
export type UlpType = (typeof ULP_TYPES)[number];

export const INTELX_BUCKETS = [
  "leaks.private.general",
  "leaks.logs",
  "whois",
  "dns",
  "darknet.tor",
] as const;
export type IntelxBucket = (typeof INTELX_BUCKETS)[number];

/** ids are numeric for shodan scans; accept string|number, sent as-is in the URL. */
export type ResourceId = string | number;

// --- Request bodies ---
export interface SearchRequest {
  query: string;
  scope?: Scope;
  plugin?: string;
  severity?: Severity;
  page?: number;
}

export interface ShodanCustomScanRequest {
  target: string;
}

export interface UlpSearchRequest {
  type: UlpType;
  value: string;
}

export interface UlpDownloadRequest {
  token: string;
}

export interface IntelxDownloadRequest {
  systemId: string;
  bucket: IntelxBucket;
}

// --- Responses (best-effort: provider documents requests only) ---
export interface UnknownRecord {
  [key: string]: unknown;
}
export type SearchResponse = UnknownRecord;
export type ShodanScan = UnknownRecord;
export type ShodanScanList = ShodanScan[] | UnknownRecord;
export type ShodanHost = UnknownRecord;
export interface UlpSearchResponse extends UnknownRecord {
  downloadToken?: string;
}
export type UlpDownloadResponse = UnknownRecord;
export type IntelxDownloadList = UnknownRecord;
export type IntelxDownloadResponse = UnknownRecord;

export interface BinaryFile {
  data: ArrayBuffer;
  contentType: string;
  filename?: string;
}
