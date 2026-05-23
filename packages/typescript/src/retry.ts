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
