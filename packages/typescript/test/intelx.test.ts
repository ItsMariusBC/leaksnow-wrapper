import { describe, it, expect, vi } from "vitest";
import { LeaksNowClient } from "../src/client.js";

function jsonResponse(body: unknown) {
  return new Response(JSON.stringify(body), { status: 200, headers: { "content-type": "application/json" } });
}

describe("client.intelx", () => {
  it("downloads lists history", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ items: [] }));
    const client = new LeaksNowClient("ms_key", { fetch: fetchMock });
    await client.intelx.downloads();
    expect(fetchMock.mock.calls[0][0]).toBe("https://leaks.now/api/v1/intelx/downloads");
    expect(fetchMock.mock.calls[0][1].method).toBe("GET");
  });

  it("download POSTs systemId+bucket", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ id: "1" }));
    const client = new LeaksNowClient("ms_key", { fetch: fetchMock });
    await client.intelx.download({ systemId: "52c30047-3a62-41c8-bb1b-769520b90957", bucket: "leaks.private.general" });
    expect(fetchMock.mock.calls[0][0]).toBe("https://leaks.now/api/v1/intelx/download");
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({
      systemId: "52c30047-3a62-41c8-bb1b-769520b90957",
      bucket: "leaks.private.general",
    });
  });

  it("getFile returns binary with content-type and filename", async () => {
    const bytes = new Uint8Array([1, 2, 3, 4]);
    const res = new Response(bytes, {
      status: 200,
      headers: {
        "content-type": "application/octet-stream",
        "content-disposition": 'attachment; filename="dump.bin"',
      },
    });
    const fetchMock = vi.fn().mockResolvedValue(res);
    const client = new LeaksNowClient("ms_key", { fetch: fetchMock });
    const file = await client.intelx.getFile(7);

    expect(fetchMock.mock.calls[0][0]).toBe("https://leaks.now/api/v1/intelx/downloads/7/file");
    expect(file.contentType).toBe("application/octet-stream");
    expect(file.filename).toBe("dump.bin");
    expect(new Uint8Array(file.data)).toEqual(bytes);
  });

  it("deleteDownload DELETEs by id", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ ok: true }));
    const client = new LeaksNowClient("ms_key", { fetch: fetchMock });
    await client.intelx.deleteDownload(7);
    expect(fetchMock.mock.calls[0][0]).toBe("https://leaks.now/api/v1/intelx/downloads/7");
    expect(fetchMock.mock.calls[0][1].method).toBe("DELETE");
  });
});
