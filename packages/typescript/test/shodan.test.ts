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
