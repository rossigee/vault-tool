package cmd

import (
	"os"
	"testing"
	"time"
)

func TestEmergencyRootTokenCommand(t *testing.T) {
	cmd := emergencyRootTokenCmd
	if cmd == nil {
		t.Fatal("emergencyRootTokenCmd is nil")
	}
	if cmd.Use != "emergency-root-token" {
		t.Errorf("expected Use='emergency-root-token', got %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("emergencyRootTokenCmd.Short should not be empty")
	}
}

func TestEmergencyRootTokenFlags(t *testing.T) {
	cmd := emergencyRootTokenCmd
	flags := cmd.Flags()

	if flags.Lookup("addr") == nil {
		t.Error("missing --addr flag")
	}
	if flags.Lookup("keys-file") == nil {
		t.Error("missing --keys-file flag")
	}
	if flags.Lookup("passphrase") == nil {
		t.Error("missing --passphrase flag")
	}
	if flags.Lookup("ttl") == nil {
		t.Error("missing --ttl flag")
	}
	if flags.Lookup("unlimited") == nil {
		t.Error("missing --unlimited flag")
	}
}

func TestEmergencyRootTokenTTLDefault(t *testing.T) {
	cmd := emergencyRootTokenCmd
	flags := cmd.Flags()

	ttlFlag := flags.Lookup("ttl")
	if ttlFlag == nil {
		t.Fatal("missing --ttl flag")
	}

	if ttlFlag.DefValue != "24h" {
		t.Errorf("expected default TTL '24h', got %q", ttlFlag.DefValue)
	}
}

func TestEmergencyRootTokenUnlimitedFlag(t *testing.T) {
	cmd := emergencyRootTokenCmd
	flags := cmd.Flags()

	unlimitedFlag := flags.Lookup("unlimited")
	if unlimitedFlag == nil {
		t.Fatal("missing --unlimited flag")
	}

	// Check it's a boolean flag
	if unlimitedFlag.Value.Type() != "bool" {
		t.Errorf("expected --unlimited to be a bool flag, got type %s", unlimitedFlag.Value.Type())
	}
}

func TestEmergencyRootTokenMutualExclusion(t *testing.T) {
	tests := []struct {
		name      string
		ttl       string
		unlimited bool
		shouldErr bool
	}{
		{"default TTL only", "24h", false, false},
		{"custom TTL only", "1h", false, false},
		{"unlimited only", "24h", true, false},
		{"both unlimited and custom TTL", "1h", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags to defaults
			emergencyTokenTTL = "24h"
			emergencyTokenUnlimited = false

			// Set test values
			if tt.ttl != "24h" {
				emergencyTokenTTL = tt.ttl
			}
			emergencyTokenUnlimited = tt.unlimited

			// Simulate the validation logic from runEmergencyRootToken
			var tokenTTL string
			shouldErr := false

			if emergencyTokenUnlimited {
				if emergencyTokenTTL != "24h" {
					shouldErr = true
				} else {
					tokenTTL = ""
				}
			} else {
				tokenTTL = emergencyTokenTTL
			}

			if shouldErr != tt.shouldErr {
				t.Errorf("expected shouldErr=%v, got %v", tt.shouldErr, shouldErr)
			}

			if !shouldErr && emergencyTokenUnlimited && tokenTTL != "" {
				t.Errorf("expected empty tokenTTL with --unlimited, got %q", tokenTTL)
			}
		})
	}
}

func TestWriteTokenToStdout(t *testing.T) {
	// Save original value
	original := emergencyTokenFile
	defer func() { emergencyTokenFile = original }()

	emergencyTokenFile = "-"
	err := writeToken("test-token-123", "24h", false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWriteTokenToFile(t *testing.T) {
	// Save original value
	original := emergencyTokenFile
	defer func() { emergencyTokenFile = original }()

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "vault-token-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	tests := []struct {
		name       string
		token      string
		ttl        string
		unlimited  bool
		shouldFail bool
	}{
		{"write token with TTL", "token-abc", "24h", false, false},
		{"write unlimited token", "token-xyz", "", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emergencyTokenFile = tmpFile.Name()

			err := writeToken(tt.token, tt.ttl, tt.unlimited)
			if tt.shouldFail && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.shouldFail {
				// Verify file contents
				content, err := os.ReadFile(tmpFile.Name())
				if err != nil {
					t.Fatalf("failed to read temp file: %v", err)
				}
				if string(content) != tt.token+"\n" {
					t.Errorf("expected token %q, got %q", tt.token+"\n", string(content))
				}
			}
		})
	}
}

func TestWriteTokenPermissions(t *testing.T) {
	// Save original value
	original := emergencyTokenFile
	defer func() { emergencyTokenFile = original }()

	tmpFile, err := os.CreateTemp("", "vault-token-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	emergencyTokenFile = tmpFile.Name()

	err = writeToken("secret-token", "1h", false)
	if err != nil {
		t.Fatalf("writeToken failed: %v", err)
	}

	// Check file permissions are restrictive (0600)
	info, err := os.Stat(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	perm := info.Mode().Perm()
	expected := os.FileMode(0600)
	if perm != expected {
		t.Errorf("expected permissions %o, got %o", expected, perm)
	}
}

func TestTTLValidation(t *testing.T) {
	tests := []struct {
		name      string
		ttl       string
		shouldErr bool
	}{
		{"valid 24h", "24h", false},
		{"valid 1h", "1h", false},
		{"valid 30m", "30m", false},
		{"valid 168h (7 days)", "168h", false},
		{"invalid duration", "invalid", true},
		{"unsupported days format", "7d", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := time.ParseDuration(tt.ttl)
			hasErr := err != nil

			if hasErr != tt.shouldErr {
				t.Errorf("expected error=%v, got %v", tt.shouldErr, hasErr)
			}
		})
	}
}
