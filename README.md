# leaksnow-wrapper

Multi-language SDKs for the [leaks.now](https://leaks.now) `/api/v1` OSINT API.

| Language | Path | Status |
|---|---|---|
| TypeScript | `packages/typescript` (`@leaksnow/sdk`) | available |
| Go | `go/` (`github.com/ItsMariusBC/leaksnow-wrapper/go`) | available |
| Rust | `rust/` | planned |

The API contract lives in [`docs/openapi.yaml`](docs/openapi.yaml).

## TypeScript

```bash
pnpm add @leaksnow/sdk   # once published
```

```ts
import { LeaksNowClient } from "@leaksnow/sdk";

const client = new LeaksNowClient(process.env.LEAKSNOW_API_KEY!, {
  // baseUrl, timeoutMs, retry are all optional
  retry: { maxRetries: 3, baseDelayMs: 500, maxDelayMs: 8000, retryOn: [429, 500, 502, 503, 504] },
});

const leaks = await client.search({ query: "host:example.com", scope: "leak", severity: "all" });
const scan = await client.shodan.customScan({ target: "scanme.nmap.org" });
const ulp = await client.ulp.search({ type: "domain", value: "example.com" });
const file = await client.intelx.getFile(7); // { data, contentType, filename }
```

### Errors

All failures throw a `LeaksNowError` (or subclass: `LeaksNowAuthError` 401/403,
`LeaksNowQuotaError` 429, `LeaksNowValidationError` 400/422, `LeaksNowServerError` 5xx).
Timeouts/network failures throw `LeaksNowError` with `code` `"timeout"`/`"network"`.

### Auth & credits

Pass your `ms_*` key. Each successful call consumes the credits listed in `docs/openapi.yaml`.
**Never commit a real key.** Use an env var.

## Development

```bash
pnpm install
pnpm test        # all packages
pnpm build
pnpm typecheck
```

## Go

```bash
go get github.com/ItsMariusBC/leaksnow-wrapper/go@latest
```

```go
import (
    "context"
    "os"

    leaksnow "github.com/ItsMariusBC/leaksnow-wrapper/go"
)

c := leaksnow.NewClient(os.Getenv("LEAKSNOW_API_KEY"))
ctx := context.Background()

raw, err := c.Search(ctx, leaksnow.SearchRequest{Query: "host:example.com", Scope: leaksnow.ScopeLeak})
scan, err := c.Shodan.CustomScan(ctx, leaksnow.ShodanCustomScanRequest{Target: "scanme.nmap.org"})
ulp, err := c.ULP.Search(ctx, leaksnow.ULPSearchRequest{Type: leaksnow.ULPTypeDomain, Value: "example.com"})
file, err := c.IntelX.GetFile(ctx, "7") // *BinaryFile{ Data, ContentType, Filename }
```

JSON responses are returned as `json.RawMessage` (the provider documents request
bodies only); decode into your own structs. Errors are typed — use `errors.As`
with `*leaksnow.AuthError`, `*leaksnow.QuotaError`, `*leaksnow.ValidationError`,
`*leaksnow.ServerError`, or `*leaksnow.TransportError` (network/timeout).

**Releasing:** Go modules are consumed straight from VCS. Tag the submodule with
the `go/` prefix, e.g. `git tag go/v0.1.0 && git push origin go/v0.1.0`, then
`go get github.com/ItsMariusBC/leaksnow-wrapper/go@go/v0.1.0`.

## Releasing `@leaksnow/sdk`

Publishing is automated by [`.github/workflows/publish.yml`](.github/workflows/publish.yml),
triggered on a `v*` git tag.

**One-time setup:** add an npm automation token as the repo secret `NPM_TOKEN`
(Settings → Secrets and variables → Actions). The token's npm account must own
the `@leaksnow` scope.

**To publish a version:**

```bash
# 1. bump packages/typescript/package.json "version" (e.g. 0.2.0) and commit
# 2. tag with the SAME version, prefixed with v
git tag v0.2.0
git push origin v0.2.0
```

The workflow checks the tag matches `package.json` version, runs tests + build,
then `npm publish --provenance --access public` (verified provenance badge on npm).
