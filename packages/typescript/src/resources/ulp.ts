import type { Transport } from "../client.js";
import type {
  UlpSearchRequest,
  UlpSearchResponse,
  UlpDownloadRequest,
  UlpDownloadResponse,
} from "../types.js";

export class UlpResource {
  readonly #transport: Transport;
  constructor(transport: Transport) {
    this.#transport = transport;
  }

  search(body: UlpSearchRequest): Promise<UlpSearchResponse> {
    return this.#transport.request<UlpSearchResponse>("POST", "/api/v1/ulp/search", { body });
  }

  download(body: UlpDownloadRequest): Promise<UlpDownloadResponse> {
    return this.#transport.request<UlpDownloadResponse>("POST", "/api/v1/ulp/download", { body });
  }
}
