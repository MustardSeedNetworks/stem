/**
 * Valibot schemas for the stem test-configuration forms.
 *
 * Each ConfigForm (RFC 2544/2889/6349, Y.1564/1731, TSN, TrafficGen)
 * used to embed validation only via HTML5 min/max attributes — a
 * browser hint that's bypassed by copy-paste or programmatic input.
 * These schemas mirror the bounds the JSX already declared, plus the
 * cross-field rules that were previously impossible to express
 * cleanly.
 *
 * The default values match the existing `default<Foo>Config` exports
 * in each ConfigForm file; keeping them here means a single source of
 * truth for both shape and seed values.
 *
 * Range bounds were derived from the HTML5 min={} max={} attributes
 * already in the JSX. Where the JSX had no explicit bound, the schema
 * uses a domain-sensible cap (e.g., MEG level 0-7, VLAN 0-4094, PCP
 * 0-7 per IEEE 802.1Q).
 */
import * as v from 'valibot';

// =============================================================================
// Reusable field schemas
// =============================================================================

/** VLAN ID per IEEE 802.1Q: 0 = untagged, 1-4094 = tagged, 4095 reserved. */
const VlanIdSchema = v.pipe(
  v.number('VLAN ID must be a number'),
  v.integer('VLAN ID must be an integer'),
  v.minValue(0, 'VLAN ID must be 0-4094'),
  v.maxValue(4094, 'VLAN ID must be 0-4094'),
);

/** Priority Code Point per IEEE 802.1p: 0-7. */
const PCPSchema = v.pipe(
  v.number('PCP must be a number'),
  v.integer('PCP must be an integer'),
  v.minValue(0, 'PCP must be 0-7'),
  v.maxValue(7, 'PCP must be 0-7'),
);

/** Standard Ethernet frame size: 64-9000 bytes (jumbo). */
const FrameSizeSchema = v.pipe(
  v.number('Frame size must be a number'),
  v.integer('Frame size must be an integer'),
  v.minValue(64, 'Frame size must be at least 64 bytes'),
  v.maxValue(9000, 'Frame size must be at most 9000 bytes'),
);

// =============================================================================
// RFC 2544 — throughput, latency, frame-loss, back-to-back
// =============================================================================

export const RFC2544ConfigSchema = v.object({
  duration: v.pipe(
    v.number(),
    v.minValue(1, 'Duration must be at least 1 second'),
    v.maxValue(3600, 'Duration must be at most 3600 seconds'),
  ),
  frameSizes: v.pipe(v.array(FrameSizeSchema), v.minLength(1, 'Select at least one frame size')),
  resolution: v.pipe(
    v.number(),
    v.minValue(0.01, 'Resolution must be at least 0.01%'),
    v.maxValue(10, 'Resolution must be at most 10%'),
  ),
  maxLoss: v.pipe(
    v.number(),
    v.minValue(0, 'Max loss must be 0% or more'),
    v.maxValue(100, 'Max loss must be 100% or less'),
  ),
  warmup: v.pipe(
    v.number(),
    v.minValue(0, 'Warmup must be 0 or more seconds'),
    v.maxValue(600, 'Warmup must be at most 600 seconds'),
  ),
  trials: v.pipe(
    v.number(),
    v.integer(),
    v.minValue(1, 'Trials must be at least 1'),
    v.maxValue(100, 'Trials must be at most 100'),
  ),
  stepSize: v.pipe(
    v.number(),
    v.minValue(0.1, 'Step size must be at least 0.1%'),
    v.maxValue(50, 'Step size must be at most 50%'),
  ),
  bidirectional: v.boolean(),
});

// =============================================================================
// RFC 2889 — LAN switch benchmarking
// =============================================================================

export const RFC2889ConfigSchema = v.object({
  frameSize: FrameSizeSchema,
  duration: v.pipe(
    v.number(),
    v.minValue(1, 'Duration must be at least 1 second'),
    v.maxValue(3600),
  ),
  warmup: v.pipe(v.number(), v.minValue(0), v.maxValue(600)),
  addressCount: v.pipe(
    v.number(),
    v.integer(),
    v.minValue(1, 'Address count must be at least 1'),
    v.maxValue(1000000, 'Address count must be at most 1,000,000'),
  ),
  acceptableLoss: v.pipe(v.number(), v.minValue(0), v.maxValue(100)),
  portCount: v.pipe(
    v.number(),
    v.integer(),
    v.minValue(2, 'Port count must be at least 2'),
    v.maxValue(128, 'Port count must be at most 128'),
  ),
  pattern: v.pipe(
    v.number(),
    v.integer(),
    v.minValue(0),
    v.maxValue(2, 'Pattern must be 0 (mesh), 1 (pair), or 2 (broadcast)'),
  ),
});

// =============================================================================
// RFC 6349 — TCP throughput
// =============================================================================

export const RFC6349ConfigSchema = v.pipe(
  v.object({
    targetRateMbps: v.pipe(
      v.number(),
      v.minValue(1, 'Target rate must be at least 1 Mbps'),
      v.maxValue(100000, 'Target rate must be at most 100,000 Mbps'),
    ),
    minRTTMs: v.pipe(v.number(), v.minValue(0)),
    maxRTTMs: v.pipe(v.number(), v.minValue(0)),
    rwndSize: v.pipe(
      v.number(),
      v.integer(),
      v.minValue(1024, 'RWND must be at least 1024 bytes'),
      v.maxValue(16777216, 'RWND must be at most 16 MB'),
    ),
    duration: v.pipe(v.number(), v.minValue(1), v.maxValue(3600)),
    parallelStreams: v.pipe(
      v.number(),
      v.integer(),
      v.minValue(1),
      v.maxValue(64, 'Parallel streams must be at most 64'),
    ),
    mss: v.pipe(
      v.number(),
      v.integer(),
      v.minValue(536, 'MSS must be at least 536 bytes'),
      v.maxValue(9000, 'MSS must be at most 9000 bytes'),
    ),
    mode: v.pipe(
      v.number(),
      v.integer(),
      v.minValue(0),
      v.maxValue(2, 'Mode must be 0 (bidi), 1 (up), or 2 (down)'),
    ),
  }),
  v.check((c) => c.minRTTMs <= c.maxRTTMs, 'Minimum RTT must be ≤ maximum RTT'),
);

// =============================================================================
// TrafficGen — custom traffic generation
// =============================================================================

/** MAC address: 6 octets, colon or dash separated. Empty string allowed (optional). */
const OptionalMacSchema = v.union([
  v.literal(''),
  v.pipe(
    v.string(),
    v.regex(
      /^[0-9A-Fa-f]{2}([:-][0-9A-Fa-f]{2}){5}$/,
      'MAC must be 6 octets like AA:BB:CC:DD:EE:FF',
    ),
  ),
]);

export const TrafficGenConfigSchema = v.object({
  frameSize: FrameSizeSchema,
  ratePct: v.pipe(
    v.number(),
    v.minValue(0.01, 'Rate must be at least 0.01%'),
    v.maxValue(100, 'Rate must be at most 100%'),
  ),
  duration: v.pipe(v.number(), v.minValue(1), v.maxValue(86400)),
  warmup: v.pipe(v.number(), v.minValue(0), v.maxValue(600)),
  streamId: v.pipe(
    v.number(),
    v.integer(),
    v.minValue(0),
    v.maxValue(65535, 'Stream ID must fit in 16 bits'),
  ),
  burstMode: v.boolean(),
  burstSize: v.pipe(
    v.number(),
    v.integer(),
    v.minValue(1),
    v.maxValue(1000000, 'Burst size must be at most 1,000,000 frames'),
  ),
  interBurstGapUs: v.pipe(
    v.number(),
    v.minValue(0),
    v.maxValue(1000000, 'IBG must be at most 1,000,000 µs (1 second)'),
  ),
  srcMac: OptionalMacSchema,
  dstMac: OptionalMacSchema,
  vlanId: VlanIdSchema,
  vlanPriority: PCPSchema,
});

// =============================================================================
// TSN — Time-Sensitive Networking
// =============================================================================

export const TSNConfigSchema = v.pipe(
  v.object({
    duration: v.pipe(v.number(), v.minValue(1), v.maxValue(3600)),
    warmup: v.pipe(v.number(), v.minValue(0), v.maxValue(600)),
    frameSize: FrameSizeSchema,
    maxLatencyNs: v.pipe(
      v.number(),
      v.minValue(0),
      v.maxValue(1e12, 'Max latency must be ≤ 1 second (1e12 ns)'),
    ),
    maxJitterNs: v.pipe(
      v.number(),
      v.minValue(0),
      v.maxValue(1e12, 'Max jitter must be ≤ 1 second (1e12 ns)'),
    ),
    requirePTPSync: v.boolean(),
    maxSyncOffsetNs: v.pipe(v.number(), v.minValue(0), v.maxValue(1e9)),
    ptpEnabled: v.boolean(),
    preemptionEnabled: v.boolean(),
    numTrafficClasses: v.pipe(
      v.number(),
      v.integer(),
      v.minValue(1),
      v.maxValue(8, 'Traffic classes must be 1-8'),
    ),
    baseTimeNs: v.pipe(v.number(), v.minValue(0)),
    cycleTimeNs: v.pipe(v.number(), v.minValue(1, 'Cycle time must be at least 1 ns')),
    trafficClass: PCPSchema,
  }),
  v.check((c) => c.maxJitterNs <= c.maxLatencyNs, 'Max jitter must be ≤ max latency'),
);

// =============================================================================
// Y.1564 / MEF — service activation
// =============================================================================

export const Y1564ConfigSchema = v.pipe(
  v.object({
    cir: v.pipe(
      v.number(),
      v.minValue(1, 'CIR must be at least 1 Mbps'),
      v.maxValue(10000, 'CIR must be at most 10,000 Mbps'),
    ),
    eir: v.pipe(
      v.number(),
      v.minValue(0, 'EIR must be 0 or more'),
      v.maxValue(10000, 'EIR must be at most 10,000 Mbps'),
    ),
    cbs: v.pipe(
      v.number(),
      v.minValue(1, 'CBS must be at least 1 KB'),
      v.maxValue(1024, 'CBS must be at most 1024 KB'),
    ),
    ebs: v.pipe(
      v.number(),
      v.minValue(0, 'EBS must be 0 or more'),
      v.maxValue(1024, 'EBS must be at most 1024 KB'),
    ),
    frameSizes: v.pipe(v.array(FrameSizeSchema), v.minLength(1, 'Select at least one frame size')),
    configStepDuration: v.pipe(v.number(), v.minValue(1), v.maxValue(300)),
    perfTestDuration: v.pipe(
      v.number(),
      v.minValue(1, 'Performance test must be at least 1 second'),
      v.maxValue(86400, 'Performance test must be at most 86,400 seconds (24h)'),
    ),
    vlanId: VlanIdSchema,
    pcp: PCPSchema,
    colorAware: v.boolean(),
    flrThreshold: v.pipe(v.number(), v.minValue(0), v.maxValue(100)),
    fdThreshold: v.pipe(v.number(), v.minValue(0)),
    fdvThreshold: v.pipe(v.number(), v.minValue(0)),
  }),
  v.check(
    (c) => c.fdvThreshold <= c.fdThreshold,
    'Frame Delay Variation threshold must be ≤ Frame Delay threshold',
  ),
);

// =============================================================================
// Y.1731 — OAM (Delay, Loss, Synthetic Loss)
// =============================================================================

export const Y1731ConfigSchema = v.object({
  mepId: v.pipe(
    v.number(),
    v.integer(),
    v.minValue(1, 'MEP ID must be at least 1'),
    v.maxValue(8191, 'MEP ID must be at most 8191'),
  ),
  megLevel: v.pipe(
    v.number(),
    v.integer(),
    v.minValue(0, 'MEG level must be 0-7'),
    v.maxValue(7, 'MEG level must be 0-7'),
  ),
  megId: v.pipe(
    v.string(),
    v.minLength(1, 'MEG ID is required'),
    v.maxLength(45, 'MEG ID is too long (max 45 chars per Y.1731)'),
  ),
  ccmInterval: v.pipe(
    v.number(),
    v.minValue(3, 'CCM interval must be at least 3 ms per ITU-T Y.1731'),
    v.maxValue(600000, 'CCM interval must be at most 10 minutes'),
  ),
  priority: PCPSchema,
  duration: v.pipe(v.number(), v.minValue(1), v.maxValue(86400)),
  intervalMs: v.pipe(
    v.number(),
    v.minValue(1, 'Measurement interval must be at least 1 ms'),
    v.maxValue(60000, 'Measurement interval must be at most 60 s'),
  ),
  count: v.pipe(
    v.number(),
    v.integer(),
    v.minValue(1),
    v.maxValue(10000, 'Count must be at most 10,000 per interval'),
  ),
  frameSize: FrameSizeSchema,
  priorityTagged: v.boolean(),
});
