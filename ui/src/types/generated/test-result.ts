/**
 * AUTO-GENERATED FILE. DO NOT EDIT BY HAND.
 *
 * Regenerate with: `npm run gen-types` (or `make schema && npm run gen-types`
 * after Go DTO changes). The schema source of truth lives at
 * docs/schemas/api/; the Go DTO source lives at internal/api/.
 */
/**
 * TestResultResponse — generated from the Go struct in internal/api; refresh with `make schema` after struct changes.
 */
export interface TestResultResponse {
  status: string;
  testType?: string;
  module?: string;
  success?: boolean;
  error?: string;
  message?: string;
  data?: unknown;
  suiteId?: string;
  steps?: RunPlanStep[];
}
export interface RunPlanStep {
  testType: string;
  module: string;
  status: 'pending' | 'running' | 'passed' | 'failed' | 'skipped' | 'cancelled';
  config?: TestConfig;
  error?: string;
  result?: TestResultResponse1;
}
export interface TestConfig {
  rfc2544?: RFC2544TestConfig;
  rfc2889?: RFC2889TestConfig;
  rfc6349?: RFC6349TestConfig;
  y1564?: Y1564TestConfig;
  y1731?: Y1731TestConfig;
  tsn?: TSNTestConfig;
  trafficGen?: TrafficGenTestConfig;
}
export interface RFC2544TestConfig {
  duration: number;
  frameSizes: number[];
  resolution: number;
  maxLoss: number;
  warmup: number;
  trials: number;
  stepSize: number;
  bidirectional: boolean;
}
export interface RFC2889TestConfig {
  frameSize: number;
  duration: number;
  warmup: number;
  addressCount: number;
  acceptableLoss: number;
  portCount: number;
  pattern: number;
}
export interface RFC6349TestConfig {
  targetRateMbps: number;
  minRTTMs: number;
  maxRTTMs: number;
  rwndSize: number;
  duration: number;
  parallelStreams: number;
  mss: number;
  mode: number;
}
export interface Y1564TestConfig {
  cir: number;
  eir: number;
  cbs: number;
  ebs: number;
  frameSizes: number[];
  configStepDuration: number;
  perfTestDuration: number;
  vlanId: number;
  pcp: number;
  colorAware: boolean;
  flrThreshold: number;
  fdThreshold: number;
  fdvThreshold: number;
}
export interface Y1731TestConfig {
  mepId: number;
  megLevel: number;
  megId: string;
  ccmInterval: number;
  priority: number;
  duration: number;
  intervalMs: number;
  count: number;
  frameSize: number;
  priorityTagged: boolean;
}
export interface TSNTestConfig {
  duration: number;
  warmup: number;
  frameSize: number;
  maxLatencyNs: number;
  maxJitterNs: number;
  requirePTPSync: boolean;
  maxSyncOffsetNs: number;
  ptpEnabled: boolean;
  preemptionEnabled: boolean;
  numTrafficClasses: number;
  baseTimeNs: number;
  cycleTimeNs: number;
  trafficClass: number;
}
export interface TrafficGenTestConfig {
  frameSize: number;
  ratePct: number;
  duration: number;
  warmup: number;
  streamId: number;
  burstMode: boolean;
  burstSize: number;
  interBurstGapUs: number;
  srcMac: string;
  dstMac: string;
  vlanId: number;
  vlanPriority: number;
}
export interface TestResultResponse1 {
  status: string;
  testType?: string;
  module?: string;
  success?: boolean;
  error?: string;
  message?: string;
  data?: unknown;
  suiteId?: string;
  steps?: RunPlanStep[];
}
