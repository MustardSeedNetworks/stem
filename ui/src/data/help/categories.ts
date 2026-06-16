/**
 * @fileoverview The Stem - Help Content Categories
 * @description Test category definitions for the WebUI help system.
 */

import type { Category } from './types';

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
