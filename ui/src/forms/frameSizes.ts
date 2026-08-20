/**
 * The Ethernet frame sizes every sweep offers, shared by the forms that sweep.
 *
 * Six of the seven config forms declared their own identical copy of this
 * list, each with the qualifier baked into an English label. The qualifier is
 * a separate key here so "min" and "jumbo" translate while the byte count
 * does not, and so a size added for one standard cannot be silently missing
 * from the others.
 */
export type FrameSizeQualifier = 'frameSizeMin' | 'frameSize' | 'frameSizeMax' | 'frameSizeJumbo';

export interface FrameSizeOption {
  value: number;
  /** Key under `settings.testConfig.common`. */
  qualifier: FrameSizeQualifier;
}

export const FRAME_SIZE_OPTIONS: FrameSizeOption[] = [
  { value: 64, qualifier: 'frameSizeMin' },
  { value: 128, qualifier: 'frameSize' },
  { value: 256, qualifier: 'frameSize' },
  { value: 512, qualifier: 'frameSize' },
  { value: 1024, qualifier: 'frameSize' },
  { value: 1280, qualifier: 'frameSize' },
  { value: 1518, qualifier: 'frameSizeMax' },
  { value: 9000, qualifier: 'frameSizeJumbo' },
];
