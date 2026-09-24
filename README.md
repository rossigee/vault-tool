# Vault Tool

A Go CLI tool for Vault emergency access and maintenance, providing sealed instance unsealing and emergency root token generation with OTP protection.

## Overview

`vault-tool` enables operators to unseal a Vault instance and generate temporary emergency root tokens when normal authentication (OIDC, tokens, etc.) is unavailable. It uses GPG-encrypted unseal keys for secure key storage and management.

**Use cases:**
- Emergency root token generation when normal auth is unavailable
- Unsealing Vault after operator lockout
- Recovery from lost authentication credentials

## Features

- **Unseal** - Decrypt GPG-encrypted unseal keys and submit them to Vault to unseal it
- **Emergency Root Token** - Generate a temporary root token with configurable TTL (default 24 hours)
- **OTP Support** - Optional one-time-password protection for tokens during secure distribution
- **Decode Token** - Decrypt OTP-wrapped tokens for use
- **Structured Logging** - Detailed operation logging with `-q` flag to suppress output

## Installation

```bash
go build -o vault-tool
```

Or with version injection:

```bash
go build -ldflags "-X git.golder.lan/rossgolderltd/vault-tool/internal/version.Version=$(cat VERSION)" -o vault-tool
```

## Setup Requirements

### GPG-encrypted Unseal Keys

Store your unseal keys in an encrypted file:

```bash
# Create encrypted file with unseal keys
printf 'key1\nkey2\nkey3\n' | gpg -e -r user@example.com -o ~/.config/vault/bankrut-unseal-keys.gpg
```

Replace `key1`, `key2`, `key3` with your actual Vault unseal keys and `user@example.com` with your GPG key ID.

## Usage

### Unseal a Vault instance

```bash
./vault-tool unseal
```

Prompts for GPG passphrase, then submits unseal keys to unseal Vault.

**Options:**
- `--addr` - Vault address (default: `https://vault.bankrut.lan`)
- `--keys-file` - Path to GPG-encrypted unseal keys (default: `$HOME/.config/vault/bankrut-unseal-keys.gpg`)
- `--passphrase` - GPG passphrase (prompts if not provided)
- `-q, --quiet` - Suppress logging output

### Generate an emergency root token

```bash
./vault-tool emergency-root-token
```

Returns a plaintext root token on stdout (auto-decodes OTP by default).

**Options:**
- `--addr` - Vault address (default: `https://vault.bankrut.lan`)
- `--keys-file` - Path to GPG-encrypted unseal keys
- `--passphrase` - GPG passphrase (prompts if not provided)
- `--ttl` - Token TTL (default: `24h`, e.g., `1h`, `72h`, `7d`)
- `--use-otp` - Output OTP and encoded token separately (for secure distribution)
- `-q, --quiet` - Suppress logging output

### Decode OTP-wrapped tokens

If using `--use-otp`, decode the token with:

```bash
./vault-tool decode-token ENCODED_TOKEN --otp OTP_PASSWORD
```

Or prompt for OTP:

```bash
./vault-tool decode-token ENCODED_TOKEN
# Enter OTP: <paste OTP>
```

## OTP Support

Tokens can be generated with OTP protection for secure distribution via separate channels:

```bash
# Generate with OTP protection
export VAULT_ROLE_ID='...'
export VAULT_SECRET_ID='...'
./vault-tool emergency-root-token --use-otp

# Output:
# OTP=b3Pi7xRBYi1bDA6BO6cBYIH0WiBY
# ENCODED_TOKEN=Qx4BAgMEBQYHCAkAUVJTVFVWhPHVo9zhPP8L7q5MuHmhadqLj+A=

# Send OTP via one channel, encoded token via another
# Recipient decodes with:
./vault-tool decode-token ENCODED_TOKEN --otp OTP
```

## Troubleshooting

### "permission denied" on generate-root/attempt

**Cause:** Vault endpoint requires unauthenticated access to be explicitly allowed.

**Solution:**

Vault's `/sys/generate-root/*` endpoints are protected and require either:

1. **Vault to be sealed** (endpoints inaccessible when sealed - this is by design)
2. **Default policy updated** to allow unauthenticated access:

```bash
vault policy write default - <<'EOF'
# [existing default policy rules]

# Emergency root token generation
path "sys/generate-root/attempt" {
  capabilities = ["create", "update"]
}

path "sys/generate-root/update" {
  capabilities = ["create", "update"]
}

path "sys/generate-root/status" {
  capabilities = ["read"]
}

path "sys/generate-root/cancel" {
  capabilities = ["create", "update"]
}

path "sys/seal-status" {
  capabilities = ["read"]
}

path "sys/unseal" {
  capabilities = ["create", "update"]
}
EOF
```

Or use `vault policy read default > /tmp/policy.hcl`, edit it, then `vault policy write default /tmp/policy.hcl`.

### "failed to decrypt keys: exit status 2"

**Cause:** GPG passphrase incorrect or unseal keys file not found.

**Solution:**
1. Verify file exists: `ls ~/.config/vault/bankrut-unseal-keys.gpg`
2. Verify passphrase: `gpg -d ~/.config/vault/bankrut-unseal-keys.gpg`
3. Create file if missing with correct keys

### "root generation already in progress"

**Cause:** Previous root token generation didn't complete.

**Solution:** Use root token to cancel:

```bash
vault write sys/generate-root/cancel
# OR
curl -X PUT -H "X-Vault-Token: $VAULT_ROOT_TOKEN" \
  https://vault.bankrut.lan/v1/sys/generate-root/cancel
```

### "token checksum verification failed - OTP may be incorrect"

**Cause:** Wrong OTP used to decode token.

**Solution:** Verify OTP matches the one from token generation and retry.

## Implementation Notes

- Uses native Go HTTP client for Vault API communication
- GPG decryption via subprocess call to `gpg` command
- TLS certificate verification disabled (for emergency use; production should verify)
- Structured logging with log/slog for operation tracing
- OTP wrapping follows Vault's documented token protection procedure
- Output to stdout enables piping and scripting

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
