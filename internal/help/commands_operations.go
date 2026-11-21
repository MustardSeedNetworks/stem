package help

// The operations verbs: the commands an operator uses to run the service
// rather than to test with it. They share one exit-status contract, spelled
// out where it matters (status).

// StatusCommand documents the status subcommand.
func StatusCommand() CommandHelp {
	return CommandHelp{
		Name:    "status",
		Summary: "Report whether the daemon is running and ready",
		Description: `The status command asks the stem daemon what it is and whether it is ready,
from /__version and /health/ready. Neither needs a credential.

The daemon is found through the descriptor it publishes in its data directory
(STEM_DATA_DIR, else the packaged location), and its certificate is trusted
exactly as published; there is no way to skip verification. For a daemon on
another host, name it with the flag --api and trust its certificate with the
flag --cacert.

Exit status:
  0  the daemon is running and ready
  1  not running: nothing is listening, no descriptor was published, or the
     daemon is up but not ready
  2  the check could not be made (a certificate this host does not trust, a
     descriptor this account may not read, a bad flag)

--json prints one object whatever happened, with a "state" of running,
not_ready, not_running or error.`,
		Usage: "stem status [--json] [--api URL] [--cacert FILE]",
		Flags: []FlagHelp{
			{
				Short:      "",
				Long:       "--json",
				Type:       TypeBoolean,
				Default:    ValueFalse,
				Required:   false,
				TechDesc:   "Print the status as a JSON object; always an object, with state, running, version and the failing checks",
				LaymanDesc: "Print the result for a script instead of a person",
			},
			{
				Short:      "",
				Long:       "--api",
				Type:       TypeString,
				Default:    "",
				Required:   false,
				TechDesc:   "Daemon base URL, for a daemon this host holds no descriptor for (default: the daemon on this host that published a descriptor)",
				LaymanDesc: "Check a daemon on another machine",
			},
			{
				Short:      "",
				Long:       "--cacert",
				Type:       TypeString,
				Default:    "",
				Required:   false,
				TechDesc:   "PEM certificate to trust for the daemon (default: the certificate the daemon published)",
				LaymanDesc: "The certificate of the daemon you are checking",
			},
		},
		Examples: []Example{
			{
				Desc:    "Check the daemon on this host",
				Command: "stem status",
				Output:  "Status:      running\nURL:         https://127.0.0.1:8444\nVersion:     v0.32.58 (commit 64e328e, built 2026-11-21T09:00:00Z)",
			},
			{
				Desc:    "Branch on the result in a script",
				Command: "stem status --json > /dev/null; echo $?",
				Output:  "0",
			},
			{
				Desc:    "Check a daemon on another host",
				Command: "stem status --api https://10.44.40.23:8444 --cacert ./stem-server.crt",
				Output:  "Status:      running",
			},
		},
		SeeAlso: []string{"logs", "support-bundle", "version"},
	}
}

// LogsCommand documents the logs subcommand.
func LogsCommand() CommandHelp {
	return CommandHelp{
		Name:    "logs",
		Summary: "Show the service log",
		Description: `The logs command shows the end of the service log, or follows it.

The daemon writes its log to standard output, so where it ends up depends on
what runs it. On Linux that is the systemd journal, and logs runs "journalctl
-u stem" (you need to be root or in the systemd-journal group). On macOS it is
<data directory>/logs/stem.log, the file the launchd service writes. On other
systems, and for a log captured some other way, name the file with --file.

A followed file survives log rotation: when the file is replaced or truncated,
logs starts again from the top of the new one.`,
		Usage: "stem logs [-f] [-n N] [--file FILE]",
		Flags: []FlagHelp{
			{
				Short:      "-f",
				Long:       "--follow",
				Type:       TypeBoolean,
				Default:    ValueFalse,
				Required:   false,
				TechDesc:   "Keep printing new lines as they are written, until interrupted",
				LaymanDesc: "Watch the log live",
			},
			{
				Short:      "-n",
				Long:       "--lines",
				Type:       TypeInteger,
				Default:    "100",
				Required:   false,
				TechDesc:   "Number of recent lines to show first",
				LaymanDesc: "How many recent lines to show",
			},
			{
				Short:      "",
				Long:       "--file",
				Type:       TypeString,
				Default:    "",
				Required:   false,
				TechDesc:   "Read this log file instead of the service's own",
				LaymanDesc: "Read a log you saved somewhere else",
			},
		},
		Examples: []Example{
			{
				Desc:    "Show the last 100 lines",
				Command: "stem logs",
				Output:  "",
			},
			{
				Desc:    "Follow the log",
				Command: "stem logs -f",
				Output:  "",
			},
			{
				Desc:    "Read a log captured from a foreground run",
				Command: "stem logs --file ./stem.log --lines 50",
				Output:  "",
			},
		},
		SeeAlso: []string{"status", "support-bundle"},
	}
}

// SupportBundleCommand documents the support-bundle subcommand.
func SupportBundleCommand() CommandHelp {
	return CommandHelp{
		Name:    "support-bundle",
		Summary: "Collect redacted diagnostics for support",
		Description: `The support-bundle command writes one archive with what support asks for
first: the build, the host, the daemon's status, the redacted settings files, an
inventory of the data directory and the end of the log.

Every credential is removed before anything is written. Settings files are
redacted over their structure, the administrator credential file is reduced to
its field names, and the log is scrubbed of the credential values found in the
settings, this command's environment and the activation state, plus anything else shaped
like a token, a password hash, a licence key or a private key. Certificates are
listed by name and never opened.

Read the bundle before sending it. It is redacted, not anonymised: addresses,
interface names and the contents of the log are still in it.`,
		Usage: "stem support-bundle [-o FILE] [--log FILE] [--lines N]",
		Flags: []FlagHelp{
			{
				Short:      "-o",
				Long:       "--output",
				Type:       TypeString,
				Default:    "",
				Required:   false,
				TechDesc:   "Bundle to write; never overwritten (default: stem-support-<time>.tar.gz in the current directory)",
				LaymanDesc: "Where to save the bundle",
			},
			{
				Short:      "",
				Long:       "--log",
				Type:       TypeString,
				Default:    "",
				Required:   false,
				TechDesc:   "Log file to include instead of the service's own",
				LaymanDesc: "A log you saved somewhere else",
			},
			{
				Short:      "",
				Long:       "--lines",
				Type:       TypeInteger,
				Default:    "2000",
				Required:   false,
				TechDesc:   "Number of recent log lines to include",
				LaymanDesc: "How much of the log to include",
			},
		},
		Examples: []Example{
			{
				Desc:    "Collect a bundle",
				Command: "sudo stem support-bundle -o stem-support.tar.gz",
				Output:  "Support bundle written to stem-support.tar.gz",
			},
			{
				Desc:    "Include a log captured from a foreground run",
				Command: "stem support-bundle --log ./stem.log",
				Output:  "Support bundle written to stem-support-20261121T090000Z.tar.gz",
			},
		},
		SeeAlso: []string{"status", "logs"},
	}
}

// BackupCommand documents the backup subcommand.
func BackupCommand() CommandHelp {
	return CommandHelp{
		Name:    "backup",
		Summary: "Archive the daemon's state",
		Description: `The backup command archives the daemon's state: the administrator credential,
the reflector and update settings, the TLS certificate and key, the activation
state and the published descriptor. The data directory is STEM_DATA_DIR, else
the packaged location.

The daemon must be stopped: the command refuses while it holds the data
directory. The TLS pair is read from <data directory>/certs, where the packaged
services keep it.

The archive holds credentials and a private key, so it is written readable by
its owner alone and is never overwritten. The instance lock and the first-run
and recovery tokens are not in it. Two backups of unchanged state are
byte-identical.`,
		Usage: "stem backup [-o FILE]",
		Flags: []FlagHelp{
			{
				Short:      "-o",
				Long:       "--output",
				Type:       TypeString,
				Default:    "",
				Required:   false,
				TechDesc:   "Archive to write; never overwritten (default: stem-backup-<time>.tar.gz in the current directory)",
				LaymanDesc: "Where to save the backup",
			},
		},
		Examples: []Example{
			{
				Desc:    "Back up the service's state",
				Command: "sudo stem backup -o /srv/backups/stem.tar.gz",
				Output:  "State in /var/lib/stem backed up to /srv/backups/stem.tar.gz",
			},
		},
		SeeAlso: []string{"restore", "support-bundle"},
	}
}

// RestoreCommand documents the restore subcommand.
func RestoreCommand() CommandHelp {
	return CommandHelp{
		Name:    "restore",
		Summary: "Restore the daemon's state from a backup",
		Description: `The restore command replaces the daemon's state with an archive written by
"stem backup". Run it as root, or as the service account, with the daemon
stopped.

The restore is refused while the daemon is running: it takes the same lock the
daemon holds for its whole life, and keeps it until it is done. Before anything
is replaced the current state is copied beside the data directory
(<data directory>.pre-restore-<time>), and the archive is checked in full
first, so a refused archive changes nothing. An entry that is not part of a stem
backup, or that names a path outside the data directory, refuses the archive.

Restored files take the data directory's owner and are readable by it alone.`,
		Usage: "stem restore ARCHIVE",
		Flags: []FlagHelp{},
		Examples: []Example{
			{
				Desc:    "Restore after stopping the service",
				Command: "sudo systemctl stop stem && sudo stem restore /srv/backups/stem.tar.gz",
				Output:  "State restored from /srv/backups/stem.tar.gz into /var/lib/stem",
			},
		},
		SeeAlso: []string{"backup"},
	}
}
