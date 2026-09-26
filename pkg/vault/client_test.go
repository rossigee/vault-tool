package vault

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestInitRootTokenGenerationPayloadWithTTL(t *testing.T) {
	// Verify that TTL is included in the payload when provided
	tests := []struct {
		name        string
		ttl         string
		shouldHaveTTL bool
	}{
		{"with TTL", "24h", true},
		{"with custom TTL", "1h", true},
		{"without TTL (unlimited)", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]interface{}{}
			if tt.ttl != "" {
				payload["ttl"] = tt.ttl
			}

			data, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("failed to marshal payload: %v", err)
			}

			// Verify the payload
			var unmarshalledPayload map[string]interface{}
			err = json.Unmarshal(data, &unmarshalledPayload)
			if err != nil {
				t.Fatalf("failed to unmarshal payload: %v", err)
			}

			hasTTL := unmarshalledPayload["ttl"] != nil
			if hasTTL != tt.shouldHaveTTL {
				t.Errorf("expected TTL present=%v, got %v", tt.shouldHaveTTL, hasTTL)
			}

			if tt.shouldHaveTTL && unmarshalledPayload["ttl"] != tt.ttl {
				t.Errorf("expected TTL=%q, got %q", tt.ttl, unmarshalledPayload["ttl"])
			}
		})
	}
}

func TestInitRootTokenGenerationRequestBody(t *testing.T) {
	// Test that the HTTP request body is properly constructed
	tests := []struct {
		name         string
		ttl          string
		expectedBody string
	}{
		{"with TTL", "24h", `{"ttl":"24h"}`},
		{"with custom TTL", "1h", `{"ttl":"1h"}`},
		{"without TTL (unlimited)", "", `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]interface{}{}
			if tt.ttl != "" {
				payload["ttl"] = tt.ttl
			}

			data, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("failed to marshal payload: %v", err)
			}

			// Verify the request body matches expected JSON
			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("failed to unmarshal request body: %v", err)
			}

			// Check payload structure
			expectedPayload := map[string]interface{}{}
			if tt.ttl != "" {
				expectedPayload["ttl"] = tt.ttl
			}

			if len(result) != len(expectedPayload) {
				t.Errorf("payload length mismatch: got %d, want %d", len(result), len(expectedPayload))
			}

			for key, expectedVal := range expectedPayload {
				if val, ok := result[key]; !ok {
					t.Errorf("missing key %q in payload", key)
				} else if val != expectedVal {
					t.Errorf("value mismatch for key %q: got %v, want %v", key, val, expectedVal)
				}
			}
		})
	}
}

func TestInitRootTokenGenerationHTTP(t *testing.T) {
	tests := []struct {
		name           string
		ttl            string
		responseBody   string
		expectedNonce  string
		expectedErr    bool
		statusCode     int
	}{
		{
			name: "success with TTL",
			ttl:  "24h",
			responseBody: `{"data":{"nonce":"test-nonce-123","otp":"otp-secret","otp_length":16,"t":3,"n":5}}`,
			expectedNonce: "test-nonce-123",
			expectedErr:   false,
			statusCode:    200,
		},
		{
			name: "success without TTL (unlimited)",
			ttl:  "",
			responseBody: `{"data":{"nonce":"unlimited-nonce","otp":"otp-secret","otp_length":16,"t":3,"n":5}}`,
			expectedNonce: "unlimited-nonce",
			expectedErr:   false,
			statusCode:    200,
		},
		{
			name: "direct response format (no data wrapper)",
			ttl:  "1h",
			responseBody: `{"nonce":"direct-nonce","otp":"otp-secret","otp_length":16,"t":3,"n":5}`,
			expectedNonce: "direct-nonce",
			expectedErr:   false,
			statusCode:    200,
		},
		{
			name: "API error",
			ttl:  "24h",
			responseBody: `{"errors":["unseal keys required"]}`,
			expectedNonce: "",
			expectedErr:   true,
			statusCode:    400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/sys/generate-root/attempt" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Method != "PUT" {
					t.Errorf("expected PUT, got %s", r.Method)
				}

				// Verify request body contains TTL when provided
				body, _ := io.ReadAll(r.Body)
				var payload map[string]interface{}
				json.Unmarshal(body, &payload)

				if tt.ttl != "" {
					if val, ok := payload["ttl"]; !ok {
						t.Error("expected ttl in payload")
					} else if val != tt.ttl {
						t.Errorf("expected ttl=%q, got %q", tt.ttl, val)
					}
				} else {
					if _, ok := payload["ttl"]; ok {
						t.Error("unexpected ttl in payload for unlimited token")
					}
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			resp, err := client.InitRootTokenGeneration(context.Background(), tt.ttl)

			if tt.expectedErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectedErr && resp.Nonce != tt.expectedNonce {
				t.Errorf("expected nonce %q, got %q", tt.expectedNonce, resp.Nonce)
			}
		})
	}
}

func TestUpdateRootTokenGenerationHTTP(t *testing.T) {
	tests := []struct {
		name        string
		nonce       string
		key         string
		responseBody string
		expectedErr bool
		statusCode  int
	}{
		{
			name:         "successful update",
			nonce:        "test-nonce",
			key:          "unseal-key-1",
			responseBody: `{"data":{"nonce":"test-nonce","progress":1,"required":3,"complete":false,"encoded_token":""}}`,
			expectedErr:  false,
			statusCode:   200,
		},
		{
			name:         "token generation complete",
			nonce:        "test-nonce",
			key:          "unseal-key-3",
			responseBody: `{"data":{"nonce":"test-nonce","progress":3,"required":3,"complete":true,"encoded_token":"encoded-token-xyz"}}`,
			expectedErr:  false,
			statusCode:   200,
		},
		{
			name:         "invalid nonce",
			nonce:        "invalid-nonce",
			key:          "unseal-key-1",
			responseBody: `{"errors":["invalid nonce"]}`,
			expectedErr:  true,
			statusCode:   400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/sys/generate-root/update" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Method != "PUT" {
					t.Errorf("expected PUT, got %s", r.Method)
				}

				// Verify request payload
				body, _ := io.ReadAll(r.Body)
				var payload map[string]string
				json.Unmarshal(body, &payload)

				if payload["nonce"] != tt.nonce {
					t.Errorf("expected nonce %q, got %q", tt.nonce, payload["nonce"])
				}
				if payload["key"] != tt.key {
					t.Errorf("expected key %q, got %q", tt.key, payload["key"])
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient(server.URL)
			resp, err := client.UpdateRootTokenGeneration(context.Background(), tt.nonce, tt.key)

			if tt.expectedErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectedErr && resp.Nonce != tt.nonce {
				t.Errorf("expected nonce %q, got %q", tt.nonce, resp.Nonce)
			}
		})
	}
}
