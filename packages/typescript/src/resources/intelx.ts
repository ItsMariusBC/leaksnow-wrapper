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
  const match = /filename\*?=(?:UTF-8''|")?([^\";]+)"?/i.exec(disposition);
  return match?.[1];
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

  async getFile(id: ResourceId): Promise<BinaryFile> {
    const res = await this.#transport.requestRaw("GET", `/api/v1/intelx/downloads/${id}/file`);
    const data = await res.arrayBuffer();
    return {
      data,
      contentType: res.headers.get("content-type") ?? "application/octet-stream",
      filename: parseFilename(res.headers.get("content-disposition")),
    };
  }

  deleteDownload(id: ResourceId): Promise<unknown> {
    return this.#transport.request<unknown>("DELETE", `/api/v1/intelx/downloads/${id}`);
  }
}
