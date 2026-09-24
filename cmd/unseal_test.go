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
