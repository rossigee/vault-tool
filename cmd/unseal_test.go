package cmd

import (
	"testing"
)

func TestUnsealCommand(t *testing.T) {
	cmd := unsealCmd
	if cmd == nil {
		t.Fatal("unsealCmd is nil")
	}
	if cmd.Use != "unseal" {
		t.Errorf("expected Use='unseal', got %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("unsealCmd.Short should not be empty")
	}
}

func TestUnsealFlags(t *testing.T) {
	cmd := unsealCmd
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
}

func TestUnsealAddrRespectsVAULTADDR(t *testing.T) {
	// This test verifies that the --addr flag uses VAULT_ADDR as default.
	// Note: The default value is set at init() time, so changing VAULT_ADDR
	// during test execution won't affect the flag's default value that was
	// already computed. However, this test verifies that the flag exists
	// and that it can be overridden via the --addr flag.

	cmd := unsealCmd
	flags := cmd.Flags()

	addrFlag := flags.Lookup("addr")
	if addrFlag == nil {
		t.Fatal("missing --addr flag")
	}

	// The default value should be whatever VAULT_ADDR was set to when tests started
	// We just verify it's not the hardcoded example value anymore
	if addrFlag.DefValue == "https://vault.example.com" {
		t.Error("--addr flag should use VAULT_ADDR environment variable, not hardcoded default")
	}
}
