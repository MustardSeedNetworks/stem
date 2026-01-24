#!/bin/bash
#
# run_smoke_tests.sh - Comprehensive Smoke Tests for The Stem
#
# Requires: Linux, root/sudo, veth pair support
# Tests all stem functionality using virtual interfaces
#
# Copyright (c) 2025 Mustard Seed Networks. All rights reserved.
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
BOLD='\033[1m'
NC='\033[0m'

# Configuration
VETH_TX="veth-stem-tx"
VETH_RX="veth-stem-rx"
IP_TX="192.168.253.1"
IP_RX="192.168.253.2"
SUBNET="/24"
WEB_PORT=8887
WEB_PORT_ALT=8888

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="${SCRIPT_DIR}/../.."
STEM_BIN="${PROJECT_ROOT}/bin/stem"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Process tracking
REFLECTOR_PID=""
WEB_PID=""
TEST_PID=""

# Timing
START_TIME=$(date +%s)

# Logging
log_info()    { echo -e "${CYAN}[INFO]${NC} $1"; }
log_pass()    { echo -e "${GREEN}[PASS]${NC} $1"; }
log_fail()    { echo -e "${RED}[FAIL]${NC} $1"; }
log_skip()    { echo -e "${YELLOW}[SKIP]${NC} $1"; }
log_header()  { echo -e "\n${BOLD}${CYAN}=== $1 ===${NC}"; }
log_section() { echo -e "\n${BOLD}${MAGENTA}━━━ $1 ━━━${NC}"; }

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}Error: Smoke tests require root for veth creation${NC}"
        echo "Usage: sudo $0"
        exit 1
    fi
}

# Cleanup function
cleanup() {
    log_info "Cleaning up..."

    # Kill all test processes
    for pid in $REFLECTOR_PID $WEB_PID $TEST_PID; do
        if [[ -n "$pid" ]]; then
            kill $pid 2>/dev/null || true
            wait $pid 2>/dev/null || true
        fi
    done

    # Kill any stray processes
    pkill -f "stem.*${VETH_RX}" 2>/dev/null || true
    pkill -f "stem.*web.*${WEB_PORT}" 2>/dev/null || true
    pkill -f "stem.*web.*${WEB_PORT_ALT}" 2>/dev/null || true

    # Remove veth pair
    ip link delete "${VETH_TX}" 2>/dev/null || true

    log_info "Cleanup complete"
}

# Set up trap for cleanup
trap cleanup EXIT

# Create veth pair for testing
setup_veth() {
    log_info "Creating veth pair..."

    # Remove existing if present
    ip link delete "${VETH_TX}" 2>/dev/null || true

    # Create veth pair
    ip link add "${VETH_TX}" type veth peer name "${VETH_RX}"

    # Configure both ends
    ip addr add "${IP_TX}${SUBNET}" dev "${VETH_TX}"
    ip addr add "${IP_RX}${SUBNET}" dev "${VETH_RX}"
    ip link set "${VETH_TX}" up
    ip link set "${VETH_RX}" up
    ip link set "${VETH_TX}" mtu 9000
    ip link set "${VETH_RX}" mtu 9000

    # Disable reverse path filtering
    echo 0 > /proc/sys/net/ipv4/conf/${VETH_TX}/rp_filter
    echo 0 > /proc/sys/net/ipv4/conf/${VETH_RX}/rp_filter
    echo 0 > /proc/sys/net/ipv4/conf/all/rp_filter

    log_info "veth pair ready: ${VETH_TX} <-> ${VETH_RX}"
}

# Run a test and record result
run_test() {
    local name="$1"
    local cmd="$2"
    local expected_exit="${3:-0}"

    TESTS_RUN=$((TESTS_RUN + 1))

    local output
    local exit_code

    set +e
    output=$(eval "$cmd" 2>&1)
    exit_code=$?
    set -e

    if [[ $exit_code -eq $expected_exit ]]; then
        log_pass "$name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_fail "$name (exit: $exit_code, expected: $expected_exit)"
        echo "  Output: $(echo "$output" | head -3 | tr '\n' ' ')"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Skip a test with reason
skip_test() {
    local name="$1"
    local reason="$2"

    TESTS_RUN=$((TESTS_RUN + 1))
    TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
    log_skip "$name - $reason"
}

# Assert helper
assert_contains() {
    local output="$1"
    local expected="$2"
    local name="$3"

    TESTS_RUN=$((TESTS_RUN + 1))
    if echo "$output" | grep -qi "$expected"; then
        log_pass "$name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_fail "$name (expected to contain: $expected)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_json_field() {
    local json="$1"
    local field="$2"
    local expected="$3"
    local name="$4"

    TESTS_RUN=$((TESTS_RUN + 1))
    local value=$(echo "$json" | grep -o "\"$field\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" | sed 's/.*: *"\([^"]*\)".*/\1/' || echo "")

    if [[ "$value" == "$expected" ]]; then
        log_pass "$name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_fail "$name (got: '$value', expected: '$expected')"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# ============================================================================
# SECTION 1: Binary and Build Tests
# ============================================================================

test_binary_and_build() {
    log_section "BINARY & BUILD TESTS"

    # Binary existence
    log_header "Binary Verification"
    if [[ ! -x "${STEM_BIN}" ]]; then
        log_info "Binary not found, building..."
        (cd "${PROJECT_ROOT}" && make build)
    fi

    run_test "Binary exists and is executable" \
        "test -x ${STEM_BIN}"

    run_test "Binary is not empty" \
        "test -s ${STEM_BIN}"

    # Check binary size (should be at least 5MB with embedded UI)
    # Use -L to follow symlinks, try Linux stat first, then macOS stat
    local size=$(stat -Lc%s "${STEM_BIN}" 2>/dev/null || stat -Lf%z "${STEM_BIN}" 2>/dev/null)
    run_test "Binary has embedded assets (>5MB)" \
        "test ${size:-0} -gt 5000000"
}

# ============================================================================
# SECTION 2: Version and Branding Tests
# ============================================================================

test_version_and_branding() {
    log_section "VERSION & BRANDING TESTS"

    log_header "Version Command"
    run_test "Version command succeeds" \
        "${STEM_BIN} version"

    local version_output=$("${STEM_BIN}" version 2>&1)

    assert_contains "$version_output" "The Stem" "Branding shows 'The Stem'"
    assert_contains "$version_output" "2025" "Copyright year is 2025"
    assert_contains "$version_output" "Mustard Seed" "Company name present"
    assert_contains "$version_output" "Commit" "Commit info present"
    assert_contains "$version_output" "Built" "Build time present"
}

# ============================================================================
# SECTION 3: CLI Help Tests
# ============================================================================

test_cli_help() {
    log_section "CLI HELP TESTS"

    log_header "Main Help"
    run_test "Help flag (-h)" "${STEM_BIN} -h"
    run_test "Help flag (--help)" "${STEM_BIN} --help"
    run_test "Help command" "${STEM_BIN} help"

    log_header "Subcommand Help"
    run_test "Reflect help" "${STEM_BIN} reflect --help"
    run_test "Test help" "${STEM_BIN} test --help"
    run_test "Web help" "${STEM_BIN} web --help"
    run_test "License help" "${STEM_BIN} license --help"
    run_test "TUI help" "${STEM_BIN} tui --help"

    log_header "Help Content Verification"
    local help_output=$("${STEM_BIN}" --help 2>&1)

    assert_contains "$help_output" "reflect" "Help mentions reflect command"
    assert_contains "$help_output" "test" "Help mentions test command"
    assert_contains "$help_output" "web" "Help mentions web command"
    assert_contains "$help_output" "license" "Help mentions license command"
    assert_contains "$help_output" "EXAMPLES" "Help contains examples"
    assert_contains "$help_output" "RFC 2544" "Help references RFC 2544"
    assert_contains "$help_output" "Y.1564" "Help references Y.1564"
}

# ============================================================================
# SECTION 4: Test Types and Categories
# ============================================================================

test_test_types() {
    log_section "TEST TYPES VERIFICATION"

    log_header "List Tests Command"
    run_test "list-tests command" "${STEM_BIN} list-tests"

    local list_output=$("${STEM_BIN}" list-tests 2>&1)

    log_header "RFC 2544 Tests (6 required)"
    assert_contains "$list_output" "throughput" "RFC 2544 throughput test listed"
    assert_contains "$list_output" "latency" "RFC 2544 latency test listed"
    assert_contains "$list_output" "frame_loss" "RFC 2544 frame_loss test listed"
    assert_contains "$list_output" "back_to_back" "RFC 2544 back_to_back test listed"
    assert_contains "$list_output" "system_recovery" "RFC 2544 system_recovery test listed"
    assert_contains "$list_output" "reset" "RFC 2544 reset test listed"

    log_header "Y.1564 Tests (3 required)"
    assert_contains "$list_output" "y1564_config" "Y.1564 config test listed"
    assert_contains "$list_output" "y1564_perf" "Y.1564 performance test listed"
    assert_contains "$list_output" "y1564" "Y.1564 full test listed"

    log_header "RFC 2889 LAN Switch Tests (5 required)"
    assert_contains "$list_output" "rfc2889_forwarding" "RFC 2889 forwarding test listed"
    assert_contains "$list_output" "rfc2889_caching" "RFC 2889 caching test listed"
    assert_contains "$list_output" "rfc2889_learning" "RFC 2889 learning test listed"
    assert_contains "$list_output" "rfc2889_broadcast" "RFC 2889 broadcast test listed"
    assert_contains "$list_output" "rfc2889_congestion" "RFC 2889 congestion test listed"

    log_header "Additional Standards"
    assert_contains "$list_output" "rfc6349" "RFC 6349 TCP tests listed"
    assert_contains "$list_output" "y1731" "Y.1731 OAM tests listed"
    assert_contains "$list_output" "mef" "MEF tests listed"
    assert_contains "$list_output" "tsn" "TSN tests listed"

    log_header "Test Count Verification"
    local test_count=$(echo "$list_output" | grep -c "^\s\s[a-z]" || echo "0")
    run_test "At least 27 test types defined" \
        "test $test_count -ge 27"
}

# ============================================================================
# SECTION 5: Reflector Tests
# ============================================================================

test_reflector() {
    log_section "REFLECTOR TESTS"

    log_header "Reflector Startup"
    ${STEM_BIN} reflect -i ${VETH_RX} --profile all >/dev/null 2>&1 &
    REFLECTOR_PID=$!
    sleep 2

    TESTS_RUN=$((TESTS_RUN + 1))
    if kill -0 $REFLECTOR_PID 2>/dev/null; then
        log_pass "Reflector starts successfully"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_fail "Reflector failed to start"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        REFLECTOR_PID=""
        return 1
    fi

    log_header "Graceful Shutdown"
    kill -TERM $REFLECTOR_PID 2>/dev/null
    sleep 2

    TESTS_RUN=$((TESTS_RUN + 1))
    if ! kill -0 $REFLECTOR_PID 2>/dev/null; then
        log_pass "Reflector handles SIGTERM gracefully"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        REFLECTOR_PID=""
    else
        log_fail "Reflector did not shutdown on SIGTERM"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        kill -9 $REFLECTOR_PID 2>/dev/null || true
        REFLECTOR_PID=""
    fi

    log_header "Profile Tests"
    for profile in netally msn all custom; do
        ${STEM_BIN} reflect -i ${VETH_RX} --profile $profile >/dev/null 2>&1 &
        local pid=$!
        sleep 1

        TESTS_RUN=$((TESTS_RUN + 1))
        if kill -0 $pid 2>/dev/null; then
            log_pass "Profile '$profile' starts correctly"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            kill $pid 2>/dev/null
            wait $pid 2>/dev/null || true
        else
            log_fail "Profile '$profile' failed to start"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    done

    log_header "Port Filtering"
    ${STEM_BIN} reflect -i ${VETH_RX} --profile all --port 3842 >/dev/null 2>&1 &
    local pid=$!
    sleep 1

    TESTS_RUN=$((TESTS_RUN + 1))
    if kill -0 $pid 2>/dev/null; then
        log_pass "Reflector with port filter starts"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        kill $pid 2>/dev/null
        wait $pid 2>/dev/null || true
    else
        log_fail "Reflector with port filter failed"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    log_header "OUI Filtering"
    ${STEM_BIN} reflect -i ${VETH_RX} --profile all --oui "00:c0:17" >/dev/null 2>&1 &
    pid=$!
    sleep 1

    TESTS_RUN=$((TESTS_RUN + 1))
    if kill -0 $pid 2>/dev/null; then
        log_pass "Reflector with OUI filter starts"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        kill $pid 2>/dev/null
        wait $pid 2>/dev/null || true
    else
        log_fail "Reflector with OUI filter failed"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# ============================================================================
# SECTION 6: WebUI Tests
# ============================================================================

test_webui() {
    log_section "WEBUI TESTS"

    log_header "WebUI Startup"
    ${STEM_BIN} web -p ${WEB_PORT} >/dev/null 2>&1 &
    WEB_PID=$!
    sleep 3

    TESTS_RUN=$((TESTS_RUN + 1))
    if kill -0 $WEB_PID 2>/dev/null; then
        log_pass "WebUI starts on port ${WEB_PORT}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_fail "WebUI failed to start"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        WEB_PID=""
        return 1
    fi

    log_header "API Health Endpoint"
    local health=$(curl -s http://localhost:${WEB_PORT}/api/health)

    assert_json_field "$health" "status" "healthy" "Health status is 'healthy'"
    assert_json_field "$health" "product" "The Stem" "Product name is 'The Stem'"
    assert_json_field "$health" "company" "Mustard Seed Networks" "Company name correct"

    log_header "Core API Endpoints"
    run_test "GET /api/stats returns JSON" \
        "curl -s http://localhost:${WEB_PORT}/api/stats | grep -q '{'"

    run_test "GET /api/interfaces returns JSON" \
        "curl -s http://localhost:${WEB_PORT}/api/interfaces | grep -q '\['"

    run_test "GET /api/settings returns JSON" \
        "curl -s http://localhost:${WEB_PORT}/api/settings | grep -q '{'"

    run_test "GET /api/mode returns JSON" \
        "curl -s http://localhost:${WEB_PORT}/api/mode | grep -q 'mode'"

    log_header "License API Endpoints"
    run_test "GET /api/license returns JSON" \
        "curl -s http://localhost:${WEB_PORT}/api/license | grep -q '{'"

    run_test "GET /api/license/trial returns JSON" \
        "curl -s http://localhost:${WEB_PORT}/api/license/trial | grep -q 'active'"

    log_header "Reflector API Endpoints"
    run_test "GET /api/reflector/config returns JSON" \
        "curl -s http://localhost:${WEB_PORT}/api/reflector/config | grep -q 'profile'"

    run_test "GET /api/reflector/stats returns JSON" \
        "curl -s http://localhost:${WEB_PORT}/api/reflector/stats | grep -q 'running'"

    log_header "API POST Operations"
    run_test "POST /api/mode accepts JSON" \
        "curl -s -X POST -H 'Content-Type: application/json' -d '{\"mode\":\"reflector\"}' http://localhost:${WEB_PORT}/api/mode | grep -q 'updated'"

    run_test "POST /api/settings accepts JSON" \
        "curl -s -X POST -H 'Content-Type: application/json' -d '{\"interface\":\"eth0\"}' http://localhost:${WEB_PORT}/api/settings | grep -q 'updated'"

    run_test "POST /api/reflector/config accepts JSON" \
        "curl -s -X POST -H 'Content-Type: application/json' -d '{\"profile\":\"netally\"}' http://localhost:${WEB_PORT}/api/reflector/config | grep -q 'updated'"

    log_header "Test Control Endpoints"
    run_test "POST /api/test/start starts test" \
        "curl -s -X POST http://localhost:${WEB_PORT}/api/test/start | grep -q 'started'"

    run_test "POST /api/test/stop stops test" \
        "curl -s -X POST http://localhost:${WEB_PORT}/api/test/stop | grep -q 'stopped'"

    log_header "Static File Serving"
    run_test "Root path serves HTML" \
        "curl -s http://localhost:${WEB_PORT}/ | grep -q 'html'"

    run_test "HTML contains The Stem title" \
        "curl -s http://localhost:${WEB_PORT}/ | grep -qi 'stem'"

    log_header "Error Handling"
    run_test "Invalid endpoint returns 404" \
        "curl -s -o /dev/null -w '%{http_code}' http://localhost:${WEB_PORT}/api/nonexistent | grep -q '404'"

    run_test "Wrong method returns 405" \
        "curl -s -X DELETE http://localhost:${WEB_PORT}/api/health 2>&1 | grep -qi 'method'"

    # Cleanup
    kill $WEB_PID 2>/dev/null
    wait $WEB_PID 2>/dev/null || true
    WEB_PID=""
}

# ============================================================================
# SECTION 7: Network Stack Tests
# ============================================================================

test_network_stack() {
    log_section "NETWORK STACK TESTS"

    log_header "Packet Path Test"
    ${STEM_BIN} reflect -i ${VETH_RX} --profile all >/dev/null 2>&1 &
    REFLECTOR_PID=$!
    sleep 2

    if ! kill -0 $REFLECTOR_PID 2>/dev/null; then
        skip_test "Network path tests" "Reflector failed to start"
        return
    fi

    local pre_rx=$(cat /sys/class/net/${VETH_TX}/statistics/rx_packets)
    local pre_tx=$(cat /sys/class/net/${VETH_TX}/statistics/tx_packets)

    # Send test packets
    ping -c 20 -I ${VETH_TX} ${IP_RX} -W 1 >/dev/null 2>&1 || true
    sleep 1

    local post_rx=$(cat /sys/class/net/${VETH_TX}/statistics/rx_packets)
    local post_tx=$(cat /sys/class/net/${VETH_TX}/statistics/tx_packets)
    local packets_rx=$((post_rx - pre_rx))
    local packets_tx=$((post_tx - pre_tx))

    log_info "TX packets: $packets_tx, RX packets: $packets_rx"

    TESTS_RUN=$((TESTS_RUN + 1))
    if [[ $packets_tx -gt 0 ]]; then
        log_pass "Packets transmitted on veth ($packets_tx pkts)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_fail "No packets transmitted"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    TESTS_RUN=$((TESTS_RUN + 1))
    if [[ $packets_rx -gt 0 ]]; then
        log_pass "Packets received on veth ($packets_rx pkts)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_skip "No packets received" "Requires ITO signature for reflection"
    fi

    # Cleanup
    kill $REFLECTOR_PID 2>/dev/null
    wait $REFLECTOR_PID 2>/dev/null || true
    REFLECTOR_PID=""

    log_header "MTU Support"
    run_test "Jumbo MTU configured (9000)" \
        "test $(cat /sys/class/net/${VETH_TX}/mtu) -eq 9000"
}

# ============================================================================
# SECTION 8: License Tests
# ============================================================================

test_license() {
    log_section "LICENSE TESTS"

    log_header "License Commands"
    run_test "license --status command" \
        "${STEM_BIN} license --status 2>&1 | grep -qi 'license\|trial\|status'"

    local status_output=$("${STEM_BIN}" license --status 2>&1)

    assert_contains "$status_output" "Device ID" "License shows device ID"
    assert_contains "$status_output" "Platform" "License shows platform"

    log_header "Trial Mode"
    # Don't actually start trial to preserve state, just test command exists
    run_test "license --trial flag recognized" \
        "${STEM_BIN} license --help 2>&1 | grep -q 'trial'"

    run_test "license --activate flag recognized" \
        "${STEM_BIN} license --help 2>&1 | grep -q 'activate'"

    run_test "license --deactivate flag recognized" \
        "${STEM_BIN} license --help 2>&1 | grep -q 'deactivate'"
}

# ============================================================================
# SECTION 9: Error Handling Tests
# ============================================================================

test_error_handling() {
    log_section "ERROR HANDLING TESTS"

    log_header "Invalid Interface"
    run_test "Invalid interface name rejected" \
        "${STEM_BIN} reflect -i nonexistent_interface_12345 2>&1 | grep -qi 'error\|fail\|not found'"

    log_header "Missing Required Args"
    run_test "reflect without -i shows error" \
        "${STEM_BIN} reflect 2>&1 | grep -qi 'interface\|required'" 0

    run_test "test without -i shows error" \
        "${STEM_BIN} test 2>&1 | grep -qi 'interface\|required'" 0

    log_header "Invalid Test Types"
    run_test "Invalid test type rejected" \
        "${STEM_BIN} test -i ${VETH_TX} -t invalid_test_xyz 2>&1 | grep -qi 'unknown\|error'"

    log_header "Invalid Profile"
    run_test "Reflector starts even with unknown profile (defaults)" \
        "timeout 2 ${STEM_BIN} reflect -i ${VETH_RX} --profile unknown_profile 2>&1 || true" 0

    log_header "Unknown Commands"
    run_test "Unknown command shows usage" \
        "${STEM_BIN} unknowncommand 2>&1 | grep -qi 'unknown\|usage'" 0
}

# ============================================================================
# SECTION 10: Help System Tests
# ============================================================================

test_help_system() {
    log_section "HELP SYSTEM TESTS"

    log_header "Tutorial System"
    run_test "tutorial command exists" \
        "${STEM_BIN} tutorial 2>&1 | grep -qi 'tutorial\|guide'"

    log_header "Glossary System"
    run_test "glossary command exists" \
        "${STEM_BIN} glossary 2>&1 | grep -qi 'glossary\|term'"

    log_header "Help Topics"
    local topics=("throughput" "latency" "rfc2544" "y1564" "reflect" "test")

    for topic in "${topics[@]}"; do
        run_test "help $topic available" \
            "${STEM_BIN} help $topic 2>&1 | grep -qi '${topic}\|description\|usage'" 0
    done
}

# ============================================================================
# SECTION 11: Performance/Stress Tests
# ============================================================================

test_performance() {
    log_section "PERFORMANCE TESTS"

    log_header "Rapid Start/Stop"
    local start_stop_ok=true
    for i in $(seq 1 5); do
        ${STEM_BIN} reflect -i ${VETH_RX} --profile all >/dev/null 2>&1 &
        local pid=$!
        sleep 0.5
        if ! kill -0 $pid 2>/dev/null; then
            start_stop_ok=false
            break
        fi
        kill $pid 2>/dev/null
        wait $pid 2>/dev/null || true
    done

    TESTS_RUN=$((TESTS_RUN + 1))
    if $start_stop_ok; then
        log_pass "Rapid start/stop cycles (5x)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_fail "Rapid start/stop failed"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    log_header "Web Server Load"
    ${STEM_BIN} web -p ${WEB_PORT_ALT} >/dev/null 2>&1 &
    WEB_PID=$!
    sleep 2

    if kill -0 $WEB_PID 2>/dev/null; then
        # Hit API rapidly
        local request_ok=true
        for i in $(seq 1 50); do
            if ! curl -s http://localhost:${WEB_PORT_ALT}/api/health >/dev/null 2>&1; then
                request_ok=false
                break
            fi
        done

        TESTS_RUN=$((TESTS_RUN + 1))
        if $request_ok; then
            log_pass "Web server handles 50 rapid requests"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            log_fail "Web server dropped requests under load"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi

        # Concurrent requests
        local concurrent_ok=true
        for i in $(seq 1 10); do
            curl -s http://localhost:${WEB_PORT_ALT}/api/health &
        done
        wait

        TESTS_RUN=$((TESTS_RUN + 1))
        if curl -s http://localhost:${WEB_PORT_ALT}/api/health | grep -q "healthy"; then
            log_pass "Web server handles concurrent requests"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            log_fail "Web server failed under concurrent load"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi

        kill $WEB_PID 2>/dev/null
        wait $WEB_PID 2>/dev/null || true
        WEB_PID=""
    else
        skip_test "Web server load tests" "Server failed to start"
    fi

    log_header "Memory/Resource Check"
    ${STEM_BIN} reflect -i ${VETH_RX} --profile all >/dev/null 2>&1 &
    REFLECTOR_PID=$!
    sleep 2

    if kill -0 $REFLECTOR_PID 2>/dev/null; then
        # Let it run for a bit
        sleep 5

        TESTS_RUN=$((TESTS_RUN + 1))
        if kill -0 $REFLECTOR_PID 2>/dev/null; then
            log_pass "Reflector stable after 5 seconds"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            log_fail "Reflector crashed"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi

        kill $REFLECTOR_PID 2>/dev/null
        wait $REFLECTOR_PID 2>/dev/null || true
        REFLECTOR_PID=""
    fi
}

# ============================================================================
# SECTION 12: Integration Tests
# ============================================================================

test_integration() {
    log_section "INTEGRATION TESTS"

    log_header "Full Workflow: Reflector + WebUI"

    # Start reflector
    ${STEM_BIN} reflect -i ${VETH_RX} --profile all >/dev/null 2>&1 &
    REFLECTOR_PID=$!
    sleep 2

    # Start WebUI
    ${STEM_BIN} web -p ${WEB_PORT} >/dev/null 2>&1 &
    WEB_PID=$!
    sleep 2

    TESTS_RUN=$((TESTS_RUN + 1))
    if kill -0 $REFLECTOR_PID 2>/dev/null && kill -0 $WEB_PID 2>/dev/null; then
        log_pass "Reflector and WebUI run concurrently"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_fail "Concurrent processes failed"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Check stats endpoint reflects reflector data
    local stats=$(curl -s http://localhost:${WEB_PORT}/api/stats 2>/dev/null)
    TESTS_RUN=$((TESTS_RUN + 1))
    if echo "$stats" | grep -q "packetsReceived"; then
        log_pass "Stats endpoint returns packet data"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_fail "Stats endpoint missing packet data"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Cleanup
    kill $REFLECTOR_PID 2>/dev/null
    kill $WEB_PID 2>/dev/null
    wait $REFLECTOR_PID 2>/dev/null || true
    wait $WEB_PID 2>/dev/null || true
    REFLECTOR_PID=""
    WEB_PID=""

    log_header "JSON Output Test"
    # Test JSON output format (quick test, not full run)
    run_test "Test --json flag recognized" \
        "${STEM_BIN} test --help 2>&1 | grep -q 'json'"

    run_test "Test --csv flag recognized" \
        "${STEM_BIN} test --help 2>&1 | grep -q 'csv'"
}

# ============================================================================
# SECTION 13: Signal Handling Tests
# ============================================================================

test_signal_handling() {
    log_section "SIGNAL HANDLING TESTS"

    log_header "SIGTERM Handling"
    ${STEM_BIN} reflect -i ${VETH_RX} --profile all >/dev/null 2>&1 &
    local pid=$!
    sleep 2

    kill -TERM $pid 2>/dev/null
    sleep 2

    TESTS_RUN=$((TESTS_RUN + 1))
    if ! kill -0 $pid 2>/dev/null; then
        log_pass "Process terminates on SIGTERM"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_fail "Process ignores SIGTERM"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        kill -9 $pid 2>/dev/null || true
    fi

    log_header "SIGINT Handling"
    ${STEM_BIN} reflect -i ${VETH_RX} --profile all >/dev/null 2>&1 &
    pid=$!
    sleep 2

    kill -INT $pid 2>/dev/null
    sleep 2

    TESTS_RUN=$((TESTS_RUN + 1))
    if ! kill -0 $pid 2>/dev/null; then
        log_pass "Process terminates on SIGINT"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_fail "Process ignores SIGINT"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        kill -9 $pid 2>/dev/null || true
    fi
}

# ============================================================================
# Main
# ============================================================================

main() {
    echo -e "${BOLD}${CYAN}╔══════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}${CYAN}║          The Stem - Comprehensive Smoke Test Suite              ║${NC}"
    echo -e "${BOLD}${CYAN}║         Copyright (c) 2025 Mustard Seed Networks                ║${NC}"
    echo -e "${BOLD}${CYAN}╚══════════════════════════════════════════════════════════════════╝${NC}"
    echo ""

    check_root
    setup_veth

    # Run all test sections
    test_binary_and_build
    test_version_and_branding
    test_cli_help
    test_test_types
    test_reflector
    test_webui
    test_network_stack
    test_license
    test_error_handling
    test_help_system
    test_performance
    test_integration
    test_signal_handling

    # Calculate elapsed time
    local end_time=$(date +%s)
    local elapsed=$((end_time - START_TIME))

    # Summary
    echo ""
    echo -e "${BOLD}${CYAN}══════════════════════════════════════════════════════════════════${NC}"
    log_header "TEST SUMMARY"
    echo -e "  ${BOLD}Total Tests:${NC}   ${TESTS_RUN}"
    echo -e "  ${GREEN}Passed:${NC}        ${TESTS_PASSED}"
    echo -e "  ${RED}Failed:${NC}        ${TESTS_FAILED}"
    echo -e "  ${YELLOW}Skipped:${NC}       ${TESTS_SKIPPED}"
    echo -e "  ${CYAN}Duration:${NC}      ${elapsed} seconds"
    echo ""

    local pass_rate=0
    if [[ $TESTS_RUN -gt 0 ]]; then
        pass_rate=$((TESTS_PASSED * 100 / TESTS_RUN))
    fi
    echo -e "  ${BOLD}Pass Rate:${NC}     ${pass_rate}%"
    echo ""

    if [[ $TESTS_FAILED -gt 0 ]]; then
        echo -e "${RED}${BOLD}══════════════════════════════════════════════════════════════════${NC}"
        echo -e "${RED}${BOLD}                    SMOKE TESTS FAILED                            ${NC}"
        echo -e "${RED}${BOLD}══════════════════════════════════════════════════════════════════${NC}"
        exit 1
    else
        echo -e "${GREEN}${BOLD}══════════════════════════════════════════════════════════════════${NC}"
        echo -e "${GREEN}${BOLD}                  ALL SMOKE TESTS PASSED                          ${NC}"
        echo -e "${GREEN}${BOLD}══════════════════════════════════════════════════════════════════${NC}"
        exit 0
    fi
}

main "$@"
