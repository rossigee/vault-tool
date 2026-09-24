package cmd

import (
	"testing"
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
