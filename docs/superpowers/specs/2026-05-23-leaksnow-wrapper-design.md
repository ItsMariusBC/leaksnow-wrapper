# leaks.now API Wrapper — Design

**Date:** 2026-05-23
**Status:** Approved (pending spec review)
**Repo:** https://github.com/ItsMariusBC/leaksnow-wrapper.git

## Goal

Multi-language SDK wrapping the leaks.now `/api/v1` OSINT API (Leak, Service,
Shodan, ULP, IntelX). TypeScript delivered first (complete + tested), then Go
and Rust ports. Single monorepo, OpenAPI spec as shared contract.

## Authentication

All `/api/v1/*` routes use the API key only (no session cookie).

```
Authorization: Bearer ms_xxxxxxxxxxxx
Content-Type: application/json
```

Key prefix: `ms_`. Base URL: `https://leaks.now`.

## Monorepo Layout

```
leaksnow-wrapper/
├── packages/typescript/      # @leaksnow/sdk  (iteration 1)
│   ├── src/
│   │   ├── client.ts         # LeaksNowClient + transport
│   │   ├── errors.ts         # LeaksNowError + subclasses
│   │   ├── retry.ts          # backoff logic
│   │   ├── types.ts          # request/response/enum types
│   │   └── resources/        # search, shodan, ulp, intelx
│   ├── test/
│   ├── package.json
│   ├── tsconfig.json
│   ├── tsup.config.ts
│   └── vitest.config.ts
├── go/                       # leaksnow  (iteration 2)
├── rust/                     # leaksnow crate (iteration 3)
├── docs/
│   ├── openapi.yaml          # single source of truth
│   └── superpowers/specs/
├── pnpm-workspace.yaml
├── turbo.json
├── package.json              # root, workspace scripts
├── .gitignore
└── README.md
```

Tooling: pnpm workspaces + Turborepo. `go/` and `rust/` are sibling dirs (not
pnpm workspaces; built/tested via their own toolchains, orchestrated by Turbo
tasks where useful). No publish pipeline this iteration — code only.

## API Surface (endpoint → method mapping)

| Method | Endpoint | TS method | Credits |
|---|---|---|---|
| POST | `/api/v1/search` | `client.search(body)` | 1 |
| GET | `/api/v1/shodan/custom-scans` | `client.shodan.customScans()` | 0 |
| POST | `/api/v1/shodan/custom-scan` | `client.shodan.customScan({ target })` | 1 |
| GET | `/api/v1/shodan/custom-scans/:id` | `client.shodan.getScan(id)` | 0 |
| DELETE | `/api/v1/shodan/custom-scans/:id` | `client.shodan.deleteScan(id)` | 0 |
| GET | `/api/v1/shodan/host/:ip` | `client.shodan.host(ip)` | 0 |
| POST | `/api/v1/ulp/search` | `client.ulp.search({ type, value })` | 1 |
| POST | `/api/v1/ulp/download` | `client.ulp.download({ token })` | 10 |
| GET | `/api/v1/intelx/downloads` | `client.intelx.downloads()` | 0 |
| POST | `/api/v1/intelx/download` | `client.intelx.download({ systemId, bucket })` | 5 |
| GET | `/api/v1/intelx/downloads/:id/file` | `client.intelx.getFile(id)` | 0 |
| DELETE | `/api/v1/intelx/downloads/:id` | `client.intelx.deleteDownload(id)` | 0 |

### `POST /api/v1/search` body

| Field | Type | Notes |
|---|---|---|
| `query` | string | OSINT syntax, or Shodan syntax when `scope: "shodan"` |
| `scope` | `"leak" \| "service" \| "shodan"` | default `"leak"` |
| `plugin` | string | OSINT plugin filter (e.g. `DotEnvConfigPlugin`); ignored for shodan |
| `severity` | `"critical" \| "high" \| "medium" \| "low" \| "info" \| "all"` | |
| `page` | number | default 0 |

### ULP

- `ulp.search`: `type` ∈ `domain | email | username | password`, plus `value`.
  Returns rows + a `downloadToken` (valid ~15 min).
- `ulp.download`: `{ token }` → JSON export of already-fetched rows.

### IntelX

- `intelx.download`: `systemId` (UUID), `bucket` ∈
  `leaks.private.general | leaks.logs | whois | dns | darknet.tor`. Server-side
  cached.
- `intelx.getFile(id)`: returns the cached binary (no re-billing) →
  `ArrayBuffer`/`Blob`.

> Response shapes are not fully documented by the provider. Types model known
> request fields strictly; response bodies are typed permissively (best-effort
> interfaces with index signatures / `unknown` fallbacks) and tightened as real
> responses are observed. This is recorded as a known limitation, not a guess.

## TypeScript Client Design

### Construction

```ts
const client = new LeaksNowClient("ms_...", {
  baseUrl?: string,        // default "https://leaks.now"
  fetch?: typeof fetch,    // injectable; default globalThis.fetch
  timeoutMs?: number,      // default 30000, via AbortController
  retry?: {                // default DISABLED
    maxRetries: number,    // e.g. 3
    baseDelayMs: number,   // e.g. 500
    maxDelayMs: number,    // e.g. 8000
    retryOn: number[],     // default [429, 500, 502, 503, 504]
  },
});
```

- Runtime-agnostic: relies only on WHATWG `fetch` (Node 18+, Bun, Deno,
  browser). Zero runtime dependencies.
- Resources namespaced: `client.search`, `client.shodan.*`, `client.ulp.*`,
  `client.intelx.*`.

### Transport & retry

- Single private `request()` builds headers (Bearer + Content-Type), serializes
  JSON, applies timeout via `AbortController`.
- Retry off by default. When enabled: exponential backoff with full jitter on
  configured status codes; respects `Retry-After` header when present. Caps at
  `maxRetries`. Non-retryable errors throw immediately.

### Errors

```
LeaksNowError              # base: status, statusText, body, requestId?
├── LeaksNowAuthError      # 401 / 403
├── LeaksNowQuotaError     # 429 (quota exhausted), exposes retryAfter
├── LeaksNowValidationError# 400 / 422
└── LeaksNowServerError    # 5xx
```

Network/timeout failures → `LeaksNowError` with `code: "network" | "timeout"`.

### Binary responses

`intelx.getFile(id)` returns `{ data: ArrayBuffer, contentType: string,
filename?: string }` (parsed from `Content-Disposition` when present).

### Build & test

- Build: `tsup` → ESM + CJS + `.d.ts`.
- Test: Vitest with an injected mock `fetch`. Coverage: every method (happy
  path), error mapping (401/429/5xx), retry/backoff sequencing (fake timers),
  timeout/abort, binary parsing.

## Go Port (iteration 2)

- Module `github.com/ItsMariusBC/leaksnow-wrapper/go`, package `leaksnow`.
- `leaksnow.NewClient(apiKey string, opts ...Option) *Client`.
- Functional options: `WithBaseURL`, `WithHTTPClient`, `WithTimeout`,
  `WithRetry`.
- Sub-structs `client.Shodan`, `client.ULP`, `client.IntelX`.
- All calls take `context.Context`. Typed errors via `errors.As` (e.g.
  `*QuotaError`). Stdlib `net/http` only.
- Tests with `httptest.Server`.

## Rust Port (iteration 3)

- Crate `leaksnow`. Async via `reqwest` + `tokio`, `serde` for models.
- `Client::new(api_key)` / builder `Client::builder()` for base URL, timeout,
  retry, custom `reqwest::Client`.
- Modules `shodan`, `ulp`, `intelx`. Errors via `thiserror` (`Error::Quota`,
  `Error::Auth`, ...).
- Optional `blocking` cargo feature.
- Tests with `wiremock`.

## Shared Contract

`docs/openapi.yaml` describes all endpoints, request bodies, enums, and the
auth scheme. It is the reference each language port implements against, keeping
the three SDKs aligned. Hand-written (provider gives no machine spec).

## Testing Strategy

Each SDK is tested in isolation against a mocked HTTP layer — no live API calls
in CI (would consume credits + leak keys). A separate, opt-in, manually-run
integration smoke test (gated behind an env var holding a real key) can hit the
live API for the zero-credit endpoints only.

## Out of Scope (YAGNI)

- Publishing to npm/crates.io/Go proxy (no CI publish pipeline now).
- CLI tool.
- Response-schema exhaustiveness beyond what the provider documents.
- Pagination auto-iteration helpers (can add later if response shape supports).

## Open / Known Limitations

- Response body shapes are best-effort (provider docs cover requests only).
- `:id` types: scan ids are numeric, IntelX download ids type TBD-from-response
  → modeled as `string | number` accepted, sent as-is.
