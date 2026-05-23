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
    Object.setPrototypeOf(this, new.target.prototype);
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
