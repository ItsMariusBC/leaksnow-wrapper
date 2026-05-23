# leaksnow-wrapper

Multi-language SDKs for the [leaks.now](https://leaks.now) `/api/v1` OSINT API.

| Language | Path | Status |
|---|---|---|
| TypeScript | `packages/typescript` (`@leaksnow/sdk`) | available |
| Go | `go/` | planned |
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
