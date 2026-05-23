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
