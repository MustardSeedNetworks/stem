/**
 * Schema-level tests for the test-configuration schemas. These cover
 * the range bounds + cross-field rules that HTML5 min/max attributes
 * couldn't enforce.
 */
import * as v from 'valibot';
import { describe, expect, it } from 'vitest';
import {
  RFC2544ConfigSchema,
  RFC2889ConfigSchema,
  RFC6349ConfigSchema,
  TrafficGenConfigSchema,
  TSNConfigSchema,
  Y1564ConfigSchema,
  Y1731ConfigSchema,
} from './configs';

describe('Y1564ConfigSchema', () => {
  const valid = {
    cir: 100,
    eir: 0,
    cbs: 12,
    ebs: 0,
    frameSizes: [64, 128, 1518],
    configStepDuration: 15,
    perfTestDuration: 900,
    vlanId: 0,
    pcp: 0,
    colorAware: false,
    flrThreshold: 0.01,
    fdThreshold: 10,
    fdvThreshold: 5,
  };

  it('accepts a default-shaped config', () => {
    expect(v.safeParse(Y1564ConfigSchema, valid).success).toBe(true);
  });

  it('rejects CIR < 1 Mbps', () => {
    const r = v.safeParse(Y1564ConfigSchema, { ...valid, cir: 0 });
    expect(r.success).toBe(false);
  });

  it('rejects empty frameSizes', () => {
    const r = v.safeParse(Y1564ConfigSchema, { ...valid, frameSizes: [] });
    expect(r.success).toBe(false);
  });

  it('rejects FDV threshold > FD threshold (cross-field)', () => {
    const r = v.safeParse(Y1564ConfigSchema, {
      ...valid,
      fdThreshold: 10,
      fdvThreshold: 20,
    });
    expect(r.success).toBe(false);
    if (!r.success) {
      const issues = r.issues.map((i) => i.message).join(' | ');
      expect(issues).toMatch(/Variation/i);
    }
  });

  it('rejects VLAN ID > 4094', () => {
    const r = v.safeParse(Y1564ConfigSchema, { ...valid, vlanId: 5000 });
    expect(r.success).toBe(false);
  });

  it('rejects PCP > 7', () => {
    const r = v.safeParse(Y1564ConfigSchema, { ...valid, pcp: 8 });
    expect(r.success).toBe(false);
  });
});

describe('RFC6349ConfigSchema', () => {
  const valid = {
    targetRateMbps: 100,
    minRTTMs: 5,
    maxRTTMs: 50,
    rwndSize: 65536,
    duration: 60,
    parallelStreams: 4,
    mss: 1460,
    mode: 0,
  };

  it('accepts a default-shaped config', () => {
    expect(v.safeParse(RFC6349ConfigSchema, valid).success).toBe(true);
  });

  it('rejects minRTT > maxRTT (cross-field)', () => {
    const r = v.safeParse(RFC6349ConfigSchema, {
      ...valid,
      minRTTMs: 100,
      maxRTTMs: 50,
    });
    expect(r.success).toBe(false);
  });
});

describe('TSNConfigSchema', () => {
  const valid = {
    duration: 60,
    warmup: 0,
    frameSize: 256,
    maxLatencyNs: 1_000_000,
    maxJitterNs: 100_000,
    requirePTPSync: true,
    maxSyncOffsetNs: 1000,
    ptpEnabled: true,
    preemptionEnabled: false,
    numTrafficClasses: 8,
    baseTimeNs: 0,
    cycleTimeNs: 1_000_000,
    trafficClass: 6,
  };

  it('accepts a default-shaped config', () => {
    expect(v.safeParse(TSNConfigSchema, valid).success).toBe(true);
  });

  it('rejects jitter > latency (cross-field)', () => {
    const r = v.safeParse(TSNConfigSchema, {
      ...valid,
      maxLatencyNs: 1000,
      maxJitterNs: 10_000,
    });
    expect(r.success).toBe(false);
  });
});

describe('TrafficGenConfigSchema', () => {
  const valid = {
    frameSize: 256,
    ratePct: 50,
    duration: 60,
    warmup: 0,
    streamId: 0,
    burstMode: false,
    burstSize: 1,
    interBurstGapUs: 0,
    srcMac: '',
    dstMac: '',
    vlanId: 0,
    vlanPriority: 0,
  };

  it('accepts a default-shaped config', () => {
    expect(v.safeParse(TrafficGenConfigSchema, valid).success).toBe(true);
  });

  it('accepts a valid MAC address', () => {
    const r = v.safeParse(TrafficGenConfigSchema, {
      ...valid,
      srcMac: 'AA:BB:CC:DD:EE:FF',
    });
    expect(r.success).toBe(true);
  });

  it('rejects malformed MAC', () => {
    const r = v.safeParse(TrafficGenConfigSchema, {
      ...valid,
      srcMac: 'not-a-mac',
    });
    expect(r.success).toBe(false);
  });

  it('rejects rate > 100%', () => {
    const r = v.safeParse(TrafficGenConfigSchema, { ...valid, ratePct: 150 });
    expect(r.success).toBe(false);
  });
});

describe('RFC2544ConfigSchema', () => {
  const valid = {
    duration: 60,
    frameSizes: [64, 1518],
    resolution: 1,
    maxLoss: 0,
    warmup: 0,
    trials: 1,
    stepSize: 1,
    bidirectional: false,
  };

  it('accepts a default-shaped config', () => {
    expect(v.safeParse(RFC2544ConfigSchema, valid).success).toBe(true);
  });

  it('rejects empty frameSizes', () => {
    const r = v.safeParse(RFC2544ConfigSchema, { ...valid, frameSizes: [] });
    expect(r.success).toBe(false);
  });
});

describe('RFC2889ConfigSchema', () => {
  const valid = {
    frameSize: 256,
    duration: 60,
    warmup: 0,
    addressCount: 100,
    acceptableLoss: 0,
    portCount: 2,
    pattern: 0,
  };

  it('accepts a default-shaped config', () => {
    expect(v.safeParse(RFC2889ConfigSchema, valid).success).toBe(true);
  });

  it('rejects pattern > 2', () => {
    const r = v.safeParse(RFC2889ConfigSchema, { ...valid, pattern: 5 });
    expect(r.success).toBe(false);
  });
});

describe('Y1731ConfigSchema', () => {
  const valid = {
    mepId: 1,
    megLevel: 0,
    megId: 'MSN-MEG',
    ccmInterval: 1000,
    priority: 0,
    duration: 60,
    intervalMs: 100,
    count: 10,
    frameSize: 256,
    priorityTagged: false,
  };

  it('accepts a default-shaped config', () => {
    expect(v.safeParse(Y1731ConfigSchema, valid).success).toBe(true);
  });

  it('rejects empty megId', () => {
    const r = v.safeParse(Y1731ConfigSchema, { ...valid, megId: '' });
    expect(r.success).toBe(false);
  });

  it('rejects MEG level > 7', () => {
    const r = v.safeParse(Y1731ConfigSchema, { ...valid, megLevel: 8 });
    expect(r.success).toBe(false);
  });

  it('rejects CCM interval < 3ms (Y.1731 minimum)', () => {
    const r = v.safeParse(Y1731ConfigSchema, { ...valid, ccmInterval: 1 });
    expect(r.success).toBe(false);
  });
});
