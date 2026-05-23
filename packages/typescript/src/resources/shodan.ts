import type { Transport } from "../client.js";
import type {
  ShodanCustomScanRequest,
  ShodanScan,
  ShodanScanList,
  ShodanHost,
  ResourceId,
} from "../types.js";

export class ShodanResource {
  readonly #transport: Transport;
  constructor(transport: Transport) {
    this.#transport = transport;
  }

  customScans(): Promise<ShodanScanList> {
    return this.#transport.request<ShodanScanList>("GET", "/api/v1/shodan/custom-scans");
  }

  customScan(body: ShodanCustomScanRequest): Promise<ShodanScan> {
    return this.#transport.request<ShodanScan>("POST", "/api/v1/shodan/custom-scan", { body });
  }

  getScan(id: ResourceId): Promise<ShodanScan> {
    return this.#transport.request<ShodanScan>("GET", `/api/v1/shodan/custom-scans/${encodeURIComponent(String(id))}`);
  }

  deleteScan(id: ResourceId): Promise<unknown> {
    return this.#transport.request<unknown>("DELETE", `/api/v1/shodan/custom-scans/${encodeURIComponent(String(id))}`);
  }

  host(ip: string): Promise<ShodanHost> {
    return this.#transport.request<ShodanHost>("GET", `/api/v1/shodan/host/${encodeURIComponent(ip)}`);
  }
}
