/**
 * @fileoverview The Stem - Help Content Tutorials
 * @description Tutorial definitions for the WebUI help system.
 */

import type { Tutorial } from './types';

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
