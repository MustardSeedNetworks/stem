/*
 * Seed Test Suite - Error Messages
 *
 * Enhanced error messages with context, causes, and solutions.
 */

package help

import (
	"fmt"
	"io"
	"maps"
	"os"
)

// GetAllErrors returns all error help entries.
func GetAllErrors() map[string]ErrorHelp {
	errors := make(map[string]ErrorHelp)

	// Add interface errors.
	maps.Copy(errors, getInterfaceErrors())

	// Add permission errors.
	maps.Copy(errors, getPermissionErrors())

	// Add license errors.
	maps.Copy(errors, getLicenseErrors())

	// Add test errors.
	maps.Copy(errors, getTestErrors())

	// Add configuration errors.
	maps.Copy(errors, getConfigErrors())

	// Add system errors.
	maps.Copy(errors, getSystemErrors())

	return errors
}

// getInterfaceErrors returns interface-related error help.
func getInterfaceErrors() map[string]ErrorHelp {
	return map[string]ErrorHelp{
		"ERR_INTERFACE_REQUIRED": {
			Code:    "ERR_INTERFACE_REQUIRED",
			Message: "Network interface is required",
			Cause:   "The --interface (-i) flag was not specified",
			Solution: `Specify the network interface to use for testing.

To see available interfaces:
  ip link show

Look for interfaces that are UP and have a carrier (connected).`,
			Examples: []Example{
				{Desc: "Reflector example", Command: "stem reflect -i eth0", Output: ""},
				{Desc: "Test example", Command: "stem test -i enp3s0 -t throughput", Output: ""},
			},
			RelatedCmd: "stem help reflect",
		},
		"ERR_INTERFACE_NOT_FOUND": {
			Code:    "ERR_INTERFACE_NOT_FOUND",
			Message: "Network interface not found",
			Cause:   "The specified interface does not exist on this system",
			Solution: `Check the interface name is correct.

List available interfaces:
  ip link show

Common interface names:
  - eth0, eth1    (traditional naming)
  - enp3s0, ens1  (predictable naming)
  - eno1          (onboard)`,
			Examples: []Example{
				{Desc: "List interfaces", Command: "ip link show", Output: ""},
			},
			RelatedCmd: "stem help reflect",
		},
		"ERR_INTERFACE_DOWN": {
			Code:    "ERR_INTERFACE_DOWN",
			Message: "Network interface is down",
			Cause:   "The specified interface exists but is not active",
			Solution: `Bring the interface up:
  sudo ip link set eth0 up

Check if cable is connected:
  ethtool eth0 | grep "Link detected"`,
			Examples: []Example{
				{Desc: "Bring interface up", Command: "sudo ip link set eth0 up", Output: ""},
				{Desc: "Check link status", Command: "ethtool eth0", Output: ""},
			},
			RelatedCmd: "",
		},
	}
}

// getPermissionErrors returns permission-related error help.
func getPermissionErrors() map[string]ErrorHelp {
	return map[string]ErrorHelp{
		"ERR_PERMISSION_DENIED": {
			Code:    "ERR_PERMISSION_DENIED",
			Message: "Permission denied - root privileges required",
			Cause:   "Network testing requires elevated privileges for raw socket access",
			Solution: `Run with sudo:
  sudo stem reflect -i eth0

Or configure capabilities (advanced):
  sudo setcap cap_net_raw,cap_net_admin+ep /usr/local/bin/stem`,
			Examples: []Example{
				{Desc: "Run with sudo", Command: "sudo stem reflect -i eth0", Output: ""},
			},
			RelatedCmd: "",
		},
	}
}

// getLicenseErrors returns license-related error help.
func getLicenseErrors() map[string]ErrorHelp {
	return map[string]ErrorHelp{
		"ERR_LICENSE_REQUIRED": {
			Code:    "ERR_LICENSE_REQUIRED",
			Message: "Valid license required",
			Cause:   "No license key has been activated",
			Solution: `Activate your license:
  stem license -k XXXX-XXXX-XXXX-XXXX

Purchase a license at: https://mustardseednetworks.com

Check current status:
  stem license --status`,
			Examples: []Example{
				{Desc: "Activate license", Command: "stem license -k ABCD-1234-EFGH-5678", Output: ""},
				{Desc: "Check status", Command: "stem license --status", Output: ""},
			},
			RelatedCmd: "stem help license",
		},
		"ERR_LICENSE_INVALID": {
			Code:    "ERR_LICENSE_INVALID",
			Message: "Invalid license key",
			Cause:   "The provided license key is not valid",
			Solution: `Check the key format: XXXX-XXXX-XXXX-XXXX

Ensure no typos or extra characters.

If you believe this is an error, contact support.`,
			Examples: []Example{
				{Desc: "Correct format", Command: "stem license -k ABCD-1234-EFGH-5678", Output: ""},
			},
			RelatedCmd: "stem help license",
		},
		"ERR_LICENSE_EXPIRED": {
			Code:    "ERR_LICENSE_EXPIRED",
			Message: "License has expired",
			Cause:   "The license key has passed its expiration date",
			Solution: `Renew your license at: https://mustardseednetworks.com

Contact sales for renewal options.`,
			Examples:   []Example{},
			RelatedCmd: "stem help license",
		},
		"ERR_FEATURE_NOT_LICENSED": {
			Code:    "ERR_FEATURE_NOT_LICENSED",
			Message: "Feature not available in current license tier",
			Cause:   "Your license does not include this feature",
			Solution: `Available tiers:
  - Reflector: Packet reflection only
  - TestSuite: Full test suite (RFC 2544, Y.1564, etc.)
  - Enterprise: All features + API + multi-user

Upgrade at: https://mustardseednetworks.com

Check your current features:
  stem license --status`,
			Examples: []Example{
				{Desc: "Check features", Command: "stem license --status", Output: ""},
			},
			RelatedCmd: "stem help license",
		},
	}
}

// getTestErrors returns test-related error help.
func getTestErrors() map[string]ErrorHelp {
	return map[string]ErrorHelp{
		"ERR_TEST_TYPE_INVALID": {
			Code:    "ERR_TEST_TYPE_INVALID",
			Message: "Unknown test type",
			Cause:   "The specified test type is not recognized",
			Solution: `See available tests:
  stem help tests

Or get help on a specific test:
  stem help throughput`,
			Examples: []Example{
				{Desc: "List tests", Command: "stem help tests", Output: ""},
				{Desc: "Throughput test", Command: "stem test -i eth0 -t throughput", Output: ""},
			},
			RelatedCmd: "stem help tests",
		},
		"ERR_REFLECTOR_UNREACHABLE": {
			Code:    "ERR_REFLECTOR_UNREACHABLE",
			Message: "Cannot reach reflector",
			Cause:   "No response from the target reflector",
			Solution: `Check:
1. Reflector is running:
   ssh user@target 'stem reflect -i eth0'

2. Network connectivity:
   ping <target-ip>

3. Firewall allows UDP port 3842:
   sudo iptables -L | grep 3842

4. Correct target IP specified:
   stem test -i eth0 -t throughput --target <correct-ip>`,
			Examples: []Example{
				{Desc: "Start remote reflector", Command: "ssh user@target 'stem reflect -i eth0'", Output: ""},
				{Desc: "Check connectivity", Command: "ping <target-ip>", Output: ""},
			},
			RelatedCmd: "stem help reflect",
		},
		"ERR_TEST_TIMEOUT": {
			Code:    "ERR_TEST_TIMEOUT",
			Message: "Test timed out",
			Cause:   "Test did not complete within expected time",
			Solution: `Possible causes:
1. Reflector stopped or crashed
2. Network connectivity lost
3. Very high packet loss (all frames dropped)

Try:
- Verify reflector status
- Run with shorter duration
- Check for network issues`,
			Examples: []Example{
				{Desc: "Shorter test", Command: "stem test -i eth0 -t throughput -d 30", Output: ""},
			},
			RelatedCmd: "",
		},
	}
}

// getConfigErrors returns configuration-related error help.
func getConfigErrors() map[string]ErrorHelp {
	return map[string]ErrorHelp{
		"ERR_CIR_REQUIRED": {
			Code:    "ERR_CIR_REQUIRED",
			Message: "CIR (Committed Information Rate) is required for this test",
			Cause:   "Y.1564 and MEF tests require a CIR to be specified",
			Solution: `Specify the CIR from your service contract:
  --cir <rate in Mbps>

Example: For a 100 Mbps service:`,
			Examples: []Example{
				{Desc: "100 Mbps service", Command: "stem test -i eth0 -t y1564_config --cir 100", Output: ""},
			},
			RelatedCmd: "stem help y1564_config",
		},
		"ERR_INVALID_FRAME_SIZE": {
			Code:    "ERR_INVALID_FRAME_SIZE",
			Message: "Invalid frame size specified",
			Cause:   "Frame size must be between 64 and 9000 bytes",
			Solution: `Valid frame sizes:
  - Minimum: 64 bytes (Ethernet minimum)
  - Standard max: 1518 bytes
  - Jumbo: up to 9000 bytes (if supported)

Standard RFC 2544 sizes: 64, 128, 256, 512, 1024, 1280, 1518`,
			Examples: []Example{
				{
					Desc:    "Standard sizes",
					Command: "stem test -i eth0 -t throughput --frame-sizes 64,128,256,512,1024,1280,1518",
					Output:  "",
				},
				{
					Desc:    "Custom sizes",
					Command: "stem test -i eth0 -t throughput --frame-sizes 64,512,1518",
					Output:  "",
				},
			},
			RelatedCmd: "stem help throughput",
		},
	}
}

// getSystemErrors returns system-related error help.
func getSystemErrors() map[string]ErrorHelp {
	return map[string]ErrorHelp{
		"ERR_AF_XDP_NOT_AVAILABLE": {
			Code:    "ERR_AF_XDP_NOT_AVAILABLE",
			Message: "AF_XDP mode not available",
			Cause:   "Kernel does not support AF_XDP or interface doesn't support XDP",
			Solution: `Requirements for AF_XDP:
  - Linux kernel 5.x or later
  - Interface with XDP support
  - Root privileges

Check kernel version:
  uname -r

Fall back to AF_PACKET mode:
  stem reflect -i eth0 --mode af_packet`,
			Examples: []Example{
				{Desc: "Check kernel", Command: "uname -r", Output: ""},
				{Desc: "Use AF_PACKET", Command: "stem reflect -i eth0 --mode af_packet", Output: ""},
			},
			RelatedCmd: "stem help reflect",
		},
		"ERR_DPDK_NOT_AVAILABLE": {
			Code:    "ERR_DPDK_NOT_AVAILABLE",
			Message: "DPDK mode not available",
			Cause:   "DPDK libraries not installed or interface not configured",
			Solution: `DPDK requires:
  - DPDK libraries installed
  - Interface bound to DPDK driver
  - Hugepages configured

Check DPDK setup:
  dpdk-devbind.py --status

Fall back to AF_XDP or AF_PACKET:
  stem reflect -i eth0 --mode af_xdp`,
			Examples: []Example{
				{Desc: "Check DPDK", Command: "dpdk-devbind.py --status", Output: ""},
				{Desc: "Use AF_XDP", Command: "stem reflect -i eth0 --mode af_xdp", Output: ""},
			},
			RelatedCmd: "stem help reflect",
		},
	}
}

// PrintError prints a formatted error message with help.
func PrintError(code string) {
	PrintErrorTo(os.Stderr, code)
}

// PrintErrorTo prints a formatted error message with help to the specified writer.
func PrintErrorTo(w io.Writer, code string) {
	errors := GetAllErrors()
	errHelp, ok := errors[code]
	if !ok {
		_, _ = fmt.Fprintf(w, "Error: Unknown error code %s\n", code)
		return
	}

	_, _ = fmt.Fprintf(w, "\n❌ Error: %s\n", errHelp.Message)
	_, _ = fmt.Fprintf(w, "\n%s\n", errHelp.Solution)

	if len(errHelp.Examples) > 0 {
		_, _ = fmt.Fprint(w, "\nExamples:\n")
		for _, ex := range errHelp.Examples {
			_, _ = fmt.Fprintf(w, "  %s\n", ex.Command)
		}
	}

	if errHelp.RelatedCmd != "" {
		_, _ = fmt.Fprintf(w, "\nFor more information: %s\n", errHelp.RelatedCmd)
	}
	_, _ = fmt.Fprintln(w)
}

// PrintErrorWithDetails prints error with custom details.
func PrintErrorWithDetails(code string, details string) {
	PrintErrorWithDetailsTo(os.Stderr, code, details)
}

// PrintErrorWithDetailsTo prints error with custom details to the specified writer.
func PrintErrorWithDetailsTo(w io.Writer, code, details string) {
	errors := GetAllErrors()
	errHelp, ok := errors[code]
	if !ok {
		_, _ = fmt.Fprintf(w, "Error: %s\n", details)
		return
	}

	_, _ = fmt.Fprintf(w, "\n❌ Error: %s\n", errHelp.Message)
	if details != "" {
		_, _ = fmt.Fprintf(w, "   %s\n", details)
	}
	_, _ = fmt.Fprintf(w, "\n%s\n", errHelp.Solution)

	if len(errHelp.Examples) > 0 {
		_, _ = fmt.Fprint(w, "\nExamples:\n")
		for _, ex := range errHelp.Examples {
			_, _ = fmt.Fprintf(w, "  %s\n", ex.Command)
		}
	}

	if errHelp.RelatedCmd != "" {
		_, _ = fmt.Fprintf(w, "\nFor more information: %s\n", errHelp.RelatedCmd)
	}
	_, _ = fmt.Fprintln(w)
}
