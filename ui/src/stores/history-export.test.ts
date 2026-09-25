import { describe, expect, it } from 'vitest';
import { exportFileName, historyToCsv, historyToJson } from './history-export';
import type { HistoricalResult } from './history-store';

const passed: HistoricalResult = {
  id: '2026-08-17T10:00:00Z-RFC 2544',
  testType: 'RFC 2544',
  module: 'benchmark',
  status: 'completed',
  startedAt: '2026-08-17T09:58:00Z',
  completedAt: '2026-08-17T10:00:00Z',
  duration: 120000,
  success: true,
  metrics: { throughput_mbps: 940, frame_size: 1518 },
};

const failed: HistoricalResult = {
  id: '2026-08-17T11:00:00Z-Y.1564',
  testType: 'Y.1564',
  module: 'servicetest',
  status: 'error',
  completedAt: '2026-08-17T11:00:00Z',
  success: false,
  error: 'peer unreachable, "no reply"',
  metrics: { loss_pct: '-0.5', latency_us: -1 },
};

const noVerdict: HistoricalResult = {
  id: 'no-verdict',
  testType: 'Traffic',
  module: 'trafficgen',
  status: 'stopped',
};

function parse(csv: string): string[] {
  expect(csv.endsWith('\r\n')).toBe(true);
  return csv.slice(0, -2).split('\r\n');
}

describe('historyToCsv', () => {
  it('writes a header plus one row per run, metrics as sorted union columns', () => {
    const lines = parse(historyToCsv([passed, failed, noVerdict]));

    expect(lines).toEqual([
      'id,testType,module,status,verdict,startedAt,completedAt,durationMs,error,frame_size,latency_us,loss_pct,throughput_mbps',
      '2026-08-17T10:00:00Z-RFC 2544,RFC 2544,benchmark,completed,PASS,2026-08-17T09:58:00Z,2026-08-17T10:00:00Z,120000,,1518,,,940',
      '2026-08-17T11:00:00Z-Y.1564,Y.1564,servicetest,error,FAIL,,2026-08-17T11:00:00Z,,"peer unreachable, ""no reply""",,-1,-0.5,',
      'no-verdict,Traffic,trafficgen,stopped,,,,,,,,,',
    ]);
  });

  it('leaves the verdict empty for a run that recorded none, never PASS', () => {
    const [, row] = parse(historyToCsv([noVerdict]));

    expect(row.split(',')[4]).toBe('');
  });

  it('keeps a formula-shaped string as text but a negative number as a number', () => {
    const lines = parse(
      historyToCsv([
        { ...noVerdict, error: '=HYPERLINK("http://x")', metrics: { note: '@SUM(A1)', d: '-3' } },
      ]),
    );

    expect(lines[0]).toBe(
      'id,testType,module,status,verdict,startedAt,completedAt,durationMs,error,d,note',
    );
    expect(lines[1]).toBe(
      `no-verdict,Traffic,trafficgen,stopped,,,,,"'=HYPERLINK(""http://x"")",-3,'@SUM(A1)`,
    );
  });

  it('quotes a cell that carries a line break', () => {
    const csv = historyToCsv([{ ...noVerdict, error: 'line one\nline two' }]);

    expect(csv).toContain(',"line one\nline two"\r\n');
  });

  it('writes only the header when there is no history', () => {
    expect(parse(historyToCsv([]))).toEqual([
      'id,testType,module,status,verdict,startedAt,completedAt,durationMs,error',
    ]);
  });
});

describe('historyToJson', () => {
  it('round-trips the stored records unchanged', () => {
    expect(JSON.parse(historyToJson([passed, failed, noVerdict]))).toEqual([
      passed,
      failed,
      noVerdict,
    ]);
  });
});

describe('exportFileName', () => {
  it('stamps the local date and time with zero padding', () => {
    expect(exportFileName(new Date(2026, 0, 5, 7, 8, 9), 'csv')).toBe(
      'stem-history-20260105-070809.csv',
    );
  });
});
