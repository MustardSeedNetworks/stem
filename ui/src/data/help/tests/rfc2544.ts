/**
 * @fileoverview The Stem - RFC 2544 Test Definitions
 * @description Help content for RFC 2544 benchmarking tests.
 */

import type { TestHelp } from '../types';

export const rfc2544Tests: Record<string, TestHelp> = {
  throughput: {
    id: 'throughput',
    name: 'Throughput Test',
    standard: 'RFC 2544 Section 26.1',
    category: 'RFC 2544',
    summary: 'Finds the maximum speed your network can handle without dropping packets.',
    techDesc: `The throughput test uses binary search to determine the maximum rate at which
the DUT (Device Under Test) can forward frames without any frame loss. Starting at the
theoretical maximum rate, the test iteratively adjusts the offered load based on whether
frames were lost, converging on the maximum lossless rate.`,
    laymanDesc: `Think of your network like a highway. This test finds out how many cars
(data packets) can travel on it before traffic jams (packet loss) start happening.
It keeps increasing traffic until packets start getting dropped, then backs off to find
the sweet spot.`,
    whenToUse: `• Validating new network equipment before deployment
• Troubleshooting slow network performance
• Verifying ISP is delivering promised bandwidth
• Baseline testing after configuration changes`,
    whenNotToUse: `• If you need latency measurements (use Latency test)
• For TCP application performance (use RFC 6349)
• For switch MAC table testing (use RFC 2889)`,
    parameters: [
      {
        name: 'Frame Sizes',
        flag: '--frame-sizes',
        type: 'comma-separated integers',
        defaultValue: '64,128,256,512,1024,1280,1518',
        required: false,
        techDesc: 'Ethernet frame sizes in bytes to test',
        laymanDesc:
          'Different packet sizes to try - small packets stress the network differently than large ones',
        example: '--frame-sizes 64,512,1518',
      },
      {
        name: 'Duration',
        flag: '--duration',
        type: 'integer (seconds)',
        defaultValue: '60',
        required: false,
        techDesc: 'Duration of each trial iteration',
        laymanDesc: 'How long to run each speed test',
        example: '--duration 30',
      },
      {
        name: 'Resolution',
        flag: '--resolution',
        type: 'float (percentage)',
        defaultValue: '0.1',
        required: false,
        techDesc: 'Binary search resolution as percentage of line rate',
        laymanDesc: 'How precisely to find the maximum speed (smaller = more precise but slower)',
        example: '--resolution 0.5',
      },
      {
        name: 'Max Loss',
        flag: '--max-loss',
        type: 'float (percentage)',
        defaultValue: '0.0',
        required: false,
        techDesc: 'Maximum acceptable frame loss percentage',
        laymanDesc: 'How much packet loss is acceptable (0% means no loss allowed)',
        example: '--max-loss 0.001',
      },
      {
        name: 'Warmup',
        flag: '--warmup',
        type: 'integer (seconds)',
        defaultValue: '2',
        required: false,
        techDesc: 'Warmup period before measurements begin',
        laymanDesc: 'Time to let the network stabilize before measuring',
        example: '--warmup 5',
      },
      {
        name: 'Trials',
        flag: '--trials',
        type: 'integer',
        defaultValue: '3',
        required: false,
        techDesc: 'Number of trial iterations per test point',
        laymanDesc: 'How many times to repeat each measurement for accuracy',
        example: '--trials 5',
      },
      {
        name: 'Step Size',
        flag: '--step-size',
        type: 'float (percentage)',
        defaultValue: '10.0',
        required: false,
        techDesc: 'Frame loss rate step size for testing',
        laymanDesc: 'How much to change the speed between tests',
        example: '--step-size 5.0',
      },
      {
        name: 'Bidirectional',
        flag: '--bidirectional',
        type: 'boolean',
        defaultValue: 'false',
        required: false,
        techDesc: 'Run tests in both directions simultaneously',
        laymanDesc: 'Test upload and download at the same time',
        example: '--bidirectional',
      },
    ],
    metrics: [
      {
        name: 'Max Rate',
        unit: '% of line rate',
        goodRange: '>95% is excellent, >80% is acceptable',
        badMeaning: 'Below 80% indicates a bottleneck or configuration issue',
      },
      {
        name: 'Throughput',
        unit: 'Mbps or Gbps',
        goodRange: 'Close to rated interface speed',
        badMeaning: 'Significantly below rated speed indicates problem',
      },
    ],
    passCriteria: 'Zero frame loss at the reported throughput rate',
    failMeaning: 'Unable to achieve any rate without frame loss',
    examples: [
      {
        desc: 'Basic throughput test',
        command: 'stem test -i eth0 -t throughput',
        output: 'Max Rate: 98.5% (985 Mbps)',
      },
    ],
    tips: [
      'Run multiple iterations for production validation',
      'Small frames (64 bytes) stress packet processing; large frames test raw bandwidth',
    ],
    seeAlso: ['latency', 'frame_loss', 'y1564_config'],
  },

  latency: {
    id: 'latency',
    name: 'Latency Test',
    standard: 'RFC 2544 Section 26.2',
    category: 'RFC 2544',
    summary: 'Measures round-trip delay time for packets at various throughput levels.',
    techDesc: `The latency test measures the time required for a frame to travel from the
originating device through the DUT and back. This is performed at the throughput rate
determined by the throughput test.`,
    laymanDesc: `This test measures "lag" - how long it takes for a message to get from
point A to point B and back. Lower numbers are better:
• Under 1ms: Excellent (good for video calls, gaming)
• 1-10ms: Good for most applications
• Over 50ms: May cause noticeable delays`,
    whenToUse: `• Validating low-latency network requirements
• VoIP and video conferencing quality assurance
• Real-time applications testing`,
    whenNotToUse: `• If you only need bandwidth measurements
• For packet loss analysis at various rates`,
    parameters: [
      {
        name: 'Duration',
        flag: '--duration',
        type: 'integer (seconds)',
        defaultValue: '120',
        required: false,
        techDesc: 'Test duration for statistical accuracy',
        laymanDesc: 'How long to collect measurements',
        example: '--duration 60',
      },
    ],
    metrics: [
      {
        name: 'Average Latency',
        unit: 'microseconds',
        goodRange: '<1000µs for most applications',
        badMeaning: 'High latency indicates congestion or distance',
      },
      {
        name: 'Jitter',
        unit: 'microseconds',
        goodRange: '<100µs for voice/video',
        badMeaning: 'High jitter causes quality issues',
      },
    ],
    passCriteria: 'Latency within acceptable range for application',
    failMeaning: 'Network may not be suitable for latency-sensitive apps',
    examples: [
      {
        desc: 'Basic latency test',
        command: 'stem test -i eth0 -t latency',
        output: 'Avg: 125µs, Jitter: 23µs',
      },
    ],
    tips: ['Test at multiple rates to understand how latency changes with load'],
    seeAlso: ['throughput', 'frame_delay'],
  },

  frame_loss: {
    id: 'frame_loss',
    name: 'Frame Loss Rate Test',
    standard: 'RFC 2544 Section 26.3',
    category: 'RFC 2544',
    summary: 'Measures what percentage of packets are lost at different network speeds.',
    techDesc: `The frame loss rate test determines the percentage of frames not forwarded
by the DUT at various offered loads, starting at 100% and decreasing until zero loss.`,
    laymanDesc: `This test answers: "How many packets get lost as I push more traffic
through the network?" It creates a stress curve showing when loss starts happening.`,
    whenToUse: `• Understanding network behavior under overload
• Capacity planning and upgrade justification
• Comparing equipment performance`,
    whenNotToUse: `• For finding maximum lossless rate (use Throughput)
• For latency analysis`,
    parameters: [],
    metrics: [
      {
        name: 'Loss Rate',
        unit: 'percentage',
        goodRange: '0% at operating rate',
        badMeaning: 'Any loss at normal rates is problematic',
      },
    ],
    passCriteria: 'Zero loss at planned operating rate',
    failMeaning: 'Network cannot sustain planned traffic levels',
    examples: [
      {
        desc: 'Frame loss test',
        command: 'stem test -i eth0 -t frame_loss',
        output: '100%: 2.3% loss, 80%: 0% loss',
      },
    ],
    tips: ['Use results to set traffic engineering thresholds'],
    seeAlso: ['throughput'],
  },

  back_to_back: {
    id: 'back_to_back',
    name: 'Back-to-Back Frames Test',
    standard: 'RFC 2544 Section 26.4',
    category: 'RFC 2544',
    summary: 'Measures how many packets can be sent in a burst without any loss.',
    techDesc: `The back-to-back frames test measures the maximum number of frames that can be
transmitted at minimum inter-frame gap before a frame is lost.`,
    laymanDesc: `This test measures "burst capacity" - how much data can be sent all at once.
Higher numbers are better for handling waves of data like video streams.`,
    whenToUse: `• Buffer sizing validation
• Video streaming infrastructure
• Burst traffic applications`,
    whenNotToUse: '• For sustained throughput (use Throughput test)',
    parameters: [],
    metrics: [
      {
        name: 'Burst Size',
        unit: 'frames',
        goodRange: 'Depends on requirements',
        badMeaning: 'Small burst size may cause issues with bursty traffic',
      },
    ],
    passCriteria: 'Burst capacity meets application requirements',
    failMeaning: 'May experience drops during traffic bursts',
    examples: [
      {
        desc: 'Back-to-back test',
        command: 'stem test -i eth0 -t back_to_back',
        output: 'Max burst: 2048 frames',
      },
    ],
    tips: ['Results indicate effective buffer size of the DUT'],
    seeAlso: ['throughput', 'congestion'],
  },

  system_recovery: {
    id: 'system_recovery',
    name: 'System Recovery Test',
    standard: 'RFC 2544 Section 26.5',
    category: 'RFC 2544',
    summary: 'Measures how quickly the network recovers after being overloaded.',
    techDesc: `The system recovery test measures how long a DUT takes to recover from an
overload condition by transmitting at 110% of max throughput then reducing to 50%.`,
    laymanDesc: `After your network gets overwhelmed, how long does it take to get back to normal?
Fast recovery (under 1 second) is good.`,
    whenToUse: `• Mission-critical network validation
• Understanding DUT behavior after congestion`,
    whenNotToUse: '• For normal operating conditions',
    parameters: [],
    metrics: [
      {
        name: 'Recovery Time',
        unit: 'milliseconds',
        goodRange: '<1000ms',
        badMeaning: 'Long recovery impacts user experience',
      },
    ],
    passCriteria: 'Recovery time within acceptable limits',
    failMeaning: 'DUT may cause extended impact after congestion',
    examples: [
      {
        desc: 'Recovery test',
        command: 'stem test -i eth0 -t system_recovery',
        output: 'Recovery time: 245ms',
      },
    ],
    tips: [],
    seeAlso: ['throughput', 'reset'],
  },

  reset: {
    id: 'reset',
    name: 'Reset Test',
    standard: 'RFC 2544 Section 26.6',
    category: 'RFC 2544',
    summary: 'Measures how long the device takes to recover from a reset.',
    techDesc: `The reset test measures the time required for a DUT to recover from
hardware or software reset events.`,
    laymanDesc: `When network equipment restarts, how long is the network down?
Lower reset times mean less disruption during maintenance.`,
    whenToUse: `• Maintenance window planning
• High-availability architecture design`,
    whenNotToUse: '• For normal performance testing',
    parameters: [],
    metrics: [
      {
        name: 'Reset Time',
        unit: 'seconds',
        goodRange: '<60s for most equipment',
        badMeaning: 'Long reset times impact availability SLAs',
      },
    ],
    passCriteria: 'Reset time meets availability requirements',
    failMeaning: 'Equipment restart takes too long',
    examples: [
      {
        desc: 'Reset test',
        command: 'stem test -i eth0 -t reset',
        output: 'Reset time: 45 seconds',
      },
    ],
    tips: [],
    seeAlso: ['system_recovery'],
  },
};
