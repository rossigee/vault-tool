# Vault Tool

A Go CLI tool for Vault emergency access and maintenance, providing sealed instance unsealing and emergency root token generation with OTP protection.

**Latest Release:** v0.2.3

## Overview

`vault-tool` is a **break-glass emergency tool** for Vault incident response. Normal admin operations should use OIDC or other per-user authentication for audit trail purposes. 

This tool enables operators to unseal a sealed Vault instance and generate temporary emergency root tokens **only when normal authentication (OIDC, tokens, etc.) is unavailable**. It uses GPG-encrypted unseal keys for secure key storage and management.

**Prerequisites:** Your Vault server must be configured with `enable_unauthenticated_access = ["generate-root"]` to allow emergency token generation. See Setup Requirements below.

**Emergency workflow:**

1. **Normal operations**: Use OIDC login for all admin access (enables audit trail, per-user accountability)
2. **Emergency (OIDC unavailable)**: 
   - Seal the Vault instance if needed
   - Unseal it using `vault-tool unseal` with encrypted unseal keys  
   - Generate temporary root token using `vault-tool emergency-root-token`
   - Perform emergency investigation/remediation
   - Revoke the emergency token when done

**Break-glass use cases only:**
- Emergency root token when OIDC/normal auth is unavailable AND Vault has been sealed
- Unsealing Vault after unexpected seal or startup failure
- Disaster recovery from authentication infrastructure failure
- **NOT for routine admin tasks** (use OIDC for those - audit trail required)

## Features

- **Unseal** - Decrypt GPG-encrypted unseal keys and submit them to Vault to unseal it
- **Emergency Root Token** - Generate temporary root token with configurable TTL (default 24 hours)
- **OTP Support** - Optional one-time-password protection for secure multi-channel token distribution
- **Decode Token** - Decrypt OTP-wrapped tokens for use
- **Flexible Output** - Token written to `~/.vault-token` by default (secure 0600 permissions), or specify custom location with `-f` flag, or stdout with `-f -`
- **Quiet by Default** - No logging output unless `-v` (verbose) or `-d` (debug) flags specified

## Installation

### Option 1: Download prebuilt binary or Debian package from GitHub Releases

Download from [Releases](https://github.com/rossigee/vault-tool/releases):

**Linux amd64:**
```bash
# Tarball
wget https://github.com/rossigee/vault-tool/releases/download/v0.2.3/vault-tool_0.2.3_linux_amd64.tar.gz
tar -xzf vault-tool_0.2.3_linux_amd64.tar.gz
sudo install -m 0755 vault-tool /usr/local/bin/

# Or .deb package
wget https://github.com/rossigee/vault-tool/releases/download/v0.2.3/vault-tool_0.2.3_linux_amd64.deb
sudo dpkg -i vault-tool_0.2.3_linux_amd64.deb
```

**Linux arm64:**
```bash
# Tarball
wget https://github.com/rossigee/vault-tool/releases/download/v0.2.3/vault-tool_0.2.3_linux_arm64.tar.gz
tar -xzf vault-tool_0.2.3_linux_arm64.tar.gz
sudo install -m 0755 vault-tool /usr/local/bin/

# Or .deb package
wget https://github.com/rossigee/vault-tool/releases/download/v0.2.3/vault-tool_0.2.3_linux_arm64.deb
sudo dpkg -i vault-tool_0.2.3_linux_arm64.deb
```

### Option 2: Install from source

```bash
go install github.com/rossigee/vault-tool@latest
```

Or build locally:

```bash
git clone https://github.com/rossigee/vault-tool.git
cd vault-tool
go build -o vault-tool
```

## Emergency Workflow

### Complete example: Generate and use an emergency root token

```bash
# 1. Set Vault address (optional, uses --addr flag if not set)
export VAULT_ADDR=https://vault.example.com

# 2. Generate emergency root token (quiet by default, writes to ~/.vault-token)
vault-tool emergency-root-token

# 3. Use the token immediately
export VAULT_TOKEN=$(cat ~/.vault-token)
vault status
vault auth list

# 4. Optional: Troubleshoot with verbose logging
vault-tool emergency-root-token -v -f /tmp/root-token.txt

# 5. Clean up - revoke when emergency is resolved
vault token revoke -self

# 6. Securely delete the token file
shred -u ~/.vault-token
```

Or retrieve token from stdout for piping:
```bash
vault-tool emergency-root-token -f - | tee >(export VAULT_TOKEN=$(cat))
```

### Secure distribution workflow: Use OTP protection for separate channels

```bash
# Alice generates token with OTP protection
alice$ vault-tool emergency-root-token --use-otp
# Output:
# OTP=EXAMPLE_OTP_VALUE_CHANGE_THIS
# ENCODED_TOKEN=EXAMPLE_ENCODED_TOKEN_CHANGE_THIS

# Alice sends OTP via Slack/Signal
# Alice sends ENCODED_TOKEN via email

# Bob receives both and decodes
bob$ vault-tool decode-token EXAMPLE_ENCODED_TOKEN_CHANGE_THIS --otp EXAMPLE_OTP_VALUE_CHANGE_THIS
# (outputs plaintext Vault token: hvs.XXXXXXXXXXXXXXXX...)
```

## Setup Requirements

### Vault Server Configuration

Enable unauthenticated access to root token generation endpoints in your Vault server configuration:

```hcl
# In vault.hcl or your Vault config
enable_unauthenticated_access = ["generate-root"]
```

Then restart Vault. This allows `vault-tool` to generate emergency root tokens using only the unseal keys, without requiring a separate authentication token.

### GPG-encrypted Unseal Keys

Store your unseal keys in an encrypted file:

```bash
# Create encrypted file with unseal keys
printf 'key1\nkey2\nkey3\n' | gpg -e -r user@example.com -o ~/.config/vault/unseal-keys.gpg
```

Replace `key1`, `key2`, `key3` with your actual Vault unseal keys and `user@example.com` with your GPG key ID.

## Usage

### Unseal a Vault instance

```bash
./vault-tool unseal
```

Prompts for GPG passphrase, then submits unseal keys to unseal Vault.

**Options:**
- `--addr` - Vault address (default: `$VAULT_ADDR` environment variable, or empty if not set)
- `--keys-file` - Path to GPG-encrypted unseal keys (default: `$HOME/.config/vault/unseal-keys.gpg`)
- `--passphrase` - GPG passphrase (prompts if not provided)
- `-q, --quiet` - Suppress logging output

### Generate an emergency root token

```bash
./vault-tool emergency-root-token
```

**Default behavior:** Quiet, decrypts unseal keys, generates root token, auto-decodes OTP wrapping, writes plaintext token to `~/.vault-token` with 0600 permissions.

**Options:**
- `--addr` - Vault address (default: `$VAULT_ADDR` environment variable, or empty if not set)
- `--keys-file` - Path to GPG-encrypted unseal keys (default: `$HOME/.config/vault/unseal-keys.gpg`)
- `--passphrase` - GPG passphrase (prompts if not provided)
- `--ttl` - Token TTL (default: `24h`, e.g., `1h`, `72h`, `7d`)
- `-f, --file` - Output file for token (default: `~/.vault-token`, use `-f -` for stdout)
- `--use-otp` - Output OTP and encoded token separately (for secure distribution over separate channels)
- `-v, --verbose` - Enable info-level logging
- `-d, --debug` - Enable debug-level logging

### OTP-Protected Token Generation

For secure distribution of root tokens via separate channels:

```bash
# Generate with OTP protection (outputs are separate)
./vault-tool emergency-root-token --use-otp

# Output:
# OTP=EXAMPLE_OTP_VALUE
# ENCODED_TOKEN=EXAMPLE_ENCODED_TOKEN

# Send OTP via one channel (e.g., Slack, Signal), encoded token via another (e.g., email)
```

### Decode OTP-wrapped tokens

Decode a token that was generated with `--use-otp`:

```bash
./vault-tool decode-token ENCODED_TOKEN --otp OTP

# Or prompt for OTP interactively:
./vault-tool decode-token ENCODED_TOKEN
# Enter OTP: <paste OTP>
```

**Options:**
- `-q, --quiet` - Suppress logging output

## Troubleshooting

### "permission denied" on generate-root/attempt

**Cause:** Vault's `/sys/generate-root/*` endpoints are authenticated by default. You must explicitly enable unauthenticated access in Vault's configuration.

**Solution:**

1. **Enable unauthenticated access in Vault config:**

Add this to your Vault configuration file (vault.hcl or similar):

```hcl
enable_unauthenticated_access = ["generate-root"]
```

2. **Restart Vault:**

```bash
# Use your deployment method (systemd, Kubernetes, etc.)
systemctl restart vault
# OR
kubectl rollout restart deployment vault
```

3. **Verify it's enabled:**

```bash
vault operator generate-root -init
# Should succeed without requiring a token
```

After configuration, `vault-tool emergency-root-token` will work using only the unseal keys for authentication.

### "failed to decrypt keys: exit status 2"

**Cause:** GPG passphrase incorrect or unseal keys file not found.

**Solution:**
1. Verify file exists: `ls ~/.config/vault/unseal-keys.gpg`
2. Verify passphrase: `gpg -d ~/.config/vault/unseal-keys.gpg`
3. Create file if missing with correct keys

### "root generation already in progress"

**Cause:** A previous root token generation is still in progress (Vault limits to one at a time).

**Solution:** 

1. If you have a root token, cancel the attempt:
```bash
vault write sys/generate-root/cancel
```

2. Or wait for the operation to timeout (check Vault logs for timeout settings)

3. If stuck, restart Vault (will clear all in-progress operations)

### "illegal base64 data at input byte X"

**Cause:** The encoded token from Vault cannot be base64-decoded.

**Solution:** This indicates a server-side issue. Verify:
- Vault version is 2.1.0 or later
- The token was actually generated successfully (check logs)
- The token string wasn't truncated or corrupted in copy/paste

### "token and OTP length mismatch"

**Cause:** The OTP and encoded token have incompatible lengths.

**Solution:**
- Ensure you're using the EXACT OTP and ENCODED_TOKEN from the generation output
- Don't modify or shorten either value
- Use copy/paste rather than manual typing

## Implementation Notes

- **Native HTTP client**: Uses Go's `net/http` for Vault API communication (no external dependencies)
- **GPG decryption**: Subprocess call to `gpg` for secure key management
- **TLS**: Full certificate verification enabled (required for all connections)
- **Structured logging**: Uses Go's `log/slog` for detailed operation tracing
- **OTP encoding**: Matches Vault's official implementation (XOR + base64.RawStdEncoding)
- **Token format**: Generates Vault root tokens with `hvs.` prefix (service token variant)
- **Output to stdout**: Enables piping to environment variables or other tools

## Architecture

```
vault-tool/
├── VERSION                       # Semantic version (injected via ldflags)
├── main.go                       # Entry point
├── cmd/
│   ├── root.go                  # Root Cobra command with subcommands
│   ├── unseal.go                # Unseal subcommand
│   ├── emergency_root_token.go  # Emergency root token generation
│   ├── decode_token.go          # OTP-wrapped token decoder
│   └── decode_token_test.go     # Tests
├── pkg/
│   ├── vault/
│   │   ├── client.go            # Vault API client (unseal, generate-root)
│   │   └── client_test.go       # Tests
│   ├── gpg/
│   │   └── decrypt.go           # GPG key decryption subprocess
│   └── otp/
│       ├── decode.go            # OTP token unwrapping (XOR + SHA1)
│       └── decode_test.go       # Tests
├── internal/
│   ├── version/
│   │   └── version.go           # Version variable for ldflags injection
│   └── logger/
│       └── logger.go            # Structured logging (log/slog)
└── debian/                       # Debian packaging
    ├── control                  # Package metadata
    ├── rules                    # Build rules
    ├── changelog                # Release notes
    ├── install                  # Install directives
    ├── source/format            # Source format (3.0 native)
    └── compat                   # Debhelper compatibility
```

**Key design choices:**
- Unauthenticated Vault API calls (requires Vault ACL policy updates for `/sys/generate-root/*`)
- OTP tokens use XOR wrapping with SHA1 checksum (per Vault documentation)
- Structured logging enables troubleshooting without verbose flags
- Subprocess GPG decryption handles keyring management automatically
- No external credentials required beyond GPG passphrase and unseal keys
