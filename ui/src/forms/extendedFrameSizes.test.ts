import { describe, expect, it } from 'vitest';

import { defaultTrafficGenConfig } from '../components/TrafficGenConfigForm';
import { defaultTSNConfig } from '../components/TSNConfigForm';
import { defaultY1564Config } from '../components/Y1564ConfigForm';
import { EXTENDED_FRAME_SIZE_OPTIONS, MIN_EXTENDED_FRAME_SIZE } from './frameSizes';

// The extended payload formats cannot carry a 64-byte frame: the daemon
// answers 400, and before stem#1250 the dataplane answered a bare -22. The
// shipped defaults are what a bare "start" submits, so they have to be legal.
describe('extended payload frame sizes', () => {
  it('ships a legal TSN default', () => {
    expect(defaultTSNConfig.frameSize).toBeGreaterThanOrEqual(MIN_EXTENDED_FRAME_SIZE);
  });

  it('ships a legal TrafficGen default', () => {
    expect(defaultTrafficGenConfig.frameSize).toBeGreaterThanOrEqual(MIN_EXTENDED_FRAME_SIZE);
  });

  it('ships a legal Y.1564 sweep', () => {
    expect(defaultY1564Config.frameSizes.length).toBeGreaterThan(0);
    for (const size of defaultY1564Config.frameSizes) {
      expect(size).toBeGreaterThanOrEqual(MIN_EXTENDED_FRAME_SIZE);
    }
  });

  it('offers no size the daemon would reject', () => {
    expect(EXTENDED_FRAME_SIZE_OPTIONS.length).toBeGreaterThan(0);
    for (const option of EXTENDED_FRAME_SIZE_OPTIONS) {
      expect(option.value).toBeGreaterThanOrEqual(MIN_EXTENDED_FRAME_SIZE);
    }
  });
});
