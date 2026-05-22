# TypeScript SDK (`@leaksnow/sdk`) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. Every implementation task is TDD: failing test first, then code.

**Goal:** Ship `@leaksnow/sdk`, a typed, runtime-agnostic TypeScript client for the leaks.now `/api/v1` API, fully unit-tested against a mocked `fetch`.

**Architecture:** A `Transport` core handles auth headers, JSON (de)serialization, timeouts via `AbortController`, error mapping, and optional retry/backoff. Four resource classes (`search` lives on the client directly; `shodan`, `ulp`, `intelx` are namespaced) call `Transport`. Zero runtime dependencies — only WHATWG `fetch`.

**Tech Stack:** TypeScript, tsup (ESM+CJS+d.ts), Vitest (mocked fetch), pnpm workspaces, Turborepo.

**Scope note:** This plan covers the TypeScript package only. The Go and Rust ports get their own plans in later iterations (each implements the same `docs/openapi.yaml` contract). The OpenAPI doc is authored in this plan (Task 11) so all three stay aligned.

---

## File Structure

```
pnpm-workspace.yaml            # workspace globs
package.json                   # root: scripts delegate to turbo
turbo.json                     # build/test/typecheck pipeline
packages/typescript/
  package.json                 # @leaksnow/sdk
  tsconfig.json
  tsup.config.ts
  vitest.config.ts
  src/
    index.ts                   # public exports
    errors.ts                  # LeaksNowError + subclasses + fromResponse()
    retry.ts                   # RetryOptions, backoffDelay(), shouldRetry()
    types.ts                   # enums, request bodies, response interfaces
    client.ts                  # Transport + LeaksNowClient
    resources/
      shodan.ts                # ShodanResource
      ulp.ts                   # UlpResource
      intelx.ts                # IntelxResource
  test/
    errors.test.ts
    retry.test.ts
    transport.test.ts
    search.test.ts
    shodan.test.ts
    ulp.test.ts
    intelx.test.ts
docs/openapi.yaml              # shared contract
README.md
```

Each file has one responsibility: `errors` = error taxonomy, `retry` = backoff math + retry predicate, `types` = data shapes, `client` = transport + assembly, one file per resource namespace.

---

## Task 1: Monorepo + package scaffold

**Files:**
- Create: `pnpm-workspace.yaml`, `package.json`, `turbo.json`
- Create: `packages/typescript/package.json`, `packages/typescript/tsconfig.json`, `packages/typescript/tsup.config.ts`, `packages/typescript/vitest.config.ts`
- Create: `packages/typescript/src/index.ts` (temporary stub)

- [ ] **Step 1: Root workspace files**

`pnpm-workspace.yaml`:
```yaml
packages:
  - "packages/*"
```

`package.json`:
```json
{
  "name": "leaksnow-wrapper",
  "private": true,
  "version": "0.0.0",
  "packageManager": "pnpm@9.12.0",
  "scripts": {
    "build": "turbo run build",
    "test": "turbo run test",
    "typecheck": "turbo run typecheck"
  },
  "devDependencies": {
    "turbo": "^2.1.0"
  }
}
```

`turbo.json`:
```json
{
  "$schema": "https://turbo.build/schema.json",
  "tasks": {
    "build": { "dependsOn": ["^build"], "outputs": ["dist/**"] },
    "test": { "dependsOn": ["^build"] },
    "typecheck": { "dependsOn": ["^build"] }
  }
}
```

- [ ] **Step 2: Package manifest**

`packages/typescript/package.json`:
```json
{
  "name": "@leaksnow/sdk",
  "version": "0.1.0",
  "description": "TypeScript SDK for the leaks.now /api/v1 OSINT API",
  "type": "module",
  "main": "./dist/index.cjs",
  "module": "./dist/index.js",
  "types": "./dist/index.d.ts",
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.js",
      "require": "./dist/index.cjs"
    }
  },
  "files": ["dist"],
  "scripts": {
    "build": "tsup",
    "test": "vitest run",
    "test:watch": "vitest",
    "typecheck": "tsc --noEmit"
  },
  "engines": { "node": ">=18" },
  "devDependencies": {
    "tsup": "^8.3.0",
    "typescript": "^5.6.0",
    "vitest": "^2.1.0"
  }
}
```

- [ ] **Step 3: tsconfig, tsup, vitest configs**

`packages/typescript/tsconfig.json`:
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "lib": ["ES2022", "DOM"],
    "strict": true,
    "declaration": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "verbatimModuleSyntax": true,
    "noUncheckedIndexedAccess": true,
    "outDir": "dist",
    "rootDir": "src"
  },
  "include": ["src"]
}
```

`packages/typescript/tsup.config.ts`:
```ts
import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.ts"],
  format: ["esm", "cjs"],
  dts: true,
  clean: true,
  sourcemap: true,
});
```

`packages/typescript/vitest.config.ts`:
```ts
import { defineConfig } from "vitest/config";

export default defineConfig({
  test: { include: ["test/**/*.test.ts"], environment: "node" },
});
```

- [ ] **Step 4: Temporary index stub**

`packages/typescript/src/index.ts`:
```ts
export const VERSION = "0.1.0";
```

- [ ] **Step 5: Install and verify**

Run: `pnpm install && pnpm --filter @leaksnow/sdk typecheck`
Expected: install succeeds, `tsc --noEmit` exits 0.

- [ ] **Step 6: Commit**

```bash
git add pnpm-workspace.yaml package.json turbo.json packages/typescript pnpm-lock.yaml
git commit -m "chore: scaffold pnpm+turborepo monorepo and @leaksnow/sdk package"
```

---

## Task 2: Error taxonomy

**Files:**
- Create: `packages/typescript/src/errors.ts`
- Test: `packages/typescript/test/errors.test.ts`

- [ ] **Step 1: Write the failing test**

`packages/typescript/test/errors.test.ts`:
```ts
import { describe, it, expect } from "vitest";
import {
  LeaksNowError,
  LeaksNowAuthError,
  LeaksNowQuotaError,
  LeaksNowValidationError,
  LeaksNowServerError,
  errorFromResponse,
} from "../src/errors.js";

describe("errorFromResponse", () => {
  it("maps 401/403 to LeaksNowAuthError", () => {
    expect(errorFromResponse(401, "Unauthorized", { m: 1 })).toBeInstanceOf(LeaksNowAuthError);
    expect(errorFromResponse(403, "Forbidden", null)).toBeInstanceOf(LeaksNowAuthError);
  });
  it("maps 429 to LeaksNowQuotaError and keeps retryAfterMs", () => {
    const err = errorFromResponse(429, "Too Many Requests", null, 2000);
    expect(err).toBeInstanceOf(LeaksNowQuotaError);
    expect(err.retryAfterMs).toBe(2000);
  });
  it("maps 400/422 to LeaksNowValidationError", () => {
    expect(errorFromResponse(400, "Bad Request", null)).toBeInstanceOf(LeaksNowValidationError);
    expect(errorFromResponse(422, "Unprocessable", null)).toBeInstanceOf(LeaksNowValidationError);
  });
  it("maps 5xx to LeaksNowServerError", () => {
    expect(errorFromResponse(500, "Server Error", null)).toBeInstanceOf(LeaksNowServerError);
    expect(errorFromResponse(503, "Unavailable", null)).toBeInstanceOf(LeaksNowServerError);
  });
  it("falls back to base LeaksNowError for other 4xx", () => {
    const err = errorFromResponse(418, "Teapot", null);
    expect(err).toBeInstanceOf(LeaksNowError);
    expect(err.constructor).toBe(LeaksNowError);
    expect(err.status).toBe(418);
  });
  it("base error carries code and body", () => {
    const err = new LeaksNowError("boom", { code: "timeout" });
    expect(err.code).toBe("timeout");
    expect(err.name).toBe("LeaksNowError");
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --filter @leaksnow/sdk exec vitest run test/errors.test.ts`
Expected: FAIL — cannot resolve `../src/errors.js`.

- [ ] **Step 3: Write minimal implementation**

`packages/typescript/src/errors.ts`:
```ts
export type LeaksNowErrorCode = "network" | "timeout" | "http";

export interface LeaksNowErrorOptions {
  status?: number;
  statusText?: string;
  body?: unknown;
  code?: LeaksNowErrorCode;
  retryAfterMs?: number;
}

export class LeaksNowError extends Error {
  readonly status?: number;
  readonly statusText?: string;
  readonly body?: unknown;
  readonly code: LeaksNowErrorCode;
  readonly retryAfterMs?: number;

  constructor(message: string, opts: LeaksNowErrorOptions = {}) {
    super(message);
    this.name = "LeaksNowError";
    this.status = opts.status;
    this.statusText = opts.statusText;
    this.body = opts.body;
    this.code = opts.code ?? "http";
    this.retryAfterMs = opts.retryAfterMs;
  }
}

export class LeaksNowAuthError extends LeaksNowError {
  constructor(message: string, opts: LeaksNowErrorOptions = {}) {
    super(message, opts);
    this.name = "LeaksNowAuthError";
  }
}
export class LeaksNowQuotaError extends LeaksNowError {
  constructor(message: string, opts: LeaksNowErrorOptions = {}) {
    super(message, opts);
    this.name = "LeaksNowQuotaError";
  }
}
export class LeaksNowValidationError extends LeaksNowError {
  constructor(message: string, opts: LeaksNowErrorOptions = {}) {
    super(message, opts);
    this.name = "LeaksNowValidationError";
  }
}
export class LeaksNowServerError extends LeaksNowError {
  constructor(message: string, opts: LeaksNowErrorOptions = {}) {
    super(message, opts);
    this.name = "LeaksNowServerError";
  }
}

export function errorFromResponse(
  status: number,
  statusText: string,
  body: unknown,
  retryAfterMs?: number,
): LeaksNowError {
  const opts: LeaksNowErrorOptions = { status, statusText, body, code: "http", retryAfterMs };
  const message = `leaks.now request failed: ${status} ${statusText}`;
  if (status === 401 || status === 403) return new LeaksNowAuthError(message, opts);
  if (status === 429) return new LeaksNowQuotaError(message, opts);
  if (status === 400 || status === 422) return new LeaksNowValidationError(message, opts);
  if (status >= 500) return new LeaksNowServerError(message, opts);
  return new LeaksNowError(message, opts);
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --filter @leaksnow/sdk exec vitest run test/errors.test.ts`
Expected: PASS (6 tests).

- [ ] **Step 5: Commit**

```bash
git add packages/typescript/src/errors.ts packages/typescript/test/errors.test.ts
git commit -m "feat(sdk): add error taxonomy and status-to-error mapping"
```

---

## Task 3: Types & enums

**Files:**
- Create: `packages/typescript/src/types.ts`
- Test: `packages/typescript/test/types.test-d.ts` is not needed; verify via `typecheck`. Add a runtime sanity test instead.
- Test: `packages/typescript/test/types.test.ts`

- [ ] **Step 1: Write the failing test**

`packages/typescript/test/types.test.ts`:
```ts
import { describe, it, expect } from "vitest";
import { SEVERITIES, SCOPES, ULP_TYPES, INTELX_BUCKETS } from "../src/types.js";

describe("enum value lists", () => {
  it("exposes scope values", () => {
    expect(SCOPES).toEqual(["leak", "service", "shodan"]);
  });
  it("exposes severity values including all", () => {
    expect(SEVERITIES).toEqual(["critical", "high", "medium", "low", "info", "all"]);
  });
  it("exposes ulp types", () => {
    expect(ULP_TYPES).toEqual(["domain", "email", "username", "password"]);
  });
  it("exposes intelx buckets", () => {
    expect(INTELX_BUCKETS).toEqual([
      "leaks.private.general",
      "leaks.logs",
      "whois",
      "dns",
      "darknet.tor",
    ]);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --filter @leaksnow/sdk exec vitest run test/types.test.ts`
Expected: FAIL — cannot resolve `../src/types.js`.

- [ ] **Step 3: Write minimal implementation**

`packages/typescript/src/types.ts`:
```ts
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --filter @leaksnow/sdk exec vitest run test/types.test.ts`
Expected: PASS (4 tests).

- [ ] **Step 5: Commit**

```bash
git add packages/typescript/src/types.ts packages/typescript/test/types.test.ts
git commit -m "feat(sdk): add request/response types and enum constants"
```

---

## Task 4: Retry math & predicate

**Files:**
- Create: `packages/typescript/src/retry.ts`
- Test: `packages/typescript/test/retry.test.ts`

- [ ] **Step 1: Write the failing test**

`packages/typescript/test/retry.test.ts`:
```ts
import { describe, it, expect, vi, afterEach } from "vitest";
import { DEFAULT_RETRY, backoffDelay, shouldRetry, type RetryOptions } from "../src/retry.js";

const opts: RetryOptions = { maxRetries: 3, baseDelayMs: 500, maxDelayMs: 8000, retryOn: [429, 500, 502, 503, 504] };

afterEach(() => vi.restoreAllMocks());

describe("backoffDelay", () => {
  it("uses retryAfterMs when provided, capped at maxDelayMs", () => {
    expect(backoffDelay(0, opts, 1000)).toBe(1000);
    expect(backoffDelay(0, opts, 999999)).toBe(8000);
  });
  it("applies full jitter on exponential base (Math.random mocked)", () => {
    vi.spyOn(Math, "random").mockReturnValue(0.5);
    // attempt 0 -> exp = min(500*1, 8000)=500 -> 0.5*500=250
    expect(backoffDelay(0, opts)).toBe(250);
    // attempt 2 -> exp = min(500*4, 8000)=2000 -> 0.5*2000=1000
    expect(backoffDelay(2, opts)).toBe(1000);
  });
  it("never exceeds maxDelayMs", () => {
    vi.spyOn(Math, "random").mockReturnValue(1);
    expect(backoffDelay(10, opts)).toBeLessThanOrEqual(8000);
  });
});

describe("shouldRetry", () => {
  it("retries listed statuses while attempts remain", () => {
    expect(shouldRetry(429, 0, opts)).toBe(true);
    expect(shouldRetry(503, 2, opts)).toBe(true);
  });
  it("does not retry once attempts are exhausted", () => {
    expect(shouldRetry(429, 3, opts)).toBe(false);
  });
  it("does not retry unlisted statuses", () => {
    expect(shouldRetry(400, 0, opts)).toBe(false);
  });
});

describe("DEFAULT_RETRY", () => {
  it("defaults to zero retries (disabled)", () => {
    expect(DEFAULT_RETRY.maxRetries).toBe(0);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --filter @leaksnow/sdk exec vitest run test/retry.test.ts`
Expected: FAIL — cannot resolve `../src/retry.js`.

- [ ] **Step 3: Write minimal implementation**

`packages/typescript/src/retry.ts`:
```ts
export interface RetryOptions {
  maxRetries: number;
  baseDelayMs: number;
  maxDelayMs: number;
  retryOn: number[];
}

export const DEFAULT_RETRY: RetryOptions = {
  maxRetries: 0,
  baseDelayMs: 500,
  maxDelayMs: 8000,
  retryOn: [429, 500, 502, 503, 504],
};

/** Full-jitter exponential backoff. Honors server retryAfterMs (capped) when present. */
export function backoffDelay(attempt: number, opts: RetryOptions, retryAfterMs?: number): number {
  if (retryAfterMs != null) return Math.min(retryAfterMs, opts.maxDelayMs);
  const exp = Math.min(opts.baseDelayMs * 2 ** attempt, opts.maxDelayMs);
  return Math.random() * exp;
}

export function shouldRetry(status: number, attempt: number, opts: RetryOptions): boolean {
  if (attempt >= opts.maxRetries) return false;
  return opts.retryOn.includes(status);
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --filter @leaksnow/sdk exec vitest run test/retry.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add packages/typescript/src/retry.ts packages/typescript/test/retry.test.ts
git commit -m "feat(sdk): add backoff math and retry predicate"
```

---

## Task 5: Transport core

**Files:**
- Create: `packages/typescript/src/client.ts` (Transport portion; LeaksNowClient added in Task 9)
- Test: `packages/typescript/test/transport.test.ts`

- [ ] **Step 1: Write the failing test**

`packages/typescript/test/transport.test.ts`:
```ts
import { describe, it, expect, vi, afterEach } from "vitest";
import { Transport } from "../src/client.js";
import { LeaksNowAuthError, LeaksNowQuotaError, LeaksNowError } from "../src/errors.js";

function jsonResponse(body: unknown, init: ResponseInit = {}) {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { "content-type": "application/json" },
    ...init,
  });
}

afterEach(() => vi.restoreAllMocks());

describe("Transport.request", () => {
  it("sends Bearer auth + JSON content-type and parses JSON", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ ok: true }));
    const t = new Transport("ms_key", { fetch: fetchMock });
    const out = await t.request("POST", "/api/v1/search", { body: { query: "x" } });

    expect(out).toEqual({ ok: true });
    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toBe("https://leaks.now/api/v1/search");
    expect(init.method).toBe("POST");
    expect(init.headers.Authorization).toBe("Bearer ms_key");
    expect(init.headers["Content-Type"]).toBe("application/json");
    expect(JSON.parse(init.body)).toEqual({ query: "x" });
  });

  it("omits body for GET", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse([]));
    const t = new Transport("ms_key", { fetch: fetchMock });
    await t.request("GET", "/api/v1/shodan/custom-scans");
    expect(fetchMock.mock.calls[0][1].body).toBeUndefined();
  });

  it("maps 401 to LeaksNowAuthError with parsed body", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ error: "bad key" }, { status: 401, statusText: "Unauthorized" }));
    const t = new Transport("ms_key", { fetch: fetchMock });
    await expect(t.request("POST", "/api/v1/search", { body: {} })).rejects.toBeInstanceOf(LeaksNowAuthError);
  });

  it("retries on 429 then succeeds when retry enabled", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse({}, { status: 429, statusText: "Too Many Requests", headers: { "retry-after": "0" } }))
      .mockResolvedValueOnce(jsonResponse({ ok: 1 }));
    const t = new Transport("ms_key", { fetch: fetchMock, retry: { maxRetries: 2, baseDelayMs: 1, maxDelayMs: 5, retryOn: [429] } });
    const out = await t.request("POST", "/api/v1/search", { body: {} });
    expect(out).toEqual({ ok: 1 });
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("throws LeaksNowQuotaError after retries exhausted", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({}, { status: 429, statusText: "Too Many Requests", headers: { "retry-after": "0" } }));
    const t = new Transport("ms_key", { fetch: fetchMock, retry: { maxRetries: 1, baseDelayMs: 1, maxDelayMs: 5, retryOn: [429] } });
    await expect(t.request("POST", "/api/v1/search", { body: {} })).rejects.toBeInstanceOf(LeaksNowQuotaError);
    expect(fetchMock).toHaveBeenCalledTimes(2); // initial + 1 retry
  });

  it("wraps abort/timeout as LeaksNowError code=timeout", async () => {
    const fetchMock = vi.fn().mockImplementation((_url, init: RequestInit) => {
      return new Promise((_resolve, reject) => {
        (init.signal as AbortSignal).addEventListener("abort", () => {
          reject(new DOMException("Aborted", "AbortError"));
        });
      });
    });
    const t = new Transport("ms_key", { fetch: fetchMock, timeoutMs: 5 });
    const err = (await t.request("GET", "/api/v1/shodan/custom-scans").catch((e) => e)) as LeaksNowError;
    expect(err).toBeInstanceOf(LeaksNowError);
    expect(err.code).toBe("timeout");
  });

  it("requestRaw returns the Response untouched", async () => {
    const res = jsonResponse({ bin: true });
    const fetchMock = vi.fn().mockResolvedValue(res);
    const t = new Transport("ms_key", { fetch: fetchMock });
    const out = await t.requestRaw("GET", "/api/v1/intelx/downloads/1/file");
    expect(out).toBe(res);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --filter @leaksnow/sdk exec vitest run test/transport.test.ts`
Expected: FAIL — `Transport` not exported from `../src/client.js`.

- [ ] **Step 3: Write minimal implementation**

`packages/typescript/src/client.ts`:
```ts
import { errorFromResponse, LeaksNowError } from "./errors.js";
import { DEFAULT_RETRY, backoffDelay, shouldRetry, type RetryOptions } from "./retry.js";

export interface ClientConfig {
  baseUrl?: string;
  fetch?: typeof fetch;
  timeoutMs?: number;
  retry?: Partial<RetryOptions>;
}

export interface RequestOptions {
  body?: unknown;
}

const DEFAULT_BASE_URL = "https://leaks.now";
const DEFAULT_TIMEOUT_MS = 30_000;

function parseRetryAfter(res: Response): number | undefined {
  const raw = res.headers.get("retry-after");
  if (raw == null) return undefined;
  const seconds = Number(raw);
  return Number.isFinite(seconds) ? seconds * 1000 : undefined;
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export class Transport {
  readonly #apiKey: string;
  readonly #baseUrl: string;
  readonly #fetch: typeof fetch;
  readonly #timeoutMs: number;
  readonly #retry: RetryOptions;

  constructor(apiKey: string, config: ClientConfig = {}) {
    if (!apiKey) throw new LeaksNowError("apiKey is required", { code: "http" });
    this.#apiKey = apiKey;
    this.#baseUrl = (config.baseUrl ?? DEFAULT_BASE_URL).replace(/\/+$/, "");
    const f = config.fetch ?? globalThis.fetch;
    if (!f) throw new LeaksNowError("no fetch implementation available", { code: "network" });
    this.#fetch = f;
    this.#timeoutMs = config.timeoutMs ?? DEFAULT_TIMEOUT_MS;
    this.#retry = { ...DEFAULT_RETRY, ...config.retry };
  }

  async requestRaw(method: string, path: string, opts: RequestOptions = {}): Promise<Response> {
    const url = `${this.#baseUrl}${path}`;
    const headers: Record<string, string> = { Authorization: `Bearer ${this.#apiKey}` };
    let body: string | undefined;
    if (opts.body !== undefined && method !== "GET" && method !== "DELETE") {
      headers["Content-Type"] = "application/json";
      body = JSON.stringify(opts.body);
    }

    let attempt = 0;
    for (;;) {
      const controller = new AbortController();
      const timer = setTimeout(() => controller.abort(), this.#timeoutMs);
      let res: Response;
      try {
        res = await this.#fetch(url, { method, headers, body, signal: controller.signal });
      } catch (cause) {
        clearTimeout(timer);
        const isAbort = cause instanceof Error && cause.name === "AbortError";
        throw new LeaksNowError(isAbort ? "request timed out" : "network request failed", {
          code: isAbort ? "timeout" : "network",
        });
      } finally {
        clearTimeout(timer);
      }

      if (res.ok) return res;

      const retryAfterMs = parseRetryAfter(res);
      if (shouldRetry(res.status, attempt, this.#retry)) {
        await sleep(backoffDelay(attempt, this.#retry, retryAfterMs));
        attempt += 1;
        continue;
      }
      const errBody = await safeJson(res);
      throw errorFromResponse(res.status, res.statusText, errBody, retryAfterMs);
    }
  }

  async request<T>(method: string, path: string, opts: RequestOptions = {}): Promise<T> {
    const res = await this.requestRaw(method, path, opts);
    return (await safeJson(res)) as T;
  }
}

async function safeJson(res: Response): Promise<unknown> {
  const text = await res.text();
  if (!text) return null;
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --filter @leaksnow/sdk exec vitest run test/transport.test.ts`
Expected: PASS (7 tests).

- [ ] **Step 5: Commit**

```bash
git add packages/typescript/src/client.ts packages/typescript/test/transport.test.ts
git commit -m "feat(sdk): add Transport with auth, timeout, retry and error mapping"
```

---

## Task 6: `search` method on client

**Files:**
- Modify: `packages/typescript/src/client.ts` (add `LeaksNowClient` with `search`)
- Test: `packages/typescript/test/search.test.ts`

- [ ] **Step 1: Write the failing test**

`packages/typescript/test/search.test.ts`:
```ts
import { describe, it, expect, vi } from "vitest";
import { LeaksNowClient } from "../src/client.js";

function jsonResponse(body: unknown) {
  return new Response(JSON.stringify(body), { status: 200, headers: { "content-type": "application/json" } });
}

describe("client.search", () => {
  it("POSTs /api/v1/search with the given body", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ results: [] }));
    const client = new LeaksNowClient("ms_key", { fetch: fetchMock });
    const out = await client.search({ query: "host:example.com", scope: "leak", severity: "all", page: 0 });

    expect(out).toEqual({ results: [] });
    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toBe("https://leaks.now/api/v1/search");
    expect(init.method).toBe("POST");
    expect(JSON.parse(init.body)).toEqual({ query: "host:example.com", scope: "leak", severity: "all", page: 0 });
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --filter @leaksnow/sdk exec vitest run test/search.test.ts`
Expected: FAIL — `LeaksNowClient` not exported.

- [ ] **Step 3: Write minimal implementation**

Append to `packages/typescript/src/client.ts`:
```ts
import type { SearchRequest, SearchResponse } from "./types.js";

export class LeaksNowClient {
  protected readonly transport: Transport;

  constructor(apiKey: string, config: ClientConfig = {}) {
    this.transport = new Transport(apiKey, config);
  }

  search(body: SearchRequest): Promise<SearchResponse> {
    return this.transport.request<SearchResponse>("POST", "/api/v1/search", { body });
  }
}
```
> Note: place the `import type` line at the top of the file with the other imports, not mid-file.

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --filter @leaksnow/sdk exec vitest run test/search.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add packages/typescript/src/client.ts packages/typescript/test/search.test.ts
git commit -m "feat(sdk): add LeaksNowClient.search"
```

---

## Task 7: Shodan resource

**Files:**
- Create: `packages/typescript/src/resources/shodan.ts`
- Modify: `packages/typescript/src/client.ts` (wire `client.shodan`)
- Test: `packages/typescript/test/shodan.test.ts`

- [ ] **Step 1: Write the failing test**

`packages/typescript/test/shodan.test.ts`:
```ts
import { describe, it, expect, vi } from "vitest";
import { LeaksNowClient } from "../src/client.js";

function jsonResponse(body: unknown) {
  return new Response(JSON.stringify(body), { status: 200, headers: { "content-type": "application/json" } });
}

function makeClient() {
  const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ ok: true }));
  const client = new LeaksNowClient("ms_key", { fetch: fetchMock });
  return { client, fetchMock };
}

describe("client.shodan", () => {
  it("customScans GETs the list", async () => {
    const { client, fetchMock } = makeClient();
    await client.shodan.customScans();
    expect(fetchMock.mock.calls[0][0]).toBe("https://leaks.now/api/v1/shodan/custom-scans");
    expect(fetchMock.mock.calls[0][1].method).toBe("GET");
  });

  it("customScan POSTs the target", async () => {
    const { client, fetchMock } = makeClient();
    await client.shodan.customScan({ target: "scanme.nmap.org" });
    expect(fetchMock.mock.calls[0][0]).toBe("https://leaks.now/api/v1/shodan/custom-scan");
    expect(fetchMock.mock.calls[0][1].method).toBe("POST");
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({ target: "scanme.nmap.org" });
  });

  it("getScan GETs by id", async () => {
    const { client, fetchMock } = makeClient();
    await client.shodan.getScan(42);
    expect(fetchMock.mock.calls[0][0]).toBe("https://leaks.now/api/v1/shodan/custom-scans/42");
    expect(fetchMock.mock.calls[0][1].method).toBe("GET");
  });

  it("deleteScan DELETEs by id", async () => {
    const { client, fetchMock } = makeClient();
    await client.shodan.deleteScan(42);
    expect(fetchMock.mock.calls[0][0]).toBe("https://leaks.now/api/v1/shodan/custom-scans/42");
    expect(fetchMock.mock.calls[0][1].method).toBe("DELETE");
  });

  it("host GETs by ip", async () => {
    const { client, fetchMock } = makeClient();
    await client.shodan.host("1.2.3.4");
    expect(fetchMock.mock.calls[0][0]).toBe("https://leaks.now/api/v1/shodan/host/1.2.3.4");
    expect(fetchMock.mock.calls[0][1].method).toBe("GET");
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --filter @leaksnow/sdk exec vitest run test/shodan.test.ts`
Expected: FAIL — `client.shodan` undefined.

- [ ] **Step 3: Write minimal implementation**

`packages/typescript/src/resources/shodan.ts`:
```ts
import type { Transport } from "../client.js";
import type {
  ShodanCustomScanRequest,
  ShodanScan,
  ShodanScanList,
  ShodanHost,
  ResourceId,
} from "../types.js";

export class ShodanResource {
  readonly #transport: Transport;
  constructor(transport: Transport) {
    this.#transport = transport;
  }

  customScans(): Promise<ShodanScanList> {
    return this.#transport.request<ShodanScanList>("GET", "/api/v1/shodan/custom-scans");
  }

  customScan(body: ShodanCustomScanRequest): Promise<ShodanScan> {
    return this.#transport.request<ShodanScan>("POST", "/api/v1/shodan/custom-scan", { body });
  }

  getScan(id: ResourceId): Promise<ShodanScan> {
    return this.#transport.request<ShodanScan>("GET", `/api/v1/shodan/custom-scans/${id}`);
  }

  deleteScan(id: ResourceId): Promise<unknown> {
    return this.#transport.request<unknown>("DELETE", `/api/v1/shodan/custom-scans/${id}`);
  }

  host(ip: string): Promise<ShodanHost> {
    return this.#transport.request<ShodanHost>("GET", `/api/v1/shodan/host/${ip}`);
  }
}
```

In `client.ts`, import and wire it. Add import at top:
```ts
import { ShodanResource } from "./resources/shodan.js";
```
Add a public readonly field and assign in the constructor of `LeaksNowClient`:
```ts
export class LeaksNowClient {
  protected readonly transport: Transport;
  readonly shodan: ShodanResource;

  constructor(apiKey: string, config: ClientConfig = {}) {
    this.transport = new Transport(apiKey, config);
    this.shodan = new ShodanResource(this.transport);
  }

  search(body: SearchRequest): Promise<SearchResponse> {
    return this.transport.request<SearchResponse>("POST", "/api/v1/search", { body });
  }
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --filter @leaksnow/sdk exec vitest run test/shodan.test.ts`
Expected: PASS (5 tests).

- [ ] **Step 5: Commit**

```bash
git add packages/typescript/src/resources/shodan.ts packages/typescript/src/client.ts packages/typescript/test/shodan.test.ts
git commit -m "feat(sdk): add shodan resource (scans, host)"
```

---

## Task 8: ULP resource

**Files:**
- Create: `packages/typescript/src/resources/ulp.ts`
- Modify: `packages/typescript/src/client.ts` (wire `client.ulp`)
- Test: `packages/typescript/test/ulp.test.ts`

- [ ] **Step 1: Write the failing test**

`packages/typescript/test/ulp.test.ts`:
```ts
import { describe, it, expect, vi } from "vitest";
import { LeaksNowClient } from "../src/client.js";

function jsonResponse(body: unknown) {
  return new Response(JSON.stringify(body), { status: 200, headers: { "content-type": "application/json" } });
}

describe("client.ulp", () => {
  it("search POSTs type+value and returns downloadToken", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ downloadToken: "tok123", rows: [] }));
    const client = new LeaksNowClient("ms_key", { fetch: fetchMock });
    const out = await client.ulp.search({ type: "domain", value: "example.com" });

    expect(out.downloadToken).toBe("tok123");
    expect(fetchMock.mock.calls[0][0]).toBe("https://leaks.now/api/v1/ulp/search");
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({ type: "domain", value: "example.com" });
  });

  it("download POSTs the token", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ lines: [] }));
    const client = new LeaksNowClient("ms_key", { fetch: fetchMock });
    await client.ulp.download({ token: "tok123" });
    expect(fetchMock.mock.calls[0][0]).toBe("https://leaks.now/api/v1/ulp/download");
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({ token: "tok123" });
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --filter @leaksnow/sdk exec vitest run test/ulp.test.ts`
Expected: FAIL — `client.ulp` undefined.

- [ ] **Step 3: Write minimal implementation**

`packages/typescript/src/resources/ulp.ts`:
```ts
import type { Transport } from "../client.js";
import type {
  UlpSearchRequest,
  UlpSearchResponse,
  UlpDownloadRequest,
  UlpDownloadResponse,
} from "../types.js";

export class UlpResource {
  readonly #transport: Transport;
  constructor(transport: Transport) {
    this.#transport = transport;
  }

  search(body: UlpSearchRequest): Promise<UlpSearchResponse> {
    return this.#transport.request<UlpSearchResponse>("POST", "/api/v1/ulp/search", { body });
  }

  download(body: UlpDownloadRequest): Promise<UlpDownloadResponse> {
    return this.#transport.request<UlpDownloadResponse>("POST", "/api/v1/ulp/download", { body });
  }
}
```

In `client.ts`: add `import { UlpResource } from "./resources/ulp.js";` at top, add `readonly ulp: UlpResource;` field, and in the constructor add `this.ulp = new UlpResource(this.transport);`.

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --filter @leaksnow/sdk exec vitest run test/ulp.test.ts`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
git add packages/typescript/src/resources/ulp.ts packages/typescript/src/client.ts packages/typescript/test/ulp.test.ts
git commit -m "feat(sdk): add ulp resource (search, download)"
```

---

## Task 9: IntelX resource (incl. binary file)

**Files:**
- Create: `packages/typescript/src/resources/intelx.ts`
- Modify: `packages/typescript/src/client.ts` (wire `client.intelx`)
- Test: `packages/typescript/test/intelx.test.ts`

- [ ] **Step 1: Write the failing test**

`packages/typescript/test/intelx.test.ts`:
```ts
import { describe, it, expect, vi } from "vitest";
import { LeaksNowClient } from "../src/client.js";

function jsonResponse(body: unknown) {
  return new Response(JSON.stringify(body), { status: 200, headers: { "content-type": "application/json" } });
}

describe("client.intelx", () => {
  it("downloads lists history", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ items: [] }));
    const client = new LeaksNowClient("ms_key", { fetch: fetchMock });
    await client.intelx.downloads();
    expect(fetchMock.mock.calls[0][0]).toBe("https://leaks.now/api/v1/intelx/downloads");
    expect(fetchMock.mock.calls[0][1].method).toBe("GET");
  });

  it("download POSTs systemId+bucket", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ id: "1" }));
    const client = new LeaksNowClient("ms_key", { fetch: fetchMock });
    await client.intelx.download({ systemId: "52c30047-3a62-41c8-bb1b-769520b90957", bucket: "leaks.private.general" });
    expect(fetchMock.mock.calls[0][0]).toBe("https://leaks.now/api/v1/intelx/download");
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({
      systemId: "52c30047-3a62-41c8-bb1b-769520b90957",
      bucket: "leaks.private.general",
    });
  });

  it("getFile returns binary with content-type and filename", async () => {
    const bytes = new Uint8Array([1, 2, 3, 4]);
    const res = new Response(bytes, {
      status: 200,
      headers: {
        "content-type": "application/octet-stream",
        "content-disposition": 'attachment; filename="dump.bin"',
      },
    });
    const fetchMock = vi.fn().mockResolvedValue(res);
    const client = new LeaksNowClient("ms_key", { fetch: fetchMock });
    const file = await client.intelx.getFile(7);

    expect(fetchMock.mock.calls[0][0]).toBe("https://leaks.now/api/v1/intelx/downloads/7/file");
    expect(file.contentType).toBe("application/octet-stream");
    expect(file.filename).toBe("dump.bin");
    expect(new Uint8Array(file.data)).toEqual(bytes);
  });

  it("deleteDownload DELETEs by id", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ ok: true }));
    const client = new LeaksNowClient("ms_key", { fetch: fetchMock });
    await client.intelx.deleteDownload(7);
    expect(fetchMock.mock.calls[0][0]).toBe("https://leaks.now/api/v1/intelx/downloads/7");
    expect(fetchMock.mock.calls[0][1].method).toBe("DELETE");
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --filter @leaksnow/sdk exec vitest run test/intelx.test.ts`
Expected: FAIL — `client.intelx` undefined.

- [ ] **Step 3: Write minimal implementation**

`packages/typescript/src/resources/intelx.ts`:
```ts
import type { Transport } from "../client.js";
import type {
  IntelxDownloadRequest,
  IntelxDownloadResponse,
  IntelxDownloadList,
  BinaryFile,
  ResourceId,
} from "../types.js";

function parseFilename(disposition: string | null): string | undefined {
  if (!disposition) return undefined;
  const match = /filename\*?=(?:UTF-8''|")?([^\";]+)"?/i.exec(disposition);
  return match?.[1];
}

export class IntelxResource {
  readonly #transport: Transport;
  constructor(transport: Transport) {
    this.#transport = transport;
  }

  downloads(): Promise<IntelxDownloadList> {
    return this.#transport.request<IntelxDownloadList>("GET", "/api/v1/intelx/downloads");
  }

  download(body: IntelxDownloadRequest): Promise<IntelxDownloadResponse> {
    return this.#transport.request<IntelxDownloadResponse>("POST", "/api/v1/intelx/download", { body });
  }

  async getFile(id: ResourceId): Promise<BinaryFile> {
    const res = await this.#transport.requestRaw("GET", `/api/v1/intelx/downloads/${id}/file`);
    const data = await res.arrayBuffer();
    return {
      data,
      contentType: res.headers.get("content-type") ?? "application/octet-stream",
      filename: parseFilename(res.headers.get("content-disposition")),
    };
  }

  deleteDownload(id: ResourceId): Promise<unknown> {
    return this.#transport.request<unknown>("DELETE", `/api/v1/intelx/downloads/${id}`);
  }
}
```

In `client.ts`: add `import { IntelxResource } from "./resources/intelx.js";` at top, add `readonly intelx: IntelxResource;` field, and in the constructor add `this.intelx = new IntelxResource(this.transport);`.

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --filter @leaksnow/sdk exec vitest run test/intelx.test.ts`
Expected: PASS (4 tests).

- [ ] **Step 5: Commit**

```bash
git add packages/typescript/src/resources/intelx.ts packages/typescript/src/client.ts packages/typescript/test/intelx.test.ts
git commit -m "feat(sdk): add intelx resource with binary file download"
```

---

## Task 10: Public exports + build

**Files:**
- Modify: `packages/typescript/src/index.ts`

- [ ] **Step 1: Replace the stub with real exports**

`packages/typescript/src/index.ts`:
```ts
export { LeaksNowClient, Transport } from "./client.js";
export type { ClientConfig, RequestOptions } from "./client.js";
export { ShodanResource } from "./resources/shodan.js";
export { UlpResource } from "./resources/ulp.js";
export { IntelxResource } from "./resources/intelx.js";
export {
  LeaksNowError,
  LeaksNowAuthError,
  LeaksNowQuotaError,
  LeaksNowValidationError,
  LeaksNowServerError,
  errorFromResponse,
} from "./errors.js";
export type { LeaksNowErrorCode, LeaksNowErrorOptions } from "./errors.js";
export { DEFAULT_RETRY, backoffDelay, shouldRetry } from "./retry.js";
export type { RetryOptions } from "./retry.js";
export {
  SCOPES,
  SEVERITIES,
  ULP_TYPES,
  INTELX_BUCKETS,
} from "./types.js";
export type {
  Scope,
  Severity,
  UlpType,
  IntelxBucket,
  ResourceId,
  SearchRequest,
  SearchResponse,
  ShodanCustomScanRequest,
  ShodanScan,
  ShodanScanList,
  ShodanHost,
  UlpSearchRequest,
  UlpSearchResponse,
  UlpDownloadRequest,
  UlpDownloadResponse,
  IntelxDownloadRequest,
  IntelxDownloadResponse,
  IntelxDownloadList,
  BinaryFile,
} from "./types.js";
```

- [ ] **Step 2: Run full test + typecheck + build**

Run: `pnpm --filter @leaksnow/sdk test && pnpm --filter @leaksnow/sdk typecheck && pnpm --filter @leaksnow/sdk build`
Expected: all tests PASS, `tsc --noEmit` exits 0, `dist/` contains `index.js`, `index.cjs`, `index.d.ts`.

- [ ] **Step 3: Commit**

```bash
git add packages/typescript/src/index.ts
git commit -m "feat(sdk): export public API surface"
```

---

## Task 11: OpenAPI contract

**Files:**
- Create: `docs/openapi.yaml`

- [ ] **Step 1: Write the spec**

`docs/openapi.yaml`:
```yaml
openapi: 3.0.3
info:
  title: leaks.now API
  version: v1
  description: OSINT API — Leak, Service, Shodan, ULP, IntelX. Hand-authored from provider docs (requests documented; response shapes best-effort).
servers:
  - url: https://leaks.now
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: "ms_*"
security:
  - bearerAuth: []
paths:
  /api/v1/search:
    post:
      summary: OSINT search (leak/service/shodan). Cost 1.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [query]
              properties:
                query: { type: string }
                scope: { type: string, enum: [leak, service, shodan], default: leak }
                plugin: { type: string }
                severity: { type: string, enum: [critical, high, medium, low, info, all] }
                page: { type: integer, default: 0 }
      responses: { "200": { description: OK } }
  /api/v1/shodan/custom-scans:
    get: { summary: List custom Shodan scans. Cost 0., responses: { "200": { description: OK } } }
  /api/v1/shodan/custom-scan:
    post:
      summary: Launch a custom Shodan scan. Cost 1.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [target]
              properties: { target: { type: string } }
      responses: { "200": { description: OK } }
  /api/v1/shodan/custom-scans/{id}:
    get:
      summary: Get scan status by id. Cost 0.
      parameters: [{ name: id, in: path, required: true, schema: { type: integer } }]
      responses: { "200": { description: OK } }
    delete:
      summary: Delete scan entry. Cost 0.
      parameters: [{ name: id, in: path, required: true, schema: { type: integer } }]
      responses: { "200": { description: OK } }
  /api/v1/shodan/host/{ip}:
    get:
      summary: Aggregated Shodan view for an IPv4. Cost 0.
      parameters: [{ name: ip, in: path, required: true, schema: { type: string } }]
      responses: { "200": { description: OK } }
  /api/v1/ulp/search:
    post:
      summary: ULP (Leak Zero) search. Cost 1.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [type, value]
              properties:
                type: { type: string, enum: [domain, email, username, password] }
                value: { type: string }
      responses: { "200": { description: OK — includes downloadToken (~15 min) } }
  /api/v1/ulp/download:
    post:
      summary: JSON export of fetched ULP rows. Cost 10.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [token]
              properties: { token: { type: string } }
      responses: { "200": { description: OK } }
  /api/v1/intelx/downloads:
    get: { summary: IntelX download history. Cost 0., responses: { "200": { description: OK } } }
  /api/v1/intelx/download:
    post:
      summary: Request an IntelX file (server-cached). Cost 5.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [systemId, bucket]
              properties:
                systemId: { type: string, format: uuid }
                bucket: { type: string, enum: [leaks.private.general, leaks.logs, whois, dns, darknet.tor] }
      responses: { "200": { description: OK } }
  /api/v1/intelx/downloads/{id}/file:
    get:
      summary: Download cached binary (no re-billing). Cost 0.
      parameters: [{ name: id, in: path, required: true, schema: { type: string } }]
      responses: { "200": { description: OK, content: { application/octet-stream: { schema: { type: string, format: binary } } } } }
  /api/v1/intelx/downloads/{id}:
    delete:
      summary: Hide an entry from IntelX history. Cost 0.
      parameters: [{ name: id, in: path, required: true, schema: { type: string } }]
      responses: { "200": { description: OK } }
```

- [ ] **Step 2: Commit**

```bash
git add docs/openapi.yaml
git commit -m "docs: add hand-authored OpenAPI contract for /api/v1"
```

---

## Task 12: README

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write the README**

`README.md`:
````markdown
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
````

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add monorepo README with TypeScript usage"
```

---

## Self-Review

**Spec coverage:**
- Auth (Bearer `ms_`, base URL) → Task 5 transport headers + Task 11 OpenAPI. ✓
- Monorepo (pnpm+turbo, sibling dirs) → Task 1. ✓ (Go/Rust dirs created in their own plans.)
- All 12 endpoints → Tasks 6–9 + Task 11. ✓
- Enums (scope/severity/ulp.type/bucket) → Task 3. ✓
- Error taxonomy → Task 2. ✓
- Retry/backoff configurable, off by default → Task 4 (`DEFAULT_RETRY.maxRetries=0`) + Task 5 wiring. ✓
- Injectable fetch + timeout via AbortController → Task 5. ✓
- Binary `getFile` → Task 9. ✓
- Zero runtime deps, ESM+CJS+d.ts build → Tasks 1, 10. ✓
- Tests mocked, no live calls in CI → all test tasks use mocked fetch. ✓
- OpenAPI shared contract → Task 11. ✓

**Placeholder scan:** No TBD/TODO/"handle edge cases". The "id type TBD" from the spec is resolved as `ResourceId = string | number` (Task 3). ✓

**Type consistency:** `Transport`/`request`/`requestRaw`, `LeaksNowClient`, resource class names, and method names (`customScans`, `customScan`, `getScan`, `deleteScan`, `host`, `search`, `download`, `downloads`, `getFile`, `deleteDownload`) are used identically across tasks and `index.ts`. `ClientConfig.retry` is `Partial<RetryOptions>` merged over `DEFAULT_RETRY`. ✓

**Out of scope (per spec):** no publish pipeline, no CLI, no pagination helpers, Go/Rust deferred to separate plans.
