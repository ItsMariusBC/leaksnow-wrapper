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
