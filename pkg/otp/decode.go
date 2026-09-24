package otp

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
)

// DecodeToken decodes an OTP-wrapped root token using the one-time password
// This follows Vault's documented procedure for root token generation with OTP protection
func DecodeToken(encodedToken, otp string) (string, error) {
	// Decode base64 encoded token
	tokenBytes, err := base64.StdEncoding.DecodeString(encodedToken)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 token: %w", err)
	}

	// Validate OTP length
	if len(otp) == 0 {
		return "", fmt.Errorf("OTP cannot be empty")
	}

	// The OTP is XORed with the token for wrapping/unwrapping
	// Both must be the same length for XOR operation
	if len(tokenBytes) != len(otp) {
		return "", fmt.Errorf("token and OTP length mismatch: token=%d bytes, otp=%d characters", len(tokenBytes), len(otp))
	}

	// XOR the token with OTP to unwrap it
	decodedToken := make([]byte, len(tokenBytes))
	for i := range tokenBytes {
		decodedToken[i] = tokenBytes[i] ^ otp[i]
	}

	// Verify the token is valid by checking SHA1 hash
	// The last 20 bytes should be the SHA1 hash of the rest
	if len(decodedToken) < 20 {
		return "", fmt.Errorf("decoded token too short: %d bytes", len(decodedToken))
	}

	tokenData := decodedToken[:len(decodedToken)-20]
	checksum := decodedToken[len(decodedToken)-20:]

	hash := sha1.Sum(tokenData)
	if string(hash[:]) != string(checksum) {
		return "", fmt.Errorf("token checksum verification failed - OTP may be incorrect")
	}

	// Return the decoded token (without the checksum)
	return string(tokenData), nil
}
