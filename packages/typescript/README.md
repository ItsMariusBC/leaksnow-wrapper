# @leaksnow/sdk

TypeScript SDK for the [leaks.now](https://leaks.now) `/api/v1` OSINT API.
Zero runtime dependencies, ESM + CJS, ships its own `.d.ts`. Node >= 18.

## Install

```bash
pnpm add @leaksnow/sdk   # or npm i / yarn add
```

## Quick start

```ts
import { LeaksNowClient } from "@leaksnow/sdk";

const client = new LeaksNowClient(process.env.LEAKSNOW_API_KEY!, {
  // all optional:
  // baseUrl: "https://leaks.now",
  // timeoutMs: 30_000,                         // per-attempt timeout
  // retry: { maxRetries: 3, retryOn: [429, 500, 502, 503, 504] },
});

const leaks = await client.search({ query: "host:example.com", scope: "leak", severity: "all" });
const scan  = await client.shodan.customScan({ target: "scanme.nmap.org" });
const ulp   = await client.ulp.search({ type: "domain", value: "example.com" });
```

## Resources

| Namespace | Methods |
|---|---|
| `client` | `search(body)` |
| `client.shodan` | `customScans()`, `customScan(body)`, `getScan(id)`, `deleteScan(id)`, `host(ip)` |
| `client.ulp` | `search(body)`, `download(body)` |
| `client.intelx` | `downloads()`, `download(body)`, `getFile(id)`, `deleteDownload(id)` |

## Binary downloads

`intelx.getFile(id)` returns a `BinaryFile`, not a parsed JSON body:

```ts
import { writeFile } from "node:fs/promises";

const file = await client.intelx.getFile(7);
// file.data       -> ArrayBuffer
// file.contentType -> e.g. "application/zip"
// file.filename    -> parsed from Content-Disposition (may be undefined)
await writeFile(file.filename ?? "download.bin", Buffer.from(file.data));
```

## Errors

Every failure throws a `LeaksNowError` (or a subclass). Inspect `.code`, `.status`,
`.body`, and `.retryAfterMs` when debugging:

| Class | Trigger |
|---|---|
| `LeaksNowAuthError` | HTTP 401 / 403 |
| `LeaksNowQuotaError` | HTTP 429 (see `.retryAfterMs`) |
| `LeaksNowValidationError` | HTTP 400 / 422 |
| `LeaksNowServerError` | HTTP 5xx |
| `LeaksNowError` (`code: "timeout"` / `"network"`) | per-attempt timeout / transport failure |

```ts
import { LeaksNowAuthError, LeaksNowError } from "@leaksnow/sdk";

try {
  await client.search({ query: "host:example.com" });
} catch (err) {
  if (err instanceof LeaksNowAuthError) {
    // bad or missing key
  } else if (err instanceof LeaksNowError) {
    console.error(err.code, err.status, err.body);
  }
}
```

## Response typing

Request bodies are fully typed. Response payloads are intentionally permissive
(`UnknownRecord` / `unknown`) because the upstream API documents request shapes
only — narrow them at the call site once you know your endpoint's contract.

## Auth & credits

Pass your `ms_*` key as the first argument. Each successful call consumes the
credits listed in the API contract (`docs/openapi.yaml` in the repo).
**Never hardcode a real key** — read it from an environment variable.

## License

MIT
