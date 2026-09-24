package vault

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://vault.example.com")
	if client.addr != "https://vault.example.com" {
		t.Errorf("Client address mismatch: got %s, want https://vault.example.com", client.addr)
	}
	if client.client == nil {
		t.Error("HTTP client not initialized")
	}
}

func TestUnsealResponseParsing(t *testing.T) {
	// Test that the struct can be unmarshaled from JSON
	resp := &UnsealResponse{
		Sealed:   true,
		T:        3,
		N:        5,
		Progress: 1,
	}

	if resp.Sealed != true || resp.T != 3 || resp.Progress != 1 {
		t.Error("UnsealResponse fields not set correctly")
	}
}

func TestInitRootTokenResponseParsing(t *testing.T) {
	resp := &InitRootTokenResponse{
		Nonce: "test-nonce",
		T:     3,
		N:     5,
	}

	if resp.Nonce != "test-nonce" || resp.T != 3 {
		t.Error("InitRootTokenResponse fields not set correctly")
	}
}

func TestUpdateRootTokenResponseParsing(t *testing.T) {
	resp := &UpdateRootTokenResponse{
		Nonce:        "test-nonce",
		Progress:     1,
		Required:     3,
		Complete:     false,
		EncodedToken: "",
	}

	if resp.Nonce != "test-nonce" || resp.Complete != false {
		t.Error("UpdateRootTokenResponse fields not set correctly")
	}
}
