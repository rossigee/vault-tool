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

### TLS Verification
vault-tool enforces full TLS certificate verification for all Vault connections. There is no option to disable certificate validation.

### Credential Management
- Never commit secrets or credentials to version control
- Unseal keys should be stored encrypted (e.g., using GPG)
- Generated root tokens should be revoked immediately after use
- Always use `shred` or equivalent to securely delete token files

### Audit Trail
This tool is intended for break-glass emergency scenarios only. For routine administrative operations, use Vault's normal authentication methods to maintain proper audit logs.

## Disclosure Timeline

- **Day 0**: Vulnerability reported
- **Day 1-7**: Initial assessment and response
- **Day 7-14**: Fix development and testing
- **Day 14+**: Release and public disclosure (if needed)

Thank you for helping keep vault-tool secure!
