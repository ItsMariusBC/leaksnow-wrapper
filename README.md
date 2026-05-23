# leaksnow-wrapper

Multi-language SDKs for the [leaks.now](https://leaks.now) `/api/v1` OSINT API.

| Language | Path | Status |
|---|---|---|
| TypeScript | `packages/typescript` (`@leaksnow/sdk`) | available |
| Go | `go/` (`github.com/ItsMariusBC/leaksnow-wrapper/go`) | available |
| Rust | `rust/` (`leaksnow` crate) | available |

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

**Releasing:** Go modules are consumed straight from VCS — there is no registry
upload. Pushing a `go/v*` tag triggers
[`.github/workflows/go-release.yml`](.github/workflows/go-release.yml), which
runs build + vet + race tests, creates a GitHub Release, and warms
`proxy.golang.org` so the version is indexed on pkg.go.dev:

```bash
git tag go/v0.1.0
git push origin go/v0.1.0
```

Then `go get github.com/ItsMariusBC/leaksnow-wrapper/go@v0.1.0`.

## Rust

```toml
[dependencies]
leaksnow = "0.1"
tokio = { version = "1", features = ["macros", "rt-multi-thread"] }
```

```rust
use leaksnow::{Client, SearchRequest, Scope};

#[tokio::main]
async fn main() -> Result<(), leaksnow::Error> {
    let client = Client::new(std::env::var("LEAKSNOW_API_KEY").unwrap());

    let leaks = client.search(SearchRequest {
        query: "host:example.com".into(),
        scope: Some(Scope::Leak),
        ..Default::default()
    }).await?;

    let file = client.intelx().get_file("7").await?; // BinaryFile { data, content_type, filename }
    let _ = (leaks, file);
    Ok(())
}
```

JSON responses are returned as `serde_json::Value` (provider documents request
bodies only). Errors are the `leaksnow::Error` enum (`Auth`, `Quota`,
`Validation`, `Server`, `Api`, `Transport`, `Decode`). Retries are off by
default; enable with `Client::builder(key).retry(RetryConfig { .. }).build()`.

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
