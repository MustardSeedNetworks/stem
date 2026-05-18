/**
 * @fileoverview The Stem - Help Content for WebUI
 * @description This file mirrors the Go pkg/help content for use in the React UI.
 *              Both technical and layman-friendly explanations are included.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

export interface TestHelp {
  id: string;
  name: string;
  standard: string;
  category: string;
  summary: string;
  techDesc: string;
  laymanDesc: string;
  whenToUse: string;
  whenNotToUse: string;
  parameters: Parameter[];
  metrics: Metric[];
  passCriteria: string;
  failMeaning: string;
  examples: Example[];
  tips: string[];
  seeAlso: string[];
}

export interface Parameter {
  name: string;
  flag: string;
  type: string;
  defaultValue: string;
  required: boolean;
  techDesc: string;
  laymanDesc: string;
  example: string;
}

export interface Metric {
  name: string;
  unit: string;
  goodRange: string;
  badMeaning: string;
}

export interface Example {
  desc: string;
  command: string;
  output?: string;
}

export interface Category {
  id: string;
  name: string;
  fullName: string;
  summary: string;
  description: string;
  tests: string[];
  whenToUse: string;
}

// Test Categories
export const categories: Record<string, Category> = {
  rfc2544: {
    id: 'rfc2544',
    name: 'RFC 2544',
    fullName: 'Benchmarking Methodology for Network Interconnect Devices',
    summary: 'The standard tests for measuring raw network equipment performance.',
    description: `RFC 2544 defines benchmarking methodology for network devices.
These tests measure fundamental performance characteristics: throughput, latency,
frame loss, and burst handling. Use these tests for equipment validation and comparison.`,
    tests: ['throughput', 'latency', 'frame_loss', 'back_to_back', 'system_recovery', 'reset'],
    whenToUse: 'Equipment benchmarking, performance validation, comparing vendors',
  },
  y1564: {
    id: 'y1564',
    name: 'Y.1564',
    fullName: 'Ethernet Service Activation Test Methodology',
    summary: 'The carrier standard for turning up ethernet services.',
    description: `ITU-T Y.1564 defines the methodology for activating and validating
carrier ethernet services. These tests verify that a service meets its SLA parameters
at progressive load levels and over extended duration.`,
    tests: ['y1564_config', 'y1564_performance', 'y1564_full'],
    whenToUse: 'Carrier service activation, SLA validation, service acceptance',
  },
  rfc2889: {
    id: 'rfc2889',
    name: 'RFC 2889',
    fullName: 'Benchmarking Methodology for LAN Switching Devices',
    summary: 'Tests specifically for switch/bridge performance characteristics.',
    description: `RFC 2889 extends RFC 2544 for testing LAN switches. These tests
measure switch-specific characteristics like forwarding rate across multiple ports,
MAC address table capacity, learning rate, and congestion handling.`,
    tests: ['forwarding', 'address_cache', 'learning_rate', 'broadcast', 'congestion'],
    whenToUse: 'Switch validation, data center planning, MAC table capacity verification',
  },
  rfc6349: {
    id: 'rfc6349',
    name: 'RFC 6349',
    fullName: 'Framework for TCP Throughput Testing',
    summary: 'Tests that measure real TCP application performance.',
    description: `RFC 6349 provides methodology for testing TCP throughput, which
represents actual application performance. These tests measure achievable TCP throughput
and help identify network factors affecting TCP performance.`,
    tests: ['tcp_throughput', 'path_analysis'],
    whenToUse: 'Application performance testing, WAN optimization, TCP troubleshooting',
  },
  y1731: {
    id: 'y1731',
    name: 'Y.1731',
    fullName: 'OAM Functions and Mechanisms for Ethernet Networks',
    summary: 'Operations, Administration, and Maintenance for carrier ethernet.',
    description: `ITU-T Y.1731 defines OAM functions for monitoring and maintaining
ethernet services. These tools provide in-service monitoring capabilities including
delay measurement, loss measurement, and connectivity verification.`,
    tests: ['frame_delay', 'y1731_frame_loss', 'synthetic_loss', 'loopback'],
    whenToUse: 'Production monitoring, SLA verification, fault isolation',
  },
  mef: {
    id: 'mef',
    name: 'MEF',
    fullName: 'Metro Ethernet Forum Service Tests',
    summary: 'Industry standard tests for carrier ethernet services.',
    description: `MEF (Metro Ethernet Forum) defines service specifications and
testing methodologies for carrier ethernet. These tests validate services against
MEF specifications including bandwidth profiles and Class of Service.`,
    tests: ['mef_config', 'mef_performance', 'mef_full'],
    whenToUse: 'MEF-certified service validation, multi-CoS testing, carrier acceptance',
  },
  tsn: {
    id: 'tsn',
    name: 'TSN',
    fullName: 'Time-Sensitive Networking',
    summary: 'Tests for deterministic, time-critical industrial networks.',
    description: `IEEE 802.1 Time-Sensitive Networking tests validate networks
requiring deterministic timing. These tests verify that time-aware shaping, traffic
isolation, and scheduled latency meet industrial automation requirements.`,
    tests: ['gate_timing', 'traffic_isolation', 'scheduled_latency', 'tsn_full'],
    whenToUse: 'Industrial automation, automotive ethernet, deterministic networking',
  },
  trafficgen: {
    id: 'trafficgen',
    name: 'TrafficGen',
    fullName: 'Custom Traffic Generation',
    summary: 'Generate custom traffic patterns for specialized testing scenarios.',
    description: `Traffic generation tools for creating custom test patterns.
Use these when standard tests (RFC 2544, Y.1564) don't cover your specific
requirements. Supports burst mode, VLAN tagging, and controlled rates.`,
    tests: ['custom_stream'],
    whenToUse: 'Custom stress testing, QoS validation, network diagnostics',
  },
};

// All test definitions
export const tests: Record<string, TestHelp> = {
  // RFC 2544 Tests
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

  // Y.1564 Tests
  y1564_config: {
    id: 'y1564_config',
    name: 'Y.1564 Service Configuration Test',
    standard: 'ITU-T Y.1564',
    category: 'Y.1564',
    summary: 'Validates carrier ethernet service at 25%, 50%, 75%, and 100% of committed rate.',
    techDesc: `The Y.1564 Service Configuration Test validates that an Ethernet service
meets its SLA parameters at progressive load steps (25%, 50%, 75%, 100% of CIR).`,
    laymanDesc: `When you buy an ethernet service from a carrier, this test verifies
you're getting what you paid for at different load levels.`,
    whenToUse: `• New service activation
• Service verification after maintenance
• SLA dispute resolution`,
    whenNotToUse: '• For raw equipment benchmarking (use RFC 2544)',
    parameters: [
      {
        name: 'CIR',
        flag: '--cir',
        type: 'float (Mbps)',
        defaultValue: '1000',
        required: true,
        techDesc: 'Committed Information Rate',
        laymanDesc: 'The speed your contract guarantees',
        example: '--cir 100',
      },
      {
        name: 'EIR',
        flag: '--eir',
        type: 'float (Mbps)',
        defaultValue: '0',
        required: false,
        techDesc: 'Excess Information Rate - bandwidth above CIR that may be available',
        laymanDesc: 'Extra burst speed you might get when network is not busy',
        example: '--eir 50',
      },
      {
        name: 'CBS',
        flag: '--cbs',
        type: 'integer (KB)',
        defaultValue: '0',
        required: false,
        techDesc: 'Committed Burst Size - maximum burst at CIR',
        laymanDesc: 'How much data can be sent in a burst at guaranteed speed',
        example: '--cbs 12',
      },
      {
        name: 'EBS',
        flag: '--ebs',
        type: 'integer (KB)',
        defaultValue: '0',
        required: false,
        techDesc: 'Excess Burst Size - maximum burst at EIR',
        laymanDesc: 'How much data can be sent in a burst at excess speed',
        example: '--ebs 12',
      },
      {
        name: 'Frame Sizes',
        flag: '--frame-sizes',
        type: 'comma-separated integers',
        defaultValue: '64,512,1518',
        required: false,
        techDesc: 'Ethernet frame sizes in bytes to test',
        laymanDesc: 'Different packet sizes to validate service with',
        example: '--frame-sizes 64,1518',
      },
      {
        name: 'Step Duration',
        flag: '--step-duration',
        type: 'integer (seconds)',
        defaultValue: '60',
        required: false,
        techDesc: 'Duration of each configuration step (25%, 50%, 75%, 100%)',
        laymanDesc: 'How long to test at each speed level',
        example: '--step-duration 30',
      },
      {
        name: 'VLAN ID',
        flag: '--vlan-id',
        type: 'integer (0-4095)',
        defaultValue: '0',
        required: false,
        techDesc: 'VLAN identifier for tagged traffic',
        laymanDesc: 'Virtual network ID if your service uses VLANs',
        example: '--vlan-id 100',
      },
      {
        name: 'PCP',
        flag: '--pcp',
        type: 'integer (0-7)',
        defaultValue: '0',
        required: false,
        techDesc: 'Priority Code Point for 802.1p CoS marking',
        laymanDesc: 'Traffic priority level (7 is highest)',
        example: '--pcp 5',
      },
      {
        name: 'Color Aware',
        flag: '--color-aware',
        type: 'boolean',
        defaultValue: 'false',
        required: false,
        techDesc: 'Enable color-aware traffic conditioning',
        laymanDesc: 'Test color marking for traffic policing',
        example: '--color-aware',
      },
      {
        name: 'FLR Threshold',
        flag: '--flr-threshold',
        type: 'float (percentage)',
        defaultValue: '0.0',
        required: false,
        techDesc: 'Frame Loss Ratio acceptance threshold',
        laymanDesc: 'Maximum acceptable packet loss percentage',
        example: '--flr-threshold 0.01',
      },
      {
        name: 'FD Threshold',
        flag: '--fd-threshold',
        type: 'float (ms)',
        defaultValue: '10.0',
        required: false,
        techDesc: 'Frame Delay acceptance threshold in milliseconds',
        laymanDesc: 'Maximum acceptable delay',
        example: '--fd-threshold 5.0',
      },
      {
        name: 'FDV Threshold',
        flag: '--fdv-threshold',
        type: 'float (ms)',
        defaultValue: '5.0',
        required: false,
        techDesc: 'Frame Delay Variation acceptance threshold in milliseconds',
        laymanDesc: 'Maximum acceptable jitter',
        example: '--fdv-threshold 2.0',
      },
    ],
    metrics: [
      {
        name: 'IR (Information Rate)',
        unit: 'Mbps',
        goodRange: 'Within 1% of configured CIR',
        badMeaning: 'Service not delivering promised bandwidth',
      },
      {
        name: 'FD (Frame Delay)',
        unit: 'milliseconds',
        goodRange: 'Below threshold at all steps',
        badMeaning: 'Exceeds SLA delay commitment',
      },
      {
        name: 'FLR (Frame Loss Ratio)',
        unit: 'percentage',
        goodRange: 'Below configured threshold',
        badMeaning: 'Packets being lost above acceptable level',
      },
      {
        name: 'FDV (Frame Delay Variation)',
        unit: 'milliseconds',
        goodRange: 'Below configured threshold',
        badMeaning: 'Jitter exceeds SLA commitment',
      },
    ],
    passCriteria: 'All metrics within thresholds at all four CIR steps',
    failMeaning: 'Service does not meet SLA',
    examples: [
      {
        desc: 'Test 100 Mbps service',
        command: 'stem test -i eth0 -t y1564_config --cir 100',
        output: 'All steps: PASS',
      },
    ],
    tips: ['CIR should match your contract exactly'],
    seeAlso: ['y1564_performance', 'mef_config'],
  },

  y1564_performance: {
    id: 'y1564_performance',
    name: 'Y.1564 Service Performance Test',
    standard: 'ITU-T Y.1564',
    category: 'Y.1564',
    summary: 'Extended duration test to validate service quality over time.',
    techDesc: `The Y.1564 Service Performance Test validates sustained performance
over an extended period (typically 15 minutes to hours).`,
    laymanDesc: `After passing the initial speed test, can your network connection
maintain that performance for hours?`,
    whenToUse: `• After passing Configuration Test
• Extended burn-in testing`,
    whenNotToUse: '• Initial service turn-up (do Config Test first)',
    parameters: [
      {
        name: 'Duration',
        flag: '--duration',
        type: 'integer (minutes)',
        defaultValue: '15',
        required: false,
        techDesc: 'Test duration in minutes',
        laymanDesc: 'How long to run the test',
        example: '--duration 60',
      },
    ],
    metrics: [
      {
        name: 'Sustained Rate',
        unit: 'Mbps',
        goodRange: 'Within 1% of CIR for entire duration',
        badMeaning: 'Performance degrades over time',
      },
    ],
    passCriteria: 'All metrics within thresholds for entire duration',
    failMeaning: 'Service shows instability over time',
    examples: [
      {
        desc: '15-minute test',
        command: 'stem test -i eth0 -t y1564_performance --cir 100',
        output: 'Performance stable',
      },
    ],
    tips: [],
    seeAlso: ['y1564_config'],
  },

  y1564_full: {
    id: 'y1564_full',
    name: 'Y.1564 Full SAC Test',
    standard: 'ITU-T Y.1564',
    category: 'Y.1564',
    summary: 'Complete Service Activation Test - Configuration followed by Performance.',
    techDesc: `The Full SAC Test combines both Configuration and Performance tests
into a complete validation sequence.`,
    laymanDesc: `The complete, official test for verifying a carrier ethernet service.
This is what carriers use to officially "turn up" a new service.`,
    whenToUse: `• Official service activation
• Complete service validation`,
    whenNotToUse: '• Quick troubleshooting',
    parameters: [],
    metrics: [],
    passCriteria: 'Both Configuration and Performance tests pass',
    failMeaning: 'Service does not meet acceptance criteria',
    examples: [
      {
        desc: 'Full SAC test',
        command: 'stem test -i eth0 -t y1564_full --cir 100',
        output: 'Service Accepted',
      },
    ],
    tips: [],
    seeAlso: ['y1564_config', 'y1564_performance'],
  },

  // Add remaining tests with shorter entries for space
  forwarding: {
    id: 'forwarding',
    name: 'Forwarding Rate Test',
    standard: 'RFC 2889 Section 5.2',
    category: 'RFC 2889',
    summary: 'Measures how fast a switch can move packets between ports.',
    techDesc: 'Measures maximum forwarding rate across multiple ports.',
    laymanDesc: 'How fast can your switch shuffle packets between all its ports at once?',
    whenToUse: 'Switch aggregate capacity testing',
    whenNotToUse: 'Single port-pair testing',
    parameters: [],
    metrics: [],
    passCriteria: 'Aggregate throughput meets switch specifications',
    failMeaning: 'Switch fabric bottleneck',
    examples: [],
    tips: [],
    seeAlso: ['throughput'],
  },

  address_cache: {
    id: 'address_cache',
    name: 'Address Caching Capacity Test',
    standard: 'RFC 2889 Section 5.5',
    category: 'RFC 2889',
    summary: 'Determines how many MAC addresses a switch can remember.',
    techDesc: 'Tests MAC address table capacity while maintaining forwarding.',
    laymanDesc: 'How many devices can this switch keep track of?',
    whenToUse: 'Large campus network planning',
    whenNotToUse: 'Small networks',
    parameters: [],
    metrics: [],
    passCriteria: 'Address capacity meets deployment requirements',
    failMeaning: 'May need larger MAC table',
    examples: [],
    tips: [],
    seeAlso: ['learning_rate'],
  },

  learning_rate: {
    id: 'learning_rate',
    name: 'Address Learning Rate Test',
    standard: 'RFC 2889 Section 5.6',
    category: 'RFC 2889',
    summary: 'Measures how fast a switch can learn new device addresses.',
    techDesc: 'Tests how quickly a switch populates its MAC table.',
    laymanDesc: 'When new devices connect, how fast can the switch register them?',
    whenToUse: 'Highly dynamic environments',
    whenNotToUse: 'Static networks',
    parameters: [],
    metrics: [],
    passCriteria: 'Learning rate meets requirements',
    failMeaning: 'May cause delays for new devices',
    examples: [],
    tips: [],
    seeAlso: ['address_cache'],
  },

  broadcast: {
    id: 'broadcast',
    name: 'Broadcast Frame Handling Test',
    standard: 'RFC 2889 Section 5.7',
    category: 'RFC 2889',
    summary: 'Tests how the switch handles broadcast traffic.',
    techDesc: 'Measures broadcast handling and impact on unicast.',
    laymanDesc: 'How does your switch handle messages to ALL devices?',
    whenToUse: 'Networks with broadcast-heavy protocols',
    whenNotToUse: 'Point-to-point testing',
    parameters: [],
    metrics: [],
    passCriteria: 'Minimal impact on unicast',
    failMeaning: 'Broadcasts affecting normal traffic',
    examples: [],
    tips: [],
    seeAlso: ['forwarding'],
  },

  congestion: {
    id: 'congestion',
    name: 'Congestion Control Test',
    standard: 'RFC 2889 Section 5.8',
    category: 'RFC 2889',
    summary: 'Tests switch behavior when ports are oversubscribed.',
    techDesc: 'Characterizes queuing, dropping, and fairness during congestion.',
    laymanDesc: 'What happens when too much traffic tries to go to the same place?',
    whenToUse: 'Server farm switch validation',
    whenNotToUse: 'Non-blocking architectures',
    parameters: [],
    metrics: [],
    passCriteria: 'Fair distribution, no HOL blocking',
    failMeaning: 'Unfair bandwidth allocation',
    examples: [],
    tips: [],
    seeAlso: ['forwarding'],
  },

  tcp_throughput: {
    id: 'tcp_throughput',
    name: 'TCP Throughput Test',
    standard: 'RFC 6349',
    category: 'RFC 6349',
    summary: 'Measures real application throughput using TCP.',
    techDesc: 'Tests TCP throughput accounting for protocol behavior.',
    laymanDesc: 'This measures REAL download/upload speeds you actually experience.',
    whenToUse: 'Application performance troubleshooting',
    whenNotToUse: 'Layer 2 equipment testing',
    parameters: [],
    metrics: [],
    passCriteria: 'TCP throughput meets requirements',
    failMeaning: 'Network may need optimization',
    examples: [],
    tips: [],
    seeAlso: ['path_analysis', 'throughput'],
  },

  path_analysis: {
    id: 'path_analysis',
    name: 'Path Analysis Test',
    standard: 'RFC 6349',
    category: 'RFC 6349',
    summary: "Analyzes what's limiting your network speed.",
    techDesc: 'Characterizes RTT, bottleneck bandwidth, and BDP.',
    laymanDesc: 'Answers WHY your connection is slow, not just HOW slow.',
    whenToUse: 'TCP troubleshooting',
    whenNotToUse: 'If you just need throughput numbers',
    parameters: [],
    metrics: [],
    passCriteria: 'Identifies optimization opportunities',
    failMeaning: 'N/A - diagnostic test',
    examples: [],
    tips: [],
    seeAlso: ['tcp_throughput'],
  },

  frame_delay: {
    id: 'frame_delay',
    name: 'Frame Delay Measurement',
    standard: 'ITU-T Y.1731',
    category: 'Y.1731',
    summary: 'Precise one-way and two-way delay measurements using OAM.',
    techDesc: 'Y.1731 DMM/DMR for precise delay measurement.',
    laymanDesc: 'Super-precise timing measurements for carrier networks.',
    whenToUse: 'SLA monitoring in production',
    whenNotToUse: 'Initial service turn-up',
    parameters: [],
    metrics: [],
    passCriteria: 'Delay within SLA',
    failMeaning: 'Exceeds SLA threshold',
    examples: [],
    tips: [],
    seeAlso: ['latency'],
  },

  y1731_frame_loss: {
    id: 'y1731_frame_loss',
    name: 'Frame Loss Measurement',
    standard: 'ITU-T Y.1731',
    category: 'Y.1731',
    summary: 'Monitors packet loss on production carrier networks.',
    techDesc: 'Y.1731 LMM/LMR for continuous loss monitoring.',
    laymanDesc: 'Continuously monitors if packets are being lost without disrupting traffic.',
    whenToUse: 'Continuous service monitoring',
    whenNotToUse: 'Initial service testing',
    parameters: [],
    metrics: [],
    passCriteria: 'Loss within SLA',
    failMeaning: 'SLA may be violated',
    examples: [],
    tips: [],
    seeAlso: ['frame_loss'],
  },

  synthetic_loss: {
    id: 'synthetic_loss',
    name: 'Synthetic Loss Measurement',
    standard: 'ITU-T Y.1731',
    category: 'Y.1731',
    summary: 'Continuous reliability monitoring using test signals.',
    techDesc: 'SLM/SLR for loss measurement independent of user traffic.',
    laymanDesc: 'Sends special test signals to continuously check network health.',
    whenToUse: 'Links with variable traffic',
    whenNotToUse: 'High-traffic links',
    parameters: [],
    metrics: [],
    passCriteria: '0% synthetic loss',
    failMeaning: 'Network path has problems',
    examples: [],
    tips: [],
    seeAlso: ['y1731_frame_loss'],
  },

  loopback: {
    id: 'loopback',
    name: 'Loopback Test',
    standard: 'ITU-T Y.1731',
    category: 'Y.1731',
    summary: 'Quick connectivity check using OAM loopback.',
    techDesc: 'Y.1731 LBM/LBR for connectivity verification.',
    laymanDesc: 'A "ping" for carrier ethernet networks.',
    whenToUse: 'Quick connectivity verification',
    whenNotToUse: 'Performance testing',
    parameters: [],
    metrics: [],
    passCriteria: 'Response received',
    failMeaning: 'Connectivity problem',
    examples: [],
    tips: [],
    seeAlso: ['frame_delay'],
  },

  mef_config: {
    id: 'mef_config',
    name: 'MEF Service Configuration Test',
    standard: 'MEF 14/48',
    category: 'MEF',
    summary: 'Validates carrier ethernet service per MEF standards.',
    techDesc: 'Tests bandwidth profiles and CoS per MEF specifications.',
    laymanDesc: 'The official carrier ethernet validation per industry standards.',
    whenToUse: 'MEF-certified service validation',
    whenNotToUse: 'Simple single-class services',
    parameters: [],
    metrics: [],
    passCriteria: 'Bandwidth and CoS compliance',
    failMeaning: 'Service not meeting MEF specs',
    examples: [],
    tips: [],
    seeAlso: ['y1564_config'],
  },

  mef_performance: {
    id: 'mef_performance',
    name: 'MEF Performance Test',
    standard: 'MEF 14/48',
    category: 'MEF',
    summary: 'Extended MEF service quality validation.',
    techDesc: 'Extended duration tests per MEF specifications.',
    laymanDesc: 'Long-running test for carrier service quality.',
    whenToUse: 'After MEF Config test passes',
    whenNotToUse: 'Quick spot checks',
    parameters: [],
    metrics: [],
    passCriteria: 'Performance within specs',
    failMeaning: 'Service shows instability',
    examples: [],
    tips: [],
    seeAlso: ['mef_config'],
  },

  mef_full: {
    id: 'mef_full',
    name: 'MEF Full Test Suite',
    standard: 'MEF 14/48',
    category: 'MEF',
    summary: 'Complete MEF service validation.',
    techDesc: 'Complete MEF validation sequence.',
    laymanDesc: 'The complete MEF certification test.',
    whenToUse: 'Official MEF service acceptance',
    whenNotToUse: 'Troubleshooting',
    parameters: [],
    metrics: [],
    passCriteria: 'All MEF tests pass',
    failMeaning: 'Service not MEF compliant',
    examples: [],
    tips: [],
    seeAlso: ['mef_config', 'mef_performance'],
  },

  gate_timing: {
    id: 'gate_timing',
    name: 'TSN Gate Timing Test',
    standard: 'IEEE 802.1Qbv',
    category: 'TSN',
    summary: 'Verifies Time-Aware Shaper gate timing accuracy.',
    techDesc: 'Validates TAS gates open and close at correct times per IEEE 802.1Qbv.',
    laymanDesc: 'Verifies network "time gates" for industrial automation.',
    whenToUse: 'Industrial automation networks',
    whenNotToUse: 'Traditional IT networks',
    parameters: [
      {
        name: 'Duration',
        flag: '--duration',
        type: 'integer (seconds)',
        defaultValue: '60',
        required: false,
        techDesc: 'Test duration for timing validation',
        laymanDesc: 'How long to run the timing test',
        example: '--duration 120',
      },
      {
        name: 'Warmup',
        flag: '--warmup',
        type: 'integer (seconds)',
        defaultValue: '5',
        required: false,
        techDesc: 'Warmup period for PTP synchronization',
        laymanDesc: 'Time to let timing stabilize before measuring',
        example: '--warmup 10',
      },
      {
        name: 'Frame Size',
        flag: '--frame-size',
        type: 'integer (bytes)',
        defaultValue: '64',
        required: false,
        techDesc: 'Ethernet frame size for TSN test traffic',
        laymanDesc: 'Size of test packets',
        example: '--frame-size 128',
      },
      {
        name: 'Max Latency',
        flag: '--max-latency-ns',
        type: 'integer (nanoseconds)',
        defaultValue: '1000000',
        required: false,
        techDesc: 'Maximum acceptable latency in nanoseconds',
        laymanDesc: 'Highest allowed delay (in billionths of a second)',
        example: '--max-latency-ns 500000',
      },
      {
        name: 'Max Jitter',
        flag: '--max-jitter-ns',
        type: 'integer (nanoseconds)',
        defaultValue: '100000',
        required: false,
        techDesc: 'Maximum acceptable jitter in nanoseconds',
        laymanDesc: 'Highest allowed timing variation',
        example: '--max-jitter-ns 50000',
      },
      {
        name: 'Require PTP Sync',
        flag: '--require-ptp-sync',
        type: 'boolean',
        defaultValue: 'true',
        required: false,
        techDesc: 'Require PTP synchronization before testing',
        laymanDesc: 'Wait for precise time sync before testing',
        example: '--require-ptp-sync=false',
      },
      {
        name: 'Base Time',
        flag: '--base-time-ns',
        type: 'integer (nanoseconds)',
        defaultValue: '0',
        required: false,
        techDesc: 'Base time for gate control list in nanoseconds',
        laymanDesc: 'When the timing schedule starts',
        example: '--base-time-ns 1000000',
      },
      {
        name: 'Cycle Time',
        flag: '--cycle-time-ns',
        type: 'integer (nanoseconds)',
        defaultValue: '1000000',
        required: false,
        techDesc: 'Gate cycle time in nanoseconds',
        laymanDesc: 'How often the timing pattern repeats',
        example: '--cycle-time-ns 500000',
      },
      {
        name: 'Traffic Class',
        flag: '--traffic-class',
        type: 'integer (0-7)',
        defaultValue: '7',
        required: false,
        techDesc: 'IEEE 802.1Q traffic class for scheduled traffic',
        laymanDesc: 'Priority level for time-critical traffic',
        example: '--traffic-class 6',
      },
    ],
    metrics: [
      {
        name: 'Gate Timing Accuracy',
        unit: 'nanoseconds',
        goodRange: 'Within 1000ns of scheduled time',
        badMeaning: 'Gates not opening/closing precisely',
      },
      {
        name: 'Latency Jitter',
        unit: 'nanoseconds',
        goodRange: 'Below max-jitter-ns threshold',
        badMeaning: 'Timing too variable for deterministic apps',
      },
    ],
    passCriteria: 'All gates within timing tolerance',
    failMeaning: 'Cannot support deterministic timing',
    examples: [
      {
        desc: 'Basic TSN gate timing test',
        command: 'stem test -i eth0 -t gate_timing --max-latency-ns 500000',
        output: 'Gate timing: PASS (max deviation: 823ns)',
      },
    ],
    tips: [
      'Ensure PTP is configured and synchronized before testing',
      'Use IEEE 802.1AS-compliant hardware for best results',
    ],
    seeAlso: ['traffic_isolation'],
  },

  traffic_isolation: {
    id: 'traffic_isolation',
    name: 'TSN Traffic Class Isolation Test',
    standard: 'IEEE 802.1Qbv/Qbu',
    category: 'TSN',
    summary: 'Verifies critical traffic is protected from other classes.',
    techDesc: 'Tests traffic class isolation and frame preemption.',
    laymanDesc: 'Verifies important packets get through regardless of other traffic.',
    whenToUse: 'Mixed traffic TSN networks',
    whenNotToUse: 'Networks without traffic classes',
    parameters: [],
    metrics: [],
    passCriteria: 'Critical traffic unaffected by background',
    failMeaning: 'Critical traffic affected',
    examples: [],
    tips: [],
    seeAlso: ['gate_timing'],
  },

  scheduled_latency: {
    id: 'scheduled_latency',
    name: 'TSN Scheduled Latency Test',
    standard: 'IEEE 802.1Qbv',
    category: 'TSN',
    summary: 'Measures if packets arrive exactly when scheduled.',
    techDesc: 'Validates deterministic latency for scheduled traffic.',
    laymanDesc: 'Verifies packets arrive at EXACTLY the right time.',
    whenToUse: 'Deterministic latency validation',
    whenNotToUse: 'Networks without timing requirements',
    parameters: [],
    metrics: [],
    passCriteria: 'Latency within scheduled window',
    failMeaning: 'Traffic not meeting timing',
    examples: [],
    tips: [],
    seeAlso: ['gate_timing'],
  },

  tsn_full: {
    id: 'tsn_full',
    name: 'TSN Full Validation Suite',
    standard: 'IEEE 802.1Qbv/Qbu',
    category: 'TSN',
    summary: 'Complete TSN network validation.',
    techDesc: 'Complete TSN validation including all timing and isolation tests.',
    laymanDesc: 'The complete test for Time-Sensitive Networks.',
    whenToUse: 'Complete TSN validation',
    whenNotToUse: 'Troubleshooting specific issues',
    parameters: [],
    metrics: [],
    passCriteria: 'All TSN tests pass',
    failMeaning: 'Network not suitable for TSN',
    examples: [],
    tips: [],
    seeAlso: ['gate_timing', 'traffic_isolation', 'scheduled_latency'],
  },

  // TrafficGen Tests
  custom_stream: {
    id: 'custom_stream',
    name: 'Custom Traffic Stream',
    standard: 'N/A',
    category: 'TrafficGen',
    summary: 'Generate custom traffic patterns for specialized testing.',
    techDesc: `Custom traffic generation allows creation of arbitrary traffic patterns
including burst mode, specific MAC addresses, VLAN tagging, and controlled rates.
Useful for stress testing, QoS validation, and network diagnostics.`,
    laymanDesc: `Create your own custom traffic patterns for specialized tests.
Useful when standard tests don't cover your specific scenario.`,
    whenToUse: `• Custom stress testing scenarios
• QoS and traffic shaping validation
• Network diagnostics and debugging
• Vendor-specific testing requirements`,
    whenNotToUse: `• Use standard tests (RFC 2544, Y.1564) when applicable
• For certification or compliance testing`,
    parameters: [
      {
        name: 'Frame Size',
        flag: '--frame-size',
        type: 'integer (bytes)',
        defaultValue: '1518',
        required: false,
        techDesc: 'Ethernet frame size in bytes (64-9216)',
        laymanDesc: 'Size of each packet to generate',
        example: '--frame-size 512',
      },
      {
        name: 'Rate Percent',
        flag: '--rate-pct',
        type: 'float (percentage)',
        defaultValue: '100.0',
        required: false,
        techDesc: 'Traffic rate as percentage of line rate',
        laymanDesc: 'How fast to send traffic (100% = maximum speed)',
        example: '--rate-pct 50.0',
      },
      {
        name: 'Duration',
        flag: '--duration',
        type: 'integer (seconds)',
        defaultValue: '60',
        required: false,
        techDesc: 'Duration of traffic generation',
        laymanDesc: 'How long to generate traffic',
        example: '--duration 300',
      },
      {
        name: 'Warmup',
        flag: '--warmup',
        type: 'integer (seconds)',
        defaultValue: '2',
        required: false,
        techDesc: 'Warmup period before measurements',
        laymanDesc: 'Time to stabilize before counting',
        example: '--warmup 5',
      },
      {
        name: 'Stream ID',
        flag: '--stream-id',
        type: 'integer',
        defaultValue: '1',
        required: false,
        techDesc: 'Unique identifier for this traffic stream',
        laymanDesc: 'Label for tracking this traffic',
        example: '--stream-id 42',
      },
      {
        name: 'Burst Mode',
        flag: '--burst-mode',
        type: 'boolean',
        defaultValue: 'false',
        required: false,
        techDesc: 'Enable burst traffic mode instead of continuous',
        laymanDesc: 'Send traffic in bursts instead of steady flow',
        example: '--burst-mode',
      },
      {
        name: 'Burst Size',
        flag: '--burst-size',
        type: 'integer (frames)',
        defaultValue: '100',
        required: false,
        techDesc: 'Number of frames per burst (when burst-mode enabled)',
        laymanDesc: 'How many packets per burst',
        example: '--burst-size 50',
      },
      {
        name: 'Inter-Burst Gap',
        flag: '--inter-burst-gap-us',
        type: 'integer (microseconds)',
        defaultValue: '1000',
        required: false,
        techDesc: 'Gap between bursts in microseconds',
        laymanDesc: 'Pause between bursts',
        example: '--inter-burst-gap-us 500',
      },
      {
        name: 'Source MAC',
        flag: '--src-mac',
        type: 'string (MAC address)',
        defaultValue: '(interface MAC)',
        required: false,
        techDesc: 'Source MAC address for generated frames',
        laymanDesc: 'Sender address on packets',
        example: '--src-mac 00:11:22:33:44:55',
      },
      {
        name: 'Destination MAC',
        flag: '--dst-mac',
        type: 'string (MAC address)',
        defaultValue: 'ff:ff:ff:ff:ff:ff',
        required: false,
        techDesc: 'Destination MAC address for generated frames',
        laymanDesc: 'Target address on packets',
        example: '--dst-mac 00:aa:bb:cc:dd:ee',
      },
      {
        name: 'VLAN ID',
        flag: '--vlan-id',
        type: 'integer (0-4095)',
        defaultValue: '0',
        required: false,
        techDesc: 'VLAN identifier for 802.1Q tagged frames',
        laymanDesc: 'Virtual network tag (0 = no VLAN)',
        example: '--vlan-id 100',
      },
      {
        name: 'VLAN Priority',
        flag: '--vlan-priority',
        type: 'integer (0-7)',
        defaultValue: '0',
        required: false,
        techDesc: 'Priority Code Point for 802.1p CoS',
        laymanDesc: 'Traffic priority (7 = highest)',
        example: '--vlan-priority 5',
      },
    ],
    metrics: [
      {
        name: 'Tx Rate',
        unit: 'Mbps',
        goodRange: 'Matches configured rate',
        badMeaning: 'Unable to achieve requested rate',
      },
      {
        name: 'Tx Packets',
        unit: 'count',
        goodRange: 'Stable count increase',
        badMeaning: 'Transmission issues',
      },
      {
        name: 'Rx Packets',
        unit: 'count',
        goodRange: 'Matches Tx when looped back',
        badMeaning: 'Packet loss detected',
      },
    ],
    passCriteria: 'Traffic generated at requested rate',
    failMeaning: 'Unable to generate or receive traffic',
    examples: [
      {
        desc: 'Generate 50% rate traffic with 512-byte frames',
        command: 'stem test -i eth0 -t custom_stream --rate-pct 50 --frame-size 512',
        output: 'Tx Rate: 500 Mbps, Tx: 1.2M pps',
      },
      {
        desc: 'Burst mode with VLAN tagging',
        command: 'stem test -i eth0 -t custom_stream --burst-mode --vlan-id 100',
        output: 'Burst Mode: 100 frames/burst, VLAN 100',
      },
    ],
    tips: [
      'Use burst mode to test switch buffer behavior',
      'VLAN tagging requires 802.1Q-capable equipment',
      'Monitor receiver to verify packet delivery',
    ],
    seeAlso: ['throughput', 'back_to_back'],
  },
};

// Helper function to get tests by category
export function getTestsByCategory(categoryId: string): TestHelp[] {
  const cat = categories[categoryId];
  if (!cat) {
    return [];
  }
  return cat.tests.map((id) => tests[id]).filter(Boolean);
}

// Helper to search tests
export function searchTests(keyword: string): TestHelp[] {
  const lower = keyword.toLowerCase();
  return Object.values(tests).filter(
    (t) =>
      t.name.toLowerCase().includes(lower) ||
      t.summary.toLowerCase().includes(lower) ||
      t.techDesc.toLowerCase().includes(lower) ||
      t.laymanDesc.toLowerCase().includes(lower),
  );
}

// ============================================================================
// GLOSSARY
// ============================================================================

export interface GlossaryEntry {
  term: string;
  fullName: string;
  category: string;
  techDef: string;
  laymanDef: string;
  related: string[];
}

export const glossary: Record<string, GlossaryEntry> = {
  cir: {
    term: 'CIR',
    fullName: 'Committed Information Rate',
    category: 'Bandwidth',
    techDef:
      'The guaranteed bandwidth as specified in a service level agreement, below which the network commits to deliver traffic with minimal loss or delay.',
    laymanDef: 'The speed your internet/network contract guarantees you will always get.',
    related: ['eir', 'bandwidth', 'sla'],
  },
  eir: {
    term: 'EIR',
    fullName: 'Excess Information Rate',
    category: 'Bandwidth',
    techDef:
      'The rate above CIR up to which the network will attempt to deliver frames, but without guarantees.',
    laymanDef: 'Bonus bandwidth you might get when the network is not busy, but no promises.',
    related: ['cir', 'bandwidth'],
  },
  bandwidth: {
    term: 'Bandwidth',
    fullName: 'Network Bandwidth',
    category: 'Bandwidth',
    techDef:
      'Maximum rate of data transfer across a network path, typically measured in bits per second.',
    laymanDef:
      'How much data can flow through your connection per second - like water through a pipe.',
    related: ['throughput', 'cir'],
  },
  throughput: {
    term: 'Throughput',
    fullName: 'Network Throughput',
    category: 'Bandwidth',
    techDef:
      'The actual rate of successful data transfer, accounting for protocol overhead and network conditions.',
    laymanDef:
      'How much useful data actually gets through - slightly less than bandwidth due to overhead.',
    related: ['bandwidth', 'goodput'],
  },
  latency: {
    term: 'Latency',
    fullName: 'Network Latency',
    category: 'Latency',
    techDef: 'The time delay for a packet to travel from source to destination.',
    laymanDef: 'The "lag" - how long it takes for your data to reach the other end.',
    related: ['rtt', 'delay', 'jitter'],
  },
  rtt: {
    term: 'RTT',
    fullName: 'Round-Trip Time',
    category: 'Latency',
    techDef:
      'The time for a signal to travel to a destination and back, including processing delays.',
    laymanDef: 'How long a round trip takes - send a message, get a reply.',
    related: ['latency', 'ping'],
  },
  jitter: {
    term: 'Jitter',
    fullName: 'Packet Delay Variation',
    category: 'Latency',
    techDef: 'The variation in latency between successive packets in a flow.',
    laymanDef: 'How much the delay wobbles around - matters for video calls and gaming.',
    related: ['latency', 'pdv'],
  },
  pdv: {
    term: 'PDV',
    fullName: 'Packet Delay Variation',
    category: 'Latency',
    techDef: 'ITU-T standard term for jitter, measuring the difference in delay between packets.',
    laymanDef: 'The technical name for jitter - timing inconsistency.',
    related: ['jitter', 'latency'],
  },
  pps: {
    term: 'pps',
    fullName: 'Packets Per Second',
    category: 'Performance',
    techDef: 'The rate at which network packets are processed or transmitted.',
    laymanDef: 'How many packets your network can handle each second.',
    related: ['throughput', 'mpps'],
  },
  mpps: {
    term: 'Mpps',
    fullName: 'Million Packets Per Second',
    category: 'Performance',
    techDef: 'Standard unit for measuring high-speed packet processing rates.',
    laymanDef: 'Millions of packets per second - for measuring fast switches and routers.',
    related: ['pps'],
  },
  mtu: {
    term: 'MTU',
    fullName: 'Maximum Transmission Unit',
    category: 'Protocol',
    techDef: 'The largest packet size (in bytes) that can be transmitted without fragmentation.',
    laymanDef: 'The biggest chunk of data you can send at once - usually 1500 bytes.',
    related: ['frame_size', 'jumbo_frames'],
  },
  jumbo_frames: {
    term: 'Jumbo Frames',
    fullName: 'Jumbo Frames',
    category: 'Protocol',
    techDef: 'Ethernet frames larger than 1500 bytes, typically up to 9000 bytes.',
    laymanDef: 'Extra-large packets for data centers - more efficient but need special support.',
    related: ['mtu', 'frame_size'],
  },
  mac: {
    term: 'MAC',
    fullName: 'Media Access Control',
    category: 'Protocol',
    techDef: 'Layer 2 hardware address that uniquely identifies network interfaces.',
    laymanDef: 'The unique hardware address burned into every network card.',
    related: ['ethernet', 'layer2'],
  },
  dut: {
    term: 'DUT',
    fullName: 'Device Under Test',
    category: 'Testing',
    techDef: 'The network device being tested and evaluated.',
    laymanDef: 'The thing you are testing.',
    related: ['sut'],
  },
  sut: {
    term: 'SUT',
    fullName: 'System Under Test',
    category: 'Testing',
    techDef: 'The complete system including all devices in the test path.',
    laymanDef: 'Everything you are testing together as a system.',
    related: ['dut'],
  },
  sla: {
    term: 'SLA',
    fullName: 'Service Level Agreement',
    category: 'Service',
    techDef: 'Contractual agreement specifying performance guarantees and penalties.',
    laymanDef: 'The contract that says what your provider promises to deliver.',
    related: ['cir', 'oam'],
  },
  oam: {
    term: 'OAM',
    fullName: 'Operations, Administration, and Maintenance',
    category: 'Service',
    techDef: 'Tools and protocols for monitoring and managing networks, including Y.1731.',
    laymanDef: 'Built-in network monitoring tools that carriers use.',
    related: ['y1731', 'cfm'],
  },
  cfm: {
    term: 'CFM',
    fullName: 'Connectivity Fault Management',
    category: 'Service',
    techDef: 'IEEE 802.1ag protocol for detecting and isolating connectivity faults.',
    laymanDef: 'System for automatically finding network problems.',
    related: ['oam', 'y1731'],
  },
  xdp: {
    term: 'XDP',
    fullName: 'eXpress Data Path',
    category: 'Technology',
    techDef:
      'Linux kernel technology for high-performance packet processing before the network stack.',
    laymanDef: 'Fast-path technology that processes packets super quickly.',
    related: ['afxdp', 'dpdk'],
  },
  afxdp: {
    term: 'AF_XDP',
    fullName: 'Address Family XDP',
    category: 'Technology',
    techDef: 'Linux socket type for zero-copy packet processing using XDP.',
    laymanDef: 'Way to handle network packets extremely fast in Linux.',
    related: ['xdp', 'dpdk'],
  },
  dpdk: {
    term: 'DPDK',
    fullName: 'Data Plane Development Kit',
    category: 'Technology',
    techDef: 'Set of libraries for fast packet processing by bypassing the kernel.',
    laymanDef: 'Technology for super-fast networking, used in carriers and data centers.',
    related: ['xdp', 'afxdp'],
  },
  tos: {
    term: 'ToS',
    fullName: 'Type of Service',
    category: 'QoS',
    techDef: 'IPv4 header field used for quality of service prioritization.',
    laymanDef: 'Tag in IP packets to mark priority level.',
    related: ['dscp', 'cos'],
  },
  dscp: {
    term: 'DSCP',
    fullName: 'Differentiated Services Code Point',
    category: 'QoS',
    techDef: '6-bit field in IP header for traffic classification and QoS.',
    laymanDef: 'Modern way to mark packet priority - tells the network how important a packet is.',
    related: ['tos', 'cos'],
  },
  cos: {
    term: 'CoS',
    fullName: 'Class of Service',
    category: 'QoS',
    techDef: '3-bit field in VLAN tag for Layer 2 traffic classification (IEEE 802.1p).',
    laymanDef: 'Priority marking at the Ethernet level (layer 2).',
    related: ['dscp', 'vlan'],
  },
  vlan: {
    term: 'VLAN',
    fullName: 'Virtual LAN',
    category: 'Protocol',
    techDef: 'Logical network partition at Layer 2, typically using IEEE 802.1Q tagging.',
    laymanDef: 'Virtual networks that keep traffic separated on the same physical equipment.',
    related: ['cos', 'qinq'],
  },
  qinq: {
    term: 'Q-in-Q',
    fullName: 'IEEE 802.1ad (Double Tagging)',
    category: 'Protocol',
    techDef: 'Stacking of VLAN tags for provider/customer separation.',
    laymanDef: 'Putting a VLAN tag inside another VLAN tag - used by carriers.',
    related: ['vlan', 'svlan'],
  },
  bdp: {
    term: 'BDP',
    fullName: 'Bandwidth-Delay Product',
    category: 'Performance',
    techDef: 'Product of bandwidth and RTT, representing data in flight for optimal TCP.',
    laymanDef: 'How much data should be "in the air" at once for maximum speed.',
    related: ['rtt', 'tcp'],
  },
  tcp: {
    term: 'TCP',
    fullName: 'Transmission Control Protocol',
    category: 'Protocol',
    techDef: 'Connection-oriented transport protocol with reliability and flow control.',
    laymanDef: 'The protocol that makes sure data arrives correctly and in order.',
    related: ['bdp', 'udp'],
  },
  udp: {
    term: 'UDP',
    fullName: 'User Datagram Protocol',
    category: 'Protocol',
    techDef: 'Connectionless transport protocol without reliability guarantees.',
    laymanDef: 'Fast but unreliable - used for video streaming and gaming.',
    related: ['tcp'],
  },
  tsn: {
    term: 'TSN',
    fullName: 'Time-Sensitive Networking',
    category: 'Technology',
    techDef: 'IEEE 802.1 standards for deterministic, low-latency Ethernet communication.',
    laymanDef: 'Network technology where packets arrive at EXACTLY the right time.',
    related: ['tas', 'gcl'],
  },
  tas: {
    term: 'TAS',
    fullName: 'Time-Aware Shaper',
    category: 'Technology',
    techDef: 'IEEE 802.1Qbv mechanism for scheduled traffic transmission.',
    laymanDef: 'Network feature that opens and closes time gates for traffic.',
    related: ['tsn', 'gcl'],
  },
  gcl: {
    term: 'GCL',
    fullName: 'Gate Control List',
    category: 'Technology',
    techDef: 'Schedule defining when traffic classes can transmit in TSN.',
    laymanDef: 'The schedule that says when each type of traffic can go.',
    related: ['tsn', 'tas'],
  },
  reflector: {
    term: 'Reflector',
    fullName: 'Packet Reflector',
    category: 'Testing',
    techDef: 'Endpoint that returns received packets for round-trip measurement.',
    laymanDef: 'Device at the far end that bounces packets back for testing.',
    related: ['loopback'],
  },
  loopback: {
    term: 'Loopback',
    fullName: 'Network Loopback',
    category: 'Testing',
    techDef: 'Interface or mechanism that returns traffic to its source.',
    laymanDef: 'Sending traffic back to where it came from for testing.',
    related: ['reflector'],
  },
  binary_search: {
    term: 'Binary Search',
    fullName: 'Binary Search Algorithm',
    category: 'Testing',
    techDef: 'Algorithm that halves the search space each iteration to find optimal rate.',
    laymanDef: 'Quickly finding the right speed by trying half as much each time.',
    related: ['throughput'],
  },
  frame_loss: {
    term: 'Frame Loss',
    fullName: 'Packet/Frame Loss',
    category: 'Performance',
    techDef: 'Percentage of transmitted frames that fail to arrive at destination.',
    laymanDef: 'Packets that got lost along the way.',
    related: ['packet_loss'],
  },
};

// Helper to search glossary
export function searchGlossary(keyword: string): GlossaryEntry[] {
  const lower = keyword.toLowerCase();
  return Object.values(glossary).filter(
    (e) =>
      e.term.toLowerCase().includes(lower) ||
      e.fullName.toLowerCase().includes(lower) ||
      e.techDef.toLowerCase().includes(lower) ||
      e.laymanDef.toLowerCase().includes(lower),
  );
}

// ============================================================================
// TUTORIALS
// ============================================================================

export interface TutorialStep {
  title: string;
  content: string;
  command?: string;
  expected?: string;
  tip?: string;
}

export interface Tutorial {
  id: string;
  title: string;
  duration: string;
  level: 'Beginner' | 'Intermediate' | 'Advanced';
  description: string;
  steps: TutorialStep[];
}

export const tutorials: Record<string, Tutorial> = {
  quickstart: {
    id: 'quickstart',
    title: 'Quick Start Guide',
    duration: '5 min',
    level: 'Beginner',
    description: 'Get started with your first network test in 5 minutes.',
    steps: [
      {
        title: 'Check System Requirements',
        content: `Before you begin, make sure you have:
- A Linux system (kernel 5.x+ for best performance)
- Two network interfaces (or one plus a remote reflector)
- Root/sudo access for raw socket operations`,
        command: 'uname -r',
        expected: '5.x.x or higher',
        tip: 'Kernel 5.x enables AF_XDP for high-performance testing',
      },
      {
        title: 'Find Your Network Interfaces',
        content:
          'List available network interfaces. Look for interfaces that are UP and have a cable connected.',
        command: 'ip link show',
        expected: 'List of interfaces with state UP',
      },
      {
        title: 'Start the Reflector',
        content:
          'On the far-end device (or a second port), start the reflector. This will echo back all received test traffic.',
        command: 'sudo stem reflect -i eth1',
        expected: 'Reflector started on eth1',
        tip: 'The reflector can run in AF_XDP mode for 10+ Mpps performance',
      },
      {
        title: 'Run Your First Test',
        content: 'Now run a basic throughput test from the tester side.',
        command: 'sudo stem test -i eth0 -t throughput',
        expected: 'Test completes with throughput results',
      },
      {
        title: 'View Results',
        content:
          'The test will display results showing maximum throughput for each frame size. Congratulations - you have completed your first network test!',
        tip: 'Use --output json for machine-readable results',
      },
    ],
  },
  reflector: {
    id: 'reflector',
    title: 'Setting Up Packet Reflection',
    duration: '10 min',
    level: 'Beginner',
    description: 'Learn how to set up packet reflectors for network testing.',
    steps: [
      {
        title: 'Understanding Packet Reflection',
        content: `Packet reflection returns received frames back to the sender, enabling round-trip measurements. The reflector can operate in several modes:

- AF_PACKET: Standard Linux sockets (1-2 Mpps)
- AF_XDP: Fast kernel bypass (5-10 Mpps)
- DPDK: Maximum performance (15-40+ Mpps)`,
      },
      {
        title: 'Basic Reflector Setup',
        content: 'Start a basic reflector using the default configuration.',
        command: 'sudo stem reflect -i eth0',
        expected: 'Reflector running on eth0',
      },
      {
        title: 'High-Performance Mode',
        content:
          'For maximum performance, use AF_XDP mode. This requires a modern kernel with XDP support.',
        command: 'sudo stem reflect -i eth0 --mode af_xdp',
        expected: 'AF_XDP reflector running',
        tip: 'Check kernel version with "uname -r" - need 5.x or later',
      },
      {
        title: 'Profile-Based Reflection',
        content: 'Reflector profiles configure which test signatures to respond to.',
        command: 'sudo stem reflect -i eth0 --profile all',
        expected: 'Reflector with all signatures enabled',
      },
    ],
  },
  rfc2544: {
    id: 'rfc2544',
    title: 'RFC 2544 Testing Deep Dive',
    duration: '20 min',
    level: 'Intermediate',
    description: 'Master RFC 2544 benchmark tests for network equipment.',
    steps: [
      {
        title: 'Understanding RFC 2544',
        content: `RFC 2544 defines standard benchmarks for network devices:

- Throughput: Maximum speed without loss
- Latency: Round-trip delay at specified rate
- Frame Loss: Loss percentage vs offered load
- Back-to-Back: Maximum burst capacity
- System Recovery: Recovery after overload
- Reset: Device restart time`,
      },
      {
        title: 'Throughput Testing',
        content: 'Throughput finds the maximum rate with 0% packet loss using binary search.',
        command: 'sudo stem test -i eth0 -t throughput',
        expected: 'Throughput results for each frame size',
        tip: 'Standard frame sizes: 64, 128, 256, 512, 1024, 1280, 1518 bytes',
      },
      {
        title: 'Latency Testing',
        content: 'Latency measures round-trip time at a specified offered load.',
        command: 'sudo stem test -i eth0 -t latency',
        expected: 'Min/avg/max latency per frame size',
      },
      {
        title: 'Frame Loss Testing',
        content: 'Frame loss measures packet loss at various load percentages.',
        command: 'sudo stem test -i eth0 -t frame_loss',
        expected: 'Loss percentage at each load level',
      },
      {
        title: 'Custom Frame Sizes',
        content: 'Specify custom frame sizes for your specific needs.',
        command: 'sudo stem test -i eth0 -t throughput --frame-sizes 64,512,1518',
        expected: 'Results for specified frame sizes only',
      },
      {
        title: 'Interpreting Results',
        content: `Good RFC 2544 results typically show:
- Throughput: >95% of line rate
- Latency: <1ms for LAN equipment
- Frame Loss: 0% at rated throughput
- Back-to-Back: Large burst capacity

Poor results indicate equipment limitations, configuration issues, or network problems.`,
      },
    ],
  },
  y1564: {
    id: 'y1564',
    title: 'Y.1564 Service Activation',
    duration: '15 min',
    level: 'Intermediate',
    description: 'Learn carrier ethernet service activation testing with Y.1564.',
    steps: [
      {
        title: 'Understanding Y.1564',
        content: `Y.1564 (EtherSAM) is the carrier standard for turning up ethernet services. It validates that a service meets its SLA by testing at progressive load levels.

The test has two phases:
1. Configuration Test: Quick validation at 25/50/75/100% of CIR
2. Performance Test: Extended duration at full CIR`,
      },
      {
        title: 'Know Your Service Parameters',
        content: `Before testing, gather from your service contract:
- CIR: Committed Information Rate (guaranteed speed)
- EIR: Excess Information Rate (burst speed above CIR)
- CBS: Committed Burst Size
- FD: Maximum Frame Delay
- FDV: Maximum Frame Delay Variation (jitter)
- FLR: Maximum Frame Loss Ratio`,
      },
      {
        title: 'Run Configuration Test',
        content: 'The config test validates service at each CIR step.',
        command: 'sudo stem test -i eth0 -t y1564_config --cir 100',
        expected: 'PASS at all four CIR levels',
      },
      {
        title: 'Run Performance Test',
        content: 'After config passes, run the extended performance test.',
        command: 'sudo stem test -i eth0 -t y1564_performance --cir 100 --duration 15',
        expected: 'Sustained performance for 15 minutes',
      },
      {
        title: 'Full SAC Test',
        content: 'Run both tests in sequence with a single command.',
        command: 'sudo stem test -i eth0 -t y1564_full --cir 100',
        expected: 'Service Activation Complete',
      },
    ],
  },
  troubleshoot: {
    id: 'troubleshoot',
    title: 'Troubleshooting Test Failures',
    duration: '15 min',
    level: 'Advanced',
    description: 'Diagnose and fix common network testing problems.',
    steps: [
      {
        title: 'Permission Denied Errors',
        content: `If you see "permission denied" or socket errors, you need root privileges.`,
        command: 'sudo stem reflect -i eth0',
        tip: 'Network testing requires raw socket access',
      },
      {
        title: 'Interface Not Found',
        content: 'Verify the interface exists and is up.',
        command: 'ip link show eth0',
        expected: 'Interface details with state UP',
        tip: 'Bring up with: sudo ip link set eth0 up',
      },
      {
        title: 'No Response from Reflector',
        content: `If tests fail with "no response", check:
1. Reflector is running on the remote end
2. Network path is connected
3. Firewalls allow the test traffic`,
        command: 'ping <reflector-ip>',
        expected: 'Ping responses',
      },
      {
        title: 'Poor Performance',
        content: `If throughput is lower than expected:
1. Check link speed: ethtool eth0
2. Check for errors: ip -s link show eth0
3. Try AF_XDP mode for better performance`,
        command: 'ethtool eth0 | grep Speed',
        expected: 'Expected link speed',
      },
      {
        title: 'AF_XDP Not Working',
        content:
          'AF_XDP requires kernel 5.x+ and driver support. Fall back to AF_PACKET if needed.',
        command: 'uname -r',
        expected: '5.x.x or later',
        tip: 'Use --mode af_packet as fallback',
      },
    ],
  },
  results: {
    id: 'results',
    title: 'Interpreting Test Results',
    duration: '10 min',
    level: 'Beginner',
    description: 'Understand what your test results mean and what to do about them.',
    steps: [
      {
        title: 'Reading Throughput Results',
        content: `Throughput is reported as a percentage of line rate.

Excellent: >95% - Equipment/network performing well
Good: 80-95% - Acceptable, minor overhead
Concerning: 60-80% - Investigate bottleneck
Poor: <60% - Significant problem`,
      },
      {
        title: 'Reading Latency Results',
        content: `Latency is reported in milliseconds.

LAN equipment: <1ms expected
Metro area: 5-20ms typical
Wide area: varies by distance

High jitter (variation) matters more than latency for real-time applications.`,
      },
      {
        title: 'Understanding Frame Loss',
        content: `Frame loss should be 0% at rated throughput.

0%: Perfect - service meets commitment
0.001-0.1%: May be acceptable for some applications
>0.1%: Problematic - investigate cause`,
      },
      {
        title: 'Export and Compare',
        content: 'Export results for trending and comparison.',
        command: 'sudo stem test -i eth0 -t throughput --output json > results.json',
        expected: 'JSON file for processing',
        tip: 'Compare baseline vs. current to spot degradation',
      },
    ],
  },
};

