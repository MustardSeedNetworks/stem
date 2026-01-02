# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please report it responsibly.

### How to Report

1. **Do NOT open a public issue** for security vulnerabilities
2. Email security concerns to the maintainer
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### What to Expect

- **Acknowledgment** within 48 hours
- **Initial assessment** within 7 days
- **Resolution timeline** communicated based on severity
- **Credit** in release notes (if desired)

### Severity Levels

| Level    | Description                         | Target Resolution |
| -------- | ----------------------------------- | ----------------- |
| Critical | Remote code execution, auth bypass  | 24-48 hours       |
| High     | Data exposure, privilege escalation | 7 days            |
| Medium   | Limited impact vulnerabilities      | 30 days           |
| Low      | Minor issues, hardening             | Next release      |

## Security Best Practices

When deploying The Stem:

### Network Security

- Deploy on isolated/management networks when possible
- Use firewall rules to restrict access to the web interface
- Consider VPN access for remote management

### Authentication

- Change default credentials immediately
- Use strong passwords (12+ characters)
- Rotate credentials periodically

### HTTPS

- Use valid TLS certificates in production
- Self-signed certificates are acceptable for isolated networks

### Updates

- Keep The Stem updated to the latest version
- Subscribe to release notifications
- Review changelogs for security fixes

## Security Features

The Stem includes:

- HTTPS support
- Password authentication
- Rate limiting on auth endpoints
- Minimal attack surface (single binary)
- No default open ports (except configured interface)

## Scope

The following are in scope for security reports:

- Authentication/authorization bypass
- Remote code execution
- Command injection
- Sensitive data exposure
- Privilege escalation

The following are out of scope:

- Denial of service (DoS)
- Social engineering
- Physical access attacks
- Issues requiring root access already

## Acknowledgments

We appreciate security researchers who help keep The Stem secure. Contributors will be acknowledged in release notes unless they prefer to remain anonymous.
