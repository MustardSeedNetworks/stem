/*
 * Seed Test Suite - Configuration Documentation
 *
 * Comprehensive documentation for all configuration options.
 */

package help

// ConfigSection represents a configuration section.
type ConfigSection struct {
	Name        string // "server"
	Description string // What this section configures
	Options     []ConfigOption
}

// ConfigOption documents a single configuration option.
type ConfigOption struct {
	Name       string   // "port"
	Path       string   // "server.port"
	Type       string   // "integer"
	Default    string   // "8444"
	EnvVar     string   // "STEM_SERVER_PORT"
	TechDesc   string   // Technical description
	LaymanDesc string   // Plain English
	Values     []string // Valid values for enums
	Example    string   // Example usage
}

// GetAllConfigSections returns documentation for all config sections.
func GetAllConfigSections() []ConfigSection {
	return []ConfigSection{
		serverConfig(),
		interfaceConfig(),
		reflectorConfig(),
		platformConfig(),
		statsConfig(),
		loggingConfig(),
		licenseConfig(),
	}
}

// GetConfigByPath returns a specific config option by path.
func GetConfigByPath(path string) (ConfigOption, bool) {
	for _, section := range GetAllConfigSections() {
		for _, opt := range section.Options {
			if opt.Path == path {
				return opt, true
			}
		}
	}
	return ConfigOption{}, false
}

func serverConfig() ConfigSection {
	return ConfigSection{
		Name:        "server",
		Description: "Web server and API settings for the Test Master interface",
		Options: []ConfigOption{
			{
				Name:       "port",
				Path:       "server.port",
				Type:       "integer",
				Default:    "8444",
				EnvVar:     "STEM_SERVER_PORT",
				TechDesc:   "TCP port for the HTTPS server. If port is in use during package installation, an alternative port is auto-selected.",
				LaymanDesc: "The port number for the web interface. Default is 8444 (like https://localhost:8444).",
				Example:    "port: 9000",
			},
			{
				Name:       "tls_port",
				Path:       "server.tls_port",
				Type:       "integer",
				Default:    "8443",
				EnvVar:     "STEM_SERVER_TLS_PORT",
				TechDesc:   "TCP port for HTTPS server. Set to 0 to disable TLS. Requires tls_cert and tls_key or auto-generates self-signed certificate.",
				LaymanDesc: "Secure (HTTPS) port. Set to 0 to turn off secure access.",
				Example:    "tls_port: 443",
			},
			{
				Name:       "bind_address",
				Path:       "server.bind_address",
				Type:       "string (IP address)",
				Default:    "0.0.0.0",
				EnvVar:     "STEM_SERVER_BIND",
				TechDesc:   "IP address to bind the server socket. 0.0.0.0 binds to all interfaces; 127.0.0.1 for local-only access.",
				LaymanDesc: "Which network interface to listen on. '0.0.0.0' means all of them, '127.0.0.1' means only this computer.",
				Example:    "bind_address: \"192.168.1.100\"",
			},
			{
				Name:       "tls_cert",
				Path:       "server.tls_cert",
				Type:       "string (file path)",
				Default:    "",
				EnvVar:     "STEM_TLS_CERT",
				TechDesc:   "Path to PEM-encoded TLS certificate. If empty and tls_port > 0, a self-signed certificate is generated.",
				LaymanDesc: "Location of your SSL certificate file for HTTPS. Leave empty to use auto-generated certificate.",
				Example:    "tls_cert: \"/etc/stem/server.crt\"",
			},
			{
				Name:       "tls_key",
				Path:       "server.tls_key",
				Type:       "string (file path)",
				Default:    "",
				EnvVar:     "STEM_TLS_KEY",
				TechDesc:   "Path to PEM-encoded TLS private key. Must match tls_cert if specified.",
				LaymanDesc: "Location of your SSL private key file. Must match the certificate.",
				Example:    "tls_key: \"/etc/stem/server.key\"",
			},
		},
	}
}

func interfaceConfig() ConfigSection {
	return ConfigSection{
		Name:        "interface",
		Description: "Default network interface for packet transmission and reception",
		Options: []ConfigOption{
			{
				Name:       "interface",
				Path:       "interface",
				Type:       "string (interface name)",
				Default:    "",
				EnvVar:     "STEM_INTERFACE",
				TechDesc:   "Default network interface for test traffic. Empty string triggers auto-detection based on link state and capability.",
				LaymanDesc: "Which network port to use for testing (e.g., eth0, ens192). Leave empty to auto-detect.",
				Example:    "interface: \"eth0\"",
			},
		},
	}
}

func reflectorConfig() ConfigSection {
	return ConfigSection{
		Name:        "reflector",
		Description: "Packet reflector settings for remote testing scenarios",
		Options: []ConfigOption{
			{
				Name:       "signature_filter",
				Path:       "reflector.signature_filter",
				Type:       "string",
				Default:    "all",
				EnvVar:     "STEM_REFLECTOR_SIGNATURE",
				TechDesc:   "Filter for test packet signatures. Controls which test traffic types are reflected.",
				LaymanDesc: "Which types of test packets to reflect back. 'all' means respond to everything.",
				Values:     []string{"all", "ito", "rfc2544", "y1564", "msn", "custom"},
				Example:    "signature_filter: \"rfc2544\"",
			},
			{
				Name:       "mode",
				Path:       "reflector.mode",
				Type:       "string",
				Default:    "all",
				EnvVar:     "STEM_REFLECTOR_MODE",
				TechDesc:   "Reflection mode: 'mac' swaps L2 addresses only, 'mac-ip' swaps L2 and L3, 'all' reflects all matching traffic.",
				LaymanDesc: "How to bounce packets back. 'all' works for most cases.",
				Values:     []string{"mac", "mac-ip", "all"},
				Example:    "mode: \"mac-ip\"",
			},
			{
				Name:       "port",
				Path:       "reflector.filtering.port",
				Type:       "integer",
				Default:    "3842",
				EnvVar:     "STEM_REFLECTOR_PORT",
				TechDesc:   "UDP port for ITO (Intelligent Test Operations) protocol. Set to 0 to accept any port.",
				LaymanDesc: "Port number for test traffic. Default 3842 is the standard test port.",
				Example:    "port: 3842",
			},
			{
				Name:       "filter_oui",
				Path:       "reflector.filtering.filter_oui",
				Type:       "boolean",
				Default:    "false",
				EnvVar:     "STEM_FILTER_OUI",
				TechDesc:   "Enable OUI-based MAC address filtering. Only reflects packets from matching source OUI.",
				LaymanDesc: "Only respond to packets from specific device manufacturers. Usually disabled.",
				Example:    "filter_oui: true",
			},
			{
				Name:       "oui",
				Path:       "reflector.filtering.oui",
				Type:       "string (MAC prefix)",
				Default:    "00:c0:17",
				EnvVar:     "STEM_OUI",
				TechDesc:   "OUI prefix (first 3 octets of MAC address) to filter on when filter_oui is enabled.",
				LaymanDesc: "The manufacturer code to filter on (e.g., 00:c0:17 for NetOptics).",
				Example:    "oui: \"00:1a:2b\"",
			},
			{
				Name:       "filter_mac",
				Path:       "reflector.filtering.filter_mac",
				Type:       "boolean",
				Default:    "false",
				EnvVar:     "STEM_FILTER_MAC",
				TechDesc:   "Enable destination MAC filtering. Only reflects packets destined for specific MAC addresses.",
				LaymanDesc: "Only respond to packets sent to specific addresses. Usually disabled.",
				Example:    "filter_mac: true",
			},
		},
	}
}

func platformConfig() ConfigSection {
	return ConfigSection{
		Name:        "platform",
		Description: "Low-level packet processing platform settings (advanced)",
		Options: []ConfigOption{
			{
				Name:       "use_dpdk",
				Path:       "platform.use_dpdk",
				Type:       "boolean",
				Default:    "false",
				EnvVar:     "STEM_USE_DPDK",
				TechDesc:   "Enable DPDK (Data Plane Development Kit) for kernel-bypass packet processing. Requires DPDK installation and hugepages configuration.",
				LaymanDesc: "Use ultra-high-speed packet processing. Only enable if DPDK is installed and configured.",
				Example:    "use_dpdk: true",
			},
			{
				Name:       "use_af_xdp",
				Path:       "platform.use_af_xdp",
				Type:       "boolean",
				Default:    "true",
				EnvVar:     "STEM_USE_AF_XDP",
				TechDesc:   "Use AF_XDP sockets for high-performance packet I/O. Faster than AF_PACKET, slower than DPDK. Requires Linux 5.0+.",
				LaymanDesc: "Use faster packet processing on Linux. Recommended to leave enabled.",
				Example:    "use_af_xdp: true",
			},
			{
				Name:       "dpdk_args",
				Path:       "platform.dpdk_args",
				Type:       "string",
				Default:    "",
				EnvVar:     "STEM_DPDK_ARGS",
				TechDesc:   "DPDK EAL (Environment Abstraction Layer) arguments. Used to configure DPDK core affinity, memory, and device binding.",
				LaymanDesc: "Advanced DPDK settings. Leave empty unless you know what you're doing.",
				Example:    "dpdk_args: \"-l 0-3 -n 4 --socket-mem 1024\"",
			},
		},
	}
}

func statsConfig() ConfigSection {
	return ConfigSection{
		Name:        "stats",
		Description: "Statistics collection and output format settings",
		Options: []ConfigOption{
			{
				Name:       "format",
				Path:       "stats.format",
				Type:       "string",
				Default:    "json",
				EnvVar:     "STEM_STATS_FORMAT",
				TechDesc:   "Output format for statistics. JSON for programmatic parsing, text for human readability, CSV for spreadsheets.",
				LaymanDesc: "How to format results. 'json' works best with the web interface.",
				Values:     []string{"text", "json", "csv"},
				Example:    "format: \"csv\"",
			},
			{
				Name:       "interval",
				Path:       "stats.interval",
				Type:       "integer (seconds)",
				Default:    "10",
				EnvVar:     "STEM_STATS_INTERVAL",
				TechDesc:   "Interval between statistics updates in seconds. Lower values provide more granular data but increase CPU usage.",
				LaymanDesc: "How often to update statistics. 10 seconds is a good balance.",
				Example:    "interval: 5",
			},
		},
	}
}

func loggingConfig() ConfigSection {
	return ConfigSection{
		Name:        "logging",
		Description: "Application logging settings",
		Options: []ConfigOption{
			{
				Name:       "level",
				Path:       "logging.level",
				Type:       "string",
				Default:    "info",
				EnvVar:     "STEM_LOG_LEVEL",
				TechDesc:   "Minimum log level to output. debug < info < warn < error. Debug includes detailed packet traces.",
				LaymanDesc: "How much detail to log. 'info' is normal, 'debug' shows everything.",
				Values:     []string{"debug", "info", "warn", "error"},
				Example:    "level: \"debug\"",
			},
			{
				Name:       "format",
				Path:       "logging.format",
				Type:       "string",
				Default:    "json",
				EnvVar:     "STEM_LOG_FORMAT",
				TechDesc:   "Log output format. JSON for structured logging (recommended for production), text for human readability.",
				LaymanDesc: "Format for log messages. 'json' works best with log management tools.",
				Values:     []string{"text", "json"},
				Example:    "format: \"text\"",
			},
			{
				Name:       "file",
				Path:       "logging.file",
				Type:       "string (file path)",
				Default:    "",
				EnvVar:     "STEM_LOG_FILE",
				TechDesc:   "Path to log file. Empty string logs to stdout (captured by systemd/journald). File rotation is not built-in.",
				LaymanDesc: "Where to save logs. Leave empty to use system logging (journalctl).",
				Example:    "file: \"/var/log/stem/stem.log\"",
			},
		},
	}
}

func licenseConfig() ConfigSection {
	return ConfigSection{
		Name:        "license",
		Description: "License activation and validation settings",
		Options: []ConfigOption{
			{
				Name:       "key",
				Path:       "license.key",
				Type:       "string",
				Default:    "",
				EnvVar:     "STEM_LICENSE_KEY",
				TechDesc:   "License key in format XXXX-XXXX-XXXX-XXXX. Determines available features (Reflector, TestSuite, Enterprise tiers).",
				LaymanDesc: "Your license key from Mustard Seed Networks. Controls which features are available.",
				Example:    "key: \"ABCD-1234-EFGH-5678\"",
			},
			{
				Name:       "server",
				Path:       "license.server",
				Type:       "string (URL)",
				Default:    "",
				EnvVar:     "STEM_LICENSE_SERVER",
				TechDesc:   "License server URL for online validation. Empty for offline license validation.",
				LaymanDesc: "License server address. Usually left empty for offline licenses.",
				Example:    "server: \"https://license.mustard-seed.net\"",
			},
		},
	}
}

// GetEnvironmentVariables returns all supported environment variables.
func GetEnvironmentVariables() []ConfigOption {
	var envVars []ConfigOption
	for _, section := range GetAllConfigSections() {
		for _, opt := range section.Options {
			if opt.EnvVar != "" {
				envVars = append(envVars, opt)
			}
		}
	}

	// Add additional environment-only variables
	envVars = append(envVars, []ConfigOption{
		{
			Name:       "Auth Username",
			Path:       "",
			Type:       "string",
			Default:    "admin",
			EnvVar:     "STEM_AUTH_USERNAME",
			TechDesc:   "Username for web interface authentication. Set in /etc/stem/environment for security.",
			LaymanDesc: "Login username for the web interface.",
			Example:    "STEM_AUTH_USERNAME=operator",
		},
		{
			Name:       "Auth Password",
			Path:       "",
			Type:       "string",
			Default:    "",
			EnvVar:     "STEM_AUTH_PASSWORD",
			TechDesc:   "Password for web interface authentication. Must be set before enabling auth. Never store in config file.",
			LaymanDesc: "Login password for the web interface. MUST be changed from default.",
			Example:    "STEM_AUTH_PASSWORD=SecureP@ssw0rd!",
		},
		{
			Name:       "JWT Secret",
			Path:       "",
			Type:       "string",
			Default:    "",
			EnvVar:     "STEM_JWT_SECRET",
			TechDesc:   "Secret key for JWT token signing. Should be cryptographically random, minimum 32 characters. Auto-generated if empty.",
			LaymanDesc: "Security key for login sessions. Auto-generated if not set.",
			Example:    "STEM_JWT_SECRET=your-64-character-random-string-here",
		},
	}...)

	return envVars
}

// ConfigHelp provides overall configuration guidance.
type ConfigHelp struct {
	FilePath    string
	Description string
	Sections    []ConfigSection
	EnvVars     []ConfigOption
}

// GetConfigHelp returns comprehensive configuration documentation.
func GetConfigHelp() ConfigHelp {
	return ConfigHelp{
		FilePath: "/etc/stem/config.yaml",
		Description: `The Stem configuration file controls all aspects of the application including
web server settings, packet processing, and logging. Configuration can also be
overridden using environment variables (useful for containers and secrets).

Environment variables take precedence over config file values.

Sensitive values (passwords, JWT secrets, license keys) should be stored in
/etc/stem/environment, which is sourced by the systemd service.`,
		Sections: GetAllConfigSections(),
		EnvVars:  GetEnvironmentVariables(),
	}
}
