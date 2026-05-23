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
