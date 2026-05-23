import type { Transport } from "../client.js";
import type {
  IntelxDownloadRequest,
  IntelxDownloadResponse,
  IntelxDownloadList,
  BinaryFile,
  ResourceId,
} from "../types.js";

function parseFilename(disposition: string | null): string | undefined {
  if (!disposition) return undefined;
  // RFC 5987 extended form takes precedence; decode percent-encoding.
  const ext = /filename\*\s*=\s*(?:UTF-8|ISO-8859-1)?''([^;]+)/i.exec(disposition);
  if (ext?.[1]) {
    const raw = ext[1].trim();
    try {
      return decodeURIComponent(raw);
    } catch {
      return raw;
    }
  }
  const basic = /filename\s*=\s*"?([^";]+)"?/i.exec(disposition);
  return basic?.[1]?.trim();
}

export class IntelxResource {
  readonly #transport: Transport;
  constructor(transport: Transport) {
    this.#transport = transport;
  }

  downloads(): Promise<IntelxDownloadList> {
    return this.#transport.request<IntelxDownloadList>("GET", "/api/v1/intelx/downloads");
  }

  download(body: IntelxDownloadRequest): Promise<IntelxDownloadResponse> {
    return this.#transport.request<IntelxDownloadResponse>("POST", "/api/v1/intelx/download", { body });
  }

  /**
   * Download a stored IntelX file by id.
   *
   * Returns raw bytes as a {@link BinaryFile}, NOT parsed JSON. Persist with e.g.
   * `await writeFile(file.filename ?? "download.bin", Buffer.from(file.data))`.
   * `filename` is best-effort parsed from `Content-Disposition` and may be undefined.
   */
  async getFile(id: ResourceId): Promise<BinaryFile> {
    const res = await this.#transport.requestRaw("GET", `/api/v1/intelx/downloads/${encodeURIComponent(String(id))}/file`);
    const data = await res.arrayBuffer();
    return {
      data,
      contentType: res.headers.get("content-type") ?? "application/octet-stream",
      filename: parseFilename(res.headers.get("content-disposition")),
    };
  }

  deleteDownload(id: ResourceId): Promise<unknown> {
    return this.#transport.request<unknown>("DELETE", `/api/v1/intelx/downloads/${encodeURIComponent(String(id))}`);
  }
}
