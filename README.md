# Vault Tool

A Go CLI tool for Vault emergency access and maintenance, providing sealed instance unsealing and emergency root token generation when normal authentication (OIDC) is unavailable.

## Features

- **Unseal** - Decrypt GPG-encrypted unseal keys and submit them to Vault to unseal it
- **Emergency Root Token** - Use unseal keys to generate a temporary root token for emergency maintenance with configurable TTL (default 24 hours)

## Installation

```bash
go build -o vault-tool
```

## Usage

### Unseal a Vault instance

```bash
./vault-tool unseal
```

Options:
- `--addr` - Vault address (default: `https://vault.bankrut.lan`)
- `--keys-file` - Path to GPG-encrypted unseal keys (default: `$HOME/.config/vault/bankrut-unseal-keys.gpg`)
- `--passphrase` - GPG passphrase (will prompt from stdin if not provided)

### Generate an emergency root token

```bash
./vault-tool emergency-root-token
```

Options:
- `--addr` - Vault address (default: `https://vault.bankrut.lan`)
- `--keys-file` - Path to GPG-encrypted unseal keys (default: `$HOME/.config/vault/bankrut-unseal-keys.gpg`)
- `--passphrase` - GPG passphrase (will prompt from stdin if not provided)
- `--ttl` - Token TTL (default: `24h`, e.g., `1h`, `72h`, `7d`)

The emergency root token is printed to stdout, making it suitable for scripting or piping to other tools.

## Setup

### Create GPG-encrypted unseal keys file

```bash
printf 'key1\nkey2\nkey3\n' | gpg -e -r ross@golder.org -o ~/.config/vault/bankrut-unseal-keys.gpg
```

Where `key1`, `key2`, and `key3` are your actual Vault unseal keys.

## Implementation Notes

- Uses native Go HTTP client for Vault API communication
- GPG decryption via subprocess call to `gpg` command (handles keyring management automatically)
- TLS certificate verification is disabled (for emergency use; production should use proper verification)
- Output to stdout for emergency root token enables easy scripting and piping

## Architecture

```
vault-tool/
├── main.go                 # Entry point
├── cmd/
│   ├── root.go            # Root command with subcommands
│   ├── unseal.go          # Unseal subcommand
│   └── emergency_root_token.go  # Emergency root token subcommand
└── pkg/
    ├── vault/
    │   └── client.go      # Vault API client
    └── gpg/
        └── decrypt.go     # GPG key decryption
```
