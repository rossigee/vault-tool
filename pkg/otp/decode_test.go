package otp

import (
	"encoding/base64"
	"testing"
)

func TestDecodeToken(t *testing.T) {
	// Test data: create a known token, wrap it with OTP, and verify decoding
	originalToken := "s.0123456789abcdef0123456789abcd"
	// OTP must be same length as token for Vault compatibility
	otp := "01234567890123456789012345678901"

	// Verify lengths match
	if len(originalToken) != len(otp) {
		t.Fatalf("Test setup error: token (%d) != otp (%d)", len(originalToken), len(otp))
	}

	// XOR with OTP
	xorData := make([]byte, len(originalToken))
	for i := range originalToken {
		xorData[i] = originalToken[i] ^ otp[i]
	}

	// Base64 encode using RawStdEncoding (matches Vault's encoding)
	encodedToken := base64.RawStdEncoding.EncodeToString(xorData)

	// Decode and verify
	decodedToken, err := DecodeToken(encodedToken, otp)
	if err != nil {
		t.Fatalf("DecodeToken failed: %v", err)
	}

	if decodedToken != originalToken {
		t.Fatalf("Decoded token mismatch: expected %q, got %q", originalToken, decodedToken)
	}
}

func TestDecodeTokenWrongOTP(t *testing.T) {
	// Test with wrong OTP - will produce wrong token
	originalToken := "s.0123456789abcdef0123456789abcd"
	correctOTP := "01234567890123456789012345678901"
	wrongOTP := "10234567890123456789012345678901"

	xorData := make([]byte, len(originalToken))
	for i := range originalToken {
		xorData[i] = originalToken[i] ^ correctOTP[i]
	}

	encodedToken := base64.RawStdEncoding.EncodeToString(xorData)

	// Decode with wrong OTP produces a different token (but doesn't error)
	decoded, err := DecodeToken(encodedToken, wrongOTP)
	if err != nil {
		t.Fatalf("DecodeToken should not error: %v", err)
	}

	if decoded == originalToken {
		t.Fatalf("Wrong OTP should produce different token")
	}
}

func TestDecodeTokenEmptyOTP(t *testing.T) {
	_, err := DecodeToken("dGVzdA", "")
	if err == nil {
		t.Fatalf("DecodeToken should have failed with empty OTP")
	}

	if err.Error() != "OTP cannot be empty" {
		t.Fatalf("Unexpected error message: %v", err)
	}
}

func TestDecodeTokenMismatchedLengths(t *testing.T) {
	// OTP must match encoded token length (after decoding)
	// 5 bytes encoded as base64 Raw = 7 chars, decodes back to 5 bytes
	encodedToken := base64.RawStdEncoding.EncodeToString([]byte("short"))
	_, err := DecodeToken(encodedToken, "toolong")
	if err == nil || err.Error() != "token and OTP length mismatch: token=5 bytes, otp=7 characters" {
		t.Fatalf("DecodeToken should fail with length mismatch: %v", err)
	}
}
