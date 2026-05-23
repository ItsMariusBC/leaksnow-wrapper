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
