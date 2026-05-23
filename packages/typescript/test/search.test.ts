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
