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
