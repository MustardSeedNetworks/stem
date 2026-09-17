/**
 * AUTO-GENERATED FILE. DO NOT EDIT BY HAND.
 *
 * Regenerate with: `npm run gen-types` (or `make schema && npm run gen-types`
 * after Go DTO changes). The schema source of truth lives at
 * docs/schemas/api/; the Go DTO source lives at internal/api/.
 */
/**
 * InterfaceInfo — generated from the Go struct in internal/api; refresh with `make schema` after struct changes.
 */
export interface InterfaceInfo {
  name: string;
  mac: string;
  speed: number;
  duplex: string;
  state: string;
  driver: string;
  physical: boolean;
  xdp: boolean;
  score: number;
  mtu: number;
  ipv4: string;
  ipv6: string;
  usable: boolean;
}
