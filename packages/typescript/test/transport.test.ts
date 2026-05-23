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
