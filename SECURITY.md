# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in vault-tool, please do NOT open a public GitHub issue. Instead, please report it responsibly using one of the following methods:

### Option 1: GitHub Security Advisory (Preferred)
Use GitHub's private vulnerability disclosure feature:
1. Go to the [Security Advisories](https://github.com/rossigee/vault-tool/security/advisories) page
2. Click "Report a vulnerability"
3. Fill out the form with details about the vulnerability
4. Submit the report privately

### Option 2: Email
Send an email to **ross@golder.org** with:
- A clear description of the vulnerability
- Steps to reproduce (if applicable)
- Potential impact
- Suggested fix (if you have one)

Please allow 7-14 days for an initial response.

## Supported Versions

Security updates are provided for the latest stable release. For previous versions, we recommend upgrading to the latest version.

| Version | Supported |
|---------|-----------|
| Latest  | ✅        |
| Older   | ⚠️        |

## Security Considerations

### Prerequisites & Risk

**⚠️ CRITICAL:** vault-tool requires Vault to be configured with unauthenticated access to the root token generation endpoint:

```hcl
enable_unauthenticated_access = ["generate-root"]
```

This is a **non-standard, high-risk Vault configuration** that should only be enabled for break-glass scenarios. This configuration:
- Allows anyone with access to the Vault API to initiate root token generation (requires unseal keys to complete)
- Bypasses Vault's normal authentication and audit logging for the initial step
- Should be strictly controlled and monitored

**Mitigation:**
- Only enable this ACL rule in high-security environments where break-glass access is essential
- Disable it immediately after resolving the emergency if possible
- Monitor Vault audit logs for unexpected calls to `/sys/generate-root/*` endpoints
- Restrict network access to Vault to trusted operators only

### TLS Verification
vault-tool enforces full TLS certificate verification for all Vault connections. There is no option to disable certificate validation. This prevents man-in-the-middle attacks.

### Credential Management
- Never commit secrets or credentials to version control
- Unseal keys should be stored encrypted (e.g., using GPG) with strong passphrases
- Generated root tokens should be revoked **immediately** after use — do not leave them active
- Always use `shred` or equivalent to securely delete token files: `shred -u ~/.vault-token`
- Consider using OTP-wrapped tokens (`--use-otp`) for secure multi-channel distribution when working in teams

### Logging Security
- By default, vault-tool is quiet (no logging) unless `-v` or `-d` flags are used
- Even with verbose logging enabled, vault-tool does NOT log plaintext tokens, OTPs, passphrases, or unseal keys
- Avoid redirecting logs to insecure files; pipe to encrypted storage if necessary
- Debug logs contain metadata (lengths, counts, nonces) but not sensitive values

### Audit Trail
This tool is intended for **break-glass emergency scenarios only**. For routine administrative operations, use Vault's normal authentication methods to maintain proper audit logs. Using this tool in place of normal authentication:
- Bypasses per-user accountability (cannot attribute actions to specific users)
- May violate compliance/regulatory requirements (SOX, PCI-DSS, HIPAA, etc.)
- Reduces incident response visibility
- Should only be used when normal auth is completely unavailable

## Disclosure Timeline

- **Day 0**: Vulnerability reported
- **Day 1-7**: Initial assessment and response
- **Day 7-14**: Fix development and testing
- **Day 14+**: Release and public disclosure (if needed)

Thank you for helping keep vault-tool secure!
