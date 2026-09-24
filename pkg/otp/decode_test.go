package otp

import (
	"crypto/sha1"
	"encoding/base64"
	"testing"
)

func TestDecodeToken(t *testing.T) {
	// Test data: create a known token, wrap it with OTP, and verify decoding
	originalToken := "s.test1234567890"
	// OTP must be same length as token+hash for Vault compatibility
	// 16 bytes token + 20 bytes SHA1 = 36 bytes total
	otp := "012345678901234567890123456789012345"

	// Build wrapped token: token_data + sha1_hash
	hash := sha1.Sum([]byte(originalToken))
	wrappedData := append([]byte(originalToken), hash[:]...)

	// Verify wrapped data length matches OTP length
	if len(wrappedData) != len(otp) {
		t.Fatalf("Test setup error: wrapped data (%d) != otp (%d)", len(wrappedData), len(otp))
	}

	// XOR with OTP
	xorData := make([]byte, len(wrappedData))
	for i := range wrappedData {
		xorData[i] = wrappedData[i] ^ otp[i]
	}

	// Base64 encode
	encodedToken := base64.StdEncoding.EncodeToString(xorData)

	// Decode and verify
	decodedToken, err := DecodeToken(encodedToken, otp)
	if err != nil {
		t.Fatalf("DecodeToken failed: %v", err)
	}

	if decodedToken != originalToken {
		t.Fatalf("Decoded token mismatch: expected %q, got %q", originalToken, decodedToken)
	}
}

func TestDecodeTokenInvalidOTP(t *testing.T) {
	// Test with wrong OTP - should fail checksum validation
	originalToken := "s.test1234567890"
	correctOTP := "012345678901234567890123456789012345"
	wrongOTP := "abcdefghijklmnopqrstuvwxyzabcdefghij"

	hash := sha1.Sum([]byte(originalToken))
	wrappedData := append([]byte(originalToken), hash[:]...)

	xorData := make([]byte, len(wrappedData))
	for i := range wrappedData {
		xorData[i] = wrappedData[i] ^ correctOTP[i]
	}

	encodedToken := base64.StdEncoding.EncodeToString(xorData)

	// Try to decode with wrong OTP
	_, err := DecodeToken(encodedToken, wrongOTP)
	if err == nil {
		t.Fatalf("DecodeToken should have failed with wrong OTP")
	}

	if err.Error() != "token checksum verification failed - OTP may be incorrect" {
		t.Fatalf("Unexpected error message: %v", err)
	}
}

func TestDecodeTokenEmptyOTP(t *testing.T) {
	_, err := DecodeToken("dGVzdA==", "")
	if err == nil {
		t.Fatalf("DecodeToken should have failed with empty OTP")
	}

	if err.Error() != "OTP cannot be empty" {
		t.Fatalf("Unexpected error message: %v", err)
	}
}

func TestDecodeTokenShortToken(t *testing.T) {
	// Token too short to contain SHA1 hash
	encodedToken := base64.StdEncoding.EncodeToString([]byte("short"))
	_, err := DecodeToken(encodedToken, "otp_pass")
	if err == nil {
		t.Fatalf("DecodeToken should have failed with short token")
	}
}
